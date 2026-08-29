//go:build integration

package usecase_test

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBootstrapSuperAdmin_Integration は「最初の運営管理者ができた瞬間に招待免除の経路が閉じる」
// という不変条件を、実 PostgreSQL への同時サインアップで固定する。
//
// 免除の突き合わせを「畳んでから比べる」方式（strings.EqualFold）にすると、
// "ops@example.com" と "OPS@example.com" はアプリでは同一・DB の一意索引では別行になり、
// どちらも免除を通過してどちらも INSERT に成功する。同じアドレスでの多重作成を止めていたのは
// 設計ではなく一意索引の副作用でしかなかったので、大小文字だけ違う入力には防壁が無かった。
// ここでは「畳めば同じだが byte 列は全て違う」12 通りを同時に流し、
// できあがる super_admin が 1 人以下であることを確かめる。
func TestBootstrapSuperAdmin_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	users := persistence.NewUserRepository(sqlDB)
	ctx := context.Background()

	const bootstrapEmail = "ops@example.com"
	// 畳めば bootstrapEmail と同じだが、byte 列は互いに異なる 12 通り。
	variants := []string{
		"ops@example.com",
		"OPS@example.com",
		"Ops@example.com",
		"oPs@example.com",
		"opS@example.com",
		"OPs@example.com",
		"oPS@example.com",
		"OpS@example.com",
		"ops@Example.com",
		"ops@eXample.com",
		"ops@EXAMPLE.COM",
		"OPS@EXAMPLE.COM",
	}

	// 1 回で必ず競合が起きるとは限らないので複数回まわす。
	const rounds = 5
	for round := range rounds {
		testsupport.TruncateAll(t, sqlDB, "users", "user_oidc_identities")
		uc := usecase.NewUpsertUserFromIDTokenUseCase(users, nil, bootstrapEmail, persistence.NewUserInvitationTransactionRunner(sqlDB))

		results := make([]*domain.User, len(variants))
		errs := make([]error, len(variants))
		start := make(chan struct{})
		var wg sync.WaitGroup
		for i, email := range variants {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start // 全員を同じ瞬間に走らせる
				results[i], errs[i] = uc.Execute(ctx, usecase.UpsertUserFromIDTokenInput{
					CognitoSub:     fmt.Sprintf("race-%d-%d", round, i),
					Email:          email,
					IsCognitoAdmin: true,
				})
			}()
		}
		close(start)
		wg.Wait()

		accepted := 0
		for i := range variants {
			assert.NoErrorf(t, errs[i], "round %d: %q でエラー", round, variants[i])
			if results[i] != nil {
				accepted++
			}
		}

		created := queryStrings(t, sqlDB,
			`SELECT u.email FROM users u
			 JOIN roles r ON r.id = u.role_id
			 WHERE r.name = 'super_admin' AND u.deleted_at IS NULL
			 ORDER BY u.id`)
		// 不変条件の検証は assert（1 回落ちても後続の round を最後まで回して全体像を出す）。
		assert.LessOrEqualf(t, len(created), 1,
			"round %d: 競合で super_admin が %d 人できた（期待: 1 人以下）: %v",
			round, len(created), created)
		assert.Equalf(t, accepted, len(created),
			"round %d: 許可した数と作成された super_admin の数が食い違う: allowed=%d created=%v",
			round, accepted, created)

		var total int64
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM users WHERE deleted_at IS NULL`).Scan(&total))
		assert.LessOrEqualf(t, total, int64(1), "round %d: users 行が %d 件できた（期待: 1 件以下）", round, total)

		if len(created) == 1 {
			// 保存されるのは生の claim 値ではなく正規形（以後の byte 一致検索・一意索引と揃う）。
			assert.Equalf(t, bootstrapEmail, created[0], "round %d: 保存された email が正規形でない", round)
		}
	}
}

// queryStrings は 1 列の文字列を全行読み出す。
func queryStrings(t *testing.T, db *sql.DB, query string, args ...any) []string {
	t.Helper()
	rows, err := db.Query(query, args...)
	require.NoError(t, err)
	defer func() { require.NoError(t, rows.Close()) }()
	out := make([]string, 0)
	for rows.Next() {
		var v string
		require.NoError(t, rows.Scan(&v))
		out = append(out, v)
	}
	require.NoError(t, rows.Err())
	return out
}
