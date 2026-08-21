package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// TemporaryPasswordCreator は Cognito に一時パスワード付きユーザーを作る抽象
// （infra/cognito.AdminUserCreator が満たす）。返す一時パスワードは呼び元で 1 度だけ提示し、
// 保存・ログ出力しない。
type TemporaryPasswordCreator interface {
	CreateWithTemporaryPassword(ctx context.Context, email, name string) (temporaryPassword string, err error)
}

// CreateTemporaryPasswordInvitationUseCase は「初期パスワード方式」の招待を作る（FRESTYLE-313）。
// pending 招待行（role/company を持ち、ログイン時の招待ゲートで適用される）を作り、
// あわせて Cognito ユーザーを一時パスワード付きで作成する。マジックリンク方式
// （CreateAdminInvitationUseCase）とは別 usecase（§2.3 単一責任）。
type CreateTemporaryPasswordInvitationUseCase struct {
	repo      repository.AdminInvitationRepository
	cognito   TemporaryPasswordCreator
	expiresIn time.Duration
}

// NewCreateTemporaryPasswordInvitationUseCase は初期パスワード方式の招待作成 usecase を組み立てる。
// cognito が nil のとき Execute は ErrTemporaryPasswordUnavailable を返す
// （COGNITO_USER_POOL_ID 未設定 = 本機能無効。マジックリンクには影響しない）。
func NewCreateTemporaryPasswordInvitationUseCase(
	r repository.AdminInvitationRepository,
	cognito TemporaryPasswordCreator,
) *CreateTemporaryPasswordInvitationUseCase {
	return &CreateTemporaryPasswordInvitationUseCase{
		repo:      r,
		cognito:   cognito,
		expiresIn: 7 * 24 * time.Hour,
	}
}

// ErrTemporaryPasswordUnavailable は初期パスワード方式が未構成のときに返す（501/400 用）。
var ErrTemporaryPasswordUnavailable = errors.New("temporary password invitation is not configured")

// CreateTemporaryPasswordInvitationOutput は招待行と、1 度だけ返す一時パスワード。
type CreateTemporaryPasswordInvitationOutput struct {
	Invitation        *domain.AdminInvitation
	TemporaryPassword string
}

// Execute は pending 招待を作成し、Cognito ユーザーを一時パスワード付きで作る。
// 順序は「招待行 → Cognito」。Cognito が既存 email で失敗した場合は
// cognito.ErrUserAlreadyExists がそのまま伝播する（handler が 409 に写す）。
func (u *CreateTemporaryPasswordInvitationUseCase) Execute(
	ctx context.Context,
	in CreateAdminInvitationInput,
) (*CreateTemporaryPasswordInvitationOutput, error) {
	if u.cognito == nil {
		return nil, ErrTemporaryPasswordUnavailable
	}
	if in.CompanyID == 0 || in.Email == "" || in.Role == "" {
		return nil, errors.New("companyID, email, role are required")
	}

	// 順序が重要: 先に Cognito ユーザーを作り、成功したときだけ pending 招待行を作る。
	// 逆順（招待行→Cognito）にすると、既存 email への 409（UsernameExists）で招待行だけが
	// 孤児として残り、その行がログイン時の email ゲート（FindPendingByEmail）で拾われて
	// 被害ユーザーの会社が黙って付け替わる経路を生む（多角レビューで確定した major）。
	// Cognito が失敗すれば招待行は作られないため、この経路自体が成立しない。
	tempPw, err := u.cognito.CreateWithTemporaryPassword(ctx, in.Email, in.Name)
	if err != nil {
		// エラー種別（既存ユーザー等）は handler がステータスへ写せるようそのまま返す。
		return nil, fmt.Errorf("create cognito user: %w", err)
	}

	token := uuid.NewString()
	inv := &domain.AdminInvitation{
		CompanyID: in.CompanyID,
		Email:     in.Email,
		Role:      in.Role,
		Name:      in.Name,
		Status:    domain.InvitationStatusPending,
		Token:     &token,
		ExpiresAt: time.Now().UTC().Add(u.expiresIn),
	}
	if err := u.repo.Create(ctx, inv); err != nil {
		// ここに来るのは Cognito 成功後の DB 失敗という稀ケース。Cognito ユーザーは残るが、
		// 招待行が無いためログイン時は招待ゲートで拒否される（fail closed）。管理者は再試行でき、
		// 再試行時は Cognito が 409 を返すので「既に存在」と分かる。
		return nil, fmt.Errorf("create invitation after cognito user: %w", err)
	}

	return &CreateTemporaryPasswordInvitationOutput{Invitation: inv, TemporaryPassword: tempPw}, nil
}
