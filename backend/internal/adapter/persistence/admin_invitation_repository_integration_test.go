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
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewAdminInvitationRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "invitations")

	future := time.Now().UTC().Add(time.Hour)
	past := time.Now().UTC().Add(-time.Hour)

	// 直接 INSERT で status / expires_at / token を作り込む。
	insert := func(id uint64, company uint64, email, status, token string, expires time.Time) {
		_, err := sqlDB.Exec(
			`INSERT INTO invitations (id, company_id, email, role, name, status, token, expires_at, created_at)
			 VALUES ($1, $2, $3, $4, 'n', $5, $6, $7, NOW())`,
			id, company, email, domain.RoleCompanyAdmin, status, token, expires,
		)
		require.NoError(t, err)
	}

	t.Run("FindPendingByToken は pending かつ未期限切れのみ返す", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "invitations")
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
		testsupport.TruncateAll(t, sqlDB, "invitations")
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
		testsupport.TruncateAll(t, sqlDB, "invitations")
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
		testsupport.TruncateAll(t, sqlDB, "invitations")
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
		testsupport.TruncateAll(t, sqlDB, "invitations")
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

// TestAdminInvitationRepository_ListByWorkspaceID_Integration は CompanyAdmin 経路
// （FRESTYLE-401）が別ワークスペースの招待を返さないこと、pending のみ返すことを
// 実 Postgres で固定する。InsertInvitation の company_id→workspace_id dual-write
// （FRESTYLE-399）に依存するため repo.Create 経由でデータを作る。
func TestAdminInvitationRepository_ListByWorkspaceID_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewAdminInvitationRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, append([]string{"invitations"}, tenantBridgeTables...)...)

	insertCompany(t, sqlDB, 1, "会社 A", true)
	insertCompany(t, sqlDB, 2, "会社 B", true)
	runStartupBackfill(ctx, t, sqlDB)
	ws1 := companyWorkspaceID(t, sqlDB, 1)
	ws2 := companyWorkspaceID(t, sqlDB, 2)
	require.True(t, ws1.Valid)
	require.True(t, ws2.Valid)
	require.NotEqual(t, ws1.UUID, ws2.UUID)

	mk := func(companyID uint64, email string) *domain.AdminInvitation {
		return &domain.AdminInvitation{
			CompanyID: companyID, Email: email, Role: domain.RoleCompanyAdmin,
			Name: "n", Status: domain.InvitationStatusPending, ExpiresAt: time.Now().UTC().Add(time.Hour),
		}
	}
	require.NoError(t, repo.Create(ctx, mk(1, "pending-a@example.com")))
	require.NoError(t, repo.Create(ctx, mk(2, "other-workspace@example.com")))

	accepted := mk(1, "accepted-a@example.com")
	require.NoError(t, repo.Create(ctx, accepted))
	require.NoError(t, repo.UpdateStatus(ctx, accepted.ID, domain.InvitationStatusAccepted))

	got, err := repo.ListByWorkspaceID(ctx, ws1.UUID.String())
	require.NoError(t, err)
	require.Len(t, got, 1, "他ワークスペースの招待・pending 以外は混ざらない")
	require.Equal(t, "pending-a@example.com", got[0].Email)

	empty, err := repo.ListByWorkspaceID(ctx, "")
	require.NoError(t, err)
	require.Empty(t, empty, "空 ID は該当なし扱い")

	invalid, err := repo.ListByWorkspaceID(ctx, "not-a-uuid")
	require.NoError(t, err)
	require.Empty(t, invalid, "不正な形式の ID も該当なし扱い")
}

// TestAdminInvitationRepository_FindByID_WorkspaceID_Integration は FindByID が
// workspace_id も返すこと（招待取消の対象側比較が使う値）を実 Postgres で固定する。
func TestAdminInvitationRepository_FindByID_WorkspaceID_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewAdminInvitationRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, append([]string{"invitations"}, tenantBridgeTables...)...)

	insertCompany(t, sqlDB, 1, "会社 A", true)
	runStartupBackfill(ctx, t, sqlDB)
	ws1 := companyWorkspaceID(t, sqlDB, 1)
	require.True(t, ws1.Valid)

	inv := &domain.AdminInvitation{
		CompanyID: 1, Email: "a@example.com", Role: domain.RoleCompanyAdmin,
		Name: "n", Status: domain.InvitationStatusPending, ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	require.NoError(t, repo.Create(ctx, inv))

	got, err := repo.FindByID(ctx, inv.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.WorkspaceID)
	require.Equal(t, ws1.UUID.String(), *got.WorkspaceID)

	// バックフィル前に作られた行を模して workspace_id を直接 NULL のまま挿入する
	// （repo.Create は dual-write するため経由できない）。domain 側は nil のまま返すこと。
	_, err = sqlDB.Exec(
		`INSERT INTO invitations (id, company_id, workspace_id, email, role, name, status, expires_at, created_at)
		 VALUES (999999, 1, NULL, 'legacy@example.com', 'company_admin', 'n', 'pending', $1, now())`,
		time.Now().UTC().Add(time.Hour),
	)
	require.NoError(t, err)
	legacy, err := repo.FindByID(ctx, 999999)
	require.NoError(t, err)
	require.NotNil(t, legacy)
	require.Nil(t, legacy.WorkspaceID, "バックフィル未到達の行は workspace_id が nil のまま")
}
