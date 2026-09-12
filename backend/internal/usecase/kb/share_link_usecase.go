package kb

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
	"golang.org/x/crypto/bcrypt"
)

// 共有リンクを開けないときの理由。どれも「開けない」だが、利用者に返す案内が変わるため分ける
// （期限切れなら再発行を頼む、パスワード違いなら入れ直す）。
var (
	ErrShareLinkRevoked          = errors.New("share link has been revoked")
	ErrShareLinkExpired          = errors.New("share link has expired")
	ErrShareLinkPasswordRequired = errors.New("share link requires a password")
	ErrShareLinkPasswordMismatch = errors.New("share link password does not match")
	// ErrShareLinkPageOutOfScope はリンクの対象ページでもその子孫でもないページを開こうとしたときに返す。
	ErrShareLinkPageOutOfScope = errors.New("page is not covered by this share link")
)

// shareLinkTokenBytes は共有 URL に載せるトークンの乱数バイト数。32 バイト（256bit）あれば
// 総当たりは現実的でなく、ハッシュを SHA-256 にできる（遅いハッシュが要るのは人が選ぶ
// 短い値＝パスワードの方）。
const shareLinkTokenBytes = 32

// hashShareLinkToken はトークンを SHA-256 で縮める。平文を DB に置かないため保存も照合もこの値で行う。
func hashShareLinkToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// IssueShareLinkUseCase はページの公開 URL を発行する。
// 戻り値の Token はこのときだけ返る平文（DB にはハッシュしか残らない）。呼び出し側は
// URL を組み立てて利用者に渡し、以後は保持しないこと。
type IssueShareLinkUseCase struct {
	repo repository.ShareLinkRepository
}

func NewIssueShareLinkUseCase(r repository.ShareLinkRepository) *IssueShareLinkUseCase {
	return &IssueShareLinkUseCase{repo: r}
}

type IssueShareLinkInput struct {
	WorkspaceID string
	PageID      string
	// Capability はリンク経由でできることの既定（view または edit）。
	Capability domain.Capability
	// Password が空でなければパスワード付きにする。
	Password string
	// ExpiresAt が nil なら無期限。
	ExpiresAt       *time.Time
	CreatedByUserID uint64
}

// IssueShareLinkOutput は発行したリンクと、その 1 回だけ返る平文トークンの組。
type IssueShareLinkOutput struct {
	Link *domain.ShareLink
	// Token は URL に載せる平文トークン。DB には SHA-256 だけが残るため、失うと
	// リンクは二度と取り出せない（再発行になる）。
	Token string
}

func (u *IssueShareLinkUseCase) Execute(ctx context.Context, in IssueShareLinkInput) (*IssueShareLinkOutput, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.PageID == "" {
		return nil, errors.New("pageID is required")
	}
	if in.CreatedByUserID == 0 {
		return nil, errors.New("createdByUserID is required")
	}
	if !in.Capability.Valid() {
		return nil, ErrInvalidCapability
	}
	if in.ExpiresAt != nil && !in.ExpiresAt.After(time.Now()) {
		return nil, errors.New("expiresAt must be in the future")
	}

	raw := make([]byte, shareLinkTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("共有リンクのトークン生成に失敗: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)

	var passwordHash *string
	if in.Password != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("共有リンクのパスワードのハッシュ化に失敗: %w", err)
		}
		s := string(h)
		passwordHash = &s
	}

	link, err := u.repo.Create(ctx, repository.ShareLinkWrite{
		WorkspaceID:     in.WorkspaceID,
		PageID:          in.PageID,
		Capability:      in.Capability,
		TokenHash:       hashShareLinkToken(token),
		PasswordHash:    passwordHash,
		ExpiresAt:       in.ExpiresAt,
		CreatedByUserID: in.CreatedByUserID,
	})
	if err != nil {
		return nil, err
	}
	return &IssueShareLinkOutput{Link: link, Token: token}, nil
}

// RevokeShareLinkUseCase は共有リンクを失効させる（冪等）。
// 行は消さず revoked_at を立てるので、誰がいつ止めたかは残る。
type RevokeShareLinkUseCase struct {
	repo repository.ShareLinkRepository
}

func NewRevokeShareLinkUseCase(r repository.ShareLinkRepository) *RevokeShareLinkUseCase {
	return &RevokeShareLinkUseCase{repo: r}
}

type RevokeShareLinkInput struct {
	WorkspaceID string
	ShareLinkID string
}

func (u *RevokeShareLinkUseCase) Execute(ctx context.Context, in RevokeShareLinkInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.ShareLinkID == "" {
		return errors.New("shareLinkID is required")
	}
	return u.repo.Revoke(ctx, in.WorkspaceID, in.ShareLinkID)
}

// VerifyShareLinkUseCase は共有 URL のトークン（とパスワード）を検証し、使えるリンクを返す。
// 返ったリンクの PrincipalID がそのアクセスの主体になり、以後の権限解決は
// CheckShareLinkPermissionUseCase が行う。
type VerifyShareLinkUseCase struct {
	repo repository.ShareLinkRepository
}

func NewVerifyShareLinkUseCase(r repository.ShareLinkRepository) *VerifyShareLinkUseCase {
	return &VerifyShareLinkUseCase{repo: r}
}

type VerifyShareLinkInput struct {
	Token    string
	Password string
}

func (u *VerifyShareLinkUseCase) Execute(ctx context.Context, in VerifyShareLinkInput) (*domain.ShareLink, error) {
	if in.Token == "" {
		return nil, repository.ErrShareLinkNotFound
	}
	link, err := u.repo.FindByTokenHash(ctx, hashShareLinkToken(in.Token))
	if err != nil {
		return nil, err
	}
	if link.RevokedAt != nil {
		return nil, ErrShareLinkRevoked
	}
	if !link.Usable(time.Now()) {
		return nil, ErrShareLinkExpired
	}
	if link.RequiresPassword() {
		if in.Password == "" {
			return nil, ErrShareLinkPasswordRequired
		}
		if err := bcrypt.CompareHashAndPassword([]byte(*link.PasswordHash), []byte(in.Password)); err != nil {
			return nil, ErrShareLinkPasswordMismatch
		}
	}
	return link, nil
}

// CheckShareLinkPermissionUseCase は検証済みの共有リンクで、あるページを閲覧 / 編集できるかを答える。
//
// できることはリンク自身の Capability だけで決まる（リンクの来訪者はワークスペースに
// 所属しないため、付与の 3 段は届かない）。共有リンクは広げる方向にしか働かない —
// ログインしていない相手へ「見せる」を足すだけで、既に見えている人から取り上げはしない。
//
// 対象ページはリンクのページ自身かその子孫でなければならない。リンクを持っているだけで
// スペース内の別のページを開けてしまわないよう、ここで必ず確かめる。
type CheckShareLinkPermissionUseCase struct {
	permissions repository.KnowledgeBasePermissionRepository
	pages       repository.KnowledgeBaseRepository
}

func NewCheckShareLinkPermissionUseCase(
	permissions repository.KnowledgeBasePermissionRepository,
	pages repository.KnowledgeBaseRepository,
) *CheckShareLinkPermissionUseCase {
	return &CheckShareLinkPermissionUseCase{permissions: permissions, pages: pages}
}

type CheckShareLinkPermissionInput struct {
	// Link は VerifyShareLinkUseCase が返した検証済みのリンク。
	Link   *domain.ShareLink
	PageID string
}

func (u *CheckShareLinkPermissionUseCase) Execute(ctx context.Context, in CheckShareLinkPermissionInput) (*domain.PagePermission, error) {
	if in.Link == nil {
		return nil, errors.New("link is required")
	}
	if in.PageID == "" {
		return nil, errors.New("pageID is required")
	}
	inScope, err := u.pages.HasDescendant(ctx, in.Link.WorkspaceID, in.Link.PageID, in.PageID)
	if err != nil {
		return nil, err
	}
	if !inScope {
		return nil, ErrShareLinkPageOutOfScope
	}
	facts, err := u.permissions.PagePermissionFactsForPrincipal(ctx, in.Link.WorkspaceID, in.PageID, in.Link.PrincipalID)
	if err != nil {
		return nil, err
	}
	capability := in.Link.Capability
	facts.ShareLinkCapability = &capability
	perm := domain.ResolvePagePermission(*facts)
	return &perm, nil
}

// ListPageShareLinksUseCase はページに発行済みの共有リンクを返す（失効済みも含む — 止めた
// 確認と、いつ誰が止めたかを追えるようにするため）。トークンは発行時の 1 回しか返らない
// （DB には SHA-256 しか残らない）ため、生きているリンクを知る手段はこれしか無い。
//
// 返す domain.ShareLink の TokenHash / PasswordHash は json:"-" で API へ出ず、
// この一覧からリンクを開く手がかりは出ない。
type ListPageShareLinksUseCase struct {
	repo repository.ShareLinkRepository
}

func NewListPageShareLinksUseCase(r repository.ShareLinkRepository) *ListPageShareLinksUseCase {
	return &ListPageShareLinksUseCase{repo: r}
}

type ListPageShareLinksInput struct {
	WorkspaceID string
	PageID      string
}

func (u *ListPageShareLinksUseCase) Execute(ctx context.Context, in ListPageShareLinksInput) ([]domain.ShareLink, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.PageID == "" {
		return nil, errors.New("pageID is required")
	}
	return u.repo.ListByPage(ctx, in.WorkspaceID, in.PageID)
}
