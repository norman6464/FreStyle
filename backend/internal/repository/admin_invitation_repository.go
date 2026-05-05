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
	// FindPendingByToken はマジックリンク受諾フロー用。token 一致 & status=pending & expires_at 未経過のみ返す。
	// 該当なしは (nil, nil) を返し、呼び出し側で「無効/期限切れ token」として扱う。
	FindPendingByToken(ctx context.Context, token string) (*domain.AdminInvitation, error)
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

func (r *adminInvitationRepository) FindPendingByToken(ctx context.Context, token string) (*domain.AdminInvitation, error) {
	if token == "" {
		return nil, nil
	}
	var row domain.AdminInvitation
	// expires_at は招待作成時に必ず未来の値を入れる前提（usecase 側で 7 日後をセット）。
	// 期限切れを DB 側で弾くことで「フロントで時刻がズレていても安全」な検証になる。
	// UTC_TIMESTAMP() を使うことで DB サーバーのローカルタイムゾーン設定に依存せず比較する
	// （RDS が JST に設定されていてもアプリは UTC で時刻を扱う前提なので統一）。
	err := r.db.WithContext(ctx).
		Where("token = ? AND status = ? AND expires_at > UTC_TIMESTAMP()", token, domain.InvitationStatusPending).
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
