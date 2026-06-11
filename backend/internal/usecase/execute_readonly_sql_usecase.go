package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// SQLConsoleMaxRows は read-only SQL コンソールが 1 回に返す行数の上限。
const SQLConsoleMaxRows = 1000

var (
	// ErrEmptyQuery はクエリが空のとき。
	ErrEmptyQuery = errors.New("クエリが空です")
	// ErrNotReadOnlyQuery は SELECT / WITH 以外（書き込み等）のとき。
	ErrNotReadOnlyQuery = errors.New("読み取り専用の単一 SELECT / WITH クエリのみ実行できます")
)

// ExecuteReadOnlySQLInput は read-only SQL コンソールの入力。
type ExecuteReadOnlySQLInput struct {
	Query string
}

// ExecuteReadOnlySQLUseCase は super_admin が入力した SQL を、読み取り専用の前提で実行する。
// バリデーション（SELECT / WITH のみ）は防御の一段目で、最終保証は repository の
// read-only トランザクション。
type ExecuteReadOnlySQLUseCase struct {
	repo repository.SQLConsoleRepository
}

// NewExecuteReadOnlySQLUseCase は ExecuteReadOnlySQLUseCase を生成する。
func NewExecuteReadOnlySQLUseCase(r repository.SQLConsoleRepository) *ExecuteReadOnlySQLUseCase {
	return &ExecuteReadOnlySQLUseCase{repo: r}
}

// Execute はクエリを検証し、読み取り専用で実行して結果を返す。
func (u *ExecuteReadOnlySQLUseCase) Execute(ctx context.Context, in ExecuteReadOnlySQLInput) (*repository.SQLQueryResult, error) {
	q := strings.TrimSpace(in.Query)
	if q == "" {
		return nil, ErrEmptyQuery
	}
	if !isReadOnlyQuery(q) {
		return nil, ErrNotReadOnlyQuery
	}
	return u.repo.RunReadOnly(ctx, q, SQLConsoleMaxRows)
}

// isReadOnlyQuery は単一の SELECT / WITH クエリのみを許可する。
// 末尾の単一セミコロンは許容するが、文中のセミコロン（複文）は拒否する。
// 最終的な書き込み拒否は repository の read-only トランザクションが担保する。
func isReadOnlyQuery(q string) bool {
	s := strings.TrimSpace(q)
	s = strings.TrimRight(s, ";")
	if strings.Contains(s, ";") {
		return false
	}
	lower := strings.ToLower(s)
	return strings.HasPrefix(lower, "select") || strings.HasPrefix(lower, "with")
}
