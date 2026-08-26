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
// クエリは sqlc 生成コード（生 SQL）で、GORM からは接続プール（*sql.DB）だけを借りる。
type companyRepository struct{ db *gorm.DB }

func NewCompanyRepository(db *gorm.DB) repository.CompanyRepository {
	return &companyRepository{db: db}
}

// queries は GORM が持つ接続プールを借りて sqlc の Queries を作る（別 pool を持たない）。
func (r *companyRepository) queries() (*sqlcgen.Queries, error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	return sqlcgen.New(sqlDB), nil
}

// withTx は 1 つのトランザクションを開き、その中でだけ有効な Queries を fn に渡す。
// fn がエラーを返せば（あるいは Commit に失敗すれば）書き込みはすべて巻き戻る。
func (r *companyRepository) withTx(ctx context.Context, fn func(qtx *sqlcgen.Queries) error) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // Commit 済みなら no-op
	if err := fn(sqlcgen.New(sqlDB).WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit()
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
	q, err := r.queries()
	if err != nil {
		return nil, err
	}
	rows, err := q.ListCompanies(ctx)
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
	q, err := r.queries()
	if err != nil {
		return nil, err
	}
	row, err := q.GetCompanyByID(ctx, id64)
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

// UpdateAiChatEnabled は ai_chat_enabled_for_trainees を更新する（updated_at も更新）。
// 対応する workspaces 行にも同じ値を写す（テナント設定の移行期間の二重書き）。
// 会社の更新と写しは同じトランザクションで行い、片方だけが確定した状態を作らない。
func (r *companyRepository) UpdateAiChatEnabled(ctx context.Context, companyID uint64, enabled bool) error {
	id64, ok := toInt64ID(companyID)
	if !ok {
		return nil // 存在し得ない id = 0 件更新と同じ（従来から件数は見ていない）
	}
	return r.withTx(ctx, func(qtx *sqlcgen.Queries) error {
		if err := qtx.UpdateCompanyAiChatEnabled(ctx, sqlcgen.UpdateCompanyAiChatEnabledParams{
			ID:                       id64,
			AiChatEnabledForTrainees: enabled,
		}); err != nil {
			return err
		}
		return qtx.MirrorCompanySettingsToWorkspace(ctx, id64)
	})
}

// UpdateActive は会社アカウントの有効/無効を更新する。false で無効化（その会社の全ユーザーが利用不可）。
// 対象会社が存在せず 0 件更新だった場合は gorm.ErrRecordNotFound を返す（handler が 404 にマップ）。
// 対応する workspaces 行にも同じ値を写す。会社の更新と写しは同じトランザクションで行う。
func (r *companyRepository) UpdateActive(ctx context.Context, companyID uint64, active bool) error {
	id64, ok := toInt64ID(companyID)
	if !ok {
		return gorm.ErrRecordNotFound // 存在し得ない id = not found
	}
	return r.withTx(ctx, func(qtx *sqlcgen.Queries) error {
		affected, err := qtx.UpdateCompanyActive(ctx, sqlcgen.UpdateCompanyActiveParams{
			ID:       id64,
			IsActive: active,
		})
		if err != nil {
			return err
		}
		// 「見つからない」の判定は companies 側の更新件数で行う。写し先（workspaces）の
		// 件数を見ると、まだ紐付いていない会社が 404 に化ける。
		if affected == 0 {
			return gorm.ErrRecordNotFound
		}
		return qtx.MirrorCompanySettingsToWorkspace(ctx, id64)
	})
}
