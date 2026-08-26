package persistence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// notificationRepository は [repository.NotificationRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、GORM からは接続プール（*sql.DB）だけを借りる。
type notificationRepository struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	uid, ok := toInt64ID(n.UserID)
	if !ok {
		return nil // 存在し得ない user_id は書き込まない
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	createdAt := n.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now() // GORM autoCreateTime 相当（ゼロのときだけ now）
	}
	row, err := sqlcgen.New(sqlDB).InsertNotification(ctx, sqlcgen.InsertNotificationParams{
		UserID:    uid,
		Type:      n.Type,
		Title:     n.Title,
		Body:      n.Body,
		IsRead:    n.IsRead,
		CreatedAt: createdAt,
	})
	if err != nil {
		return err
	}
	n.ID = uint64(row.ID)
	n.CreatedAt = row.CreatedAt
	return nil
}

// createManyItem は json_to_recordset に渡す 1 行分。キー名は SQL 側の列名と一致させる。
type createManyItem struct {
	UserID uint64 `json:"user_id"`
	Type   string `json:"type"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	IsRead bool   `json:"is_read"`
}

// CreateMany は複数件を 1 回の INSERT でまとめて作成する。
// 宛先が 1 人増えるごとに DB との往復が 1 回増える形だと件数に比例して遅くなるため、
// json 配列 1 個を渡して json_to_recordset で展開し 1 文にまとめる。
func (r *notificationRepository) CreateMany(ctx context.Context, ns []domain.Notification) error {
	if len(ns) == 0 {
		return nil // 宛先 0 件は何もしない（呼び出し側で件数を気にしなくてよいように）
	}
	items := make([]createManyItem, 0, len(ns))
	for _, n := range ns {
		items = append(items, createManyItem{
			UserID: n.UserID,
			Type:   n.Type,
			Title:  n.Title,
			Body:   n.Body,
			IsRead: n.IsRead,
		})
	}
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return err
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlcgen.New(sqlDB).CreateNotifications(ctx, itemsJSON)
}

func (r *notificationRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.Notification, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return []domain.Notification{}, nil // 存在し得ない user_id = 0 件
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListNotificationsByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Notification, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.Notification{
			ID:        uint64(row.ID),
			UserID:    uint64(row.UserID),
			Type:      row.Type,
			Title:     row.Title,
			Body:      row.Body,
			IsRead:    row.IsRead,
			CreatedAt: row.CreatedAt,
		})
	}
	return out, nil
}

func (r *notificationRepository) MarkRead(ctx context.Context, userID, id uint64) error {
	uid, ok := toInt64ID(userID)
	if !ok {
		return nil // 存在し得ない user_id = 対象なし
	}
	nid, ok := toInt64ID(id)
	if !ok {
		return nil // 存在し得ない id = 対象なし
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlcgen.New(sqlDB).MarkNotificationRead(ctx, sqlcgen.MarkNotificationReadParams{
		ID:     nid,
		UserID: uid,
	})
}

func (r *notificationRepository) MarkAllRead(ctx context.Context, userID uint64) error {
	uid, ok := toInt64ID(userID)
	if !ok {
		return nil // 存在し得ない user_id = 対象なし
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlcgen.New(sqlDB).MarkAllNotificationsRead(ctx, uid)
}

func (r *notificationRepository) CountUnread(ctx context.Context, userID uint64) (int64, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return 0, nil
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return 0, err
	}
	return sqlcgen.New(sqlDB).CountUnreadNotifications(ctx, uid)
}

// stubSnsPublisher は [repository.SnsPublisher] の no-op 実装（本番の SNS 実装は別 PR）。
type stubSnsPublisher struct{}

func NewStubSnsPublisher() repository.SnsPublisher { return &stubSnsPublisher{} }

func (p *stubSnsPublisher) Publish(_ context.Context, _ uint64, _, _ string) error { return nil }
