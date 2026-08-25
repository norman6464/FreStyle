package usecase

import (
	"context"
	"errors"
	"unicode/utf8"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/pkg/fracindex"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ナレッジ基盤のページ操作で共通のビジネスルール違反。
// handler はこれらを errors.Is で判定して HTTP ステータスにマップする（400/409 相当）。
var (
	// ErrPageArchived はアーカイブ済みページへの変更操作（改名・移動・本文書き換え）に返す。
	ErrPageArchived = errors.New("page is archived")
	// ErrPageParentArchived はアーカイブ済みページを親に指定した操作に返す。
	// アーカイブ済みの親に現役の子ができるとツリーに現れない「迷子ページ」になるため入口で塞ぐ。
	ErrPageParentArchived = errors.New("parent page is archived")
	// ErrPageParentSpaceMismatch は指定スペースと親ページの所属スペースが食い違うときに返す。
	// ページの木はスペースの中で閉じる（DB の複合 FK と同じ規則を入口でも検証する）。
	ErrPageParentSpaceMismatch = errors.New("parent page belongs to a different space")
	// ErrPageCycle は自分自身または自分の子孫の下への移動に返す（木が壊れる）。
	ErrPageCycle = errors.New("cannot move a page under itself or its descendant")
)

// kbPageTitleMaxLen は pages.title (varchar(200)) の上限。DB エラーの前に入口で弾く。
const kbPageTitleMaxLen = 200

// CreatePageUseCase はスペース直下または親ページの下に新しいページを作る。
// 兄弟の末尾へ分数インデックスで採番し、closure（page_paths）もページと同時に張られる。
type CreatePageUseCase struct {
	repo repository.KnowledgeBaseRepository
}

func NewCreatePageUseCase(r repository.KnowledgeBaseRepository) *CreatePageUseCase {
	return &CreatePageUseCase{repo: r}
}

type CreatePageInput struct {
	WorkspaceID string
	SpaceID     string
	// ParentID が nil ならスペース直下（ルート）に作る。
	ParentID        *string
	Title           string
	CreatedByUserID uint64
}

func (u *CreatePageUseCase) Execute(ctx context.Context, in CreatePageInput) (*domain.Page, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return nil, errors.New("spaceID is required")
	}
	if in.CreatedByUserID == 0 {
		return nil, errors.New("createdByUserID is required")
	}
	if utf8.RuneCountInString(in.Title) > kbPageTitleMaxLen {
		return nil, errors.New("title is too long")
	}
	if _, err := u.repo.FindSpace(ctx, in.WorkspaceID, in.SpaceID); err != nil {
		return nil, err
	}
	if in.ParentID != nil {
		parent, err := u.repo.FindPage(ctx, in.WorkspaceID, *in.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.SpaceID != in.SpaceID {
			return nil, ErrPageParentSpaceMismatch
		}
		if parent.ArchivedAt != nil {
			return nil, ErrPageParentArchived
		}
	}
	last, err := u.repo.LastActiveSiblingPosition(ctx, in.WorkspaceID, in.SpaceID, in.ParentID)
	if err != nil {
		return nil, err
	}
	pos, err := fracindex.Between(last, "")
	if err != nil {
		return nil, err
	}
	page := &domain.Page{
		WorkspaceID:     in.WorkspaceID,
		SpaceID:         in.SpaceID,
		ParentID:        in.ParentID,
		Position:        pos,
		Title:           in.Title,
		CreatedByUserID: in.CreatedByUserID,
	}
	if err := u.repo.CreatePage(ctx, page); err != nil {
		return nil, err
	}
	return page, nil
}

// GetPageUseCase はページ 1 件とその本文（ProseMirror doc）を返す。
// 本文は snapshot（読み取りキャッシュ）を優先し、無ければ blocks から組み立てる。
// snapshot は本文書き換えと同一トランザクションで焼き直されるため常に blocks と同期しており、
// 「新しい順」を判断する必要はない。blocks からの組み立ては未保存の新規ページ
// （snapshot がまだ無い）へのフォールバック。
type GetPageUseCase struct {
	repo repository.KnowledgeBaseRepository
}

func NewGetPageUseCase(r repository.KnowledgeBaseRepository) *GetPageUseCase {
	return &GetPageUseCase{repo: r}
}

type GetPageInput struct {
	WorkspaceID string
	PageID      string
}

// GetPageOutput はページのメタ情報と本文の組。
type GetPageOutput struct {
	Page domain.Page `json:"page"`
	// Doc は ProseMirror ドキュメント（JSON 文字列）。API へは handler の response 型で
	// json.RawMessage に変換して出す。
	Doc string `json:"-"`
}

func (u *GetPageUseCase) Execute(ctx context.Context, in GetPageInput) (*GetPageOutput, error) {
	page, err := u.repo.FindPage(ctx, in.WorkspaceID, in.PageID)
	if err != nil {
		return nil, err
	}
	snap, err := u.repo.GetPageSnapshot(ctx, in.WorkspaceID, in.PageID)
	if err == nil {
		return &GetPageOutput{Page: *page, Doc: snap.Doc}, nil
	}
	if !errors.Is(err, repository.ErrPageSnapshotNotFound) {
		return nil, err
	}
	blocks, err := u.repo.ListBlocksByPage(ctx, in.WorkspaceID, in.PageID)
	if err != nil {
		return nil, err
	}
	tree, err := treeFromBlocks(blocks)
	if err != nil {
		return nil, err
	}
	doc, err := renderPageDoc(tree)
	if err != nil {
		return nil, err
	}
	return &GetPageOutput{Page: *page, Doc: doc}, nil
}

// GetPageTreeUseCase はスペース配下の現役ページを木構造（表示順）で返す。
type GetPageTreeUseCase struct {
	repo repository.KnowledgeBaseRepository
}

func NewGetPageTreeUseCase(r repository.KnowledgeBaseRepository) *GetPageTreeUseCase {
	return &GetPageTreeUseCase{repo: r}
}

type GetPageTreeInput struct {
	WorkspaceID string
	SpaceID     string
}

func (u *GetPageTreeUseCase) Execute(ctx context.Context, in GetPageTreeInput) ([]*PageTreeNode, error) {
	if _, err := u.repo.FindSpace(ctx, in.WorkspaceID, in.SpaceID); err != nil {
		return nil, err
	}
	pages, err := u.repo.ListActivePagesBySpace(ctx, in.WorkspaceID, in.SpaceID)
	if err != nil {
		return nil, err
	}
	// 一覧はスペースの現役ページ全件で、権限でふるいにかけていない。ここで親が欠けるのは
	// アーカイブ運用の不変条件（サブツリーごと archive）が崩れた行だけなので、
	// データを隠さないようルート扱いで見せる（表示が乱れても本文は失わない）。
	return BuildPageTree(pages, PageTreeOrphanAsRoot), nil
}

// RenamePageUseCase はページのタイトルを変更する。
type RenamePageUseCase struct {
	repo repository.KnowledgeBaseRepository
}

func NewRenamePageUseCase(r repository.KnowledgeBaseRepository) *RenamePageUseCase {
	return &RenamePageUseCase{repo: r}
}

type RenamePageInput struct {
	WorkspaceID string
	PageID      string
	Title       string
}

func (u *RenamePageUseCase) Execute(ctx context.Context, in RenamePageInput) (*domain.Page, error) {
	if utf8.RuneCountInString(in.Title) > kbPageTitleMaxLen {
		return nil, errors.New("title is too long")
	}
	page, err := u.repo.FindPage(ctx, in.WorkspaceID, in.PageID)
	if err != nil {
		return nil, err
	}
	if page.ArchivedAt != nil {
		return nil, ErrPageArchived
	}
	return u.repo.UpdatePageTitle(ctx, in.WorkspaceID, in.PageID, in.Title)
}
