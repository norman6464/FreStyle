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
// 対象行が無ければ domain.ErrNotFound を返す（handler が 404 にマップ）。
//
// 0 行更新を成功にしてはいけない理由:
//
//	UPDATE は 1 行も一致しなくても SQL としては成功する。ここで nil を返すと handler は
//	そのまま 200 を返し、利用者にはタイトルが変わったように見えるのに DB には何も
//	書かれていない。行が無いことは「保存できた」ではなく「対象が無い」なので 404 で伝える。
func (r *aiChatSessionRepository) UpdateTitle(ctx context.Context, id uint64, title string) error {
	id64, ok := toInt64ID(id)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = 対象なし
	}
	// :execrows なので実際に書き換わった行数が返る（:exec だと 0 行でも成功と区別が付かない）。
	affected, err := sqlcgen.New(r.db).UpdateAiChatSessionTitle(ctx, sqlcgen.UpdateAiChatSessionTitleParams{
		ID:    id64,
		Title: title,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete はセッションを物理削除する。対象行が無ければ domain.ErrNotFound を返す。
//
// DELETE でも 0 行を成功にしない理由:
//
//	「消えている」という事後条件だけなら 0 行削除も満たしている。それでも not-found を
//	返すのは、この経路が「自分の履歴一覧から 1 件選んで消す」操作で、呼び出し側
//	（DeleteAiChatSessionUseCase）が FindByID で所有者を先に確かめているから。
//	所有者確認を通ったのに 0 行というのは「確認と削除のあいだにセッションが消えた」競合で、
//	成功として返すと消したのが自分の操作なのか競合相手なのかが区別できなくなる。
func (r *aiChatSessionRepository) Delete(ctx context.Context, id uint64) error {
	id64, ok := toInt64ID(id)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = 対象なし
	}
	affected, err := sqlcgen.New(r.db).DeleteAiChatSession(ctx, id64)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
