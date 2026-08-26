//go:build integration

package persistence_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestAdminInvitationRepository_Auth_Integration は認可に関わる読み書き
// （token・有効期限・status の判定、ID 検索の not-found、status 遷移、token 秘匿）を
// 実 Postgres で固定する。移行で認可判定が緩まないことの根拠にする。
func TestAdminInvitationRepository_Auth_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	sqlDB := testsupport.SQLDB(t, db)
	repo := persistence.NewAdminInvitationRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "invitations")

	future := time.Now().UTC().Add(time.Hour)
	past := time.Now().UTC().Add(-time.Hour)

	// 直接 INSERT で status / expires_at / token を作り込む。
	insert := func(id uint64, company uint64, email, status, token string, expires time.Time) {
		require.NoError(t, db.Exec(
			`INSERT INTO invitations (id, company_id, email, role, name, status, token, expires_at, created_at)
			 VALUES (?, ?, ?, ?, 'n', ?, ?, ?, NOW())`,
			id, company, email, domain.RoleCompanyAdmin, status, token, expires,
		).Error)
	}

	t.Run("FindPendingByToken は pending かつ未期限切れのみ返す", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "invitations")
		insert(1, 1, "ok@example.com", domain.InvitationStatusPending, "tok-ok", future)
		insert(2, 1, "exp@example.com", domain.InvitationStatusPending, "tok-expired", past)
		insert(3, 1, "acc@example.com", domain.InvitationStatusAccepted, "tok-accepted", future)
		insert(4, 1, "can@example.com", domain.InvitationStatusCanceled, "tok-canceled", future)

		got, err := repo.FindPendingByToken(ctx, "tok-ok")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, uint64(1), got.ID)

		for _, tok := range []string{"tok-expired", "tok-accepted", "tok-canceled", "no-such-token"} {
			got, err := repo.FindPendingByToken(ctx, tok)
			require.NoError(t, err)
			require.Nil(t, got, "token=%s は受理されてはいけない", tok)
		}

		// 空 token は DB を引かずに nil, nil。
		got, err = repo.FindPendingByToken(ctx, "")
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("FindByID は該当なし・id=0 で nil,nil", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "invitations")
		insert(10, 2, "byid@example.com", domain.InvitationStatusAccepted, "tok-byid", future)

		got, err := repo.FindByID(ctx, 10)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, uint64(2), got.CompanyID, "会社スコープ認可に使う company_id を返す")
		require.Equal(t, domain.InvitationStatusAccepted, got.Status, "status に関わらず ID で引ける")

		none, err := repo.FindByID(ctx, 99999)
		require.NoError(t, err)
		require.Nil(t, none)

		zero, err := repo.FindByID(ctx, 0)
		require.NoError(t, err)
		require.Nil(t, zero)
	})

	t.Run("UpdateStatus は対象 1 行だけ status を変える", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "invitations")
		insert(20, 1, "a@example.com", domain.InvitationStatusPending, "tok-a", future)
		insert(21, 1, "b@example.com", domain.InvitationStatusPending, "tok-b", future)

		require.NoError(t, repo.UpdateStatus(ctx, 20, domain.InvitationStatusCanceled))

		got20, err := repo.FindByID(ctx, 20)
		require.NoError(t, err)
		require.Equal(t, domain.InvitationStatusCanceled, got20.Status)
		got21, err := repo.FindByID(ctx, 21)
		require.NoError(t, err)
		require.Equal(t, domain.InvitationStatusPending, got21.Status, "他の行は変わらない")

		// canceled は pending ではないので token でも引けなくなる。
		none, err := repo.FindPendingByToken(ctx, "tok-a")
		require.NoError(t, err)
		require.Nil(t, none)
	})

	t.Run("FindPendingByEmail は pending のみ返し expires は問わない（token とは非対称）", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "invitations")
		// pending だが期限切れでも email 検索では返る（GORM 版に expires フィルタは無い）。
		insert(30, 1, "pend@example.com", domain.InvitationStatusPending, "tok-30", past)
		insert(31, 1, "done@example.com", domain.InvitationStatusAccepted, "tok-31", future)

		got, err := repo.FindPendingByEmail(ctx, "pend@example.com")
		require.NoError(t, err)
		require.NotNil(t, got, "pending は期限切れでも email 検索で返る")
		require.Equal(t, uint64(30), got.ID)

		none, err := repo.FindPendingByEmail(ctx, "done@example.com")
		require.NoError(t, err)
		require.Nil(t, none, "accepted は返らない")
	})

	t.Run("token は JSON へ露出しない（秘匿）", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "invitations")
		insert(40, 1, "secret@example.com", domain.InvitationStatusPending, "super-secret-token", future)
		got, err := repo.FindPendingByToken(ctx, "super-secret-token")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.Token)
		require.Equal(t, "super-secret-token", *got.Token, "内部では token を保持する")

		b, err := json.Marshal(got)
		require.NoError(t, err)
		require.NotContains(t, string(b), "super-secret-token", "token を JSON で返してはいけない")
		require.NotContains(t, strings.ToLower(string(b)), "\"token\"", "token フィールド自体を出さない")
	})
}
