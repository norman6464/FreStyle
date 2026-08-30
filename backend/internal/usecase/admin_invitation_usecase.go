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
// 全社横断（ListAll）とワークスペース単位（ListByWorkspaceID）の 2 系統を公開する集約 read usecase
//
//naminglint:allow 全社横断 ListAll・ワークスペース単位 ListByWorkspaceID・会社指定 ListByCompanyID
type ListAdminInvitationsUseCase struct {
	repo      repository.AdminInvitationRepository
	companies repository.CompanyRepository
}

func NewListAdminInvitationsUseCase(
	r repository.AdminInvitationRepository,
	companies repository.CompanyRepository,
) *ListAdminInvitationsUseCase {
	return &ListAdminInvitationsUseCase{repo: r, companies: companies}
}

// ListAll は全社横断で招待一覧を返す（SuperAdmin 専用、認可は handler 層）。
func (u *ListAdminInvitationsUseCase) ListAll(ctx context.Context) ([]domain.AdminInvitation, error) {
	return u.repo.ListAll(ctx)
}

// ListByWorkspaceID は指定ワークスペースの招待一覧を返す（CompanyAdmin の自社一覧と、
// SuperAdmin が任意の会社を指定するときの両方で使う。認可は handler 層）。
func (u *ListAdminInvitationsUseCase) ListByWorkspaceID(ctx context.Context, workspaceID string) ([]domain.AdminInvitation, error) {
	if workspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	return u.repo.ListByWorkspaceID(ctx, workspaceID)
}

// ListByCompanyID は指定 company の招待一覧を返す（SuperAdmin の ?companyId= 絞り込み）。
//
// invitations は company_id を持たず workspace_id だけを持つため、会社に対応する
// ワークスペースへ読み替えてから引く（companies 1 : 1 workspaces）。会社が見つからない・
// ワークスペース未紐付けの会社は 0 件（その会社宛の招待は存在し得ない）。
func (u *ListAdminInvitationsUseCase) ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.AdminInvitation, error) {
	if companyID == 0 {
		return nil, errors.New("companyID is required")
	}
	if u.companies == nil {
		return nil, errors.New("company repository not configured")
	}
	company, err := u.companies.FindByID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("find company: %w", err)
	}
	if company == nil || company.WorkspaceID == nil {
		return []domain.AdminInvitation{}, nil
	}
	return u.repo.ListByWorkspaceID(ctx, *company.WorkspaceID)
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
	companies   repository.CompanyRepository
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
	companies repository.CompanyRepository,
	sender MagicLinkSender,
	buildLink LinkBuilder,
	buildMail MailBuilder,
) *CreateAdminInvitationUseCase {
	return &CreateAdminInvitationUseCase{
		repo:      r,
		companies: companies,
		sender:    sender,
		buildLink: buildLink,
		buildMail: buildMail,
		expiresIn: 7 * 24 * time.Hour,
	}
}

type CreateAdminInvitationInput struct {
	// CompanyID は SuperAdmin が「どの会社へ招くか」を company の id で指定する入口。
	// TargetWorkspace が設定されているときは使わない。
	CompanyID uint64
	// TargetWorkspace は CompanyAdmin が自社へ招くときの所属ワークスペース。
	// 設定されていればこちらを優先し、会社からの引き直しを省く
	// （actor 自身の所属なので、company を経由するまでもなく確定している）。
	TargetWorkspace domain.WorkspaceRef
	Email           string
	Role            domain.RoleName
	Name            string
}

// resolveInvitationWorkspace は招待先 company に対応する workspace_id を引く。
//
// invitations の所属参照は workspace_id ただ 1 つで、company_id は撤去済み。一方 API は
// 「どの会社へ招くか」を company の id で受ける（クライアントにテナント ID を送らせない）。
// その 2 つを繋ぐのがこの解決で、招待作成の入口 2 経路（マジックリンク / 一時パスワード）が
// 共有する。会社が見つからない・まだワークスペースに紐付いていない場合はエラーにする
// （NULL のまま招待を作ると、受諾しても所属が決まらない招待になる）。
func resolveInvitationWorkspace(
	ctx context.Context,
	companies repository.CompanyRepository,
	in CreateAdminInvitationInput,
) (*string, error) {
	// CompanyAdmin 経路は actor 自身の所属がそのまま招待先なので、会社を引き直さない。
	if wid, ok := in.TargetWorkspace.WorkspaceID(); ok {
		return &wid, nil
	}
	companyID := in.CompanyID
	if companies == nil {
		return nil, errors.New("company repository not configured")
	}
	company, err := companies.FindByID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("find company: %w", err)
	}
	if company == nil {
		return nil, fmt.Errorf("company %d not found", companyID)
	}
	if company.WorkspaceID == nil {
		return nil, fmt.Errorf("company %d has no workspace", companyID)
	}
	return company.WorkspaceID, nil
}

// Execute は token 発行 → invitations を pending で保存 → 受諾リンクメール送信、の順で招待を作る。
// sender 未設定ならメール送信はスキップ。メール送信失敗はエラーとして返す。
func (u *CreateAdminInvitationUseCase) Execute(ctx context.Context, in CreateAdminInvitationInput) (*domain.AdminInvitation, error) {
	// email はここで 1 度だけ正規形へ畳み、保存・照会・送信すべてでこの値を使う。
	// 生のまま保存すると、ログイン時の招待ゲート（正規形の OIDC メールで引く）と
	// 突き合わせられず「招待したのに招待が見つからない」状態になる。
	in.Email = domain.NormalizeEmail(in.Email)
	if _, hasWorkspace := in.TargetWorkspace.WorkspaceID(); !hasWorkspace && in.CompanyID == 0 {
		return nil, errors.New("companyID or targetWorkspace is required")
	}
	if in.Email == "" || in.Role == "" {
		return nil, errors.New("email, role are required")
	}

	workspaceID, err := resolveInvitationWorkspace(ctx, u.companies, in)
	if err != nil {
		return nil, err
	}

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
	subject, htmlBody, textBody := u.buildMail(link, in.Name, u.companyName, string(in.Role))
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
