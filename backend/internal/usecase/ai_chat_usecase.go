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

// GetAiChatSessionUseCase は単一セッションを返す。
type GetAiChatSessionUseCase struct {
	sessions repository.AiChatSessionRepository
}

func NewGetAiChatSessionUseCase(s repository.AiChatSessionRepository) *GetAiChatSessionUseCase {
	return &GetAiChatSessionUseCase{sessions: s}
}

func (u *GetAiChatSessionUseCase) Execute(ctx context.Context, id uint64) (*domain.AiChatSession, error) {
	return u.sessions.FindByID(ctx, id)
}

// UpdateAiChatSessionTitleUseCase はセッションタイトルを更新する。
type UpdateAiChatSessionTitleUseCase struct {
	sessions repository.AiChatSessionRepository
}

func NewUpdateAiChatSessionTitleUseCase(s repository.AiChatSessionRepository) *UpdateAiChatSessionTitleUseCase {
	return &UpdateAiChatSessionTitleUseCase{sessions: s}
}

func (u *UpdateAiChatSessionTitleUseCase) Execute(ctx context.Context, id uint64, title string) error {
	if title == "" {
		return errors.New("title is required")
	}
	return u.sessions.UpdateTitle(ctx, id, title)
}

// DeleteAiChatSessionUseCase はセッションを削除する。
type DeleteAiChatSessionUseCase struct {
	sessions repository.AiChatSessionRepository
}

func NewDeleteAiChatSessionUseCase(s repository.AiChatSessionRepository) *DeleteAiChatSessionUseCase {
	return &DeleteAiChatSessionUseCase{sessions: s}
}

func (u *DeleteAiChatSessionUseCase) Execute(ctx context.Context, id uint64, userID uint64) error {
	s, err := u.sessions.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if s.UserID != userID {
		return errors.New("forbidden")
	}
	return u.sessions.Delete(ctx, id)
}

// GetAiChatMessagesUseCase は DynamoDB からセッションのメッセージ一覧を返す。
type GetAiChatMessagesUseCase struct {
	messages repository.AiChatMessageRepository
}

func NewGetAiChatMessagesUseCase(m repository.AiChatMessageRepository) *GetAiChatMessagesUseCase {
	return &GetAiChatMessagesUseCase{messages: m}
}

func (u *GetAiChatMessagesUseCase) Execute(ctx context.Context, sessionID uint64) ([]domain.AiChatMessage, error) {
	if u.messages == nil {
		return nil, errors.New("message repository unavailable")
	}
	return u.messages.ListBySessionID(ctx, sessionID)
}
