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

// profileRepository は [repository.ProfileRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、GORM からは接続プール（*sql.DB）だけを借りる。
type profileRepository struct{ db *gorm.DB }

func NewProfileRepository(db *gorm.DB) repository.ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) FindByUserID(ctx context.Context, userID uint64) (*domain.Profile, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return nil, nil // 存在し得ない user_id = 未作成扱い
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	row, err := sqlcgen.New(sqlDB).GetProfileByUserID(ctx, uid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // 未作成は (nil, nil)。usecase が空表示にフォールバックする
	}
	if err != nil {
		return nil, err
	}
	return &domain.Profile{
		UserID:        uint64(row.UserID),
		Bio:           row.Bio,
		AvatarURL:     row.AvatarUrl,
		StatusMessage: row.StatusMessage,
		UpdatedAt:     row.UpdatedAt,
	}, nil
}

func (r *profileRepository) Upsert(ctx context.Context, p *domain.Profile) error {
	uid, ok := toInt64ID(p.UserID)
	if !ok {
		return nil // 存在し得ない user_id は書き込まない
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	updatedAt, err := sqlcgen.New(sqlDB).UpsertProfile(ctx, sqlcgen.UpsertProfileParams{
		UserID:        uid,
		Bio:           p.Bio,
		AvatarUrl:     p.AvatarURL,
		StatusMessage: p.StatusMessage,
	})
	if err != nil {
		return err
	}
	p.UpdatedAt = updatedAt // GORM Save 相当の書き戻し
	return nil
}
