//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/adapter/persistence"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/testsupport"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
	"github.com/norman6464/frestyle/backend/internal/usecase/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserActiveAndRetire_Integration は段 7（停止・退会）の usecase を実 PostgreSQL に対して
// 検証する。SetUserActiveUseCase / RetireSelfUseCase はいずれも UserRepository と
// KnowledgeBasePermissionRepository という別々の repository の書き込みを
// TxManager.DoInTx で 1 つのトランザクションに束ねる（package persistence 内の
// txKey/getTx/withTx が両 repository で共有されているため成立する）。
// この「別 repository をまたいだ合成が本当に 1 トランザクションになるか」は
// fake/mock では確かめようがなく、実 PostgreSQL でしか検証できない
// （usecase/user の単体テストは spy で経路だけを固定している）。
func TestUserActiveAndRetire_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	users := persistence.NewUserRepository(sqlDB)
	txManager := persistence.NewTxManager(sqlDB)
	setActive := user.NewSetUserActiveUseCase(users, persistence.NewKnowledgeBasePermissionRepository(sqlDB), txManager)
	retireSelf := user.NewRetireSelfUseCase(users, persistence.NewKnowledgeBasePermissionRepository(sqlDB), txManager)

	t.Run("停止すると status が変わり membership_events にも記録される", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		f.principalFor(ctx, t, f.alice) // actor 側もワークスペースに実在させる
		f.principalFor(ctx, t, f.bob)

		require.NoError(t, setActive.Execute(ctx, user.SetUserActiveInput{
			WorkspaceID: f.ws, TargetUserID: f.bob, ActorUserID: f.alice, Active: false,
		}))

		got, err := users.FindByID(ctx, f.bob)
		require.NoError(t, err)
		assert.Equal(t, domain.UserStatusSuspended, got.Status)

		events, err := f.perm.ListMembershipEvents(ctx, f.ws)
		require.NoError(t, err)
		require.NotEmpty(t, events)
		e := events[0]
		assert.Equal(t, domain.MembershipEventSuspended, e.Action)
		assert.Equal(t, f.bob, e.TargetUserID)
		assert.Equal(t, f.alice, e.ActorUserID)
		require.NotNil(t, e.OldLabel)
		require.NotNil(t, e.NewLabel)
		assert.Equal(t, "active", *e.OldLabel)
		assert.Equal(t, "suspended", *e.NewLabel)
	})

	t.Run("復帰すると active に戻り新ラベルもactiveで記録される", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		f.principalFor(ctx, t, f.alice)
		f.principalFor(ctx, t, f.bob)
		require.NoError(t, setActive.Execute(ctx, user.SetUserActiveInput{
			WorkspaceID: f.ws, TargetUserID: f.bob, ActorUserID: f.alice, Active: false,
		}))

		require.NoError(t, setActive.Execute(ctx, user.SetUserActiveInput{
			WorkspaceID: f.ws, TargetUserID: f.bob, ActorUserID: f.alice, Active: true,
		}))

		got, err := users.FindByID(ctx, f.bob)
		require.NoError(t, err)
		assert.Equal(t, domain.UserStatusActive, got.Status)

		events, err := f.perm.ListMembershipEvents(ctx, f.ws)
		require.NoError(t, err)
		require.NotEmpty(t, events)
		e := events[0] // 新しい順の先頭 = 直近の復帰
		assert.Equal(t, domain.MembershipEventSuspended, e.Action)
		require.NotNil(t, e.OldLabel)
		require.NotNil(t, e.NewLabel)
		assert.Equal(t, "suspended", *e.OldLabel)
		assert.Equal(t, "active", *e.NewLabel)
	})

	t.Run("対象がワークスペースのメンバーでなければ拒否しstatusは変えない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		f.principalFor(ctx, t, f.alice)
		// bob はどこにも所属させない。

		err := setActive.Execute(ctx, user.SetUserActiveInput{
			WorkspaceID: f.ws, TargetUserID: f.bob, ActorUserID: f.alice, Active: false,
		})
		assert.ErrorIs(t, err, user.ErrTargetNotWorkspaceMember)

		got, err := users.FindByID(ctx, f.bob)
		require.NoError(t, err)
		assert.Equal(t, domain.UserStatusActive, got.Status, "拒否されたので変わらない")
	})

	t.Run("退会は所属する全ワークスペースを退出してから匿名化する", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		// bob をどちらのワークスペースでも admin ではない側にしておく
		// （alice が両方で admin なので、bob が抜けても最後の admin にはならない）。
		alicePrincipalWS, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		f.makeActiveMember(t, f.ws, f.alice)
		_, err = f.perm.UpsertWorkspaceGrant(ctx, f.ws, alicePrincipalWS.ID, domain.GrantRoleAdmin, f.alice)
		require.NoError(t, err)

		aliceOther, err := f.perm.EnsureUserPrincipal(ctx, f.otherWS, f.alice)
		require.NoError(t, err)
		f.makeActiveMember(t, f.otherWS, f.alice)
		_, err = f.perm.UpsertWorkspaceGrant(ctx, f.otherWS, aliceOther.ID, domain.GrantRoleAdmin, f.alice)
		require.NoError(t, err)

		f.principalFor(ctx, t, f.bob)
		f.makeActiveMember(t, f.otherWS, f.bob)

		require.NoError(t, retireSelf.Execute(ctx, f.bob))

		member, err := f.perm.IsWorkspaceMember(ctx, f.ws, f.bob)
		require.NoError(t, err)
		assert.False(t, member, "元のワークスペースから退出している")
		member, err = f.perm.IsWorkspaceMember(ctx, f.otherWS, f.bob)
		require.NoError(t, err)
		assert.False(t, member, "もう一方のワークスペースからも退出している")

		// SoftDelete 後は FindByID からは引けなくなる（user_repository_tx_integration_test.go の
		// 「論理削除後は引けない」と同じ契約）。ここでは status そのものを直接見る。
		got, err := users.FindByID(ctx, f.bob)
		require.NoError(t, err)
		assert.Nil(t, got, "退会後は通常の検索からは引けない")

		var status string
		require.NoError(t, sqlDB.QueryRow(`SELECT status FROM users WHERE id = $1`, f.bob).Scan(&status))
		assert.Equal(t, string(domain.UserStatusDeactivated), status)
	})

	t.Run("最後のadminのワークスペースがあれば退会全体を断り何も変えない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		// bob は f.ws の唯一の admin。f.otherWS には（admin ではない形で）別に所属。
		bobPrincipalWS, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)
		f.makeActiveMember(t, f.ws, f.bob)
		_, err = f.perm.UpsertWorkspaceGrant(ctx, f.ws, bobPrincipalWS.ID, domain.GrantRoleAdmin, f.bob)
		require.NoError(t, err)

		// principalFor は常に f.ws を対象にする小道具なので、f.otherWS 側は
		// EnsureUserPrincipal + makeActiveMember を直接組み立てる
		// （principal 行が無いまま workspace_members だけ active にすると、
		// LeaveWorkspaceMembership が「principal 無し = 招待中や非メンバー相当」の
		// 経路を通ってしまい、実際のメンバーの形を再現できない）。
		_, err = f.perm.EnsureUserPrincipal(ctx, f.otherWS, f.bob)
		require.NoError(t, err)
		f.makeActiveMember(t, f.otherWS, f.bob)

		err = retireSelf.Execute(ctx, f.bob)
		assert.ErrorIs(t, err, repository.ErrLastWorkspaceAdmin)

		// 1 トランザクションに束ねているので、先に処理された側の退出も丸ごと戻っているはず。
		member, err := f.perm.IsWorkspaceMember(ctx, f.ws, f.bob)
		require.NoError(t, err)
		assert.True(t, member, "最後の admin なので居残っている")
		member, err = f.perm.IsWorkspaceMember(ctx, f.otherWS, f.bob)
		require.NoError(t, err)
		assert.True(t, member, "全体がロールバックされ、もう一方の退出も戻っている")

		got, err := users.FindByID(ctx, f.bob)
		require.NoError(t, err)
		assert.Equal(t, domain.UserStatusActive, got.Status, "断られたので status も変わらない")
	})
}
