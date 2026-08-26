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
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// --- stub: repository（handler の応答だけを見たいので DB は使わない） ---

type stubSessionNoteRepoForHandler struct {
	n *domain.SessionNote
}

func (s *stubSessionNoteRepoForHandler) FindBySessionID(_ context.Context, _ uint64) (*domain.SessionNote, error) {
	return s.n, nil
}

func (s *stubSessionNoteRepoForHandler) Upsert(_ context.Context, n *domain.SessionNote) error {
	s.n = n
	return nil
}

// stubAiChatSessionRepoForHandler はセッションの所有者だけを持つ stub。
// session が nil のときは「そのセッションは存在しない」を表し、本物と同じく
// domain.ErrNotFound を返す。
type stubAiChatSessionRepoForHandler struct {
	session *domain.AiChatSession
}

func (s *stubAiChatSessionRepoForHandler) ListByUserID(_ context.Context, _ uint64) ([]domain.AiChatSession, error) {
	return nil, nil
}

func (s *stubAiChatSessionRepoForHandler) FindByID(_ context.Context, _ uint64) (*domain.AiChatSession, error) {
	if s.session == nil {
		return nil, domain.ErrNotFound
	}
	return s.session, nil
}

func (s *stubAiChatSessionRepoForHandler) Create(_ context.Context, _ *domain.AiChatSession) error {
	return nil
}

func (s *stubAiChatSessionRepoForHandler) UpdateTitle(_ context.Context, _ uint64, _ string) error {
	return nil
}

func (s *stubAiChatSessionRepoForHandler) Delete(_ context.Context, _ uint64) error { return nil }

// sessionNoteCtx は sessionId パス パラメータ付きのテスト用 gin.Context を作る。
// uid が 0 のときは current user を入れない（未認証の経路を再現する）。
func sessionNoteCtx(method, body string, uid uint64, sessionID string) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	if uid != 0 {
		c.Set(middleware.ContextKeyCurrentUserID, uid)
	}
	c.Params = gin.Params{{Key: "sessionId", Value: sessionID}}
	return w, c
}

// newSessionNoteHandler は所有者が owner のセッション（owner が 0 なら存在しないセッション）と、
// 既存メモ note を持つ handler を組み立てる。
func newSessionNoteHandler(owner uint64, note *domain.SessionNote) *SessionNoteHandler {
	notes := &stubSessionNoteRepoForHandler{n: note}
	sessions := &stubAiChatSessionRepoForHandler{}
	if owner != 0 {
		sessions.session = &domain.AiChatSession{ID: 1, UserID: owner}
	}
	return NewSessionNoteHandler(
		usecase.NewGetSessionNoteUseCase(notes),
		usecase.NewUpsertSessionNoteUseCase(notes, sessions),
	)
}

func Test_セッションノートハンドラ_保存_未認証(t *testing.T) {
	w, c := sessionNoteCtx(http.MethodPut, `{}`, 0, "1")
	(&SessionNoteHandler{}).Upsert(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func Test_セッションノートハンドラ_保存_不正なJSON(t *testing.T) {
	w, c := sessionNoteCtx(http.MethodPut, `not-json`, 7, "1")
	(&SessionNoteHandler{}).Upsert(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func Test_セッションノートハンドラ_保存_自分のセッションなら200(t *testing.T) {
	w, c := sessionNoteCtx(http.MethodPut, `{"content":"hello"}`, 7, "1")
	newSessionNoteHandler(7, nil).Upsert(c)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String())
	}
}

// 他人のセッションと存在しないセッションは、ステータスも本文も同じでなければならない。
// どちらか一方だけ別の応答にすると、その差から session_id の実在を総当たりで判別できてしまう。
func Test_セッションノートハンドラ_保存_書けないセッションは404(t *testing.T) {
	// 応答本文の比較用。読み出し側の 404 と同じ定数を使っていることも間接的に確かめる。
	wantBody := map[string]string{"error": sessionNoteNotFoundMsg}

	assert404 := func(t *testing.T, w *httptest.ResponseRecorder) {
		t.Helper()
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d body=%s", w.Code, w.Body.String())
		}
		var got map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("応答が JSON ではない: %v (%s)", err, w.Body.String())
		}
		if got["error"] != wantBody["error"] {
			t.Fatalf("want body %v, got %v", wantBody, got)
		}
	}

	t.Run("他人のセッション_メモ未作成", func(t *testing.T) {
		// メモがまだ無い他人のセッション。SQL の ON CONFLICT ... WHERE では止まらない経路。
		w, c := sessionNoteCtx(http.MethodPut, `{"content":"攻撃者が作成"}`, 999, "1")
		newSessionNoteHandler(7, nil).Upsert(c)
		assert404(t, w)
	})

	t.Run("他人のセッション_メモ作成済み", func(t *testing.T) {
		note := &domain.SessionNote{SessionID: 1, UserID: 7, Content: "所有者のメモ"}
		w, c := sessionNoteCtx(http.MethodPut, `{"content":"攻撃者が上書き"}`, 999, "1")
		newSessionNoteHandler(7, note).Upsert(c)
		assert404(t, w)
		if note.Content != "所有者のメモ" {
			t.Fatalf("内容が変わっている: %q", note.Content)
		}
	})

	t.Run("存在しないセッション", func(t *testing.T) {
		w, c := sessionNoteCtx(http.MethodPut, `{"content":"x"}`, 7, "12345")
		newSessionNoteHandler(0, nil).Upsert(c)
		assert404(t, w)
	})
}

// 読み出し側の 404 と本文が一致していること（存在の有無を応答差から推測させない）。
func Test_セッションノートハンドラ_取得_未作成は404で本文が書き込み側と同じ(t *testing.T) {
	w, c := sessionNoteCtx(http.MethodGet, "", 7, "1")
	newSessionNoteHandler(7, nil).Get(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("応答が JSON ではない: %v (%s)", err, w.Body.String())
	}
	if got["error"] != sessionNoteNotFoundMsg {
		t.Fatalf("want %q, got %q", sessionNoteNotFoundMsg, got["error"])
	}
}
