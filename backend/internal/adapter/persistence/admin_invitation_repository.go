package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// adminInvitationRepository は [repository.AdminInvitationRepository] の GORM 実装。
type adminInvitationRepository struct{ db *gorm.DB }

func NewAdminInvitationRepository(db *gorm.DB) repository.AdminInvitationRepository {
	return &adminInvitationRepository{db: db}
}

// pending 以外（accepted / canceled）は除外する（行は物理削除せず status のみ更新）。
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
		// 該当なしは招待ユーザーでない通常サインアップなので nil, nil を返す。
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *adminInvitationRepository) FindPendingByToken(ctx context.Context, token string) (*domain.AdminInvitation, error) {
	if token == "" {
		return nil, nil
	}
	var row domain.AdminInvitation
	// 期限切れは DB 側で弾く。比較は DB 関数でなく Go の UTC 現在時刻をバインドする
	// （DB エンジン非依存 / ローカル TZ 設定に左右されない）。
	err := r.db.WithContext(ctx).
		Where("token = ? AND status = ? AND expires_at > ?", token, domain.InvitationStatusPending, time.Now().UTC()).
		First(&row).Error
	if err != nil {
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
