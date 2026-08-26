package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// noteRepository は [repository.NoteRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、GORM からは接続プール（*sql.DB）だけを借りる。
type noteRepository struct{ db *gorm.DB }

func NewNoteRepository(db *gorm.DB) repository.NoteRepository { return &noteRepository{db: db} }

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
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListNotesByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}
	notes := make([]domain.Note, 0, len(rows))
	for _, row := range rows {
		notes = append(notes, toDomainNote(row))
	}
	return notes, nil
}

func (r *noteRepository) FindByID(ctx context.Context, id uint64) (*domain.Note, error) {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil, domain.ErrNotFound // 存在し得ない id = not found
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	row, err := sqlcgen.New(sqlDB).GetNoteByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound // 404 シグナルを維持
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
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
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
	row, err := sqlcgen.New(sqlDB).InsertNote(ctx, sqlcgen.InsertNoteParams{
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

func (r *noteRepository) Update(ctx context.Context, n *domain.Note) error {
	id64, ok := toInt64ID(n.ID)
	if !ok {
		return nil // 存在し得ない id = 対象なし
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	updatedAt, err := sqlcgen.New(sqlDB).UpdateNote(ctx, sqlcgen.UpdateNoteParams{
		ID:       id64,
		Title:    n.Title,
		Content:  n.Content,
		IsPublic: n.IsPublic,
		IsPinned: n.IsPinned,
	})
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
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlcgen.New(sqlDB).DeleteNote(ctx, sqlcgen.DeleteNoteParams{
		ID:     id64,
		UserID: uid,
	})
}
