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

// ListAdminInvitationsUseCase は 招待 一覧 を 取得 する Use Case。
// 依存 port: [repository.AdminInvitationRepository] (一覧 取得 を 委譲)。
type ListAdminInvitationsUseCase struct {
	repo repository.AdminInvitationRepository
}

func NewListAdminInvitationsUseCase(r repository.AdminInvitationRepository) *ListAdminInvitationsUseCase {
	return &ListAdminInvitationsUseCase{repo: r}
}

// ListAll は全社横断で招待一覧を返す。SuperAdmin (運営側) からのみ呼ばれる想定。
// 認可は handler 層で current user.role を見て行う。
func (u *ListAdminInvitationsUseCase) ListAll(ctx context.Context) ([]domain.AdminInvitation, error) {
	return u.repo.ListAll(ctx)
}

// ListByCompanyID は指定 company の招待一覧を返す。CompanyAdmin が自社のみを見る用。
func (u *ListAdminInvitationsUseCase) ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.AdminInvitation, error) {
	if companyID == 0 {
		return nil, errors.New("companyID is required")
	}
	return u.repo.ListByCompanyID(ctx, companyID)
}

// MagicLinkSender は invitation メール送信を usecase 側から呼ぶための抽象。
// infra/ses.Client が満たす想定。テストでは fake を差し込む。
//
// usecase 側に interface を置いている理由（DIP）:
//   - usecase は infra を知らず、infra/ses が「usecase が要求する I/F」を満たす設計にする
//   - 結果として infra/ses の API 変更が usecase に波及しない
//
// magicLink には ?token=<UUID> 付き受諾画面 URL を渡す。token 自体はメール本文には含めず、
// 必ず magicLink 経由（URL の一部）でしか露出させない。
type MagicLinkSender interface {
	SendInvitationEmail(ctx context.Context, to, subject, htmlBody, textBody string) error
}

// MailBuilder は招待メールの subject / HTML / text を組み立てる関数。
// infra/ses.BuildInvitationMail を直接渡す（受諾画面 URL の組み立ても呼び出し側で済ませてから渡す）。
type MailBuilder func(magicLink, displayName, companyName, role string) (subject, htmlBody, textBody string)

// LinkBuilder は invitation の token から「受諾画面の絶対 URL」を組み立てる関数。
// infra/ses.MagicLinkURL を直接渡す。
type LinkBuilder func(token string) string

// CreateAdminInvitationUseCase は 新規 招待 を 発行 + マジック リンク メール を 送る Use Case。
// 依存 port: [repository.AdminInvitationRepository] (DB 永続化) + [MagicLinkSender] (SES 経由 の メール 送信)。
type CreateAdminInvitationUseCase struct {
	repo        repository.AdminInvitationRepository
	sender      MagicLinkSender
	buildLink   LinkBuilder
	buildMail   MailBuilder
	expiresIn   time.Duration
	companyName string // 任意。空なら本文から省略。会社マスタから引いて渡す未来用に拡張可能。
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
	CompanyID   uint64
	Email       string
	Role        string
	DisplayName string
}

// Execute は招待を作成する。手順:
//  1. UUID v4 トークンを発行
//  2. invitations 行を pending + expires_at=今+7日 で保存
//  3. SES で受諾画面マジックリンクメールを送信（sender が nil ならスキップ）
//
// メール送信失敗は invitation 自体の作成成功と切り離さず、エラーとして返す。
// （部分的に成功してリンクが分からなくなる事故を避けるため、トランザクション境界はここで揃える）
func (u *CreateAdminInvitationUseCase) Execute(ctx context.Context, in CreateAdminInvitationInput) (*domain.AdminInvitation, error) {
	if in.CompanyID == 0 || in.Email == "" || in.Role == "" {
		return nil, errors.New("companyID, email, role are required")
	}

	token := uuid.NewString()
	inv := &domain.AdminInvitation{
		CompanyID:   in.CompanyID,
		Email:       in.Email,
		Role:        in.Role,
		DisplayName: in.DisplayName,
		Status:      domain.InvitationStatusPending,
		Token:       &token,
		ExpiresAt:   time.Now().UTC().Add(u.expiresIn),
	}
	if err := u.repo.Create(ctx, inv); err != nil {
		return nil, fmt.Errorf("create invitation: %w", err)
	}

	if u.sender == nil || u.buildLink == nil || u.buildMail == nil {
		// ローカル開発などで SES が無いときは、リンクをログ出力してフローを止めない。
		// 本番では sender が必須なので、APP_BASE_URL / SES 設定がある前提で起動する。
		log.Printf("admin_invitation: sender not configured — skipping email. token=%s email=%s", token, in.Email)
		return inv, nil
	}

	link := u.buildLink(token)
	subject, htmlBody, textBody := u.buildMail(link, in.DisplayName, u.companyName, in.Role)
	if err := u.sender.SendInvitationEmail(ctx, in.Email, subject, htmlBody, textBody); err != nil {
		// 送信失敗は呼び出し側にエラーで返す。invitation 自体は DB に残るので、UI から再送信機能を作るときに使える。
		return nil, fmt.Errorf("send invitation email: %w", err)
	}
	return inv, nil
}

// CancelAdminInvitationUseCase は 既存 招待 の status を canceled に 更新 する Use Case。
// 依存 port: [repository.AdminInvitationRepository]。
type CancelAdminInvitationUseCase struct {
	repo repository.AdminInvitationRepository
}

func NewCancelAdminInvitationUseCase(r repository.AdminInvitationRepository) *CancelAdminInvitationUseCase {
	return &CancelAdminInvitationUseCase{repo: r}
}

func (u *CancelAdminInvitationUseCase) Execute(ctx context.Context, id uint64) error {
	if id == 0 {
		return errors.New("id is required")
	}
	return u.repo.UpdateStatus(ctx, id, domain.InvitationStatusCanceled)
}
