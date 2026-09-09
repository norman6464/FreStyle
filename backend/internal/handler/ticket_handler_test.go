package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ticketFixture は fake repository と、本番と同じ wiring で組んだルータの組。
// ワークスペース・スペースは kb_page_handler_test.go の定数（kbWorkspaceID 等）を
// そのまま使う（同じ package handler のテストなので再宣言しない）。
type ticketFixture struct {
	tickets *ticketFakeRepo
	pages   *kbFakePages
	perms   *kbFakePerms
	router  *gin.Engine
}

// newTicketFixture はワークスペース 2 つ・スペース 1 つの下ごしらえをして、
// registerTicketRoutesWith で本番と同じルートを張る。uid が 0 なら current user を
// 注入せず未認証を再現する。role が空なら kbUserID にはどの役割も届かない
// （CanView すら false — kbFakePerms.rolesAt は明示的な setScopeRole が無ければ空集合を返す）。
func newTicketFixture(uid uint64, role domain.GrantRole) ticketFixture {
	gin.SetMode(gin.TestMode)
	pages := newKbFakePages()
	pages.addWorkspace(kbWorkspaceID, kbWorkspaceSlug)
	pages.addWorkspace(kbOtherWorkspaceID, kbOtherWorkspaceSlug)
	pages.addSpace(kbWorkspaceID, kbSpaceID)

	perms := newKbFakePerms(pages, domain.PagePermission{})
	perms.addMember(kbWorkspaceID, kbUserID)
	if role != "" {
		perms.setScopeRole(kbSpaceID, kbUserID, role)
	}

	tickets := newTicketFakeRepo()
	users := newKbFakeUsers()

	r := gin.New()
	g := r.Group("/api/v2")
	if uid != 0 {
		g.Use(func(c *gin.Context) {
			c.Set(middleware.ContextKeyCurrentUserID, uid)
			c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: uid})
			c.Next()
		})
	}
	registerTicketRoutesWith(g, tickets, perms, pages, users, fakeTxManager{})
	return ticketFixture{tickets: tickets, pages: pages, perms: perms, router: r}
}

func (f ticketFixture) do(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

const ticketAPIBase = "/api/v2/kb/workspaces/" + kbWorkspaceSlug

func decodeJSON[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &v))
	return v
}

// --- 権限の撃ち分け（Get と Create の 2 入口で確かめれば、requireTicketPermission /
// requireTicketSpacePermission という共有ヘルパーの正しさとしては十分。同じヘルパーを
// 他の全エンドポイントも通る） ---

func Test_チケット取得_役割が無いメンバーは404(t *testing.T) {
	// メンバーではあるが、このスペースにどの役割も届いていない（setScopeRole 未設定 —
	// newKbFakePerms の既定は「役割 0 件」で、addMember だけでは CanView にならない）。
	f := newTicketFixture(kbUserID, "")
	ticket := f.tickets.addTicket(domain.Ticket{
		ID: "ticket-1", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, Title: "x",
	})

	w := f.do(t, http.MethodGet, ticketAPIBase+"/tickets/"+ticket.ID, "")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_チケット取得_閲覧のみで200(t *testing.T) {
	f := newTicketFixture(kbUserID, domain.GrantRoleViewer)
	ticketID := f.tickets.addTicket(domain.Ticket{
		ID: "ticket-1", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, Title: "本文", Number: 1,
	}).ID

	w := f.do(t, http.MethodGet, ticketAPIBase+"/tickets/"+ticketID, "")
	require.Equal(t, http.StatusOK, w.Code)
	got := decodeJSON[domain.Ticket](t, w)
	assert.Equal(t, "本文", got.Title)
}

func Test_チケット取得_他ワークスペースのチケットは404(t *testing.T) {
	f := newTicketFixture(kbUserID, domain.GrantRoleEditor)
	other := f.tickets.addTicket(domain.Ticket{ID: "ticket-x", WorkspaceID: kbOtherWorkspaceID, SpaceID: "other-space", Title: "x"})

	w := f.do(t, http.MethodGet, ticketAPIBase+"/tickets/"+other.ID, "")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_チケット作成_閲覧だけでは403(t *testing.T) {
	f := newTicketFixture(kbUserID, domain.GrantRoleViewer)
	f.tickets.addType(domain.TicketType{ID: "type-1", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, Name: "タスク", IsDefault: true})
	f.tickets.addStatus(domain.TicketStatus{ID: "status-1", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, Name: "To Do", Category: domain.TicketStatusCategoryTodo, IsInitial: true})

	w := f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/tickets", `{"title":"新規"}`)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- チケットのライフサイクル一式 ---

func Test_チケット一式_有効化から作成取得一覧更新状態変更移動担当履歴まで(t *testing.T) {
	f := newTicketFixture(kbUserID, domain.GrantRoleEditor)

	// 1) 有効化（最小構成）。
	w := f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/tickets/enable", "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	statuses, err := f.tickets.ListTicketStatuses(context.Background(), kbWorkspaceID, kbSpaceID, false)
	require.NoError(t, err)
	require.Len(t, statuses, 3)
	types, err := f.tickets.ListTicketTypes(context.Background(), kbWorkspaceID, kbSpaceID, false)
	require.NoError(t, err)
	require.Len(t, types, 1)

	// 2 度目の有効化は 409。
	w = f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/tickets/enable", "")
	assert.Equal(t, http.StatusConflict, w.Code)

	// 2) 作成（既定の種別・状態を解決）。
	w = f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/tickets", `{"title":"最初のチケット"}`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	created := decodeJSON[domain.Ticket](t, w)
	assert.Equal(t, "最初のチケット", created.Title)
	assert.EqualValues(t, 1, created.Number)
	assert.EqualValues(t, domain.TicketPriorityDefault, created.Priority)

	// 3) 取得。
	w = f.do(t, http.MethodGet, ticketAPIBase+"/tickets/"+created.ID, "")
	require.Equal(t, http.StatusOK, w.Code)

	// キーからの解決（FRESTYLE-1 相当。spaceKey はこの fake では spaceID と同一視する）。
	w = f.do(t, http.MethodGet, ticketAPIBase+"/spaces/"+kbSpaceID+"/tickets/key/"+strings.ToUpper(kbSpaceID)+"-1", "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	// 4) 一覧。
	w = f.do(t, http.MethodGet, ticketAPIBase+"/spaces/"+kbSpaceID+"/tickets", "")
	require.Equal(t, http.StatusOK, w.Code)
	list := decodeJSON[ticketListResponse](t, w)
	require.Len(t, list.Tickets, 1)

	// 5) 更新（PUT 相当。現在値を全部送る）。
	w = f.do(t, http.MethodPut, ticketAPIBase+"/tickets/"+created.ID,
		`{"title":"更新後","doc":{"type":"doc","content":[]},"typeId":"`+created.TypeID+`","priority":1}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	updated := decodeJSON[domain.Ticket](t, w)
	assert.Equal(t, "更新後", updated.Title)
	assert.EqualValues(t, domain.TicketPriorityHigh, updated.Priority)

	// 履歴に 2 項目（title・priority）が積まれている。
	w = f.do(t, http.MethodGet, ticketAPIBase+"/tickets/"+created.ID+"/history", "")
	require.Equal(t, http.StatusOK, w.Code)
	hist := decodeJSON[ticketHistoryResponse](t, w)
	require.Len(t, hist.Groups, 1)
	assert.Len(t, hist.Groups[0].Items, 2)

	// 6) 状態変更。
	doneStatus, err := f.tickets.FindTicketStatus(context.Background(), kbWorkspaceID, kbSpaceID, statusIDByCategory(statuses, domain.TicketStatusCategoryDone))
	require.NoError(t, err)
	w = f.do(t, http.MethodPost, ticketAPIBase+"/tickets/"+created.ID+"/status", `{"statusId":"`+doneStatus.ID+`"}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	closed := decodeJSON[domain.Ticket](t, w)
	require.NotNil(t, closed.ClosedAt)
	require.NotNil(t, closed.Resolution)
	assert.Equal(t, domain.TicketResolutionDone, *closed.Resolution)

	// 7) 2 件目を作って並び替え（1 件目の直後へ）。
	w = f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/tickets", `{"title":"2件目"}`)
	require.Equal(t, http.StatusCreated, w.Code)
	second := decodeJSON[domain.Ticket](t, w)
	w = f.do(t, http.MethodPost, ticketAPIBase+"/tickets/"+second.ID+"/move", `{"anchorTicketId":"`+created.ID+`","anchorAfter":false}`)
	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())

	// 8) 担当の設定・解除。
	w = f.do(t, http.MethodPut, ticketAPIBase+"/tickets/"+created.ID+"/assignee", `{"assigneePrincipalId":"principal-1"}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assignment := decodeJSON[domain.TicketAssignment](t, w)
	assert.Equal(t, "principal-1", assignment.AssigneePrincipalID)
	w = f.do(t, http.MethodDelete, ticketAPIBase+"/tickets/"+created.ID+"/assignee", "")
	require.Equal(t, http.StatusNoContent, w.Code)

	// 9) アーカイブ・復元。
	w = f.do(t, http.MethodPost, ticketAPIBase+"/tickets/"+created.ID+"/archive", "")
	require.Equal(t, http.StatusOK, w.Code)
	archived := decodeJSON[domain.Ticket](t, w)
	require.NotNil(t, archived.ArchivedAt)
	w = f.do(t, http.MethodPost, ticketAPIBase+"/tickets/"+created.ID+"/restore", "")
	require.Equal(t, http.StatusOK, w.Code)
	restored := decodeJSON[domain.Ticket](t, w)
	assert.Nil(t, restored.ArchivedAt)
}

func statusIDByCategory(statuses []domain.TicketStatus, category domain.TicketStatusCategory) string {
	for _, s := range statuses {
		if s.Category == category {
			return s.ID
		}
	}
	return ""
}

// --- 親子・入力検証 ---

func Test_チケット作成_担当が存在しなければ400(t *testing.T) {
	f := newTicketFixture(kbUserID, domain.GrantRoleEditor)
	f.tickets.addType(domain.TicketType{ID: "type-1", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, Name: "タスク", IsDefault: true})
	f.tickets.addStatus(domain.TicketStatus{ID: "status-1", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, Name: "To Do", Category: domain.TicketStatusCategoryTodo, IsInitial: true})
	created := postTicket(t, f, `{"title":"x"}`)

	w := f.do(t, http.MethodPut, ticketAPIBase+"/tickets/"+created.ID+"/assignee", `{"assigneePrincipalId":"`+ticketFakeMissingPrincipalID+`"}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var body errorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "invalid_assignee", body.Error)
}

func Test_チケット作成_開始日が期限より後なら400(t *testing.T) {
	f := newTicketFixture(kbUserID, domain.GrantRoleEditor)
	f.tickets.addType(domain.TicketType{ID: "type-1", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, Name: "タスク", IsDefault: true})
	f.tickets.addStatus(domain.TicketStatus{ID: "status-1", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, Name: "To Do", Category: domain.TicketStatusCategoryTodo, IsInitial: true})

	w := f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/tickets",
		`{"title":"x","startDate":"2026-09-10","dueDate":"2026-09-01"}`)
	require.Equal(t, http.StatusBadRequest, w.Code)
	var body errorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "invalid_date_range", body.Error)
}

func Test_チケット作成_日付の形式が不正なら400(t *testing.T) {
	f := newTicketFixture(kbUserID, domain.GrantRoleEditor)
	f.tickets.addType(domain.TicketType{ID: "type-1", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, Name: "タスク", IsDefault: true})
	f.tickets.addStatus(domain.TicketStatus{ID: "status-1", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, Name: "To Do", Category: domain.TicketStatusCategoryTodo, IsInitial: true})

	w := f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/tickets", `{"title":"x","dueDate":"2026/09/10"}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_チケット作成_優先度が範囲外なら400(t *testing.T) {
	f := newTicketFixture(kbUserID, domain.GrantRoleEditor)
	f.tickets.addType(domain.TicketType{ID: "type-1", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, Name: "タスク", IsDefault: true})
	f.tickets.addStatus(domain.TicketStatus{ID: "status-1", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, Name: "To Do", Category: domain.TicketStatusCategoryTodo, IsInitial: true})

	w := f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/tickets", `{"title":"x","priority":9}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_チケット親子_階層規則に反すると409(t *testing.T) {
	f := newTicketFixture(kbUserID, domain.GrantRoleEditor)
	subType := f.tickets.addType(domain.TicketType{ID: "type-sub", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, Name: "小作業", HierarchyLevel: -1})
	f.tickets.addStatus(domain.TicketStatus{ID: "status-1", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, Name: "To Do", Category: domain.TicketStatusCategoryTodo, IsInitial: true})
	parent := f.tickets.addTicket(domain.Ticket{ID: "parent-1", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, TypeID: subType.ID, Title: "親", Number: 1})

	w := f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/tickets",
		`{"title":"子","parentId":"`+parent.ID+`","typeId":"`+subType.ID+`"}`)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func postTicket(t *testing.T, f ticketFixture, body string) domain.Ticket {
	t.Helper()
	w := f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/tickets", body)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	return decodeJSON[domain.Ticket](t, w)
}

// --- 状態・種別マスタの管理 ---

func Test_状態マスタ_作成更新初期化アーカイブ復元(t *testing.T) {
	f := newTicketFixture(kbUserID, domain.GrantRoleEditor)

	w := f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-statuses",
		`{"name":"レビュー中","category":"in_progress","color":"#2f6b47"}`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	status := decodeJSON[domain.TicketStatus](t, w)

	w = f.do(t, http.MethodPut, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-statuses/"+status.ID,
		`{"name":"レビュー中2","category":"in_progress","color":"#a0661a"}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	w = f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-statuses/"+status.ID+"/set-initial", "")
	require.Equal(t, http.StatusNoContent, w.Code)

	w = f.do(t, http.MethodGet, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-statuses", "")
	require.Equal(t, http.StatusOK, w.Code)
	list := decodeJSON[ticketStatusListResponse](t, w)
	require.Len(t, list.Statuses, 1)
	assert.True(t, list.Statuses[0].IsInitial)

	// 現役チケットが参照していれば 409。
	f.tickets.addTicket(domain.Ticket{ID: "t-in-use", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, StatusID: status.ID, TypeID: "type-x", Title: "使用中"})
	w = f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-statuses/"+status.ID+"/archive", "")
	assert.Equal(t, http.StatusConflict, w.Code)

	delete(f.tickets.tickets, "t-in-use")
	w = f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-statuses/"+status.ID+"/archive", "")
	require.Equal(t, http.StatusNoContent, w.Code)

	w = f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-statuses/"+status.ID+"/restore", "")
	require.Equal(t, http.StatusNoContent, w.Code)
}

func Test_種別マスタ_作成更新既定アーカイブ復元(t *testing.T) {
	f := newTicketFixture(kbUserID, domain.GrantRoleEditor)

	w := f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-types",
		`{"name":"バグ","hierarchyLevel":0,"color":"#9a3b2e"}`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	typ := decodeJSON[domain.TicketType](t, w)

	w = f.do(t, http.MethodPut, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-types/"+typ.ID,
		`{"name":"バグ2","hierarchyLevel":1,"color":"#2f6b47"}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	w = f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-types/"+typ.ID+"/set-default", "")
	require.Equal(t, http.StatusNoContent, w.Code)

	w = f.do(t, http.MethodGet, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-types", "")
	require.Equal(t, http.StatusOK, w.Code)
	list := decodeJSON[ticketTypeListResponse](t, w)
	require.Len(t, list.Types, 1)
	assert.True(t, list.Types[0].IsDefault)

	w = f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-types/"+typ.ID+"/archive", "")
	require.Equal(t, http.StatusNoContent, w.Code)
	w = f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-types/"+typ.ID+"/restore", "")
	require.Equal(t, http.StatusNoContent, w.Code)
}

func Test_状態作成_不正な色は400(t *testing.T) {
	f := newTicketFixture(kbUserID, domain.GrantRoleEditor)
	w := f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-statuses",
		`{"name":"x","category":"todo","color":"not-a-color"}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_種別作成_範囲外のhierarchyLevelは400(t *testing.T) {
	f := newTicketFixture(kbUserID, domain.GrantRoleEditor)
	w := f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-types",
		`{"name":"x","hierarchyLevel":5,"color":"#2f6b47"}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_状態種別マスタ_閲覧のみでは編集操作に403(t *testing.T) {
	f := newTicketFixture(kbUserID, domain.GrantRoleViewer)
	w := f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-statuses",
		`{"name":"x","category":"todo","color":"#2f6b47"}`)
	assert.Equal(t, http.StatusForbidden, w.Code)

	w = f.do(t, http.MethodPost, ticketAPIBase+"/spaces/"+kbSpaceID+"/ticket-types",
		`{"name":"x","hierarchyLevel":0,"color":"#2f6b47"}`)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
