package handler

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_ナレッジAPI_最近見たページ は /api/v2/kb/me/recent-pages（段2）を確かめる。
// ワークスペースを横断する経路なので Test_ナレッジAPI_登録済みルートは全て認可テストの対象になっている
// の表からは外し、ここで直接叩く。
func Test_ナレッジAPI_最近見たページ(t *testing.T) {
	t.Run("未認証なら401", func(t *testing.T) {
		f := newKbFixture(kbCanView, 0)
		w := f.do(t, http.MethodGet, "/api/v2/kb/me/recent-pages", "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("候補が無ければ空配列", func(t *testing.T) {
		f := newKbFixture(kbCanView, kbUserID)
		w := f.do(t, http.MethodGet, "/api/v2/kb/me/recent-pages", "")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, "[]", strings.TrimSpace(w.Body.String()))
	})

	// hiddenPageID は実在するページ（kbDestPageID）で、明示的に CanView を落としてある
	// （存在しない ID だと CheckPagePermissionUseCase 自体が「page not found」で先に
	// continue してしまい、下の !perm.CanView 分岐を経由しないまま偶然テストが通ってしまう
	// ため、必ず実在するページを使う）。
	//
	// 変異確認: ListMyRecentPagesUseCase.Execute の `if !perm.CanView { continue }` を
	// 外すと、見えないはずの hiddenPageID がここに出てきてこのテストが落ちる。
	t.Run("見えないページは可視判定でふるわれて落ちる", func(t *testing.T) {
		f := newKbFixture(kbCanView, kbUserID)
		hiddenPageID := kbDestPageID
		f.perms.setPagePermission(hiddenPageID, kbUserID, domain.PagePermission{})
		f.views.recentFor[kbUserID] = []domain.RecentPage{
			{
				PageID: kbRootPageID, WorkspaceID: kbWorkspaceID, WorkspaceSlug: kbWorkspaceSlug,
				Title: "root", SpaceID: kbSpaceID, SpaceName: "space", ViewedAt: time.Now(),
			},
			{
				PageID: hiddenPageID, WorkspaceID: kbWorkspaceID, WorkspaceSlug: kbWorkspaceSlug,
				Title: "hidden", SpaceID: kbSpaceID, SpaceName: "space", ViewedAt: time.Now(),
			},
		}
		w := f.do(t, http.MethodGet, "/api/v2/kb/me/recent-pages", "")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		body := w.Body.String()
		assert.Contains(t, body, `"pageId":"`+kbRootPageID+`"`)
		assert.NotContains(t, body, hiddenPageID)
	})
}
