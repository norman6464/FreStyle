package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ListNotesByUserIDUseCase は current user のノート一覧を返す。
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

// CreateNoteUseCase は新規ノートを作成する。
// 作成後に user_daily_activities をベストエフォートでインクリメントする。
type CreateNoteUseCase struct {
	repo     repository.NoteRepository
	activity repository.UserDailyActivityRepository
}

func NewCreateNoteUseCase(r repository.NoteRepository, activity repository.UserDailyActivityRepository) *CreateNoteUseCase {
	return &CreateNoteUseCase{repo: r, activity: activity}
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
	if err := u.activity.Increment(ctx, in.UserID, time.Now().UTC(), repository.UserDailyActivityIncrement{
		NoteCount: 1,
	}); err != nil {
		slog.WarnContext(ctx, "user_daily_activities increment failed", "userID", in.UserID, "err", err)
	}
	return n, nil
}

// UpdateNoteUseCase はノートを更新する（所有者検証込み）。
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

// Execute は current user 所有の note だけを更新する。
//
// 「他人の note」と「存在しない note」はどちらも domain.ErrNotFound に畳んで返す。
// ここを撃ち分けると、存在オラクル（どの ID が実在するかを外から数え上げられる状態）になるため。
// notes.id は連番（bigserial）なので、ログイン済みなら誰でも 1, 2, 3 … と ID を順に叩ける。
// このとき「他人の note なので拒否」と「そんな note は無い」を別のエラーで返すと、
// 呼び出し元はレスポンスの違いだけで「この ID は実在する（＝誰かが書いた）」と判定でき、
// 本文を一切読めなくても、社内に何件の note があり ID 空間のどこが埋まっているかを全数把握できる。
// 存在の有無そのものが他人の情報なので、区別できないように 1 本の番兵へ寄せる。
//
// 同じ畳み方は session_notes（GetSessionNoteUseCase）が既に採っており、notes もそれに揃える。
func (u *UpdateNoteUseCase) Execute(ctx context.Context, in UpdateNoteInput) (*domain.Note, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	if in.ID == 0 {
		return nil, errors.New("id is required")
	}
	// repository は WHERE user_id で絞るので、他人の note はここで domain.ErrNotFound になる
	// （SQL・usecase・handler の 3 層で同じ結末に寄せる。どれか 1 つが将来外れても漏れない）。
	existing, err := u.repo.FindByID(ctx, in.UserID, in.ID)
	if err != nil {
		return nil, err
	}
	if existing.UserID != in.UserID {
		// 「無い」と同じ番兵にする。errors.New("forbidden") のような別エラーにすると、
		// handler がそれを 403 へ振り分けた時点で上記の撃ち分けが復活する。
		return nil, domain.ErrNotFound
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

// DeleteNoteUseCase はノートを削除する（repo 側で userID により所有者検証）。
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
