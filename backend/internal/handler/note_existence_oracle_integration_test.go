//go:build integration

package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// 存在オラクルとは:
//
//	応答の違いから「その ID のレコードが実在するか」を外部から判定できてしまう状態のこと。
//	notes.id は連番（bigserial）なので、ログイン済みユーザーは 1, 2, 3 … と ID を順に叩ける。
//	このとき「他人のノートなので拒否（403）」と「そんなノートは無い（400/404）」で応答が
//	違えば、本文を 1 文字も読めなくても、どの ID が埋まっているか＝社内にノートが何件あり
//	誰がいつ書き始めたかを全数把握できる。存在の有無そのものが他人の情報なので、
//	他人のノートと存在しないノートは「ステータスも本文もまったく同じ」でなければならない。
//
// このファイルは routes_note.go に登録された全経路を実 PostgreSQL 上で叩き、
// その同一性をバイト単位で固定する。ステータスだけ揃えても本文が違えば穴は残るため、
// 比較は必ず生のレスポンスボディ（[]byte）で行う。

// noteOracleRouter は本番と同じ registerNoteRoutes でルートを張り、userID を current user として注入する。
// 認証済みであること以外は本番と同じ経路を通るので、「ログインしていれば他人のノートの
// 存在を数え上げられるか」をエンドポイント単位で確かめられる。
//
// cfg は presigner の初期化に読まれる。NOTE_IMAGES_BUCKET 相当が空なので stub presigner に落ち、
// テストが S3 に依存しない（/notes/images/upload-url も本番と同じ登録経路で叩ける）。
func noteOracleRouter(db *sql.DB, userID uint64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("")
	g.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyCurrentUserID, userID)
		c.Next()
	})
	registerNoteRoutes(g, &routeDeps{db: db, cfg: &config.Config{}})
	return r
}

// noteOracleCall は 1 リクエストを投げ、ステータスと生のボディを返す。
// ボディは文字列比較ではなくバイト列そのものを見るために *httptest.ResponseRecorder から取る。
func noteOracleCall(t *testing.T, r *gin.Engine, method, path, body string) (int, []byte) {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code, w.Body.Bytes()
}

// TestNoteExistenceOracle_Integration は notes に対する全 HTTP 経路（routes_note.go）で
// 「他人の実在ノート」と「存在しないノート」が区別できないことを実 PostgreSQL 上で固定する。
//
// 対象経路（routes_note.go の登録順そのまま）:
//
//	GET    /notes                    … ID を受け取らない（列挙対象なし）。他人のノートが混ざらないことだけ見る
//	POST   /notes                    … ID を受け取らない（列挙対象なし）。current user 名義で作られることだけ見る
//	PUT    /notes/:id                … ★ID を受け取る。撃ち分けを潰した本命
//	DELETE /notes/:id                … ★ID を受け取る。SQL で user_id 固定。他人と不在が同じ 404 で揃うことを見る
//	POST   /notes/images/upload-url  … ノート ID を受け取らない（列挙対象なし）
//	GET    /sessions/:sessionId/note … ★ID を受け取る。session_notes 側。既に 404 へ畳まれていることを見る
//	PUT    /sessions/:sessionId/note … ★ID を受け取る。session_notes 側。ここでは固定しない（下記の注記を参照）
func TestNoteExistenceOracle_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "notes", "session_notes")

	const (
		me    uint64 = 7 // 攻撃者役（ログイン済みの一般ユーザー）
		other uint64 = 8 // 被害者役（他人）
	)

	noteRepo := persistence.NewNoteRepository(sqlDB)
	mkNote := func(owner uint64, title, content string) *domain.Note {
		n := &domain.Note{UserID: owner, Title: title, Content: content}
		require.NoError(t, noteRepo.Create(ctx, n))
		return n
	}

	theirNote := mkNote(other, "他人のノート", "他人にだけ見えるべき本文")
	myNote := mkNote(me, "自分のノート", "自分の本文")

	// 実在しない ID。TRUNCATE ... RESTART IDENTITY 済みなので採番値は小さく、
	// ここまで飛ばせば確実に空き番になる。
	missingID := theirNote.ID + 1_000_000

	theirPath := "/notes/" + strconv.FormatUint(theirNote.ID, 10)
	missingPath := "/notes/" + strconv.FormatUint(missingID, 10)
	attacker := noteOracleRouter(sqlDB, me)

	t.Run("PUT /notes/:id は他人の実在ノートと存在しないノートで同じ応答を返す", func(t *testing.T) {
		const body = `{"title":"上書き","content":"上書き本文"}`

		foreignCode, foreignBody := noteOracleCall(t, attacker, http.MethodPut, theirPath, body)
		missingCode, missingBody := noteOracleCall(t, attacker, http.MethodPut, missingPath, body)

		require.Equal(t, http.StatusNotFound, foreignCode, "他人のノートは 404（403 だと実在が分かる）")
		require.Equal(t, http.StatusNotFound, missingCode, "存在しないノートも 404")
		// ステータスが同じでも本文が違えば、本文の差だけで実在を判定できる。
		// 文字列ではなくバイト列で比較して、空白 1 つの差も見逃さない。
		require.Equal(t, missingBody, foreignBody,
			"本文が撃ち分けられている: 他人=%q 不在=%q", foreignBody, missingBody)
	})

	t.Run("PUT /notes/:id は他人のノートを書き換えない", func(t *testing.T) {
		got, err := noteRepo.FindByID(ctx, other, theirNote.ID)
		require.NoError(t, err)
		require.Equal(t, "他人のノート", got.Title, "404 を返すだけでなく実際に書き換わっていないこと")
		require.Equal(t, "他人にだけ見えるべき本文", got.Content)
	})

	t.Run("PUT /notes/:id 自分のノートは従来どおり更新できる", func(t *testing.T) {
		myPath := "/notes/" + strconv.FormatUint(myNote.ID, 10)
		code, body := noteOracleCall(t, attacker, http.MethodPut, myPath,
			`{"title":"更新後","content":"更新後の本文","isPublic":false,"isPinned":true}`)
		require.Equal(t, http.StatusOK, code, string(body))

		var updated domain.Note
		require.NoError(t, json.Unmarshal(body, &updated))
		require.Equal(t, "更新後", updated.Title)
		require.True(t, updated.IsPinned)

		persisted, err := noteRepo.FindByID(ctx, me, myNote.ID)
		require.NoError(t, err)
		require.Equal(t, "更新後の本文", persisted.Content, "DB にも反映されている")
	})

	// 期待値を 204 から 404 へ更新した理由:
	//   以前は「他人・不在・自分」のすべてで 204 を返していた。0 行削除まで成功にしていると、
	//   呼び出し側は「自分のノートを 1 件消せた」と「1 行も消えなかった」を区別できず、
	//   削除できていないのに画面からは行が消える（保存されていないものを保存済みに見せるのと同じ）。
	//   そこで 0 行削除を domain.ErrNotFound へ翻訳し、handler が 404 を返すようにした。
	//   存在オラクルは開かない: DELETE の WHERE に user_id が入っているため「他人のノート」も
	//   「存在しない id」もどちらも 0 行 = 同じ 404・同じ本文になる。応答が分かれるのは
	//   「自分のノートを実際に消せたか」だけで、それは自分の情報。PUT /notes/:id が
	//   既にこの畳み方（他人も不在も 404）になっており、更新と削除で結末が揃う。
	t.Run("DELETE /notes/:id は他人と不在で同じ 404 を返し、自分のノートだけ消える", func(t *testing.T) {
		foreignCode, foreignBody := noteOracleCall(t, attacker, http.MethodDelete, theirPath, "")
		missingCode, missingBody := noteOracleCall(t, attacker, http.MethodDelete, missingPath, "")

		require.Equal(t, http.StatusNotFound, foreignCode, "他人のノートは 404（204 だと消せたと誤認する）")
		require.Equal(t, http.StatusNotFound, missingCode, "存在しないノートも 404")
		// ステータスが同じでも本文が違えば、本文の差だけで実在を判定できる。
		// 文字列ではなくバイト列で比較して、空白 1 つの差も見逃さない。
		require.Equal(t, missingBody, foreignBody,
			"本文が撃ち分けられている: 他人=%q 不在=%q", foreignBody, missingBody)

		// 404 を返すだけでなく、他人のノートが実際に消えていないことも見る（WHERE user_id が効いている）。
		survivor, err := noteRepo.FindByID(ctx, other, theirNote.ID)
		require.NoError(t, err, "他人のノートは残る")
		require.Equal(t, "他人のノート", survivor.Title)

		// 自分のノートは 204 で消える。ここだけ応答が分かれるが、分かるのは自分の情報だけ。
		myPath := "/notes/" + strconv.FormatUint(myNote.ID, 10)
		mineCode, mineBody := noteOracleCall(t, attacker, http.MethodDelete, myPath, "")
		require.Equal(t, http.StatusNoContent, mineCode)
		require.Empty(t, mineBody, "204 は本文なし")
		_, err = noteRepo.FindByID(ctx, me, myNote.ID)
		require.ErrorIs(t, err, domain.ErrNotFound, "自分のノートは実際に消えている")

		// 同じノートをもう一度消すと、既に無いので不在と同じ 404 に畳まれる。
		againCode, againBody := noteOracleCall(t, attacker, http.MethodDelete, myPath, "")
		require.Equal(t, http.StatusNotFound, againCode)
		require.Equal(t, missingBody, againBody)
	})

	t.Run("GET /notes は他人のノートを一切返さない", func(t *testing.T) {
		code, body := noteOracleCall(t, attacker, http.MethodGet, "/notes", "")
		require.Equal(t, http.StatusOK, code, string(body))

		var rows []domain.Note
		require.NoError(t, json.Unmarshal(body, &rows))
		for _, row := range rows {
			require.Equal(t, me, row.UserID, "一覧は current user のノートだけ")
		}
	})

	t.Run("POST /notes は ID を受け取らず current user 名義で作る", func(t *testing.T) {
		code, body := noteOracleCall(t, attacker, http.MethodPost, "/notes",
			`{"title":"新規","content":"本文"}`)
		require.Equal(t, http.StatusCreated, code, string(body))

		var created domain.Note
		require.NoError(t, json.Unmarshal(body, &created))
		require.Equal(t, me, created.UserID, "body で userId を指定できない（ID 空間に触れない）")
	})

	t.Run("POST /notes/images/upload-url はノート ID を受け取らない", func(t *testing.T) {
		// 経路の網羅として叩く。パスにも body にもノート ID が無いので、
		// この経路から ID の実在を問い合わせる方法自体が存在しない。
		code, body := noteOracleCall(t, attacker, http.MethodPost, "/notes/images/upload-url", `{}`)
		require.Equal(t, http.StatusOK, code, string(body))
	})
}

// TestSessionNoteExistenceOracle_Integration は routes_note.go が同時に登録する
// session_notes 側の 2 経路について、同じ観点（他人のもの と 存在しない が区別できないこと）を固定する。
// notes とはテーブルが別だが、列挙のもれを防ぐため同じファイルで押さえておく。
func TestSessionNoteExistenceOracle_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "session_notes")

	const (
		me    uint64 = 7
		other uint64 = 8
	)

	sessionRepo := persistence.NewSessionNoteRepository(sqlDB)
	theirSession := uint64(4001)
	require.NoError(t, sessionRepo.Upsert(ctx, &domain.SessionNote{
		SessionID: theirSession, UserID: other, Content: "他人のセッションメモ",
	}))
	missingSession := uint64(4002) // 誰もメモを書いていないセッション

	attacker := noteOracleRouter(sqlDB, me)
	theirPath := "/sessions/" + strconv.FormatUint(theirSession, 10) + "/note"
	missingPath := "/sessions/" + strconv.FormatUint(missingSession, 10) + "/note"

	// PUT /sessions/:sessionId/note（Upsert）はここでは固定しない。
	// 実測（本ブランチで確認）では、他人のセッションに対する PUT が 200 を返したうえで
	// ON CONFLICT (session_id) DO UPDATE により他人のメモ本文を実際に上書きしてしまう
	// （行の user_id は元の所有者のまま残り、content だけ書き換わる）。
	// これは「存在が漏れる」より重い越権書き込みで、session_notes 側の別の欠陥。
	// notes の存在オラクルを塞ぐ本チケットの範囲外なので、ここで現状の挙動を
	// テストとして固定すると欠陥を仕様として固めてしまう。別チケットで扱う。
	t.Run("GET /sessions/:sessionId/note は他人のメモと未作成で同じ応答を返す", func(t *testing.T) {
		foreignCode, foreignBody := noteOracleCall(t, attacker, http.MethodGet, theirPath, "")
		missingCode, missingBody := noteOracleCall(t, attacker, http.MethodGet, missingPath, "")

		require.Equal(t, http.StatusNotFound, foreignCode)
		require.Equal(t, http.StatusNotFound, missingCode)
		require.Equal(t, missingBody, foreignBody,
			"本文が撃ち分けられている: 他人=%q 未作成=%q", foreignBody, missingBody)
	})
}
