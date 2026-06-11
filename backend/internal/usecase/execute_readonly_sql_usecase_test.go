package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSQLConsoleRepo は呼び出されたクエリ / 上限を記録するだけの fake。
type fakeSQLConsoleRepo struct {
	gotQuery   string
	gotMaxRows int
	called     bool
	result     *repository.SQLQueryResult
}

func (f *fakeSQLConsoleRepo) RunReadOnly(_ context.Context, query string, maxRows int) (*repository.SQLQueryResult, error) {
	f.called = true
	f.gotQuery = query
	f.gotMaxRows = maxRows
	if f.result != nil {
		return f.result, nil
	}
	return &repository.SQLQueryResult{Columns: []string{}, Rows: [][]any{}}, nil
}

func TestExecuteReadOnlySQL_RejectsEmpty(t *testing.T) {
	repo := &fakeSQLConsoleRepo{}
	uc := usecase.NewExecuteReadOnlySQLUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.ExecuteReadOnlySQLInput{Query: "   "})

	require.ErrorIs(t, err, usecase.ErrEmptyQuery)
	assert.False(t, repo.called, "空クエリで repository を呼んではならない")
}

func TestExecuteReadOnlySQL_RejectsNonSelect(t *testing.T) {
	cases := []string{
		"DELETE FROM users",
		"UPDATE users SET role = 'super_admin'",
		"INSERT INTO users (id) VALUES (1)",
		"DROP TABLE users",
		"TRUNCATE users",
		"SELECT 1; DELETE FROM users", // 複文（文中セミコロン）は拒否
		"  alter table users add column x int",
	}
	for _, q := range cases {
		t.Run(q, func(t *testing.T) {
			repo := &fakeSQLConsoleRepo{}
			uc := usecase.NewExecuteReadOnlySQLUseCase(repo)

			_, err := uc.Execute(context.Background(), usecase.ExecuteReadOnlySQLInput{Query: q})

			require.ErrorIs(t, err, usecase.ErrNotReadOnlyQuery)
			assert.False(t, repo.called, "非 SELECT で repository を呼んではならない: %q", q)
		})
	}
}

func TestExecuteReadOnlySQL_AllowsSelectAndWith(t *testing.T) {
	cases := []string{
		"SELECT 1",
		"select id, email from users limit 10",
		"WITH t AS (SELECT 1 AS n) SELECT n FROM t",
		"SELECT count(*) FROM master_exercises;", // 末尾セミコロンは許容
	}
	for _, q := range cases {
		t.Run(q, func(t *testing.T) {
			repo := &fakeSQLConsoleRepo{}
			uc := usecase.NewExecuteReadOnlySQLUseCase(repo)

			_, err := uc.Execute(context.Background(), usecase.ExecuteReadOnlySQLInput{Query: q})

			require.NoError(t, err)
			require.True(t, repo.called, "SELECT/WITH は repository を呼ぶ: %q", q)
			assert.Equal(t, usecase.SQLConsoleMaxRows, repo.gotMaxRows)
		})
	}
}
