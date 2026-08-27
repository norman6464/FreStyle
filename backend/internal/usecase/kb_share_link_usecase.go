package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"golang.org/x/crypto/bcrypt"
)

// 共有リンクを開けないときの理由。どれも「開けない」だが、利用者に返す案内が変わるため分ける
// （期限切れなら再発行を頼む、パスワード違いなら入れ直す）。
var (
	// ErrShareLinkRevoked は失効させたリンクを使ったときに返す。
	ErrShareLinkRevoked = errors.New("share link has been revoked")
	// ErrShareLinkExpired は期限を過ぎたリンクを使ったときに返す。
	ErrShareLinkExpired = errors.New("share link has expired")
	// ErrShareLinkPasswordRequired はパスワード付きリンクにパスワード無しで来たときに返す。
	ErrShareLinkPasswordRequired = errors.New("share link requires a password")
	// ErrShareLinkPasswordMismatch はパスワードが違うときに返す。
	ErrShareLinkPasswordMismatch = errors.New("share link password does not match")
	// ErrShareLinkPageOutOfScope はリンクの対象ページでも その子孫でもないページを
	// そのリンクで開こうとしたときに返す。
	ErrShareLinkPageOutOfScope = errors.New("page is not covered by this share link")
)

// shareLinkTokenBytes は共有 URL に載せるトークンの乱数バイト数。
// 32 バイト（256 bit）あれば総当たりは現実的でなく、ハッシュを SHA-256 にできる
// （遅いハッシュで守らなければならないのは、人が選ぶ短い値＝パスワードの方）。
const shareLinkTokenBytes = 32

// hashShareLinkToken はトークン文字列を SHA-256 で 32 バイトへ縮める。
// 平文を DB に置かないため、保存も照合もこの値で行う。
func hashShareLinkToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// IssueShareLinkUseCase はページの公開 URL を発行する。
//
// 戻り値の Token は**このときだけ**返る平文（DB にはハッシュしか残らない）。
// 呼び出し側は URL を組み立てて利用者に渡し、以後は保持しないこと。
type IssueShareLinkUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewIssueShareLinkUseCase(r repository.KnowledgeBasePermissionRepository) *IssueShareLinkUseCase {
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
	ExpiresAt *time.Time
	// CreatedByUserID は発行者。
	CreatedByUserID uint64
}

// IssueShareLinkOutput は発行したリンクと、その 1 回だけ返る平文トークンの組。
type IssueShareLinkOutput struct {
	Link *domain.ShareLink
	// Token は URL に載せる平文トークン。DB には SHA-256 だけが残るため、
	// この値を失うとリンクは二度と取り出せない（再発行になる）。
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

	link, err := u.repo.CreateShareLink(ctx, repository.ShareLinkWrite{
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
	repo repository.KnowledgeBasePermissionRepository
}

func NewRevokeShareLinkUseCase(r repository.KnowledgeBasePermissionRepository) *RevokeShareLinkUseCase {
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
	return u.repo.RevokeShareLink(ctx, in.WorkspaceID, in.ShareLinkID)
}

// VerifyShareLinkUseCase は共有 URL のトークン（とパスワード）を検証し、使えるリンクを返す。
// 返ったリンクの PrincipalID がそのアクセスの主体になり、以後の権限解決は
// CheckShareLinkPermissionUseCase が行う。
type VerifyShareLinkUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewVerifyShareLinkUseCase(r repository.KnowledgeBasePermissionRepository) *VerifyShareLinkUseCase {
	return &VerifyShareLinkUseCase{repo: r}
}

type VerifyShareLinkInput struct {
	// Token は URL に載っていた平文トークン。
	Token string
	// Password はパスワード付きリンクのときに要る。
	Password string
}

func (u *VerifyShareLinkUseCase) Execute(ctx context.Context, in VerifyShareLinkInput) (*domain.ShareLink, error) {
	if in.Token == "" {
		return nil, repository.ErrShareLinkNotFound
	}
	link, err := u.repo.FindShareLinkByTokenHash(ctx, hashShareLinkToken(in.Token))
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
// リンクの既定（Capability）を出発点に、ページに張られた例外（page_restrictions）を
// メンバーのときと同じ規則で適用する。これにより「ページ全体を公開しつつ、
// 1 枚の子ページだけ deny で隠す」が書ける。
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
	Link *domain.ShareLink
	// PageID は開こうとしているページ。
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

// ListPageShareLinksUseCase はページに発行済みの共有リンクを返す（失効済みも含む）。
//
// 発行しっぱなしを防ぐための口。トークンは発行時の 1 回しか返らず（DB には SHA-256 しか
// 残らない）、あとから「今どのリンクが生きているか」を知る手段がこれしか無い。
// 失効済みも返すのは、止めたことの確認と、いつ誰が止めたかを追えるようにするため。
//
// 返す domain.ShareLink の TokenHash / PasswordHash は json:"-" で API へ出ない。
// handler 側も平文トークンを持っていない（保存していない）ので、この一覧から
// リンクを開く手がかりは出ない。
type ListPageShareLinksUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListPageShareLinksUseCase(r repository.KnowledgeBasePermissionRepository) *ListPageShareLinksUseCase {
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
	return u.repo.ListPageShareLinks(ctx, in.WorkspaceID, in.PageID)
}
