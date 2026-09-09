//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// insertTicketStatus / insertTicketType は raw SQL でスキーマ検証用の行を直接作る
// （repository を経由しない — CHECK / FK が効くかどうかそのものを見るテストのため）。
func insertTicketStatus(db *sql.DB, id, workspaceID, spaceID, name, category, color, position string, isInitial bool) error {
	_, err := db.Exec(
		`INSERT INTO ticket_statuses (id, workspace_id, space_id, name, category, color, "position", is_initial)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, workspaceID, spaceID, name, category, color, position, isInitial,
	)
	return err
}

func insertTicketType(db *sql.DB, id, workspaceID, spaceID, name string, hierarchyLevel int, color, position string, isDefault bool) error {
	_, err := db.Exec(
		`INSERT INTO ticket_types (id, workspace_id, space_id, name, hierarchy_level, color, "position", is_default)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, workspaceID, spaceID, name, hierarchyLevel, color, position, isDefault,
	)
	return err
}

// seedTicketMaster はチケット 1 件の FK が要求する状態・種別を最小構成で用意する。
func seedTicketMaster(t *testing.T, db *sql.DB, workspaceID, spaceID string) (statusID, typeID string) {
	t.Helper()
	statusID, typeID = newID(), newID()
	require.NoError(t, insertTicketStatus(db, statusID, workspaceID, spaceID, "To Do", "todo", "#5b6b7a", "a0", true))
	require.NoError(t, insertTicketType(db, typeID, workspaceID, spaceID, "タスク", 0, "#2f6b47", "a0", true))
	return statusID, typeID
}

// insertTicketRaw は 1 件を直接 INSERT する。number / position は呼び出し側が決める
// （同じスペースを共有する複数の subtest から呼んでも uq_tickets_space_number /
// uq_tickets_space_position とぶつからないように、固定値にしない）。
func insertTicketRaw(
	db *sql.DB, id, workspaceID, spaceID string, number int64, position, typeID, statusID string,
	parentID *string, priority int, startDate, dueDate *string,
) error {
	_, err := db.Exec(
		`INSERT INTO tickets (id, workspace_id, space_id, number, type_id, status_id, parent_id,
			title, doc, plain_text, priority, start_date, due_date, "position", created_by_user_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'テスト', '{"type":"doc","content":[]}'::jsonb, '', $8, $9, $10, $11, 1)`,
		id, workspaceID, spaceID, number, typeID, statusID, parentID, priority, startDate, dueDate, position,
	)
	return err
}

// TestTicketSchema_Integration は明示 DDL（schema.hcl のチケット表）が張る制約を実 Postgres で
// 固定する。usecase 層の検証（handler に届く前に断る）とは別に、DB そのものが最後の砦として
// 同じ規則を持っていることを確かめる。
func TestTicketSchema_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	testsupport.TruncateAll(t, sqlDB, kbTables...)
	ws := createWorkspace(t, sqlDB, "tk-schema")
	space := createSpace(t, sqlDB, ws, "eng")
	statusID, typeID := seedTicketMaster(t, sqlDB, ws, space)

	t.Run("優先度は1_2_3のみ", func(t *testing.T) {
		err := insertTicketRaw(sqlDB, newID(), ws, space, 101, "a1", typeID, statusID, nil, 9, nil, nil)
		requirePgError(t, err, sqlStateCheckViolation, "ck_tickets_priority")
	})

	t.Run("開始日は期限を超えられない", func(t *testing.T) {
		start, due := "2026-09-10", "2026-09-01"
		err := insertTicketRaw(sqlDB, newID(), ws, space, 102, "a2", typeID, statusID, nil, 2, &start, &due)
		requirePgError(t, err, sqlStateCheckViolation, "ck_tickets_dates_ordered")
	})

	t.Run("closed_atとresolutionは対で持つ", func(t *testing.T) {
		id := newID()
		require.NoError(t, insertTicketRaw(sqlDB, id, ws, space, 1, "a3", typeID, statusID, nil, 2, nil, nil))
		_, err := sqlDB.Exec(`UPDATE tickets SET closed_at = now() WHERE id = $1`, id)
		requirePgError(t, err, sqlStateCheckViolation, "ck_tickets_closed_pair")
	})

	t.Run("状態のcategoryは3枠のみ", func(t *testing.T) {
		err := insertTicketStatus(sqlDB, newID(), ws, space, "変な状態", "unknown", "#5b6b7a", "a5", false)
		requirePgError(t, err, sqlStateCheckViolation, "ck_ticket_statuses_category")
	})

	t.Run("種別の階層レベルは_1_0_1のみ", func(t *testing.T) {
		err := insertTicketType(sqlDB, newID(), ws, space, "変な種別", 5, "#2f6b47", "a5", false)
		requirePgError(t, err, sqlStateCheckViolation, "ck_ticket_types_hierarchy_level")
	})

	t.Run("担当は別ワークスペースのprincipalを指せない", func(t *testing.T) {
		otherWS := createWorkspace(t, sqlDB, "tk-schema-other")
		otherUser := createUser(t, sqlDB, "outsider")
		perm := persistence.NewKnowledgeBasePermissionRepository(sqlDB)
		outsider, err := perm.EnsureUserPrincipal(context.Background(), otherWS, otherUser)
		require.NoError(t, err)

		ticketID := newID()
		require.NoError(t, insertTicketRaw(sqlDB, ticketID, ws, space, 2, "a4", typeID, statusID, nil, 2, nil, nil))
		_, err = sqlDB.Exec(
			`INSERT INTO ticket_assignments (workspace_id, ticket_id, assignee_principal_id, assigned_by_user_id)
			 VALUES ($1, $2, $3, 1)`,
			ws, ticketID, outsider.ID,
		)
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_ticket_assignments_principal")
	})
}

// TestTicketRepository_Integration は persistence.ticketRepository を実 Postgres 相手に検証する。
// usecase 層の業務規則（親子の階層・深さ・周期など）はモックで既に固定済みなので、ここでは
// 「SQL がその規則の材料を正しく読み書きするか」（採番・簡易プロトコルの日付・
// テナント境界の WHERE ガード）に絞る。
func TestTicketRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewTicketRepository(sqlDB)
	ctx := context.Background()

	setup := func(t *testing.T) (ws, space string) {
		t.Helper()
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		ws = createWorkspace(t, sqlDB, "tk-repo")
		space = createSpace(t, sqlDB, ws, "eng")
		return ws, space
	}

	t.Run("状態種別の名前は現役の中で大文字小文字を無視して一意", func(t *testing.T) {
		ws, space := setup(t)
		require.NoError(t, repo.InsertTicketStatus(ctx, &domain.TicketStatus{
			WorkspaceID: ws, SpaceID: space, Name: "To Do", Category: domain.TicketStatusCategoryTodo, Color: "#5b6b7a", Position: "a0",
		}))
		err := repo.InsertTicketStatus(ctx, &domain.TicketStatus{
			WorkspaceID: ws, SpaceID: space, Name: "TO DO", Category: domain.TicketStatusCategoryTodo, Color: "#5b6b7a", Position: "a1",
		})
		require.ErrorIs(t, err, repository.ErrTicketStatusNameTaken)

		require.NoError(t, repo.InsertTicketType(ctx, &domain.TicketType{
			WorkspaceID: ws, SpaceID: space, Name: "タスク", Color: "#2f6b47", Position: "a0",
		}))
		err = repo.InsertTicketType(ctx, &domain.TicketType{
			WorkspaceID: ws, SpaceID: space, Name: "タスク", Color: "#2f6b47", Position: "a1",
		})
		require.ErrorIs(t, err, repository.ErrTicketTypeNameTaken)
	})

	t.Run("採番はスペースごとに1から連番", func(t *testing.T) {
		ws, space := setup(t)
		status := &domain.TicketStatus{WorkspaceID: ws, SpaceID: space, Name: "To Do", Category: domain.TicketStatusCategoryTodo, Color: "#5b6b7a", Position: "a0", IsInitial: true}
		require.NoError(t, repo.InsertTicketStatus(ctx, status))
		typ := &domain.TicketType{WorkspaceID: ws, SpaceID: space, Name: "タスク", Color: "#2f6b47", Position: "a0", IsDefault: true}
		require.NoError(t, repo.InsertTicketType(ctx, typ))

		other := createSpace(t, sqlDB, ws, "other")
		otherStatus := &domain.TicketStatus{WorkspaceID: ws, SpaceID: other, Name: "To Do", Category: domain.TicketStatusCategoryTodo, Color: "#5b6b7a", Position: "a0", IsInitial: true}
		require.NoError(t, repo.InsertTicketStatus(ctx, otherStatus))
		otherType := &domain.TicketType{WorkspaceID: ws, SpaceID: other, Name: "タスク", Color: "#2f6b47", Position: "a0", IsDefault: true}
		require.NoError(t, repo.InsertTicketType(ctx, otherType))

		mk := func(spaceID, statusID, typeID, position string) *domain.Ticket {
			created, err := repo.CreateTicket(ctx, repository.TicketCreateInput{
				WorkspaceID: ws, SpaceID: spaceID, TypeID: typeID, StatusID: statusID,
				Title: "x", Doc: []byte(`{"type":"doc","content":[]}`), Position: position, Priority: domain.TicketPriorityDefault, CreatedByUserID: 1,
			})
			require.NoError(t, err)
			return created
		}
		t1 := mk(space, status.ID, typ.ID, "a0")
		t2 := mk(space, status.ID, typ.ID, "a1")
		o1 := mk(other, otherStatus.ID, otherType.ID, "a0")
		assert.EqualValues(t, 1, t1.Number)
		assert.EqualValues(t, 2, t2.Number, "同じスペース内は連番")
		assert.EqualValues(t, 1, o1.Number, "スペースが違えば1から採番し直す")
	})

	t.Run("親チェーンは根に近い順_自分は含まない", func(t *testing.T) {
		ws, space := setup(t)
		statusID, typeID := seedTicketMasterViaRepo(ctx, t, repo, ws, space)
		mk := func(parentID *string, position string) *domain.Ticket {
			created, err := repo.CreateTicket(ctx, repository.TicketCreateInput{
				WorkspaceID: ws, SpaceID: space, TypeID: typeID, StatusID: statusID, ParentID: parentID,
				Title: "x", Doc: []byte(`{"type":"doc","content":[]}`), Position: position, Priority: domain.TicketPriorityDefault, CreatedByUserID: 1,
			})
			require.NoError(t, err)
			return created
		}
		root := mk(nil, "a0")
		child := mk(&root.ID, "a1")
		grand := mk(&child.ID, "a2")

		chain, err := repo.ListTicketParentChain(ctx, ws, grand.ID)
		require.NoError(t, err)
		require.Len(t, chain, 2)
		assert.Equal(t, root.ID, chain[0].ID)
		assert.Equal(t, child.ID, chain[1].ID)

		empty, err := repo.ListTicketParentChain(ctx, ws, root.ID)
		require.NoError(t, err)
		assert.Empty(t, empty)
	})

	t.Run("担当の設定はテナント境界のWHEREガードで守られる", func(t *testing.T) {
		ws, space := setup(t)
		statusID, typeID := seedTicketMasterViaRepo(ctx, t, repo, ws, space)
		created, err := repo.CreateTicket(ctx, repository.TicketCreateInput{
			WorkspaceID: ws, SpaceID: space, TypeID: typeID, StatusID: statusID,
			Title: "x", Doc: []byte(`{"type":"doc","content":[]}`), Position: "a0", Priority: domain.TicketPriorityDefault, CreatedByUserID: 1,
		})
		require.NoError(t, err)

		perm := persistence.NewKnowledgeBasePermissionRepository(sqlDB)
		alice := createUser(t, sqlDB, "alice")
		principal, err := perm.EnsureUserPrincipal(ctx, ws, alice)
		require.NoError(t, err)

		require.NoError(t, repo.UpsertTicketAssignment(ctx, &domain.TicketAssignment{
			WorkspaceID: ws, TicketID: created.ID, AssigneePrincipalID: principal.ID, AssignedByUserID: alice,
		}))
		got, err := repo.FindTicketAssignment(ctx, ws, created.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, principal.ID, got.AssigneePrincipalID)

		// 同じ担当を UpdateTicket 経由の ON CONFLICT で置き換えても、workspace_id は
		// 自分自身のままである（EXCLUDED.workspace_id と一致しているので通常どおり通る）。
		bob := createUser(t, sqlDB, "bob")
		bobPrincipal, err := perm.EnsureUserPrincipal(ctx, ws, bob)
		require.NoError(t, err)
		require.NoError(t, repo.UpsertTicketAssignment(ctx, &domain.TicketAssignment{
			WorkspaceID: ws, TicketID: created.ID, AssigneePrincipalID: bobPrincipal.ID, AssignedByUserID: bob,
		}))
		got, err = repo.FindTicketAssignment(ctx, ws, created.ID)
		require.NoError(t, err)
		assert.Equal(t, bobPrincipal.ID, got.AssigneePrincipalID, "上書きで置き換わる")

		require.NoError(t, repo.DeleteTicketAssignment(ctx, ws, created.ID))
		got, err = repo.FindTicketAssignment(ctx, ws, created.ID)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("状態変更はclosed_atとresolutionをまとめて書き換える", func(t *testing.T) {
		ws, space := setup(t)
		statusID, typeID := seedTicketMasterViaRepo(ctx, t, repo, ws, space)
		created, err := repo.CreateTicket(ctx, repository.TicketCreateInput{
			WorkspaceID: ws, SpaceID: space, TypeID: typeID, StatusID: statusID,
			Title: "x", Doc: []byte(`{"type":"doc","content":[]}`), Position: "a0", Priority: domain.TicketPriorityDefault, CreatedByUserID: 1,
		})
		require.NoError(t, err)
		doneStatus := &domain.TicketStatus{WorkspaceID: ws, SpaceID: space, Name: "完了", Category: domain.TicketStatusCategoryDone, Color: "#2f6b47", Position: "a1"}
		require.NoError(t, repo.InsertTicketStatus(ctx, doneStatus))

		now := time.Now().UTC().Truncate(time.Second)
		resolution := domain.TicketResolutionDone
		updated, err := repo.ChangeTicketStatus(ctx, ws, created.ID, doneStatus.ID, &now, &resolution)
		require.NoError(t, err)
		assert.Equal(t, doneStatus.ID, updated.StatusID)
		require.NotNil(t, updated.ClosedAt)
		require.NotNil(t, updated.Resolution)
		assert.Equal(t, domain.TicketResolutionDone, *updated.Resolution)

		// 未完了へ戻すと usecase 側が (nil, nil) を渡す想定 — repository はそのまま書く。
		todoStatus := &domain.TicketStatus{WorkspaceID: ws, SpaceID: space, Name: "差し戻し", Category: domain.TicketStatusCategoryTodo, Color: "#5b6b7a", Position: "a2"}
		require.NoError(t, repo.InsertTicketStatus(ctx, todoStatus))
		reverted, err := repo.ChangeTicketStatus(ctx, ws, created.ID, todoStatus.ID, nil, nil)
		require.NoError(t, err)
		assert.Nil(t, reverted.ClosedAt)
		assert.Nil(t, reverted.Resolution)
	})

	t.Run("アーカイブと復元", func(t *testing.T) {
		ws, space := setup(t)
		statusID, typeID := seedTicketMasterViaRepo(ctx, t, repo, ws, space)
		created, err := repo.CreateTicket(ctx, repository.TicketCreateInput{
			WorkspaceID: ws, SpaceID: space, TypeID: typeID, StatusID: statusID,
			Title: "x", Doc: []byte(`{"type":"doc","content":[]}`), Position: "a0", Priority: domain.TicketPriorityDefault, CreatedByUserID: 1,
		})
		require.NoError(t, err)

		require.NoError(t, repo.ArchiveTicket(ctx, ws, created.ID))
		got, err := repo.FindTicket(ctx, ws, created.ID)
		require.NoError(t, err)
		require.NotNil(t, got.ArchivedAt)

		require.NoError(t, repo.RestoreTicket(ctx, ws, created.ID, "b0"))
		got, err = repo.FindTicket(ctx, ws, created.ID)
		require.NoError(t, err)
		assert.Nil(t, got.ArchivedAt)
		assert.Equal(t, "b0", got.Position)
	})

	t.Run("変更履歴はグループと項目をまとめて書き新しい順で返す", func(t *testing.T) {
		ws, space := setup(t)
		statusID, typeID := seedTicketMasterViaRepo(ctx, t, repo, ws, space)
		created, err := repo.CreateTicket(ctx, repository.TicketCreateInput{
			WorkspaceID: ws, SpaceID: space, TypeID: typeID, StatusID: statusID,
			Title: "x", Doc: []byte(`{"type":"doc","content":[]}`), Position: "a0", Priority: domain.TicketPriorityDefault, CreatedByUserID: 1,
		})
		require.NoError(t, err)

		old, new := "旧", "新"
		require.NoError(t, repo.InsertTicketChangeGroup(ctx, &domain.TicketChangeGroup{
			WorkspaceID: ws, TicketID: created.ID, ActorUserID: 1,
			Items: []domain.TicketChangeItem{{Field: domain.TicketChangeFieldTitle, OldValue: &old, NewValue: &new}},
		}))
		old2, new2 := "新", "新2"
		require.NoError(t, repo.InsertTicketChangeGroup(ctx, &domain.TicketChangeGroup{
			WorkspaceID: ws, TicketID: created.ID, ActorUserID: 1,
			Items: []domain.TicketChangeItem{{Field: domain.TicketChangeFieldTitle, OldValue: &old2, NewValue: &new2}},
		}))

		groups, err := repo.ListTicketChangeGroups(ctx, ws, created.ID)
		require.NoError(t, err)
		require.Len(t, groups, 2)
		require.Len(t, groups[0].Items, 1)
		assert.Equal(t, "新2", *groups[0].Items[0].NewValue, "新しい順")
		assert.Equal(t, "新", *groups[1].Items[0].NewValue)
	})

	t.Run("派生表は本文保存のたびに張り替わる", func(t *testing.T) {
		ws, space := setup(t)
		statusID, typeID := seedTicketMasterViaRepo(ctx, t, repo, ws, space)
		src, err := repo.CreateTicket(ctx, repository.TicketCreateInput{
			WorkspaceID: ws, SpaceID: space, TypeID: typeID, StatusID: statusID,
			Title: "x", Doc: []byte(`{"type":"doc","content":[]}`), Position: "a0", Priority: domain.TicketPriorityDefault, CreatedByUserID: 1,
		})
		require.NoError(t, err)
		target, err := repo.CreateTicket(ctx, repository.TicketCreateInput{
			WorkspaceID: ws, SpaceID: space, TypeID: typeID, StatusID: statusID,
			Title: "参照先", Doc: []byte(`{"type":"doc","content":[]}`), Position: "a1", Priority: domain.TicketPriorityDefault, CreatedByUserID: 1,
		})
		require.NoError(t, err)

		require.NoError(t, repo.ReplaceTicketTicketLinks(ctx, ws, src.ID, []string{target.ID}))
		links, err := repo.ListTicketTicketLinks(ctx, ws, src.ID)
		require.NoError(t, err)
		require.Len(t, links, 1)
		assert.Equal(t, target.ID, links[0].TargetTicketID)

		back, err := repo.ListTicketsReferencingTicket(ctx, ws, target.ID)
		require.NoError(t, err)
		require.Len(t, back, 1)
		assert.Equal(t, src.ID, back[0].SourceTicketID)

		// 張り替え（空へ）で消える。実在しない ID は黙って除外される
		// （リンク切れ 1 本のために保存全体を失敗させない方針。doc.go の doc 参照）。
		require.NoError(t, repo.ReplaceTicketTicketLinks(ctx, ws, src.ID, []string{newID()}))
		links, err = repo.ListTicketTicketLinks(ctx, ws, src.ID)
		require.NoError(t, err)
		assert.Empty(t, links)
	})
}

// seedTicketMasterViaRepo は repository 経由で状態・種別を 1 つずつ用意する
// （seedTicketMaster の raw SQL 版と違い、repository の採番・検証を経由する）。
func seedTicketMasterViaRepo(ctx context.Context, t *testing.T, repo repository.TicketRepository, ws, space string) (statusID, typeID string) {
	t.Helper()
	// Position は repository が採番しない（usecase が fracindex で決めて渡す設計 — 段 1 の
	// CreateTicketStatusUseCase / CreateTicketTypeUseCase 参照）。ここでは 1 件ずつしか
	// 作らないので固定値で足りる。
	status := &domain.TicketStatus{WorkspaceID: ws, SpaceID: space, Name: "To Do", Category: domain.TicketStatusCategoryTodo, Color: "#5b6b7a", Position: "a0", IsInitial: true}
	require.NoError(t, repo.InsertTicketStatus(ctx, status))
	typ := &domain.TicketType{WorkspaceID: ws, SpaceID: space, Name: "タスク", Color: "#2f6b47", Position: "a0", IsDefault: true}
	require.NoError(t, repo.InsertTicketType(ctx, typ))
	return status.ID, typ.ID
}

// TestTicketSimpleProtocol_Integration は simple query protocol（本番の transaction pooler と
// 同じ経路）でチケットの日付（date → string override）と doc（jsonb）が正しく往復することを
// 固定する。段 0 で sqlc.yaml に date → string の override を足した理由そのものの回帰テスト
// （extended protocol では検出できない — pgx が time.Time を渡すと 1 日ずれる本番限定の不具合
// だったため。設計 Ⅳ-K）。
func TestTicketSimpleProtocol_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDBSimpleProtocol(t)
	repo := persistence.NewTicketRepository(sqlDB)
	ctx := context.Background()

	testsupport.TruncateAll(t, sqlDB, kbTables...)
	ws := createWorkspace(t, sqlDB, "tk-simple")
	space := createSpace(t, sqlDB, ws, "eng")
	statusID, typeID := seedTicketMasterViaRepo(ctx, t, repo, ws, space)

	start, due := "2026-09-01", "2026-09-30"
	doc := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"simple protocol 経由"}]}]}`
	created, err := repo.CreateTicket(ctx, repository.TicketCreateInput{
		WorkspaceID: ws, SpaceID: space, TypeID: typeID, StatusID: statusID,
		Title: "x", Doc: []byte(doc), Position: "a0", Priority: domain.TicketPriorityDefault, CreatedByUserID: 1,
		StartDate: &start, DueDate: &due,
	})
	require.NoError(t, err)
	require.NotNil(t, created.StartDate)
	require.NotNil(t, created.DueDate)
	assert.Equal(t, start, *created.StartDate, "simple protocol でも日付がずれない")
	assert.Equal(t, due, *created.DueDate)

	got, err := repo.FindTicket(ctx, ws, created.ID)
	require.NoError(t, err)
	require.NotNil(t, got.StartDate)
	assert.Equal(t, start, *got.StartDate)
	assert.JSONEq(t, doc, string(got.Doc))
}
