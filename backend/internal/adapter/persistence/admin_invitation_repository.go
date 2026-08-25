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
// created_at は一意でないため、以下の一覧・単一取得はいずれも id をタイブレークに置いて順序を固定する
// （特に FindPendingByEmail は 1 件しか返さず、順序が揺れると「どの招待が受理されるか」が変わる）。
func (r *adminInvitationRepository) ListAll(ctx context.Context) ([]domain.AdminInvitation, error) {
	rows := make([]domain.AdminInvitation, 0)
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.InvitationStatusPending).
		Order("created_at DESC, id DESC").Find(&rows).Error
	return rows, err
}

func (r *adminInvitationRepository) ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.AdminInvitation, error) {
	rows := make([]domain.AdminInvitation, 0)
	err := r.db.WithContext(ctx).
		Where("company_id = ? AND status = ?", companyID, domain.InvitationStatusPending).
		Order("created_at DESC, id DESC").Find(&rows).Error
	return rows, err
}

// FindPendingByEmail は保留中の招待を email で引く。
// 突き合わせは lower() 同士で行う。ユーザー側の email は domain.NormalizeEmail で畳んだ値に
// 揃えて保存・照会するため、招待作成時に大文字混じりで入力された行を byte 一致で探すと
// 「招待したのに招待が見つからない」状態になる（同じアドレスの解釈が 2 つある状態を作らない）。
func (r *adminInvitationRepository) FindPendingByEmail(ctx context.Context, email string) (*domain.AdminInvitation, error) {
	var row domain.AdminInvitation
	err := r.db.WithContext(ctx).
		Where("lower(email) = lower(?) AND status = ?", email, domain.InvitationStatusPending).
		Order("created_at DESC, id DESC").First(&row).Error
	if err != nil {
		// 該当なしは招待ユーザーでない通常サインアップなので nil, nil を返す。
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *adminInvitationRepository) FindByID(ctx context.Context, id uint64) (*domain.AdminInvitation, error) {
	if id == 0 {
		return nil, nil
	}
	var row domain.AdminInvitation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
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
