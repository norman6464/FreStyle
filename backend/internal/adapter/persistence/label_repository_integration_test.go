//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLabelRepository_Integration は labelRepository（段 4: ラベル・チケットへの付け外し）を
// 実 Postgres で検証する。
func TestLabelRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewLabelRepository(sqlDB)
	tickets := persistence.NewTicketRepository(sqlDB)
	ctx := context.Background()

	setup := func(t *testing.T) (ws, space, ticketID string) {
		t.Helper()
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		ws = createWorkspace(t, sqlDB, "tk-labels")
		space = createSpace(t, sqlDB, ws, "eng")
		statusID, typeID := seedTicketMasterViaRepo(ctx, t, tickets, ws, space)
		created, err := tickets.CreateTicket(ctx, repository.TicketCreateInput{
			WorkspaceID: ws, SpaceID: space, TypeID: typeID, StatusID: statusID,
			Title: "x", Doc: []byte(`{"type":"doc","content":[]}`), Position: "a0", Priority: domain.TicketPriorityDefault, CreatedByUserID: 1,
		})
		require.NoError(t, err)
		return ws, space, created.ID
	}

	t.Run("作成_取得_一覧", func(t *testing.T) {
		ws, space, _ := setup(t)
		l := &domain.Label{WorkspaceID: ws, SpaceID: space, Name: "緊急", Color: "#ff0000"}
		require.NoError(t, repo.CreateLabel(ctx, l))
		require.NotEmpty(t, l.ID)

		got, err := repo.FindLabel(ctx, ws, l.ID)
		require.NoError(t, err)
		assert.Equal(t, "緊急", got.Name)
		assert.Equal(t, "#ff0000", got.Color)

		list, err := repo.ListLabels(ctx, ws, space)
		require.NoError(t, err)
		require.Len(t, list, 1)
	})

	t.Run("同名はスペース内で作れない_大文字小文字違いも含む", func(t *testing.T) {
		ws, space, _ := setup(t)
		require.NoError(t, repo.CreateLabel(ctx, &domain.Label{WorkspaceID: ws, SpaceID: space, Name: "Urgent", Color: "#ff0000"}))
		err := repo.CreateLabel(ctx, &domain.Label{WorkspaceID: ws, SpaceID: space, Name: "Urgent", Color: "#00ff00"})
		require.ErrorIs(t, err, repository.ErrLabelNameTaken, "完全一致")

		// name_key（lower(btrim(name))）が大文字小文字違いを畳む。name 自体の前後空白は
		// ck_labels_name_trimmed が挿入時点で既に禁じている（トリムは usecase の責務 —
		// label_usecase_test.go の Test_ラベル作成_名前をトリムし色を正規化してから保存する 参照）。
		err = repo.CreateLabel(ctx, &domain.Label{WorkspaceID: ws, SpaceID: space, Name: "URGENT", Color: "#0000ff"})
		require.ErrorIs(t, err, repository.ErrLabelNameTaken, "大文字小文字違い")

		// 別スペースなら同名でも作れる（一意制約は space_id 込み）。
		other := createSpace(t, sqlDB, ws, "other")
		err = repo.CreateLabel(ctx, &domain.Label{WorkspaceID: ws, SpaceID: other, Name: "Urgent", Color: "#0000ff"})
		require.NoError(t, err, "別スペースなら同名でも作れる")
	})

	t.Run("更新", func(t *testing.T) {
		ws, space, _ := setup(t)
		l := &domain.Label{WorkspaceID: ws, SpaceID: space, Name: "旧", Color: "#ff0000"}
		require.NoError(t, repo.CreateLabel(ctx, l))

		update := &domain.Label{ID: l.ID, WorkspaceID: ws, SpaceID: space, Name: "新", Color: "#00ff00"}
		require.NoError(t, repo.UpdateLabel(ctx, update))
		got, err := repo.FindLabel(ctx, ws, l.ID)
		require.NoError(t, err)
		assert.Equal(t, "新", got.Name)
		assert.Equal(t, "#00ff00", got.Color)
	})

	// UPDATE / DELETE は space_id でも絞る。呼び出し側が権限を確かめる相手は URL のスペースで、
	// そこから外れた行に届いてしまうと「確かめた相手」と「触った相手」が別物になる。
	t.Run("別スペースのラベルは更新も削除もできない", func(t *testing.T) {
		ws, space, _ := setup(t)
		other := createSpace(t, sqlDB, ws, "other")
		l := &domain.Label{WorkspaceID: ws, SpaceID: other, Name: "隣の", Color: "#ff0000"}
		require.NoError(t, repo.CreateLabel(ctx, l))

		update := &domain.Label{ID: l.ID, WorkspaceID: ws, SpaceID: space, Name: "改名", Color: "#00ff00"}
		require.ErrorIs(t, repo.UpdateLabel(ctx, update), repository.ErrLabelNotFound)
		require.ErrorIs(t, repo.DeleteLabel(ctx, ws, space, l.ID), repository.ErrLabelNotFound)

		got, err := repo.FindLabel(ctx, ws, l.ID)
		require.NoError(t, err, "行はそのまま残っている")
		assert.Equal(t, "隣の", got.Name)
	})

	t.Run("削除でticket_labelsも一緒に消える", func(t *testing.T) {
		ws, space, ticketID := setup(t)
		l := &domain.Label{WorkspaceID: ws, SpaceID: space, Name: "緊急", Color: "#ff0000"}
		require.NoError(t, repo.CreateLabel(ctx, l))
		require.NoError(t, repo.AddTicketLabel(ctx, ws, ticketID, l.ID))

		require.NoError(t, repo.DeleteLabel(ctx, ws, space, l.ID))
		labels, err := repo.ListLabelsByTicket(ctx, ws, ticketID)
		require.NoError(t, err)
		assert.Empty(t, labels, "ON DELETE CASCADE でticket_labelsの行も消える")

		require.ErrorIs(t, repo.DeleteLabel(ctx, ws, space, l.ID), repository.ErrLabelNotFound, "二重削除は404相当")
	})

	t.Run("付け外しは冪等", func(t *testing.T) {
		ws, space, ticketID := setup(t)
		l := &domain.Label{WorkspaceID: ws, SpaceID: space, Name: "緊急", Color: "#ff0000"}
		require.NoError(t, repo.CreateLabel(ctx, l))

		require.NoError(t, repo.AddTicketLabel(ctx, ws, ticketID, l.ID))
		require.NoError(t, repo.AddTicketLabel(ctx, ws, ticketID, l.ID), "同じラベルの二重追加はエラーにならない")
		labels, err := repo.ListLabelsByTicket(ctx, ws, ticketID)
		require.NoError(t, err)
		require.Len(t, labels, 1, "複合主キーが重複を1件に吸収する")

		require.NoError(t, repo.RemoveTicketLabel(ctx, ws, ticketID, l.ID))
		require.NoError(t, repo.RemoveTicketLabel(ctx, ws, ticketID, l.ID), "付いていないラベルを外そうとしてもエラーにならない")
		labels, err = repo.ListLabelsByTicket(ctx, ws, ticketID)
		require.NoError(t, err)
		assert.Empty(t, labels)
	})

	t.Run("ListLabelsByTicketIDsはチケットごとにまとめて返す", func(t *testing.T) {
		ws, space, ticketA := setup(t)
		// setup が既に作った状態・種別を使い回す（同じスペースへ二重に seed すると
		// uq_ticket_statuses_space_name / uq_ticket_types_space_name に引っかかる）。
		existing, err := tickets.FindTicket(ctx, ws, ticketA)
		require.NoError(t, err)
		ticketBTicket, err := tickets.CreateTicket(ctx, repository.TicketCreateInput{
			WorkspaceID: ws, SpaceID: space, TypeID: existing.TypeID, StatusID: existing.StatusID,
			Title: "y", Doc: []byte(`{"type":"doc","content":[]}`), Position: "a1", Priority: domain.TicketPriorityDefault, CreatedByUserID: 1,
		})
		require.NoError(t, err)
		ticketB := ticketBTicket.ID

		l1 := &domain.Label{WorkspaceID: ws, SpaceID: space, Name: "緊急", Color: "#ff0000"}
		require.NoError(t, repo.CreateLabel(ctx, l1))
		l2 := &domain.Label{WorkspaceID: ws, SpaceID: space, Name: "バグ", Color: "#00ff00"}
		require.NoError(t, repo.CreateLabel(ctx, l2))
		require.NoError(t, repo.AddTicketLabel(ctx, ws, ticketA, l1.ID))
		require.NoError(t, repo.AddTicketLabel(ctx, ws, ticketB, l2.ID))

		byTicket, err := repo.ListLabelsByTicketIDs(ctx, ws, []string{ticketA, ticketB})
		require.NoError(t, err)
		require.Len(t, byTicket[ticketA], 1)
		assert.Equal(t, "緊急", byTicket[ticketA][0].Name)
		require.Len(t, byTicket[ticketB], 1)
		assert.Equal(t, "バグ", byTicket[ticketB][0].Name)
	})

	t.Run("色はhex形式でなければならない", func(t *testing.T) {
		ws, space, _ := setup(t)
		// 列幅（varchar(7)）以内でなければ CHECK より先に 22001（値が長すぎる）になるため、
		// 7 文字ちょうどで hex ではない値を使う。
		_, err := sqlDB.Exec(
			`INSERT INTO labels (id, workspace_id, space_id, name, color) VALUES ($1, $2, $3, 'x', 'zzzzzzz')`,
			newID(), ws, space,
		)
		requirePgError(t, err, sqlStateCheckViolation, "ck_labels_color_hex")
	})

	t.Run("名前は空文字にできない", func(t *testing.T) {
		ws, space, _ := setup(t)
		_, err := sqlDB.Exec(
			`INSERT INTO labels (id, workspace_id, space_id, name, color) VALUES ($1, $2, $3, '', '#ff0000')`,
			newID(), ws, space,
		)
		requirePgError(t, err, sqlStateCheckViolation, "ck_labels_name_trimmed")
	})
}
