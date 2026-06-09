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

// sessionNoteRepository は [repository.SessionNoteRepository] の実装。
// 読み取り（FindBySessionID）は sqlc 生成コード（生 SQL）、書き込み（Upsert）は GORM。
type sessionNoteRepository struct{ db *gorm.DB }

func NewSessionNoteRepository(db *gorm.DB) repository.SessionNoteRepository {
	return &sessionNoteRepository{db: db}
}

func (r *sessionNoteRepository) FindBySessionID(ctx context.Context, sessionID uint64) (*domain.SessionNote, error) {
	sid, ok := toInt64ID(sessionID)
	if !ok {
		return nil, nil // 存在し得ない session_id = 未作成扱い
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	row, err := sqlcgen.New(sqlDB).GetSessionNoteBySessionID(ctx, sid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &domain.SessionNote{
		ID:        uint64(row.ID),
		SessionID: uint64(row.SessionID),
		UserID:    uint64(row.UserID),
		Content:   row.Content,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *sessionNoteRepository) Upsert(ctx context.Context, n *domain.SessionNote) error {
	return r.db.WithContext(ctx).Save(n).Error
}
