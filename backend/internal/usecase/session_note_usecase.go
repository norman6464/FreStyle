package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetSessionNoteUseCase は AI チャットセッション紐付きのノートを取得する。
type GetSessionNoteUseCase struct {
	repo repository.SessionNoteRepository
}

func NewGetSessionNoteUseCase(r repository.SessionNoteRepository) *GetSessionNoteUseCase {
	return &GetSessionNoteUseCase{repo: r}
}

// GetSessionNoteInput は取得対象と、それを要求している利用者を表す。
// UserID は所有者検証に使う（他人のノートを sessionID の総当たりで読めないようにする）。
type GetSessionNoteInput struct {
	SessionID uint64
	UserID    uint64
}

// Execute は所有者本人のノートのみ返す。
// 他人のノートは「存在しない」と同じ扱い（nil, nil）にして、
// そのセッションにノートがあること自体を漏らさない。
func (u *GetSessionNoteUseCase) Execute(ctx context.Context, in GetSessionNoteInput) (*domain.SessionNote, error) {
	if in.SessionID == 0 || in.UserID == 0 {
		return nil, errors.New("sessionID and userID are required")
	}
	n, err := u.repo.FindBySessionID(ctx, in.SessionID)
	if err != nil || n == nil {
		return nil, err
	}
	if n.UserID != in.UserID {
		return nil, nil
	}
	return n, nil
}

// UpsertSessionNoteUseCase はセッションノートを upsert する。
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
