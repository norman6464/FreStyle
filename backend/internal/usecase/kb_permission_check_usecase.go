package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ErrPagePermissionDenied はページに対する操作が実効権限で許されていないときに返す。
// handler はこれを 403（あるいは存在自体を隠すなら 404）にマップする。
var ErrPagePermissionDenied = errors.New("permission denied for this page")

// CheckPagePermissionUseCase は「このユーザーはこのページを閲覧 / 編集できるか」に答える。
// ナレッジ基盤の認可はすべてここを通す（呼び出し側に判定規則を写経させない）。
//
// 段 1-b の各 usecase（GetPageUseCase / RenamePageUseCase / ReplacePageBlocksUseCase …）への
// 組み込みは handler の段で行う。組み込み方は次のとおり:
//
//	perm, err := check.Execute(ctx, usecase.CheckPagePermissionInput{
//	    WorkspaceID: workspaceID, PageID: pageID, UserID: currentUserID,
//	})
//	if err != nil {
//	    return err // ページが無い場合は repository.ErrPageNotFound がそのまま来る
//	}
//	if !perm.CanView { // 書き込み系なら !perm.CanEdit
//	    return usecase.ErrPagePermissionDenied
//	}
//
// ツリー取得のように複数ページを扱う経路では、ページごとにこれを呼ばず
// ListViewablePagesUseCase を使う（1 ページ 1 往復にしないため）。
type CheckPagePermissionUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewCheckPagePermissionUseCase(r repository.KnowledgeBasePermissionRepository) *CheckPagePermissionUseCase {
	return &CheckPagePermissionUseCase{repo: r}
}

type CheckPagePermissionInput struct {
	WorkspaceID string
	PageID      string
	UserID      uint64
}

func (u *CheckPagePermissionUseCase) Execute(ctx context.Context, in CheckPagePermissionInput) (*domain.PagePermission, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.PageID == "" {
		return nil, errors.New("pageID is required")
	}
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	facts, err := u.repo.PagePermissionFactsForUser(ctx, in.WorkspaceID, in.PageID, in.UserID)
	if err != nil {
		return nil, err
	}
	perm := domain.ResolvePagePermission(*facts)
	return &perm, nil
}

// IsWorkspaceMemberUseCase は「このユーザーはこのワークスペースのメンバーか」に答える。
// 所属は principals（kind='user'）の行の有無がすべてで、専用のメンバーシップ表は無い。
type IsWorkspaceMemberUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewIsWorkspaceMemberUseCase(r repository.KnowledgeBasePermissionRepository) *IsWorkspaceMemberUseCase {
	return &IsWorkspaceMemberUseCase{repo: r}
}

type IsWorkspaceMemberInput struct {
	WorkspaceID string
	UserID      uint64
}

func (u *IsWorkspaceMemberUseCase) Execute(ctx context.Context, in IsWorkspaceMemberInput) (bool, error) {
	if in.WorkspaceID == "" {
		return false, errors.New("workspaceID is required")
	}
	if in.UserID == 0 {
		return false, errors.New("userID is required")
	}
	return u.repo.IsWorkspaceMember(ctx, in.WorkspaceID, in.UserID)
}

// ListViewablePagesUseCase はスペース配下の現役ページのうち、そのユーザーが閲覧できるものを返す。
// ツリー取得の土台。ページ数によらず問い合わせは 1 回で、一覧の閲覧判定は
// domain.ResolvePageView に集約する（1 ページ解決の CheckPagePermissionUseCase は
// domain.ResolvePagePermission を通る。入口の関数は違うが、既定と例外の突き合わせは
// どちらも同じ 1 つの実装へ落ちる）。
//
// 答えられるのは閲覧可否だけ。編集可否が要る画面は CheckPagePermissionUseCase を使う
// （一覧のクエリは編集の例外を集めていない）。
type ListViewablePagesUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListViewablePagesUseCase(r repository.KnowledgeBasePermissionRepository) *ListViewablePagesUseCase {
	return &ListViewablePagesUseCase{repo: r}
}

type ListViewablePagesInput struct {
	WorkspaceID string
	SpaceID     string
	UserID      uint64
}

// HiddenChildrenRootKey は ListViewablePagesOutput.HiddenChildCount で
// 「スペース直下（親を持たない段）」を指すキー。ページ ID は必ず非空なので衝突しない。
const HiddenChildrenRootKey = ""

// ListViewablePagesOutput は閲覧できるページと、その各段で伏せた件数の組。
//
// HiddenChildCount は「**見えている**段の直下にある、見えない子の数」。キーは親ページの ID で、
// スペース直下の分は HiddenChildrenRootKey に入る。0 件の段はキーごと入らない。
//
// なぜ件数を出すのか: 見えない子を黙って消すと、木に穴が空いた理由が利用者に分からず
// 「壊れている」と読まれる。件数だけを出し、題名は出さない。
//
// これは意図的な情報開示であることを明記しておく。件数からは「自分に見えていないページが
// 何枚あるか」が分かる（題名・作成者・更新日時は分からない）。伏せた側の意図を部分的に
// 損なうので、開示を止めるならこの map を作らないか、handler 側で 0/1 に丸める
// （どちらも 1 箇所の変更で足りるように、数える処理をここへ閉じてある）。
type ListViewablePagesOutput struct {
	Pages            []domain.Page
	HiddenChildCount map[string]int
}

func (u *ListViewablePagesUseCase) Execute(ctx context.Context, in ListViewablePagesInput) (ListViewablePagesOutput, error) {
	if in.WorkspaceID == "" {
		return ListViewablePagesOutput{}, errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return ListViewablePagesOutput{}, errors.New("spaceID is required")
	}
	if in.UserID == 0 {
		return ListViewablePagesOutput{}, errors.New("userID is required")
	}
	rows, err := u.repo.ListSpacePageViewFacts(ctx, in.WorkspaceID, in.SpaceID, in.UserID)
	if err != nil {
		return ListViewablePagesOutput{}, err
	}

	pages := make([]domain.Page, 0, len(rows))
	viewable := make(map[string]bool, len(rows))
	for _, row := range rows {
		if domain.ResolvePageView(row.Facts) {
			viewable[row.Page.ID] = true
			pages = append(pages, row.Page)
		}
	}

	// 1 件も見えないなら件数も返さない。
	//
	// ここを外すと実在オラクルが開く。ツリー取得は「存在しないスペース」と「中身が 1 件も
	// 見えないスペース」を撃ち分けないことになっているが、前者は 0 件・後者は N 件を返して
	// しまい、スペース ID の総当たりで実在が分かる。
	//
	// 逆に 1 件でも見えていれば、スペースの実在はその時点で既に分かっている。だから
	// 「見えている段の直下に何枚伏せてあるか」を足しても、実在については何も増えない。
	// 件数を出してよい条件は **利用者が既にその段を見ていること** であり、スペース直下は
	// 見えるページが 1 枚も無い場合に限りその足場が無い。
	if len(pages) == 0 {
		return ListViewablePagesOutput{Pages: pages, HiddenChildCount: map[string]int{}}, nil
	}

	hidden := make(map[string]int)
	for _, row := range rows {
		if viewable[row.Page.ID] {
			continue
		}
		if row.Page.ParentID == nil {
			hidden[HiddenChildrenRootKey]++
			continue
		}
		// 親も見えないなら数えない。数えると「見えない枝の中に何枚あるか」まで漏れ、
		// 見えない親の子を根へ昇格させない（PageTreeOrphanHidden）判断と食い違う。
		// 数えてよいのは、利用者が現に見ている段の直下だけ。
		if !viewable[*row.Page.ParentID] {
			continue
		}
		hidden[*row.Page.ParentID]++
	}

	return ListViewablePagesOutput{Pages: pages, HiddenChildCount: hidden}, nil
}

// CanEditPageSubtreeUseCase は「このユーザーは、このページと全子孫を編集できるか」に答える。
// ページを名指しして子孫ごと書き換える操作（アーカイブ / 復帰）の入口で使う。
//
// 根 1 枚の CheckPagePermissionUseCase では足りない。子孫には親と違う例外が張られている
// ことがあり、根だけを見て通すと「その子を直接 rename すると 403 なのに、親のアーカイブ
// 経由なら書き換えられる」という、同じ編集判定が経路で食い違う状態になる。
//
// 問い合わせはページ数によらず 1 回（サブツリーの事実をまとめて集める）。判定は
// domain.ResolvePagePermission を 1 ページずつ通す — 1 枚解決と同じ規則を使い、
// ここには写経しない。
type CanEditPageSubtreeUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewCanEditPageSubtreeUseCase(r repository.KnowledgeBasePermissionRepository) *CanEditPageSubtreeUseCase {
	return &CanEditPageSubtreeUseCase{repo: r}
}

type CanEditPageSubtreeInput struct {
	WorkspaceID string
	PageID      string
	UserID      uint64
}

func (u *CanEditPageSubtreeUseCase) Execute(ctx context.Context, in CanEditPageSubtreeInput) (bool, error) {
	if in.WorkspaceID == "" {
		return false, errors.New("workspaceID is required")
	}
	if in.PageID == "" {
		return false, errors.New("pageID is required")
	}
	if in.UserID == 0 {
		return false, errors.New("userID is required")
	}
	rows, err := u.repo.ListSubtreePagePermissionFacts(ctx, in.WorkspaceID, in.PageID, in.UserID)
	if err != nil {
		return false, err
	}
	if len(rows) == 0 {
		// closure は自分自身（depth 0）を必ず含むので、0 行は「ページが無い」を意味する。
		// 許可には倒さない（呼び出し側は先に根の権限を確かめている前提で、ここは安全弁）。
		return false, nil
	}
	for _, row := range rows {
		if !domain.ResolvePagePermission(row.Facts).CanEdit {
			return false, nil
		}
	}
	return true, nil
}
