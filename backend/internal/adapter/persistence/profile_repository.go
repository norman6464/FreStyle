package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/norman6464/frestyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// profileRepository は [repository.ProfileRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type profileRepository struct{ baseRepository }

func NewProfileRepository(db *sql.DB) repository.ProfileRepository {
	return &profileRepository{baseRepository{db: db}}
}

func (r *profileRepository) FindByUserID(ctx context.Context, userID uint64) (*domain.Profile, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return nil, nil // 存在し得ない user_id = 未作成扱い
	}
	row, err := sqlcgen.New(r.dbtx(ctx)).GetProfileByUserID(ctx, uid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // 未作成は (nil, nil)。usecase が空表示にフォールバックする
	}
	if err != nil {
		return nil, err
	}
	return toDomainProfile(row), nil
}

func (r *profileRepository) Upsert(ctx context.Context, p *domain.Profile) error {
	uid, ok := toInt64ID(p.UserID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		return outOfRangeIDError("user_id", p.UserID)
	}
	updatedAt, err := sqlcgen.New(r.dbtx(ctx)).UpsertProfile(ctx, sqlcgen.UpsertProfileParams{
		UserID:     uid,
		Bio:        p.Bio,
		AvatarUrl:  p.AvatarURL,
		StatusText: p.StatusText,
	})
	if err != nil {
		return err
	}
	p.UpdatedAt = updatedAt
	return nil
}

func (r *profileRepository) UpdateStatus(ctx context.Context, userID uint64, emoji, text string, expiresAt *time.Time) (*domain.Profile, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return nil, outOfRangeIDError("user_id", userID)
	}
	row, err := sqlcgen.New(r.dbtx(ctx)).UpsertProfileStatus(ctx, sqlcgen.UpsertProfileStatusParams{
		UserID:          uid,
		StatusEmoji:     emoji,
		StatusText:      text,
		StatusExpiresAt: nullTime(expiresAt),
	})
	if err != nil {
		return nil, err
	}
	return toDomainProfile(row), nil
}

func toDomainProfile(row sqlcgen.Profile) *domain.Profile {
	p := &domain.Profile{
		UserID:     uint64(row.UserID),
		Bio:        row.Bio,
		AvatarURL:  row.AvatarUrl,
		StatusText: row.StatusText,
		UpdatedAt:  row.UpdatedAt,
	}
	if row.StatusExpiresAt.Valid {
		t := row.StatusExpiresAt.Time
		p.StatusExpiresAt = &t
	}
	p.StatusEmoji, p.StatusText = domain.ClearExpiredStatus(row.StatusEmoji, row.StatusText, p.StatusExpiresAt, time.Now())
	return p
}
