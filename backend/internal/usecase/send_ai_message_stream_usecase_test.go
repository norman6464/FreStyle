package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/bedrock"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- mock: AiChatSessionRepository ---

type mockSessionRepo struct{ mock.Mock }

func (m *mockSessionRepo) ListByUserID(ctx context.Context, userID uint64) ([]domain.AiChatSession, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.AiChatSession), args.Error(1)
}
func (m *mockSessionRepo) FindByID(ctx context.Context, id uint64) (*domain.AiChatSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AiChatSession), args.Error(1)
}
func (m *mockSessionRepo) Create(ctx context.Context, s *domain.AiChatSession) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}
func (m *mockSessionRepo) UpdateTitle(ctx context.Context, id uint64, title string) error {
	return m.Called(ctx, id, title).Error(0)
}
func (m *mockSessionRepo) Delete(ctx context.Context, id uint64) error {
	return m.Called(ctx, id).Error(0)
}

// --- mock: AiChatMessageRepository ---

type mockMsgRepo struct{ mock.Mock }

func (m *mockMsgRepo) Save(ctx context.Context, msg *domain.AiChatMessage) error {
	return m.Called(ctx, msg).Error(0)
}
func (m *mockMsgRepo) ListBySessionID(ctx context.Context, sessionID uint64) ([]domain.AiChatMessage, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]domain.AiChatMessage), args.Error(1)
}

// --- mock: StreamingBedrockClient ---

type mockStreamBedrock struct{ mock.Mock }

func (m *mockStreamBedrock) ConverseStream(ctx context.Context, systemPrompt string, history []domain.AiChatMessage) (<-chan bedrock.StreamEvent, error) {
	args := m.Called(ctx, systemPrompt, history)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan bedrock.StreamEvent), args.Error(1)
}

// 既存セッションの場合: Delta が逐次流れて、最後に FinalMessage で完了する。
func TestSendAiMessageStream_ExistingSession_StreamsTokensAndSavesFinal(t *testing.T) {
	sessionRepo := new(mockSessionRepo)
	msgRepo := new(mockMsgRepo)
	bc := new(mockStreamBedrock)

	// 履歴: ユーザー発話 1 件
	history := []domain.AiChatMessage{
		{SessionID: 5, MessageID: "u1", Role: domain.AiChatRoleUser, Content: "Hello", CreatedAt: time.Now()},
	}
	msgRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.AiChatMessage")).Return(nil)
	msgRepo.On("ListBySessionID", mock.Anything, uint64(5)).Return(history, nil)

	// Bedrock の擬似 stream を作る
	src := make(chan bedrock.StreamEvent, 4)
	src <- bedrock.StreamEvent{Delta: "Hi"}
	src <- bedrock.StreamEvent{Delta: " there"}
	src <- bedrock.StreamEvent{Done: true}
	close(src)
	var ro <-chan bedrock.StreamEvent = src
	bc.On("ConverseStream", mock.Anything, mock.AnythingOfType("string"), history).Return(ro, nil)

	uc := usecase.NewSendAiMessageStreamUseCase(sessionRepo, msgRepo, bc, nil)
	ch, err := uc.Execute(context.Background(), usecase.SendAiMessageInput{
		UserID:    7,
		SessionID: 5,
		Content:   "Hello",
	})
	require.NoError(t, err)

	var deltas []string
	var final *domain.AiChatMessage
	for ev := range ch {
		require.NoError(t, ev.Err)
		if ev.Delta != "" {
			deltas = append(deltas, ev.Delta)
		}
		if ev.FinalMessage != nil {
			final = ev.FinalMessage
		}
	}

	assert.Equal(t, []string{"Hi", " there"}, deltas)
	require.NotNil(t, final)
	assert.Equal(t, "Hi there", final.Content)
	assert.Equal(t, domain.AiChatRoleAssistant, final.Role)
	// ユーザー msg + アシスタント msg の 2 回保存
	msgRepo.AssertNumberOfCalls(t, "Save", 2)
}

// 新規セッションの場合: NewSession イベントが先に流れて、token + Final が続く。
func TestSendAiMessageStream_NewSession_EmitsSessionFirst(t *testing.T) {
	sessionRepo := new(mockSessionRepo)
	msgRepo := new(mockMsgRepo)
	bc := new(mockStreamBedrock)

	// セッション作成: ID を埋める
	sessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AiChatSession")).
		Run(func(args mock.Arguments) {
			s := args.Get(1).(*domain.AiChatSession)
			s.ID = 99
			s.CreatedAt = time.Now()
		}).Return(nil)
	msgRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.AiChatMessage")).Return(nil)
	msgRepo.On("ListBySessionID", mock.Anything, uint64(99)).Return([]domain.AiChatMessage{}, nil)

	src := make(chan bedrock.StreamEvent, 2)
	src <- bedrock.StreamEvent{Delta: "ok"}
	src <- bedrock.StreamEvent{Done: true}
	close(src)
	var ro <-chan bedrock.StreamEvent = src
	bc.On("ConverseStream", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(ro, nil)

	uc := usecase.NewSendAiMessageStreamUseCase(sessionRepo, msgRepo, bc, nil)
	ch, err := uc.Execute(context.Background(), usecase.SendAiMessageInput{
		UserID:  7,
		Content: "新しい会話を始めたい",
	})
	require.NoError(t, err)

	first := <-ch
	require.NotNil(t, first.NewSession)
	assert.Equal(t, uint64(99), first.NewSession.ID)

	// 残りはイベントが流れて FinalMessage で終わる
	for ev := range ch {
		require.NoError(t, ev.Err)
	}
}

// 添付付き発話: Downloader が S3 から bytes を取得し、Bedrock 呼び出し前に
// 履歴の最新 user メッセージへ BlobData が詰められる。
func TestSendAiMessageStream_WithAttachment_FetchesBlobAndSavesMetadata(t *testing.T) {
	sessionRepo := new(mockSessionRepo)
	msgRepo := new(mockMsgRepo)
	bc := new(mockStreamBedrock)
	dl := &fakeDownloader{
		bytes: map[string][]byte{
			"ai-chat/7/abc.png": []byte("fake-png-bytes"),
		},
	}

	// 履歴: 今回送ったユーザー発話だけ（attachments 付き）
	saved := &domain.AiChatMessage{
		SessionID:   5,
		MessageID:   "u1",
		Role:        domain.AiChatRoleUser,
		Content:     "この画像なに？",
		Attachments: []domain.Attachment{{Key: "ai-chat/7/abc.png", Format: "png", Kind: "image", ContentType: "image/png", SizeBytes: 14}},
		CreatedAt:   time.Now(),
	}
	msgRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.AiChatMessage")).Return(nil)
	msgRepo.On("ListBySessionID", mock.Anything, uint64(5)).Return([]domain.AiChatMessage{*saved}, nil)

	src := make(chan bedrock.StreamEvent, 2)
	src <- bedrock.StreamEvent{Delta: "猫です"}
	src <- bedrock.StreamEvent{Done: true}
	close(src)
	var ro <-chan bedrock.StreamEvent = src

	// Bedrock に渡る history の末尾要素に BlobData が詰まっていることを assert する。
	bc.On("ConverseStream", mock.Anything, mock.AnythingOfType("string"),
		mock.MatchedBy(func(history []domain.AiChatMessage) bool {
			if len(history) == 0 {
				return false
			}
			last := history[len(history)-1]
			if len(last.Attachments) != 1 {
				return false
			}
			return string(last.Attachments[0].BlobData) == "fake-png-bytes"
		})).Return(ro, nil)

	uc := usecase.NewSendAiMessageStreamUseCase(sessionRepo, msgRepo, bc, dl)
	ch, err := uc.Execute(context.Background(), usecase.SendAiMessageInput{
		UserID:    7,
		SessionID: 5,
		Content:   "この画像なに？",
		Attachments: []domain.Attachment{
			{Key: "ai-chat/7/abc.png", Format: "png", Kind: "image", ContentType: "image/png", SizeBytes: 14, Filename: "abc.png"},
		},
	})
	require.NoError(t, err)
	for ev := range ch {
		require.NoError(t, ev.Err)
	}
	bc.AssertExpectations(t)
}

// fakeDownloader は in-memory key→bytes map を返す簡易実装。
type fakeDownloader struct {
	bytes map[string][]byte
}

func (f *fakeDownloader) Download(_ context.Context, key string) ([]byte, error) {
	b, ok := f.bytes[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return b, nil
}

// Bedrock 呼び出しで stream エラー: ev.Err が channel に流れて使い手に伝わる。
func TestSendAiMessageStream_StreamError_PropagatesErr(t *testing.T) {
	sessionRepo := new(mockSessionRepo)
	msgRepo := new(mockMsgRepo)
	bc := new(mockStreamBedrock)

	msgRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.AiChatMessage")).Return(nil)
	msgRepo.On("ListBySessionID", mock.Anything, uint64(3)).Return([]domain.AiChatMessage{}, nil)

	src := make(chan bedrock.StreamEvent, 1)
	src <- bedrock.StreamEvent{Err: errors.New("bedrock down")}
	close(src)
	var ro <-chan bedrock.StreamEvent = src
	bc.On("ConverseStream", mock.Anything, mock.Anything, mock.Anything).Return(ro, nil)

	uc := usecase.NewSendAiMessageStreamUseCase(sessionRepo, msgRepo, bc, nil)
	ch, err := uc.Execute(context.Background(), usecase.SendAiMessageInput{
		UserID: 7, SessionID: 3, Content: "x",
	})
	require.NoError(t, err)

	var sawErr bool
	for ev := range ch {
		if ev.Err != nil {
			sawErr = true
		}
	}
	assert.True(t, sawErr, "expected ev.Err to surface to caller")
	// FinalMessage の Save は呼ばれない（エラー前 return）
	msgRepo.AssertNumberOfCalls(t, "Save", 1)
}
