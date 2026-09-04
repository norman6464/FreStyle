//go:build integration

package persistence_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
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
	testsupport.TruncateAll(t, sqlDB, append([]string{"invitations"}, workspaceWriteTables...)...)

	// 所属参照（workspace_id）は workspaces への FK なので、実在するワークスペースを用意する。
	w1, w2 := uuid.New(), uuid.New()
	insertWorkspaceWithActive(t, sqlDB, w1, "ワークスペース A", true)
	insertWorkspaceWithActive(t, sqlDB, w2, "ワークスペース B", true)
	ws1 := w1.String()
	ws2 := w2.String()

	future := time.Now().UTC().Add(time.Hour)
	past := time.Now().UTC().Add(-time.Hour)

	// 直接 INSERT で status / expires_at / token を作り込む。
	insert := func(id uint64, workspaceID string, email, status, token string, expires time.Time) {
		_, err := sqlDB.Exec(
			`INSERT INTO invitations (id, workspace_id, email, role, name, status, token, expires_at, created_at)
			 VALUES ($1, $2, $3, $4, 'n', $5, $6, $7, NOW())`,
			id, workspaceID, email, domain.RoleCompanyAdmin, status, token, expires,
		)
		require.NoError(t, err)
	}

	t.Run("FindPendingByToken は pending かつ未期限切れのみ返す", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "invitations")
		insert(1, ws1, "ok@example.com", domain.InvitationStatusPending, "tok-ok", future)
		insert(2, ws1, "exp@example.com", domain.InvitationStatusPending, "tok-expired", past)
		insert(3, ws1, "acc@example.com", domain.InvitationStatusAccepted, "tok-accepted", future)
		insert(4, ws1, "can@example.com", domain.InvitationStatusCanceled, "tok-canceled", future)

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
		insert(10, ws2, "byid@example.com", domain.InvitationStatusAccepted, "tok-byid", future)

		got, err := repo.FindByID(ctx, 10)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.WorkspaceID)
		require.Equal(t, ws2, *got.WorkspaceID, "テナントスコープ認可に使う workspace_id を返す")
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
		insert(20, ws1, "a@example.com", domain.InvitationStatusPending, "tok-a", future)
		insert(21, ws1, "b@example.com", domain.InvitationStatusPending, "tok-b", future)

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
		insert(30, ws1, "pend@example.com", domain.InvitationStatusPending, "tok-30", past)
		insert(31, ws1, "done@example.com", domain.InvitationStatusAccepted, "tok-31", future)

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
		insert(40, ws1, "secret@example.com", domain.InvitationStatusPending, "super-secret-token", future)
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

// TestAdminInvitationRepository_ListByWorkspaceID_Integration は会社管理者の一覧経路が
// 別ワークスペースの招待を返さないこと、pending のみ返すことを実 Postgres で固定する。
func TestAdminInvitationRepository_ListByWorkspaceID_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewAdminInvitationRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, append([]string{"invitations"}, workspaceWriteTables...)...)

	ws1, ws2 := uuid.New(), uuid.New()
	insertWorkspaceWithActive(t, sqlDB, ws1, "ワークスペース A", true)
	insertWorkspaceWithActive(t, sqlDB, ws2, "ワークスペース B", true)

	mk := func(workspaceID string, email string) *domain.AdminInvitation {
		return &domain.AdminInvitation{
			WorkspaceID: &workspaceID, Email: email, Role: domain.RoleCompanyAdmin,
			Name: "n", Status: domain.InvitationStatusPending, ExpiresAt: time.Now().UTC().Add(time.Hour),
		}
	}
	require.NoError(t, repo.Create(ctx, mk(ws1.String(), "pending-a@example.com")))
	require.NoError(t, repo.Create(ctx, mk(ws2.String(), "other-workspace@example.com")))

	accepted := mk(ws1.String(), "accepted-a@example.com")
	require.NoError(t, repo.Create(ctx, accepted))
	require.NoError(t, repo.UpdateStatus(ctx, accepted.ID, domain.InvitationStatusAccepted))

	got, err := repo.ListByWorkspaceID(ctx, ws1.String())
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
	testsupport.TruncateAll(t, sqlDB, append([]string{"invitations"}, workspaceWriteTables...)...)

	ws1 := uuid.New()
	insertWorkspaceWithActive(t, sqlDB, ws1, "ワークスペース A", true)

	ws1Str := ws1.String()
	inv := &domain.AdminInvitation{
		WorkspaceID: &ws1Str, Email: "a@example.com", Role: domain.RoleCompanyAdmin,
		Name: "n", Status: domain.InvitationStatusPending, ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	require.NoError(t, repo.Create(ctx, inv))

	got, err := repo.FindByID(ctx, inv.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.WorkspaceID)
	require.Equal(t, ws1Str, *got.WorkspaceID)

	// 所属が入っていない行（運営が作った未割り当ての招待など）を直接挿入する。
	// domain 側は nil のまま返すこと。
	_, err = sqlDB.Exec(
		`INSERT INTO invitations (id, workspace_id, email, role, name, status, expires_at, created_at)
		 VALUES (999999, NULL, 'legacy@example.com', 'company_admin', 'n', 'pending', $1, now())`,
		time.Now().UTC().Add(time.Hour),
	)
	require.NoError(t, err)
	legacy, err := repo.FindByID(ctx, 999999)
	require.NoError(t, err)
	require.NotNil(t, legacy)
	require.Nil(t, legacy.WorkspaceID, "所属未設定の行は workspace_id が nil のまま")
}
