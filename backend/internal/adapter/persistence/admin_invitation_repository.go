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

// NewAdminInvitationRepository は GORM ベース の [repository.AdminInvitationRepository] を 返す。
func NewAdminInvitationRepository(db *gorm.DB) repository.AdminInvitationRepository {
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

func (r *adminInvitationRepository) FindPendingByToken(ctx context.Context, token string) (*domain.AdminInvitation, error) {
	if token == "" {
		return nil, nil
	}
	var row domain.AdminInvitation
	// expires_at は招待作成時に必ず未来の値を入れる前提（usecase 側で 7 日後をセット）。
	// 期限切れを DB 側で弾くことで「フロントで時刻がズレていても安全」な検証になる。
	//
	// 時刻比較は DB 関数（NOW() / UTC_TIMESTAMP()）ではなく Go 側で生成した UTC 現在時刻を
	// パラメータバインドで渡す。理由:
	//   - DB エンジン横断のポータビリティ（PostgreSQL には UTC_TIMESTAMP() が無い）
	//   - GORM が time.Time を timestamptz として扱うため、DB のローカル TZ 設定に依存しない
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
