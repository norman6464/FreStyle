package usecase

import (
	"context"
	"errors"
	"strings"

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
	// Archived が true ならアーカイブ済みのページを返す（既定は現役）。
	// 権限の見方は現役とまったく同じ — 同じクエリの絞り込みだけが変わる。
	Archived bool
}

// HiddenChildrenRootKey は ListViewablePagesOutput.HasHiddenChildren で
// 「スペース直下（親を持たない段）」を指すキー。ページ ID は必ず非空なので衝突しない。
const HiddenChildrenRootKey = ""

// ListViewablePagesOutput は閲覧できるページと、「その段に見えない子が居るか」の組。
//
// HasHiddenChildren のキーは親ページの ID で、スペース直下の分は HiddenChildrenRootKey に入る。
// 居ない段はキーごと入らない。
//
// なぜ知らせるのか: 見えない子を黙って消すと、木に穴が空いた理由が利用者に分からず
// 「壊れている」と読まれる。居ることだけを知らせ、題名は出さない。
//
// # なぜ枚数ではなく有無なのか
//
// 利用者にとって「2 枚」と「7 枚」の差は行動を何も変えない（知りたいのは「ここに見えない
// ものがある」だけ）。一方で枚数を出すと、伏せた量に比例して漏れる情報が増える
// （例: 採用の記録のスペースで「12 ページ」と出れば、採用の動きの規模が読める）。
// 得るものが定数で、失うものが伏せた量に比例するので、割に合わない。
//
// **枚数はどこにも作らない。** 数えてから丸めるのではなく、最初の 1 枚で true にして打ち切る。
// 変数として存在しなければ、うっかり応答に載る経路も生まれない。
type ListViewablePagesOutput struct {
	Pages             []domain.Page
	HasHiddenChildren map[string]bool
	// ParentArchived は「親がアーカイブ済み」のページの ID。事実であって判断ではない。
	//
	// アーカイブ済みの一覧で、その行を復帰できるかを呼び出し側が決めるのに使う
	// （規則は UnarchivePageUseCase が持つ: 親がアーカイブ中なら断る）。
	// 現役の一覧では常に空（現役ページの親がアーカイブ済みになることは無い）。
	ParentArchived map[string]bool
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
	rows, err := u.repo.ListSpacePageViewFacts(ctx, in.WorkspaceID, in.SpaceID, in.UserID, in.Archived)
	if err != nil {
		return ListViewablePagesOutput{}, err
	}

	pages := make([]domain.Page, 0, len(rows))
	viewable := make(map[string]bool, len(rows))
	parentArchived := make(map[string]bool)
	for _, row := range rows {
		if domain.ResolvePageView(row.Facts) {
			viewable[row.Page.ID] = true
			pages = append(pages, row.Page)
			if row.ParentArchived {
				parentArchived[row.Page.ID] = true
			}
		}
	}

	// 画面に 1 行も出ないなら、見えない子の有無も返さない。
	//
	// ここを外すと、応答の差から**そのスペースが実在するかどうか**が分かってしまう。
	// ツリー取得は「存在しないスペース」と「中身が 1 行も出ないスペース」を撃ち分けない
	// ことになっているが、前者は false・後者は true を返してしまい、スペース ID を
	// 総当たりするだけで実在を数え上げられる（存在の有無そのものが他人の情報）。
	//
	// 逆に 1 行でも出ていれば、スペースの実在はその時点で既に分かっている。だから
	// 「見えている段の直下に伏せたものが在るか」を足しても、実在については何も増えない。
	// 知らせてよい条件は **利用者が既にその段を見ていること**。
	//
	// # 「見えるページが 0 枚か」で判定してはいけない
	//
	// pages には**孤児**（自分は見えるが親が見えないページ）も入る。孤児は木に繋がらないので
	// BuildPageTree(PageTreeOrphanHidden) が丸ごと落とし、画面には 1 行も出ない。
	// つまり pages が非空でも木が空になることがある。
	//
	// 実際に踏んだ形: 根が非公開で子だけ閲覧できるとき、pages=[子] なので 0 枚判定は通り抜け、
	// 一方 hidden[""] は（見えない根を数えて）true になる。結果 {"pages":[],"hasHiddenChildren":true}
	// が返り、存在しないスペースの {"pages":[],"hasHiddenChildren":false} と撃ち分けられた。
	//
	// 木が空になるのは**見える根が 1 つも無いとき**（BuildPageTree の根は「親を持たない見えるページ」）
	// なので、そこで判定する。
	hasVisibleRoot := false
	for i := range pages {
		if pages[i].ParentID == nil {
			hasVisibleRoot = true
			break
		}
	}
	if !hasVisibleRoot {
		return ListViewablePagesOutput{Pages: pages, HasHiddenChildren: map[string]bool{}, ParentArchived: parentArchived}, nil
	}

	hidden := make(map[string]bool)
	for _, row := range rows {
		if viewable[row.Page.ID] {
			continue
		}
		if row.Page.ParentID == nil {
			hidden[HiddenChildrenRootKey] = true
			continue
		}
		// 親も見えないなら数えない。数えると「見えない枝の中に何枚あるか」まで漏れ、
		// 見えない親の子を根へ昇格させない（PageTreeOrphanHidden）判断と食い違う。
		// 数えてよいのは、利用者が現に見ている段の直下だけ。
		if !viewable[*row.Page.ParentID] {
			continue
		}
		hidden[*row.Page.ParentID] = true
	}

	return ListViewablePagesOutput{Pages: pages, HasHiddenChildren: hidden, ParentArchived: parentArchived}, nil
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

// SearchViewablePagesUseCase はワークスペース全体を題名で検索し、閲覧できるページだけを返す。
//
// ふるいは一覧（ListViewablePages）とまったく同じ domain.ResolvePageView。
// 検索だけ別の判定を持つと「一覧には出ないのに検索では出る」というずれ方をして、
// 伏せてあるページの実在が検索から漏れる。
//
// Limit は応答の件数。候補の計算量の天井（200 件）は SQL 側が持っていて別物。
type SearchViewablePagesUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewSearchViewablePagesUseCase(r repository.KnowledgeBasePermissionRepository) *SearchViewablePagesUseCase {
	return &SearchViewablePagesUseCase{repo: r}
}

type SearchViewablePagesInput struct {
	WorkspaceID string
	UserID      uint64
	// Query は題名の部分一致（大文字小文字は区別しない）。空白だけは呼び出し側で弾く。
	Query string
	// Limit は返す最大件数。0 以下なら既定の 20。上限 50。
	Limit int
}

func (u *SearchViewablePagesUseCase) Execute(ctx context.Context, in SearchViewablePagesInput) ([]domain.Page, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return nil, errors.New("query is required")
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	rows, err := u.repo.SearchWorkspacePageViewFacts(ctx, in.WorkspaceID, in.UserID, query)
	if err != nil {
		return nil, err
	}
	// 確保量は行数（SQL 側の天井 200 以下）で決める。利用者由来の limit を確保量に
	// 使わない — 上で挟んでいても、確保だけ大きくする余地を入力に持たせない。
	pages := make([]domain.Page, 0, len(rows))
	for _, row := range rows {
		if !domain.ResolvePageView(row.Facts) {
			continue
		}
		pages = append(pages, row.Page)
		if len(pages) >= limit {
			break
		}
	}
	return pages, nil
}
