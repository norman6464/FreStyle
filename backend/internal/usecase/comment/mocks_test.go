package comment_test

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/mock"
)

// usecase テストで共有する repository interface の testify/mock 実装。
// 書き方の見本は internal/usecase/kb/mocks_test.go と同じ流儀。

// --- mock: CommentRepository ---

type mockCommentRepo struct{ mock.Mock }

var _ repository.CommentRepository = (*mockCommentRepo)(nil)

func (m *mockCommentRepo) CreateCommentThread(
	ctx context.Context, workspaceID, pageID string, createdByUserID uint64, anchor repository.CommentAnchor,
) (*domain.CommentThread, error) {
	args := m.Called(ctx, workspaceID, pageID, createdByUserID, anchor)
	t, _ := args.Get(0).(*domain.CommentThread)
	return t, args.Error(1)
}

func (m *mockCommentRepo) BlockExistsInPage(ctx context.Context, workspaceID, pageID, blockID string) (bool, error) {
	args := m.Called(ctx, workspaceID, pageID, blockID)
	return args.Bool(0), args.Error(1)
}

func (m *mockCommentRepo) CreateComment(
	ctx context.Context, threadID string, authorUserID uint64, body string,
) (*domain.Comment, error) {
	args := m.Called(ctx, threadID, authorUserID, body)
	c, _ := args.Get(0).(*domain.Comment)
	return c, args.Error(1)
}

func (m *mockCommentRepo) GetCommentThread(ctx context.Context, workspaceID, pageID, threadID string) (*domain.CommentThread, error) {
	args := m.Called(ctx, workspaceID, pageID, threadID)
	t, _ := args.Get(0).(*domain.CommentThread)
	return t, args.Error(1)
}

func (m *mockCommentRepo) ListCommentThreadsByPage(ctx context.Context, workspaceID, pageID string) ([]domain.CommentThread, error) {
	args := m.Called(ctx, workspaceID, pageID)
	rows, _ := args.Get(0).([]domain.CommentThread)
	return rows, args.Error(1)
}

func (m *mockCommentRepo) ListCommentsByThreads(ctx context.Context, threadIDs []string) ([]domain.Comment, error) {
	args := m.Called(ctx, threadIDs)
	rows, _ := args.Get(0).([]domain.Comment)
	return rows, args.Error(1)
}

func (m *mockCommentRepo) ResolveCommentThread(
	ctx context.Context, workspaceID, pageID, threadID string, resolvedByUserID uint64,
) (*domain.CommentThread, error) {
	args := m.Called(ctx, workspaceID, pageID, threadID, resolvedByUserID)
	t, _ := args.Get(0).(*domain.CommentThread)
	return t, args.Error(1)
}

func (m *mockCommentRepo) ReopenCommentThread(ctx context.Context, workspaceID, pageID, threadID string) (*domain.CommentThread, error) {
	args := m.Called(ctx, workspaceID, pageID, threadID)
	t, _ := args.Get(0).(*domain.CommentThread)
	return t, args.Error(1)
}

// --- fake: TxManager ---

// txMarkerKey は fakeTxManager が DoInTx の中で ctx に埋め込む印。
// 本物の *sql.Tx を持たないテストで「repository がトランザクションの中で呼ばれたか」を
// mock.MatchedBy(inTx) で確かめられるようにするためだけの値。
type txMarkerKeyType struct{}

var txMarkerKey = txMarkerKeyType{}

// inTx は mock.MatchedBy に渡す述語。ctx に txMarkerKey が乗っていれば
// fakeTxManager.DoInTx の fn の中（＝トランザクションの中）から呼ばれたことを意味する。
func inTx(ctx context.Context) bool {
	v, _ := ctx.Value(txMarkerKey).(bool)
	return v
}

// fakeTxManager は repository.TxManager のテスト用実装。実 DB もトランザクションも
// 介さず fn(ctx) をそのまま呼ぶが、呼び出し回数と「tx の中で呼んだ」印だけは残す。
type fakeTxManager struct {
	calls int
}

var _ repository.TxManager = (*fakeTxManager)(nil)

func (f *fakeTxManager) DoInTx(ctx context.Context, fn func(context.Context) error) error {
	f.calls++
	return fn(context.WithValue(ctx, txMarkerKey, true))
}
