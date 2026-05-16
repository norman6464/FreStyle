package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// AiChatSseHandler は AI チャット用の Server-Sent Events エンドポイント。
//
// 送信フォーマット (text/event-stream):
//
//	event: session
//	data: {"id": 42, "title": "...", ...}
//
//	event: token
//	data: {"sessionId": 42, "delta": "ご"}
//
//	event: done
//	data: {"sessionId": 42, "id": "uuid", "createdAt": "...", "content": "ご質問..."}
//
//	event: error
//	data: {"message": "..."}
//
// クライアントは fetch + ReadableStream で line-based に読む（標準の EventSource は POST に
// 対応しないため、本実装では fetch ベースで送る）。
type AiChatSseHandler struct {
	sendStream *usecase.SendAiMessageStreamUseCase
}

func NewAiChatSseHandler(sendStream *usecase.SendAiMessageStreamUseCase) *AiChatSseHandler {
	return &AiChatSseHandler{sendStream: sendStream}
}

type sseRequestBody struct {
	SessionID   int64                  `json:"sessionId"`
	Content     string                 `json:"content"`
	Scene       string                 `json:"scene"`
	SessionType string                 `json:"sessionType"`
	ScenarioID  *uint64                `json:"scenarioId"`
	Attachments []sseAttachmentRequest `json:"attachments"`
}

// sseAttachmentRequest はフロントが事前に S3 へアップロード済みの添付ファイルを参照する。
// 実体バイトは含まず、key とメタだけ送って backend が S3 GetObject で取り出す。
type sseAttachmentRequest struct {
	Key         string `json:"key"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}

// 1 リクエストあたりの添付上限。Bedrock 仕様（image 20 / document 5）に対し、
// UI / 課金観点で 4 枚に絞っている（仕様検討で確定）。
const maxAttachmentsPerMessage = 4

// @Summary      AI チャット SSE ストリーミング
// @Description  Bedrock Claude へ メッセージ を 送信 し、 token を SSE (text/event-stream) で 配信。 OpenAPI は SSE を 完全 表現 でき ない ので レスポンス は 200 string と して 簡略 化。 実際 の イベント 形式 は session / token / done / error の 4 種。
// @Tags         ai-chat
// @Accept       json
// @Produce      plain
// @Param        body  body  sseRequestBody  true  "sessionId / content / scene / sessionType / scenarioId / attachments"
// @Success      200   {string}  string  "SSE stream (text/event-stream)"
// @Failure      400   {object}  errorResponse  "バリデーション"
// @Failure      401   {object}  errorResponse  "未 認証"
// @Failure      503   {object}  errorResponse  "Bedrock / DynamoDB 未 設定 (dev/stub)"
// @Router       /ai-chat/stream [post]
// @Security     CookieAuth
func (h *AiChatSseHandler) Handle(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if h.sendStream == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ai_chat_unavailable"})
		return
	}

	var body sseRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Content == "" && len(body.Attachments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content_required"})
		return
	}
	if len(body.Attachments) > maxAttachmentsPerMessage {
		c.JSON(http.StatusBadRequest, gin.H{"error": "too_many_attachments"})
		return
	}
	attachments, err := buildAttachmentsFromRequest(uid, body.Attachments)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// SSE ヘッダ
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	flushOrPanic(c.Writer)

	// usecase の context は client 切断で cancel される c.Request.Context() を渡す。
	// goroutine が長時間生きないように切断検知を伝播させる。
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	stream, err := h.sendStream.Execute(ctx, usecase.SendAiMessageInput{
		UserID:      uid,
		SessionID:   uint64(body.SessionID),
		Content:     body.Content,
		Scene:       body.Scene,
		SessionType: body.SessionType,
		ScenarioID:  body.ScenarioID,
		Attachments: attachments,
	})
	if err != nil {
		writeSSEEvent(c.Writer, "error", map[string]string{"message": "メッセージの送信に失敗しました"})
		return
	}

	// keepalive: 15 秒ごとにコメント行 ":\n\n" を送って ALB / CloudFront のアイドルタイムアウト
	// （既定 60 秒）で切断されるのを防ぐ。
	keep := time.NewTicker(15 * time.Second)
	defer keep.Stop()

	clientGone := c.Writer.CloseNotify()

	for {
		select {
		case <-clientGone:
			return
		case <-keep.C:
			if _, err := c.Writer.Write([]byte(": keepalive\n\n")); err != nil {
				return
			}
			c.Writer.Flush()
		case ev, ok := <-stream:
			if !ok {
				// usecase 側の channel が close した = 完了（FinalMessage で done 通知済）
				return
			}
			if ev.Err != nil {
				log.Printf("AiChat SSE: usecase error: %v", ev.Err)
				writeSSEEvent(c.Writer, "error", map[string]string{"message": "メッセージの送信に失敗しました"})
				return
			}
			if ev.NewSession != nil {
				writeSSEEvent(c.Writer, "session", map[string]any{
					"id":          ev.NewSession.ID,
					"title":       ev.NewSession.Title,
					"sessionType": ev.NewSession.SessionType,
					"scenarioId":  ev.NewSession.ScenarioID,
					"createdAt":   ev.NewSession.CreatedAt.Format(time.RFC3339),
				})
				continue
			}
			if ev.Delta != "" {
				writeSSEEvent(c.Writer, "token", map[string]any{
					"delta": ev.Delta,
				})
				continue
			}
			if ev.FinalMessage != nil {
				writeSSEEvent(c.Writer, "done", map[string]any{
					"sessionId": ev.FinalMessage.SessionID,
					"id":        ev.FinalMessage.MessageID,
					"role":      ev.FinalMessage.Role,
					"content":   ev.FinalMessage.Content,
					"createdAt": ev.FinalMessage.CreatedAt.Format(time.RFC3339),
				})
				return
			}
		}
	}
}

// writeSSEEvent は 1 イベントを SSE フォーマットで書き出す。
//
// 仕様: `event:` 行の後に空行 / `data:` 行 / 空行で 1 イベント完結。
// 改行を含む payload は `data:` 行を分割する必要があるが、ここでは JSON を 1 行で
// 出すので分割不要。
func writeSSEEvent(w gin.ResponseWriter, event string, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("AiChat SSE: marshal failed: %v", err)
		return
	}
	if _, err := w.Write([]byte("event: " + event + "\n")); err != nil {
		return
	}
	if _, err := w.Write([]byte("data: ")); err != nil {
		return
	}
	if _, err := w.Write(body); err != nil {
		return
	}
	if _, err := w.Write([]byte("\n\n")); err != nil {
		return
	}
	w.Flush()
}

// flushOrPanic は最初のヘッダ送信で flush できないと SSE 自体機能しないので
// 開発時の bug を早期発見するため panic させる。本番では gin の ResponseWriter は
// 必ず Flusher を実装するので発生しない。
func flushOrPanic(w gin.ResponseWriter) {
	if f, ok := any(w).(interface{ Flush() }); ok {
		f.Flush()
		return
	}
	if _, ok := any(w).(io.Writer); !ok {
		panic("ai_chat_sse_handler: ResponseWriter is not flushable")
	}
}

// buildAttachmentsFromRequest はリクエスト body の attachments[] を domain.Attachment に変換する。
//
// 各添付について:
//   - key / contentType 必須
//   - contentType は usecase.AllowedAttachmentContentTypes に含まれる必要あり
//   - sizeBytes は MIME ごとの上限以下
//   - key は **userID 配下の S3 prefix** `ai-chat/{userID}/` に必ず属する
//     （他ユーザーの添付や他 prefix の S3 オブジェクトをサーバ側で読まされるのを防止）
//
// バリデーションを SSE handler 側で 1 回通すのは「presigned URL 取得を経由したか」を
// 軽く再確認する目的（プロンプトインジェクションで任意 S3 オブジェクトを読み出させない）。
func buildAttachmentsFromRequest(userID uint64, reqs []sseAttachmentRequest) ([]domain.Attachment, error) {
	if len(reqs) == 0 {
		return nil, nil
	}
	expectedPrefix := fmt.Sprintf("ai-chat/%d/", userID)
	out := make([]domain.Attachment, 0, len(reqs))
	for _, r := range reqs {
		if r.Key == "" || r.ContentType == "" {
			return nil, sseAttachmentError("attachment_invalid")
		}
		if !strings.HasPrefix(r.Key, expectedPrefix) {
			return nil, sseAttachmentError("attachment_key_not_allowed")
		}
		rule, ok := usecase.AllowedAttachmentContentTypes[r.ContentType]
		if !ok {
			return nil, sseAttachmentError("attachment_unsupported_type")
		}
		if r.SizeBytes <= 0 || r.SizeBytes > rule.Max {
			return nil, sseAttachmentError("attachment_too_large")
		}
		out = append(out, domain.Attachment{
			Key:         r.Key,
			Filename:    r.Filename,
			ContentType: r.ContentType,
			Format:      rule.Format,
			Kind:        rule.Kind,
			SizeBytes:   r.SizeBytes,
		})
	}
	return out, nil
}

// sseAttachmentError は handler 内で 400 を返すための軽量エラー型。
type sseAttachmentError string

func (e sseAttachmentError) Error() string { return string(e) }
