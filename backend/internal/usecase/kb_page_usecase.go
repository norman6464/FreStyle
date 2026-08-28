package usecase

import (
	"context"
	"errors"
	"unicode/utf8"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/pkg/fracindex"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ノートのページ操作で共通のビジネスルール違反。
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
	// ErrPageAnchorNotSibling は、移動で指定された「隣のページ」が移動先の現役の子で
	// なかったときに返す。不在・別の親・別スペース・アーカイブ済みを区別しない。
	//
	// 黙って末尾へ落とさないのは、**利用者が落とした場所と違う場所に入り、しかも
	// 成功したように見える**ため。断って、やり直せるようにする。
	ErrPageAnchorNotSibling = errors.New("anchor page is not a sibling under the destination")
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
	// 親があるときは **スペースを引かない**。
	//
	// 引くと、応答の差から「URL に書いた spaceID のスペースが実在するか」が分かってしまう。
	// 落ちる順序がそのまま応答になるため:
	//
	//	実在しない ID → FindSpace が ErrSpaceNotFound → 404 not_found
	//	実在する別の ID → FindSpace は通り、親との不一致 → 400 parent_space_mismatch
	//
	// この差は、他の 3 経路が塞いでいるものを作成経路だけで開ける。スペース一覧は
	// 閲覧できないスペースを 1 件も返さず、ツリー取得は未存在も不可視も同じ応答に揃え、
	// handler の requireSpacePermission は両方 404 に畳んでいる。
	//
	// 親があるなら引く必要も無い。**親が在ることがスペースが在ることの証明**で、
	// しかも handler が親の編集権限を先に確かめているので、呼び出し側は既にその親を
	// 見えている（＝そのスペースの実在は本人にとって既知）。
	// 残る仕事は「URL のスペースが親のスペースと同じか」の文字列比較だけで、
	// 実在しない ID も別の実在する ID も同じ 400 に落ちる。
	//
	// 親が無いときは比較する相手がいないので、これまでどおり引いて確かめる。
	// そちらは handler の requireSpacePermission が先に通っており、不在も無権限も
	// 同じ 404 に畳まれているので差は生まれない。
	if in.ParentID == nil {
		if _, err := u.repo.FindSpace(ctx, in.WorkspaceID, in.SpaceID); err != nil {
			return nil, err
		}
	} else {
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

// FindPageUseCase はページ 1 件のメタ情報だけを引く（本文は読まない）。
//
// GetPageUseCase と分けているのは、読む量と目的が違うため。あちらは snapshot か blocks から
// ProseMirror doc を組み立てて返す「本文を画面に出すための口」で、こちらが要るのは
// ページが属するスペースと現在の状態だけ。権限操作 API（ページの例外・共有リンク）の
// 認可がこれを使う — ページに対する権限を変えてよいかは「そのページが属するスペースの
// admin か」で決まるので、まずページからスペースを知る必要がある。
//
// 本文を読まないことには意味がある。この口は認可より前に呼ばれる（スペースが分からないと
// 認可判定ができない）ため、通れば必ず中身が見えるのでは困る。返すのはメタ情報だけで、
// 呼び出し側は認可に落ちた場合それすら応答に出さない。
type FindPageUseCase struct {
	repo repository.KnowledgeBaseRepository
}

func NewFindPageUseCase(r repository.KnowledgeBaseRepository) *FindPageUseCase {
	return &FindPageUseCase{repo: r}
}

type FindPageInput struct {
	WorkspaceID string
	PageID      string
}

func (u *FindPageUseCase) Execute(ctx context.Context, in FindPageInput) (*domain.Page, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.PageID == "" {
		return nil, repository.ErrPageNotFound
	}
	return u.repo.FindPage(ctx, in.WorkspaceID, in.PageID)
}

// ResolvePageLocationUseCase は URL の /p/{pageId} からページの居場所（ワークスペース）を
// 特定する。URL にテナントを出さない（ユーザー決定 2026-08-28: URL は UUID だけ）ための口。
//
// これはテナント確定**前**に呼ばれる唯一のページ読みなので、ここでは何も判定しない。
// 権限（見えるか・編集できるか）は handler が返した WorkspaceID で
// CheckPagePermissionUseCase を通す。この usecase の答えを判定なしで応答に使わないこと。
type ResolvePageLocationUseCase struct {
	repo repository.KnowledgeBaseRepository
}

func NewResolvePageLocationUseCase(r repository.KnowledgeBaseRepository) *ResolvePageLocationUseCase {
	return &ResolvePageLocationUseCase{repo: r}
}

type ResolvePageLocationOutput struct {
	Page      domain.Page
	Workspace domain.Workspace
}

func (u *ResolvePageLocationUseCase) Execute(ctx context.Context, pageID string) (*ResolvePageLocationOutput, error) {
	if pageID == "" {
		return nil, repository.ErrPageNotFound
	}
	page, err := u.repo.FindPageByIDAcrossWorkspaces(ctx, pageID)
	if err != nil {
		return nil, err
	}
	ws, err := u.repo.FindWorkspaceByID(ctx, page.WorkspaceID)
	if err != nil {
		return nil, err
	}
	return &ResolvePageLocationOutput{Page: *page, Workspace: *ws}, nil
}

// DeletePageUseCase はページを子孫ごと物理削除する。
// アーカイブ（隠すだけ・戻せる）とは別の操作で、こちらは戻せない。
// 誰が消せるか（根と子孫全部の編集権限）は handler の入口が確かめる。
type DeletePageUseCase struct {
	repo repository.KnowledgeBaseRepository
}

func NewDeletePageUseCase(r repository.KnowledgeBaseRepository) *DeletePageUseCase {
	return &DeletePageUseCase{repo: r}
}

type DeletePageInput struct {
	WorkspaceID string
	PageID      string
}

func (u *DeletePageUseCase) Execute(ctx context.Context, in DeletePageInput) error {
	return u.repo.DeletePageSubtree(ctx, in.WorkspaceID, in.PageID)
}
