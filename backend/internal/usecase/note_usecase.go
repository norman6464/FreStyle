package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

var ErrNoteForbidden = errors.New("forbidden")

// ListNotesByUserIDUseCase は current user の ノート 一覧 を 返す。
// 依存 port: [repository.NoteRepository]。
type ListNotesByUserIDUseCase struct {
	repo repository.NoteRepository
}

func NewListNotesByUserIDUseCase(r repository.NoteRepository) *ListNotesByUserIDUseCase {
	return &ListNotesByUserIDUseCase{repo: r}
}

func (u *ListNotesByUserIDUseCase) Execute(ctx context.Context, userID uint64) ([]domain.Note, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListByUserID(ctx, userID)
}

// CreateNoteUseCase は 新規 ノート を 作成 する。
// 依存 port: [repository.NoteRepository]。
type CreateNoteUseCase struct {
	repo repository.NoteRepository
}

func NewCreateNoteUseCase(r repository.NoteRepository) *CreateNoteUseCase {
	return &CreateNoteUseCase{repo: r}
}

type CreateNoteInput struct {
	UserID   uint64
	Title    string
	Content  string
	IsPublic bool
	IsPinned bool
}

func (u *CreateNoteUseCase) Execute(ctx context.Context, in CreateNoteInput) (*domain.Note, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	if in.Title == "" {
		return nil, errors.New("title is required")
	}
	n := &domain.Note{
		UserID:   in.UserID,
		Title:    in.Title,
		Content:  in.Content,
		IsPublic: in.IsPublic,
		IsPinned: in.IsPinned,
	}
	if err := u.repo.Create(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

// UpdateNoteUseCase は ノート を 更新 する (所有者 検証 込み)。
// 依存 port: [repository.NoteRepository] (FindByID で 所有者 検証 → Update)。
type UpdateNoteUseCase struct {
	repo repository.NoteRepository
}

func NewUpdateNoteUseCase(r repository.NoteRepository) *UpdateNoteUseCase {
	return &UpdateNoteUseCase{repo: r}
}

type UpdateNoteInput struct {
	UserID   uint64
	ID       uint64
	Title    string
	Content  string
	IsPublic bool
	IsPinned bool
}

// Execute は所有者検証込みで note を更新する。
// 既存 note の UserID が input.UserID と一致しなければ ErrNoteForbidden を返す。
func (u *UpdateNoteUseCase) Execute(ctx context.Context, in UpdateNoteInput) (*domain.Note, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	if in.ID == 0 {
		return nil, errors.New("id is required")
	}
	existing, err := u.repo.FindByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if existing.UserID != in.UserID {
		return nil, ErrNoteForbidden
	}
	existing.Title = in.Title
	existing.Content = in.Content
	existing.IsPublic = in.IsPublic
	existing.IsPinned = in.IsPinned
	if err := u.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// DeleteNoteUseCase は ノート を 削除 する。 repo 側 で 所有者 検証 を 行う。
// 依存 port: [repository.NoteRepository] (Delete に userID を 渡して WHERE 絞り込み)。
type DeleteNoteUseCase struct {
	repo repository.NoteRepository
}

func NewDeleteNoteUseCase(r repository.NoteRepository) *DeleteNoteUseCase {
	return &DeleteNoteUseCase{repo: r}
}

func (u *DeleteNoteUseCase) Execute(ctx context.Context, userID, id uint64) error {
	if userID == 0 {
		return errors.New("userID is required")
	}
	if id == 0 {
		return errors.New("id is required")
	}
	return u.repo.Delete(ctx, userID, id)
}
