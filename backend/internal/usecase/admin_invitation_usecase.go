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
//naminglint:allow 全社横断 ListAll と自社 ListByCompanyID の 2 系統を公開する集約 read usecase
type ListAdminInvitationsUseCase struct {
	repo repository.AdminInvitationRepository
}

func NewListAdminInvitationsUseCase(r repository.AdminInvitationRepository) *ListAdminInvitationsUseCase {
	return &ListAdminInvitationsUseCase{repo: r}
}

// ListAll は全社横断で招待一覧を返す（SuperAdmin 専用、認可は handler 層）。
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
	CompanyID uint64
	Email     string
	Role      string
	Name      string
}

// Execute は token 発行 → invitations を pending で保存 → 受諾リンクメール送信、の順で招待を作る。
// sender 未設定ならメール送信はスキップ。メール送信失敗はエラーとして返す。
func (u *CreateAdminInvitationUseCase) Execute(ctx context.Context, in CreateAdminInvitationInput) (*domain.AdminInvitation, error) {
	if in.CompanyID == 0 || in.Email == "" || in.Role == "" {
		return nil, errors.New("companyID, email, role are required")
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
		log.Printf("CreateAdminInvitation: repo.Create failed email=%s role=%s companyID=%d: %v",
			in.Email, in.Role, in.CompanyID, err)
		return nil, fmt.Errorf("create invitation: %w", err)
	}

	if u.sender == nil || u.buildLink == nil || u.buildMail == nil {
		// SES 未設定時はリンクをログ出力してフローを止めない（本番では sender 必須）。
		log.Printf("admin_invitation: sender not configured — skipping email. token=%s email=%s", token, in.Email)
		return inv, nil
	}

	link := u.buildLink(token)
	subject, htmlBody, textBody := u.buildMail(link, in.Name, u.companyName, in.Role)
	if err := u.sender.SendInvitationEmail(ctx, in.Email, subject, htmlBody, textBody); err != nil {
		// 送信失敗はエラーで返す（invitation は DB に残るので再送に使える）。SES エラー種別を判定できるよう詳細をログに残す。
		log.Printf("CreateAdminInvitation: SES SendInvitationEmail failed to=%s subject=%q: %v",
			in.Email, subject, err)
		return nil, fmt.Errorf("send invitation email: %w", err)
	}
	log.Printf("CreateAdminInvitation: invitation sent ok id=%d to=%s role=%s companyID=%d",
		inv.ID, in.Email, in.Role, in.CompanyID)
	return inv, nil
}

// CancelAdminInvitationUseCase は既存招待の status を canceled に更新する。
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
