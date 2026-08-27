package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// companyRepository は [repository.CompanyRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type companyRepository struct{ db *sql.DB }

func NewCompanyRepository(db *sql.DB) repository.CompanyRepository {
	return &companyRepository{db: db}
}

// queries は保持している接続プールで sqlc の Queries を作る（別 pool を持たない）。
func (r *companyRepository) queries() *sqlcgen.Queries {
	return sqlcgen.New(r.db)
}

// withTx は 1 つのトランザクションを開き、その中でだけ有効な Queries を fn に渡す。
// fn がエラーを返せば（あるいは Commit に失敗すれば）書き込みはすべて巻き戻る。
func (r *companyRepository) withTx(ctx context.Context, fn func(qtx *sqlcgen.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // Commit 済みなら no-op
	if err := fn(sqlcgen.New(r.db).WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit()
}

func toDomainCompany(row sqlcgen.Company) domain.Company {
	return domain.Company{
		ID:                       uint64(row.ID),
		Name:                     row.Name,
		IsActive:                 row.IsActive,
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
	}
}

func (r *companyRepository) ListAll(ctx context.Context) ([]domain.Company, error) {
	q := r.queries()
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
		return nil, domain.ErrNotFound // 存在し得ない id = not found
	}
	q := r.queries()
	row, err := q.GetCompanyByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		// AiChatEnabledForUserUseCase が domain.ErrNotFound を見て「会社行なし = 既定 true」にする契約を維持。
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	c := toDomainCompany(row)
	return &c, nil
}


// UpdateActive は会社アカウントの有効/無効を更新する。false で無効化（その会社の全ユーザーが利用不可）。
// 対象会社が存在せず 0 件更新だった場合は domain.ErrNotFound を返す（handler が 404 にマップ）。
// 対応する workspaces 行にも同じ値を写す。会社の更新と写しは同じトランザクションで行う。
func (r *companyRepository) UpdateActive(ctx context.Context, companyID uint64, active bool) error {
	id64, ok := toInt64ID(companyID)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = not found
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
			return domain.ErrNotFound
		}
		return qtx.MirrorCompanySettingsToWorkspace(ctx, id64)
	})
}
