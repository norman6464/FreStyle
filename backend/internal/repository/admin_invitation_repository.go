package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type AdminInvitationRepository interface {
	// ListAll は全社横断で招待を返す。SuperAdmin (運営) のダッシュボード用。
	ListAll(ctx context.Context) ([]domain.AdminInvitation, error)
	// ListByCompanyID は CompanyAdmin が自社の招待のみを見るための query。
	ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.AdminInvitation, error)
	Create(ctx context.Context, inv *domain.AdminInvitation) error
	UpdateStatus(ctx context.Context, id uint64, status string) error
}

type adminInvitationRepository struct{ db *gorm.DB }

func NewAdminInvitationRepository(db *gorm.DB) AdminInvitationRepository {
	return &adminInvitationRepository{db: db}
}

func (r *adminInvitationRepository) ListAll(ctx context.Context) ([]domain.AdminInvitation, error) {
	var rows []domain.AdminInvitation
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *adminInvitationRepository) ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.AdminInvitation, error) {
	var rows []domain.AdminInvitation
	err := r.db.WithContext(ctx).Where("company_id = ?", companyID).Order("created_at DESC").Find(&rows).Error
	return rows, err
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
