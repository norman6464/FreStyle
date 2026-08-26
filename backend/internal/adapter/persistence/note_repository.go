package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// noteRepository は [repository.NoteRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type noteRepository struct{ db *sql.DB }

func NewNoteRepository(db *sql.DB) repository.NoteRepository { return &noteRepository{db: db} }

func toDomainNote(row sqlcgen.Note) domain.Note {
	return domain.Note{
		ID:        uint64(row.ID),
		UserID:    uint64(row.UserID),
		Title:     row.Title,
		Content:   row.Content,
		IsPublic:  row.IsPublic,
		IsPinned:  row.IsPinned,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func (r *noteRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.Note, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return []domain.Note{}, nil // 存在し得ない user_id = 0 件
	}
	rows, err := sqlcgen.New(r.db).ListNotesByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}
	notes := make([]domain.Note, 0, len(rows))
	for _, row := range rows {
		notes = append(notes, toDomainNote(row))
	}
	return notes, nil
}

// FindByID は id と user_id の両方で絞って 1 件返す。
// 他人の note はクエリが 0 行を返すので、存在しない id と同じ domain.ErrNotFound になる
// （「行は有るが持ち主が違う」という中間状態を呼び出し側へ渡さない）。
func (r *noteRepository) FindByID(ctx context.Context, userID, id uint64) (*domain.Note, error) {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil, domain.ErrNotFound // 存在し得ない id = not found
	}
	uid, ok := toInt64ID(userID)
	if !ok {
		return nil, domain.ErrNotFound // 存在し得ない user_id = 該当なし
	}
	row, err := sqlcgen.New(r.db).GetNoteByID(ctx, sqlcgen.GetNoteByIDParams{ID: id64, UserID: uid})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound // 404 シグナルを維持（他人の note もここに落ちる）
	}
	if err != nil {
		return nil, err
	}
	n := toDomainNote(row)
	return &n, nil
}

func (r *noteRepository) Create(ctx context.Context, n *domain.Note) error {
	uid, ok := toInt64ID(n.UserID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		return outOfRangeIDError("user_id", n.UserID)
	}
	now := time.Now()
	createdAt := n.CreatedAt
	if createdAt.IsZero() {
		createdAt = now // GORM autoCreateTime 相当（ゼロのときだけ now）
	}
	updatedAt := n.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now // GORM autoUpdateTime 相当（ゼロのときだけ now）
	}
	row, err := sqlcgen.New(r.db).InsertNote(ctx, sqlcgen.InsertNoteParams{
		UserID:    uid,
		Title:     n.Title,
		Content:   n.Content,
		IsPublic:  n.IsPublic,
		IsPinned:  n.IsPinned,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	})
	if err != nil {
		return err
	}
	n.ID = uint64(row.ID)
	n.CreatedAt = row.CreatedAt
	n.UpdatedAt = row.UpdatedAt
	return nil
}

// Update は id と user_id の両方で絞って更新する。
// 0 行更新（他人の note / 存在しない id）は sql.ErrNoRows になるので domain.ErrNotFound へ寄せ、
// 「更新できたのか何も起きなかったのか」を呼び出し側が取り違えないようにする。
func (r *noteRepository) Update(ctx context.Context, n *domain.Note) error {
	id64, ok := toInt64ID(n.ID)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = 対象なし
	}
	uid, ok := toInt64ID(n.UserID)
	if !ok {
		return domain.ErrNotFound // 存在し得ない user_id = 対象なし
	}
	updatedAt, err := sqlcgen.New(r.db).UpdateNote(ctx, sqlcgen.UpdateNoteParams{
		ID:       id64,
		UserID:   uid,
		Title:    n.Title,
		Content:  n.Content,
		IsPublic: n.IsPublic,
		IsPinned: n.IsPinned,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound // 他人の note もここに落ちる（存在しない id と同じ結末）
	}
	if err != nil {
		return err
	}
	n.UpdatedAt = updatedAt // GORM Save 相当の書き戻し
	return nil
}

func (r *noteRepository) Delete(ctx context.Context, userID, id uint64) error {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil // 存在し得ない id = 対象なし
	}
	uid, ok := toInt64ID(userID)
	if !ok {
		return nil // 存在し得ない user_id = 対象なし
	}
	return sqlcgen.New(r.db).DeleteNote(ctx, sqlcgen.DeleteNoteParams{
		ID:     id64,
		UserID: uid,
	})
}
