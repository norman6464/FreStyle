package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// sqlConsoleRepository は read-only SQL コンソールの実装。
// クエリは read-only トランザクション内で実行し、statement_timeout と行数上限を課す。
type sqlConsoleRepository struct {
	db *gorm.DB
}

// NewSQLConsoleRepository は SQLConsoleRepository 実装を生成する。
func NewSQLConsoleRepository(db *gorm.DB) repository.SQLConsoleRepository {
	return &sqlConsoleRepository{db: db}
}

// RunReadOnly は read-only トランザクション内で query を実行し、最大 maxRows 行を返す。
// SET TRANSACTION READ ONLY により、万一 write クエリが渡っても Postgres が拒否する（最終防壁）。
func (r *sqlConsoleRepository) RunReadOnly(ctx context.Context, query string, maxRows int) (*repository.SQLQueryResult, error) {
	result := &repository.SQLQueryResult{Columns: []string{}, Rows: [][]any{}}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 読み取り専用 + タイムアウトをトランザクション全体に課す。
		if err := tx.Exec("SET TRANSACTION READ ONLY").Error; err != nil {
			return err
		}
		if err := tx.Exec("SET LOCAL statement_timeout = 5000").Error; err != nil {
			return err
		}

		rows, err := tx.Raw(query).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			return err
		}
		result.Columns = cols

		for rows.Next() {
			if len(result.Rows) >= maxRows {
				result.Truncated = true
				break
			}
			cells := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range cells {
				ptrs[i] = &cells[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				return err
			}
			// []byte（text / bytea 等）は JSON で扱いやすいよう文字列化する。
			for i, c := range cells {
				if b, ok := c.([]byte); ok {
					cells[i] = string(b)
				}
			}
			result.Rows = append(result.Rows, cells)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
