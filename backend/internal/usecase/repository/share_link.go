package repository

import (
	"context"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ErrShareLinkNotFound は対象の共有リンクが存在しないときに返す。
// トークンが違う場合もこれを返す（存在の有無自体を漏らさない）。
var ErrShareLinkNotFound = errors.New("share link not found")

// ShareLinkWrite は共有リンクの発行に渡す値。
//
// ID と PrincipalID を持たないのは、どちらも採番が repository の責務のため
// （主体の作成と共有リンクの作成は同じトランザクションで行う）。
type ShareLinkWrite struct {
	WorkspaceID string
	PageID      string
	// Capability はリンク経由でできることの既定。
	Capability domain.Capability
	// TokenHash は共有 URL に載るトークンの SHA-256（32 バイト）。平文は渡さない。
	TokenHash []byte
	// PasswordHash はパスワードの bcrypt ハッシュ。nil ならパスワード無し。
	PasswordHash *string
	// ExpiresAt は有効期限。nil なら無期限。
	ExpiresAt *time.Time
	// CreatedByUserID は発行者。
	CreatedByUserID uint64
}

// ShareLinkRepository は共有リンク（share_links、および発行時に紐づく
// kind='share_link' の principal）へのアクセスを提供する。
//
// KnowledgeBasePermissionRepository から分けているのは、他のどの grant/principal
// 操作からも独立して呼ばれるため（IssueShareLinkUseCase 等、kb_share_link_usecase.go の
// 4 つの usecase 以外の消費者を持たない）。Create の内部でだけ principal + share_link を
// 1 トランザクションで作る（主体だけが残る／リンクだけが残る状態を作らない）が、
// これは 1 メソッドに閉じたままの操作で、usecase 層で束ねる対象ではない。
type ShareLinkRepository interface {
	// Create は共有リンクを発行する。kind='share_link' の主体の採番と作成も
	// 同じトランザクションで行う。
	Create(ctx context.Context, in ShareLinkWrite) (*domain.ShareLink, error)
	// Revoke は共有リンクを失効させる。既に失効済みなら何もしない（冪等）。
	// 対象が無い・別ワークスペースなら ErrShareLinkNotFound。
	Revoke(ctx context.Context, workspaceID, shareLinkID string) error
	// FindByTokenHash はトークンの SHA-256 から共有リンクを引く。
	// 期限切れ・失効も含めて返す（判定は usecase 側）。無ければ ErrShareLinkNotFound。
	FindByTokenHash(ctx context.Context, tokenHash []byte) (*domain.ShareLink, error)
	// ListByPage はページに発行された共有リンクの一覧を返す（失効済みも含む）。
	ListByPage(ctx context.Context, workspaceID, pageID string) ([]domain.ShareLink, error)
}
