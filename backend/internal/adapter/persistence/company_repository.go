package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// companyRepository は [repository.CompanyRepository] の実装。
// 読み取りは sqlc 生成コード（生 SQL）、書き込み（UpdateAiChatEnabled）は生 SQL の Exec。
type companyRepository struct{ db *gorm.DB }

func NewCompanyRepository(db *gorm.DB) repository.CompanyRepository {
	return &companyRepository{db: db}
}

func toDomainCompany(row sqlcgen.Company) domain.Company {
	return domain.Company{
		ID:                       uint64(row.ID),
		Name:                     row.Name,
		AiChatEnabledForTrainees: row.AiChatEnabledForTrainees,
		IsActive:                 row.IsActive,
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
	}
}

func (r *companyRepository) ListAll(ctx context.Context) ([]domain.Company, error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListCompanies(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Company, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainCompany(row))
	}
	return out, nil
}

func (r *companyRepository) FindByID(ctx context.Context, id uint64) (*domain.Company, error) {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil, gorm.ErrRecordNotFound // 存在し得ない id = not found
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	row, err := sqlcgen.New(sqlDB).GetCompanyByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		// AiChatEnabledForUserUseCase が ErrRecordNotFound を見て「会社行なし = 既定 true」にする契約を維持。
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	c := toDomainCompany(row)
	return &c, nil
}

// UpdateAiChatEnabled は ai_chat_enabled_for_trainees を更新する（生 SQL 直書き / updated_at も更新）。
// 対応する workspaces 行にも同じ値を写す（テナント設定の移行期間の二重書き）。
func (r *companyRepository) UpdateAiChatEnabled(ctx context.Context, companyID uint64, enabled bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		const q = `UPDATE companies SET ai_chat_enabled_for_trainees = ?, updated_at = NOW() WHERE id = ?`
		if err := tx.Exec(q, enabled, companyID).Error; err != nil {
			return err
		}
		return mirrorCompanySettingsTx(tx, companyID)
	})
}

// UpdateActive は会社アカウントの有効/無効を更新する。false で無効化（その会社の全ユーザーが利用不可）。
// 対象会社が存在せず 0 件更新だった場合は gorm.ErrRecordNotFound を返す（handler が 404 にマップ）。
// 対応する workspaces 行にも同じ値を写す。
func (r *companyRepository) UpdateActive(ctx context.Context, companyID uint64, active bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		const q = `UPDATE companies SET is_active = ?, updated_at = NOW() WHERE id = ?`
		res := tx.Exec(q, active, companyID)
		if res.Error != nil {
			return res.Error
		}
		// 「見つからない」の判定は companies 側の更新件数で行う。写し先（workspaces）の
		// 件数を見ると、まだ紐付いていない会社が 404 に化ける。
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return mirrorCompanySettingsTx(tx, companyID)
	})
}

// mirrorCompanySettingsTx は会社のテナント設定を、対応する workspaces 行へ写す。
//
// ai_chat_enabled_for_trainees / is_active は最終的に workspaces の列になる。今は companies が
// 正本で、workspaces 側はその写し（読み取りはどこも見ていない）。設定を書く経路が増えても
// 写し忘れないよう、2 列まとめて companies から写す 1 か所に集約する。
// まだワークスペースに紐付いていない会社（workspace_id IS NULL）は 0 件更新で、写す先が無い。
func mirrorCompanySettingsTx(db *gorm.DB, companyID uint64) error {
	return db.Exec(
		`UPDATE workspaces w
		 SET ai_chat_enabled_for_trainees = c.ai_chat_enabled_for_trainees,
		     is_active = c.is_active,
		     updated_at = NOW()
		 FROM companies c
		 WHERE c.id = ? AND c.workspace_id = w.id`, companyID,
	).Error
}
