package repository

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type AdminInvitationRepository interface {
	// ListAll は全社横断で招待を返す。SuperAdmin (運営) のダッシュボード用。
	ListAll(ctx context.Context) ([]domain.AdminInvitation, error)
	// ListByCompanyID は CompanyAdmin が自社の招待のみを見るための query。
	ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.AdminInvitation, error)
	// FindPendingByEmail は招待受諾フローで「ログインしてきたユーザーが招待ユーザーか」
	// を判定するために使う。同一 email で複数 pending があれば最新を返す。
	FindPendingByEmail(ctx context.Context, email string) (*domain.AdminInvitation, error)
	Create(ctx context.Context, inv *domain.AdminInvitation) error
	UpdateStatus(ctx context.Context, id uint64, status string) error
}

type adminInvitationRepository struct{ db *gorm.DB }

func NewAdminInvitationRepository(db *gorm.DB) AdminInvitationRepository {
	return &adminInvitationRepository{db: db}
}

// 一覧 API は「未承諾の招待」を返すため pending 以外（accepted / canceled）は除外する。
// 監査目的で行は物理削除せず status のみ更新している。
func (r *adminInvitationRepository) ListAll(ctx context.Context) ([]domain.AdminInvitation, error) {
	var rows []domain.AdminInvitation
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.InvitationStatusPending).
		Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *adminInvitationRepository) ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.AdminInvitation, error) {
	var rows []domain.AdminInvitation
	err := r.db.WithContext(ctx).
		Where("company_id = ? AND status = ?", companyID, domain.InvitationStatusPending).
		Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *adminInvitationRepository) FindPendingByEmail(ctx context.Context, email string) (*domain.AdminInvitation, error) {
	var row domain.AdminInvitation
	err := r.db.WithContext(ctx).
		Where("email = ? AND status = ?", email, domain.InvitationStatusPending).
		Order("created_at DESC").First(&row).Error
	if err != nil {
		// レコードが無いケースは「招待ユーザーではない通常サインアップ」なので nil, nil を返す
		// （呼び出し側でデフォルトロールにフォールバックする想定）
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *adminInvitationRepository) Create(ctx context.Context, inv *domain.AdminInvitation) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *adminInvitationRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return r.db.WithContext(ctx).Model(&domain.AdminInvitation{}).Where("id = ?", id).Update("status", status).Error
}

// CognitoAdminClient は Cognito AdminCreateUser など管理 API を抽象化する interface。
// AWS SDK 連携は別 PR で実装する。
type CognitoAdminClient interface {
	InviteUser(ctx context.Context, email, displayName, role string) (cognitoSub string, err error)
}

type stubCognitoAdmin struct{}

func NewStubCognitoAdminClient() CognitoAdminClient { return &stubCognitoAdmin{} }

func (c *stubCognitoAdmin) InviteUser(_ context.Context, email, _, _ string) (string, error) {
	return "stub-sub-" + email, nil
}
