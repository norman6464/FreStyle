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

// AiChatSseHandler は AI チャット用の SSE エンドポイント。
// イベントは session / token / done / error の 4 種。
// EventSource は POST 非対応なのでクライアントは fetch + ReadableStream で読む。
type AiChatSseHandler struct {
	sendStream *usecase.SendAiMessageStreamUseCase
}

func NewAiChatSseHandler(sendStream *usecase.SendAiMessageStreamUseCase) *AiChatSseHandler {
	return &AiChatSseHandler{sendStream: sendStream}
}

type sseRequestBody struct {
	SessionID   uint64                 `json:"sessionId"`
	Content     string                 `json:"content"`
	Scene       string                 `json:"scene"`
	SessionType string                 `json:"sessionType"`
	ScenarioID  *uint64                `json:"scenarioId"`
	Attachments []sseAttachmentRequest `json:"attachments"`
}

// sseAttachmentRequest は S3 へアップロード済みの添付参照（key とメタのみ、実体は backend が取得）。
type sseAttachmentRequest struct {
	Key         string `json:"key"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}

// maxAttachmentsPerMessage は 1 リクエストの添付上限（UI / 課金観点で 4 枚に絞る）。
const maxAttachmentsPerMessage = 4

// @Summary      AI チャット SSE ストリーミング
// @Description  Bedrock Claude へ メッセージ を 送信 し、 token を SSE で 配信。 OpenAPI は SSE の カスタム イベント を 完全 表現 でき ない ので レスポンス は string と して 簡略 表現。 実際 の イベント 形式 は session / token / done / error の 4 種 (詳細 は handler コメント 参照)。 エラー 系 (400/401/503) は 通常 の application/json で 返る。
// @Tags         ai-chat
// @Accept       json
// @Produce      text/event-stream
// @Param        body  body  sseRequestBody  true  "sessionId / content / scene / sessionType / scenarioId / attachments (最大 4 件)"
// @Success      200   {string}  string  "SSE stream (text/event-stream)"
// @Failure      400   {object}  errorResponse  "バリデーション (application/json)"
// @Failure      401   {object}  errorResponse  "未 認証 (application/json)"
// @Failure      503   {object}  errorResponse  "Bedrock / DynamoDB 未 設定 (dev/stub、 application/json)"
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

	// client 切断で cancel される ctx を usecase に渡し、goroutine リークを防ぐ。
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	stream, err := h.sendStream.Execute(ctx, usecase.SendAiMessageInput{
		UserID:      uid,
		SessionID:   body.SessionID,
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

	// keepalive: 15 秒ごとにコメント行を送り、ALB / CloudFront のアイドルタイムアウトを防ぐ。
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

// writeSSEEvent は 1 イベントを SSE フォーマット（event: 行 + data: 行 + 空行）で書き出す。
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

// flushOrPanic は flush 不可なら panic する（SSE が機能しない bug を開発時に早期発見するため）。
func flushOrPanic(w gin.ResponseWriter) {
	if f, ok := any(w).(interface{ Flush() }); ok {
		f.Flush()
		return
	}
	if _, ok := any(w).(io.Writer); !ok {
		panic("ai_chat_sse_handler: ResponseWriter is not flushable")
	}
}

// buildAttachmentsFromRequest はリクエストの attachments[] を検証して domain.Attachment に変換する。
// contentType / sizeBytes の検証に加え、key が ai-chat/{userID}/ prefix に属することを必須にして
// 他ユーザーや他 prefix の S3 オブジェクトをサーバ側で読まされるのを防ぐ。
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
