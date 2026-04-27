package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

// GetAiChatSessionsByUserIDUseCase は指定ユーザーの AI チャットセッション一覧を返す。
type GetAiChatSessionsByUserIDUseCase struct {
	sessions repository.AiChatSessionRepository
}

func NewGetAiChatSessionsByUserIDUseCase(s repository.AiChatSessionRepository) *GetAiChatSessionsByUserIDUseCase {
	return &GetAiChatSessionsByUserIDUseCase{sessions: s}
}

func (u *GetAiChatSessionsByUserIDUseCase) Execute(ctx context.Context, userID uint64) ([]domain.AiChatSession, error) {
	return u.sessions.ListByUserID(ctx, userID)
}

// CreateAiChatSessionUseCase は新規セッションを作成する。
type CreateAiChatSessionUseCase struct {
	sessions repository.AiChatSessionRepository
}

func NewCreateAiChatSessionUseCase(s repository.AiChatSessionRepository) *CreateAiChatSessionUseCase {
	return &CreateAiChatSessionUseCase{sessions: s}
}

type CreateAiChatSessionInput struct {
	UserID      uint64
	Title       string
	SessionType string
	ScenarioID  *uint64
}

func (u *CreateAiChatSessionUseCase) Execute(ctx context.Context, in CreateAiChatSessionInput) (*domain.AiChatSession, error) {
	if in.Title == "" {
		return nil, errors.New("title is required")
	}
	if in.SessionType == "" {
		in.SessionType = domain.AiChatSessionTypeFree
	}
	s := &domain.AiChatSession{
		UserID:      in.UserID,
		Title:       in.Title,
		SessionType: in.SessionType,
		ScenarioID:  in.ScenarioID,
	}
	if err := u.sessions.Create(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}
