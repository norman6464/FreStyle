package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// noteRepository は [repository.NoteRepository] の実装。
// 読み取りは sqlc 生成コード（生 SQL）、書き込みは GORM（autoTime / 採番）を使う。
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
		return nil, gorm.ErrRecordNotFound // 存在し得ない id = not found
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	row, err := sqlcgen.New(sqlDB).GetNoteByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, gorm.ErrRecordNotFound // 404 シグナルを維持
	}
	if err != nil {
		return nil, err
	}
	n := toDomainNote(row)
	return &n, nil
}

func (r *noteRepository) Create(ctx context.Context, n *domain.Note) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *noteRepository) Update(ctx context.Context, n *domain.Note) error {
	return r.db.WithContext(ctx).Save(n).Error
}

func (r *noteRepository) Delete(ctx context.Context, userID, id uint64) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&domain.Note{}).Error
}
