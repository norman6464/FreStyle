package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// aiChatSessionRepository は [repository.AiChatSessionRepository] の実装。
// 読み書きとも sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type aiChatSessionRepository struct{ db *sql.DB }

func NewAiChatSessionRepository(db *sql.DB) repository.AiChatSessionRepository {
	return &aiChatSessionRepository{db: db}
}

// toDomainAiChatSession は sqlc 生成モデル → domain への詰め替え。
func toDomainAiChatSession(row sqlcgen.AiChatSession) domain.AiChatSession {
	s := domain.AiChatSession{
		ID:          uint64(row.ID),
		UserID:      uint64(row.UserID),
		Title:       row.Title,
		SessionType: row.SessionType,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.ScenarioID.Valid {
		id := uint64(row.ScenarioID.Int64)
		s.ScenarioID = &id
	}
	return s
}

// ListByUserID は自分のセッションを新しい順で返す。
func (r *aiChatSessionRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.AiChatSession, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return make([]domain.AiChatSession, 0), nil // 存在し得ない user_id = 0 件
	}
	rows, err := sqlcgen.New(r.db).ListAiChatSessionsByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]domain.AiChatSession, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainAiChatSession(row))
	}
	return out, nil
}

// FindByID は単一セッションを返す。未存在は domain.ErrNotFound（handler が 404 に分岐）。
func (r *aiChatSessionRepository) FindByID(ctx context.Context, id uint64) (*domain.AiChatSession, error) {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil, domain.ErrNotFound // 存在し得ない id = not found
	}
	row, err := sqlcgen.New(r.db).GetAiChatSessionByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound // 404 シグナルを維持
	}
	if err != nil {
		return nil, err
	}
	s := toDomainAiChatSession(row)
	return &s, nil
}

// Create はセッションを 1 件作成し、採番 id と時刻を引数の構造体へ書き戻す。
func (r *aiChatSessionRepository) Create(ctx context.Context, s *domain.AiChatSession) error {
	uid, ok := toInt64ID(s.UserID)
	if !ok {
		// 何も書かずに nil を返すと「保存できた」と誤認される。書けないことをエラーで伝える。
		return fmt.Errorf("user id %d が int64 の範囲外です", s.UserID)
	}
	var scenario sql.NullInt64
	if s.ScenarioID != nil {
		sid, ok := toInt64ID(*s.ScenarioID)
		if !ok {
			return fmt.Errorf("scenario id %d が int64 の範囲外です", *s.ScenarioID)
		}
		scenario = sql.NullInt64{Int64: sid, Valid: true}
	}
	now := time.Now()
	createdAt := s.CreatedAt
	if createdAt.IsZero() {
		createdAt = now // GORM autoCreateTime 相当（ゼロのときだけ now）
	}
	updatedAt := s.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now // GORM autoUpdateTime 相当（ゼロのときだけ now）
	}
	row, err := sqlcgen.New(r.db).InsertAiChatSession(ctx, sqlcgen.InsertAiChatSessionParams{
		UserID:      uid,
		Title:       s.Title,
		SessionType: s.SessionType,
		ScenarioID:  scenario,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	})
	if err != nil {
		return err
	}
	s.ID = uint64(row.ID)
	s.CreatedAt = row.CreatedAt
	s.UpdatedAt = row.UpdatedAt
	return nil
}

// UpdateTitle はタイトルだけを更新する（updated_at は now() へ進む）。
// 該当行が無くてもエラーにしない（GORM 版と同じ契約）。
func (r *aiChatSessionRepository) UpdateTitle(ctx context.Context, id uint64, title string) error {
	id64, ok := toInt64ID(id)
	if !ok {
		// 0 行更新（該当なし）と区別が付かない nil を返さず、書けないことをエラーで伝える。
		return fmt.Errorf("session id %d が int64 の範囲外です", id)
	}
	return sqlcgen.New(r.db).UpdateAiChatSessionTitle(ctx, sqlcgen.UpdateAiChatSessionTitleParams{
		ID:    id64,
		Title: title,
	})
}

// Delete はセッションを物理削除する。該当行が無くてもエラーにしない（GORM 版と同じ契約）。
func (r *aiChatSessionRepository) Delete(ctx context.Context, id uint64) error {
	id64, ok := toInt64ID(id)
	if !ok {
		// 0 行削除（該当なし）と区別が付かない nil を返さず、消せないことをエラーで伝える。
		return fmt.Errorf("session id %d が int64 の範囲外です", id)
	}
	return sqlcgen.New(r.db).DeleteAiChatSession(ctx, id64)
}
