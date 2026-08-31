package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ListAdminInvitationsUseCase は招待一覧を取得する。
//
// 横断（ListAll）とワークスペース単位（ListByWorkspaceID）の 2 系統を公開する集約 read usecase
//
//naminglint:allow 横断 ListAll・ワークスペース単位 ListByWorkspaceID
type ListAdminInvitationsUseCase struct {
	repo repository.AdminInvitationRepository
}

func NewListAdminInvitationsUseCase(r repository.AdminInvitationRepository) *ListAdminInvitationsUseCase {
	return &ListAdminInvitationsUseCase{repo: r}
}

// ListAll は全ワークスペース横断で招待一覧を返す（認可は handler 層）。
func (u *ListAdminInvitationsUseCase) ListAll(ctx context.Context) ([]domain.AdminInvitation, error) {
	return u.repo.ListAll(ctx)
}

// ListByWorkspaceID は指定ワークスペースの招待一覧を返す（認可は handler 層）。
func (u *ListAdminInvitationsUseCase) ListByWorkspaceID(ctx context.Context, workspaceID string) ([]domain.AdminInvitation, error) {
	if workspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	return u.repo.ListByWorkspaceID(ctx, workspaceID)
}

// MagicLinkSender は invitation メール送信の抽象（infra/ses.Client が満たす）。
// usecase 側に置くことで infra の API 変更を波及させない（DIP）。
type MagicLinkSender interface {
	SendInvitationEmail(ctx context.Context, to, subject, htmlBody, textBody string) error
}

// MailBuilder は招待メールの subject / HTML / text を組み立てる関数。
type MailBuilder func(magicLink, displayName, companyName, role string) (subject, htmlBody, textBody string)

// LinkBuilder は token から受諾画面の絶対 URL を組み立てる関数。
type LinkBuilder func(token string) string

// CreateAdminInvitationUseCase は新規招待を発行し、マジックリンクメールを送る。
type CreateAdminInvitationUseCase struct {
	repo        repository.AdminInvitationRepository
	sender      MagicLinkSender
	buildLink   LinkBuilder
	buildMail   MailBuilder
	expiresIn   time.Duration
	companyName string // 任意。空なら本文から省略。
}

// NewCreateAdminInvitationUseCase は SES マジックリンク方式の招待作成 usecase を組み立てる。
// sender が nil のときはメール送信をスキップする（ローカル開発時のフォールバック）。
func NewCreateAdminInvitationUseCase(
	r repository.AdminInvitationRepository,
	sender MagicLinkSender,
	buildLink LinkBuilder,
	buildMail MailBuilder,
) *CreateAdminInvitationUseCase {
	return &CreateAdminInvitationUseCase{
		repo:      r,
		sender:    sender,
		buildLink: buildLink,
		buildMail: buildMail,
		expiresIn: 7 * 24 * time.Hour,
	}
}

type CreateAdminInvitationInput struct {
	// TargetWorkspace は招待先ワークスペース。招待できるのは自分の所属先だけなので、
	// handler が actor 自身の所属を入れる。テナントを横断して招く経路は無い。
	TargetWorkspace domain.WorkspaceRef
	Email           string
	Role            domain.RoleName
	Name            string
}

// Execute は token 発行 → invitations を pending で保存 → 受諾リンクメール送信、の順で招待を作る。
// sender 未設定ならメール送信はスキップ。メール送信失敗はエラーとして返す。
func (u *CreateAdminInvitationUseCase) Execute(ctx context.Context, in CreateAdminInvitationInput) (*domain.AdminInvitation, error) {
	// email はここで 1 度だけ正規形へ畳み、保存・照会・送信すべてでこの値を使う。
	// 生のまま保存すると、ログイン時の招待ゲート（正規形の OIDC メールで引く）と
	// 突き合わせられず「招待したのに招待が見つからない」状態になる。
	in.Email = domain.NormalizeEmail(in.Email)
	// 所属の決まらない招待は作らせない（受諾しても行き先が無い招待になる）。
	wid, hasWorkspace := in.TargetWorkspace.WorkspaceID()
	if !hasWorkspace {
		return nil, errors.New("targetWorkspace is required")
	}
	if in.Email == "" || in.Role == "" {
		return nil, errors.New("email, role are required")
	}
	workspaceID := &wid

	token := uuid.NewString()
	inv := &domain.AdminInvitation{
		WorkspaceID: workspaceID,
		Email:       in.Email,
		Role:        in.Role,
		Name:        in.Name,
		Status:      domain.InvitationStatusPending,
		Token:       &token,
		ExpiresAt:   time.Now().UTC().Add(u.expiresIn),
	}
	if err := u.repo.Create(ctx, inv); err != nil {
		log.Printf("CreateAdminInvitation: repo.Create failed email=%s role=%s workspaceID=%s: %v",
			in.Email, in.Role, wid, err)
		return nil, fmt.Errorf("create invitation: %w", err)
	}

	if u.sender == nil || u.buildLink == nil || u.buildMail == nil {
		// SES 未設定時はリンクをログ出力してフローを止めない（本番では sender 必須）。
		log.Printf("admin_invitation: sender not configured — skipping email. token=%s email=%s", token, in.Email)
		return inv, nil
	}

	link := u.buildLink(token)
	subject, htmlBody, textBody := u.buildMail(link, in.Name, u.companyName, string(in.Role))
	if err := u.sender.SendInvitationEmail(ctx, in.Email, subject, htmlBody, textBody); err != nil {
		// 送信失敗はエラーで返す（invitation は DB に残るので再送に使える）。SES エラー種別を判定できるよう詳細をログに残す。
		log.Printf("CreateAdminInvitation: SES SendInvitationEmail failed to=%s subject=%q: %v",
			in.Email, subject, err)
		return nil, fmt.Errorf("send invitation email: %w", err)
	}
	log.Printf("CreateAdminInvitation: invitation sent ok id=%d to=%s role=%s workspaceID=%s",
		inv.ID, in.Email, in.Role, wid)
	return inv, nil
}

// CancelAdminInvitationUseCase は既存招待の status を canceled に更新する。
type CancelAdminInvitationUseCase struct {
	repo repository.AdminInvitationRepository
}

func NewCancelAdminInvitationUseCase(r repository.AdminInvitationRepository) *CancelAdminInvitationUseCase {
	return &CancelAdminInvitationUseCase{repo: r}
}

// CancelAdminInvitationInput は取消対象と、取消を要求している管理者を表す。
type CancelAdminInvitationInput struct {
	ID             uint64
	ActorRole      domain.RoleName
	ActorWorkspace domain.WorkspaceRef
}

// ErrInvitationNotFound は対象の招待が存在しない場合に返す。
var ErrInvitationNotFound = errors.New("invitation not found")

// Execute は招待を canceled にする。
// super_admin は全社、company_admin は自社の招待のみ取消できる（それ以外は ErrForbidden）。
// 他社の招待は存在を漏らさないため ErrInvitationNotFound として扱う。
func (u *CancelAdminInvitationUseCase) Execute(ctx context.Context, in CancelAdminInvitationInput) error {
	if in.ID == 0 {
		return errors.New("id is required")
	}
	if in.ActorRole != domain.RoleSuperAdmin && in.ActorRole != domain.RoleCompanyAdmin {
		return ErrForbidden
	}

	inv, err := u.repo.FindByID(ctx, in.ID)
	if err != nil {
		return fmt.Errorf("find invitation: %w", err)
	}
	if inv == nil {
		return ErrInvitationNotFound
	}
	// 未所属の company_admin・招待の workspace_id 未設定はどちらも一致し得ないため、
	// 常に「見つからない」扱いになる。
	wid, ok := inv.WorkspaceRef().WorkspaceID()
	if in.ActorRole == domain.RoleCompanyAdmin && (!ok || !in.ActorWorkspace.Matches(wid)) {
		return ErrInvitationNotFound
	}

	// 0 行更新（= FindByID と UpdateStatus のあいだに招待が消えた）は repository が
	// domain.ErrNotFound で返す。ここで handler 用の番兵へ翻訳し、「取り消せていないのに 204」
	// にならないようにする。
	if err := u.repo.UpdateStatus(ctx, in.ID, domain.InvitationStatusCanceled); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ErrInvitationNotFound
		}
		return err
	}
	return nil
}
