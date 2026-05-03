package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
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

// --- mock: BedrockClient ---

type mockBedrock struct{ mock.Mock }

func (m *mockBedrock) Converse(ctx context.Context, systemPrompt string, history []domain.AiChatMessage) (string, error) {
	args := m.Called(ctx, systemPrompt, history)
	return args.String(0), args.Error(1)
}

func TestSendAiMessageUseCase_NewSession(t *testing.T) {
	sessionRepo := new(mockSessionRepo)
	msgRepo := new(mockMsgRepo)
	bc := new(mockBedrock)

	sessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AiChatSession")).
		Run(func(args mock.Arguments) {
			s := args.Get(1).(*domain.AiChatSession)
			s.ID = 1
			s.CreatedAt = time.Now()
		}).Return(nil)
	history := []domain.AiChatMessage{
		{SessionID: 1, MessageID: "u1", Role: domain.AiChatRoleUser, Content: "こんにちは", CreatedAt: time.Now()},
	}
	msgRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.AiChatMessage")).Return(nil)
	msgRepo.On("ListBySessionID", mock.Anything, uint64(1)).Return(history, nil)
	bc.On("Converse", mock.Anything, mock.AnythingOfType("string"), history).Return("こんにちは！何かお手伝いできますか？", nil)

	uc := usecase.NewSendAiMessageUseCase(sessionRepo, msgRepo, bc)
	out, err := uc.Execute(context.Background(), usecase.SendAiMessageInput{
		UserID:    42,
		SessionID: 0,
		Content:   "こんにちは",
	})

	require.NoError(t, err)
	require.NotNil(t, out.NewSession, "新規セッションが返されること")
	assert.Equal(t, uint64(1), out.NewSession.ID)
	assert.Equal(t, "こんにちは", out.UserMsg.Content)
	assert.Equal(t, domain.AiChatRoleUser, out.UserMsg.Role)
	assert.Equal(t, "こんにちは！何かお手伝いできますか？", out.AiMsg.Content)
	assert.Equal(t, domain.AiChatRoleAssistant, out.AiMsg.Role)
}

func TestSendAiMessageUseCase_ExistingSession(t *testing.T) {
	sessionRepo := new(mockSessionRepo)
	msgRepo := new(mockMsgRepo)
	bc := new(mockBedrock)

	history := []domain.AiChatMessage{
		{SessionID: 5, MessageID: "u1", Role: domain.AiChatRoleUser, Content: "以前の発話", CreatedAt: time.Now()},
		{SessionID: 5, MessageID: "a1", Role: domain.AiChatRoleAssistant, Content: "以前の返答", CreatedAt: time.Now()},
		{SessionID: 5, MessageID: "u2", Role: domain.AiChatRoleUser, Content: "続きの質問", CreatedAt: time.Now()},
	}
	msgRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.AiChatMessage")).Return(nil)
	msgRepo.On("ListBySessionID", mock.Anything, uint64(5)).Return(history, nil)
	bc.On("Converse", mock.Anything, mock.AnythingOfType("string"), history).Return("続きの返答", nil)

	uc := usecase.NewSendAiMessageUseCase(sessionRepo, msgRepo, bc)
	out, err := uc.Execute(context.Background(), usecase.SendAiMessageInput{
		UserID:    42,
		SessionID: 5,
		Content:   "続きの質問",
	})

	require.NoError(t, err)
	assert.Nil(t, out.NewSession, "既存セッション時は新規セッションなし")
	assert.Equal(t, uint64(5), out.AiMsg.SessionID)
	assert.Equal(t, "続きの返答", out.AiMsg.Content)
}

func TestSendAiMessageUseCase_TitleTruncation(t *testing.T) {
	// 30文字を超えるコンテンツは切り詰められること
	long := "あいうえおかきくけこさしすせそたちつてとなにぬねのはひふへほまみむ" // 32文字
	sessionRepo := new(mockSessionRepo)
	msgRepo := new(mockMsgRepo)
	bc := new(mockBedrock)

	var capturedTitle string
	sessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AiChatSession")).
		Run(func(args mock.Arguments) {
			s := args.Get(1).(*domain.AiChatSession)
			capturedTitle = s.Title
			s.ID = 2
			s.CreatedAt = time.Now()
		}).Return(nil)
	msgRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.AiChatMessage")).Return(nil)
	msgRepo.On("ListBySessionID", mock.Anything, uint64(2)).Return([]domain.AiChatMessage{
		{SessionID: 2, MessageID: "u1", Role: domain.AiChatRoleUser, Content: long, CreatedAt: time.Now()},
	}, nil)
	bc.On("Converse", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return("返答", nil)

	uc := usecase.NewSendAiMessageUseCase(sessionRepo, msgRepo, bc)
	_, err := uc.Execute(context.Background(), usecase.SendAiMessageInput{
		UserID: 1, SessionID: 0, Content: long,
	})

	require.NoError(t, err)
	assert.LessOrEqual(t, len([]rune(capturedTitle)), 31, "タイトルは30文字＋省略記号以下")
}
