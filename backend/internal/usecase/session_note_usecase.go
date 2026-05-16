package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetSessionNoteUseCase は AI チャット セッション 紐付き の ノート を 取得 する。
// 依存 port: [repository.SessionNoteRepository]。
type GetSessionNoteUseCase struct {
	repo repository.SessionNoteRepository
}

func NewGetSessionNoteUseCase(r repository.SessionNoteRepository) *GetSessionNoteUseCase {
	return &GetSessionNoteUseCase{repo: r}
}

func (u *GetSessionNoteUseCase) Execute(ctx context.Context, sessionID uint64) (*domain.SessionNote, error) {
	if sessionID == 0 {
		return nil, errors.New("sessionID is required")
	}
	return u.repo.FindBySessionID(ctx, sessionID)
}

// UpsertSessionNoteUseCase は セッション ノート を upsert する (新規 or 更新)。
// 依存 port: [repository.SessionNoteRepository]。
type UpsertSessionNoteUseCase struct {
	repo repository.SessionNoteRepository
}

func NewUpsertSessionNoteUseCase(r repository.SessionNoteRepository) *UpsertSessionNoteUseCase {
	return &UpsertSessionNoteUseCase{repo: r}
}

type UpsertSessionNoteInput struct {
	SessionID uint64
	UserID    uint64
	Content   string
}

func (u *UpsertSessionNoteUseCase) Execute(ctx context.Context, in UpsertSessionNoteInput) (*domain.SessionNote, error) {
	if in.SessionID == 0 || in.UserID == 0 {
		return nil, errors.New("sessionID and userID are required")
	}
	n := &domain.SessionNote{SessionID: in.SessionID, UserID: in.UserID, Content: in.Content}
	if err := u.repo.Upsert(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}
