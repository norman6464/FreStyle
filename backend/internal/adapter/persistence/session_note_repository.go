package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// sessionNoteRepository は [repository.SessionNoteRepository] の実装。
// 読み書きとも sqlc 生成コード（生 SQL）。書き込みは session_id の一意制約に当てた upsert。
type sessionNoteRepository struct{ db *sql.DB }

func NewSessionNoteRepository(db *sql.DB) repository.SessionNoteRepository {
	return &sessionNoteRepository{db: db}
}

func (r *sessionNoteRepository) FindBySessionID(ctx context.Context, sessionID uint64) (*domain.SessionNote, error) {
	sid, ok := toInt64ID(sessionID)
	if !ok {
		return nil, nil // 存在し得ない session_id = 未作成扱い
	}
	row, err := sqlcgen.New(r.db).GetSessionNoteBySessionID(ctx, sid)
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
	sid, ok := toInt64ID(n.SessionID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		return outOfRangeIDError("session_id", n.SessionID)
	}
	uid, ok := toInt64ID(n.UserID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		return outOfRangeIDError("user_id", n.UserID)
	}
	// session_id の一意制約に当てて upsert する。初回は INSERT、以降は content /
	// updated_at のみ UPDATE。採番 id と時刻（作成時刻は保持・更新時刻は now()）を書き戻す。
	row, err := sqlcgen.New(r.db).UpsertSessionNote(ctx, sqlcgen.UpsertSessionNoteParams{
		SessionID: sid,
		UserID:    uid,
		Content:   n.Content,
	})
	// 他人が所有するセッションへ書こうとすると、クエリ側の WHERE で DO UPDATE が発火せず
	// 0 行になる（＝相手のメモは無傷のまま）。読み出し側が他人のメモを「無い」ものとして
	// 扱うのと同じ見え方に揃えるため、not-found に翻訳する。
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	n.ID = uint64(row.ID)
	n.CreatedAt = row.CreatedAt
	n.UpdatedAt = row.UpdatedAt
	return nil
}
