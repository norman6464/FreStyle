package kb

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// ErrPagePermissionDenied はページに対する操作が実効権限で許されていないときに返す。
// handler はこれを 403（あるいは存在自体を隠すなら 404）にマップする。
var ErrPagePermissionDenied = errors.New("permission denied for this page")

// CheckPagePermissionUseCase は「このユーザーはこのページを閲覧 / 編集できるか」に答える。
// ナレッジの認可はすべてここを通す（呼び出し側に判定規則を写経させない）。
//
// 段 1-b の各 usecase（GetPageUseCase / RenamePageUseCase / ReplacePageBlocksUseCase …）への
// 組み込みは handler の段で行う。組み込み方は次のとおり:
//
//	perm, err := check.Execute(ctx, kb.CheckPagePermissionInput{
//	    WorkspaceID: workspaceID, PageID: pageID, UserID: currentUserID,
//	})
//	if err != nil {
//	    return err // ページが無い場合は repository.ErrPageNotFound がそのまま来る
//	}
//	if !perm.CanView { // 書き込み系なら !perm.CanEdit
//	    return kb.ErrPagePermissionDenied
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
// domain.ResolvePagePermission を通る。入口の関数は違うが、役割から可否を出す規則は
// どちらも同じ 1 つの実装へ落ちる）。
//
// 答えられるのは閲覧可否だけ。編集可否が要る画面は CheckPagePermissionUseCase を使う
// （一覧のクエリは所属（Member）まで集めていない）。
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
		if domain.ResolvePageView(row.Role, row.Page.Visibility, row.Page.CreatedByUserID == in.UserID) {
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
// # いまの権限モデルでは、この検査は断らない
//
// 権限は 3 段の付与を足し合わせて「届いた中で最も強い役割」で決まり、打ち消す層が無い。
// 子孫の経路は親の経路を必ず含み（page_paths）、スペースは親子で揃う（fk_pages_parent が
// (workspace_id, space_id, parent_id) で参照するため DB が強制する）。
// つまり**役割は木を下るほど弱くならない**ので、根を編集できるなら全子孫も編集できる。
//
// それでも残しているのは、これが**事実を集めるクエリの回帰を捕まえる最後の網**だから。
// 経路の辿り方を取り違える（祖先ではなく子孫を集めてしまう等）と、1 枚解決と一覧で
// 答えが食い違い、根だけ見て通す実装では気づけない。1 スペース 5,000 ページで 3 ms、
// 呼ぶのはアーカイブ / 復帰の 1 回だけなので、置いておく代償は小さい。
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

// SearchMatchFieldTitle / SearchMatchFieldBody は SearchViewablePageResult.MatchField の値。
// どちらでヒットしたかをフロントが区別する（題名一致は抜粋を出さない・本文一致は
// 抜粋とヒット位置を出す）。
const (
	SearchMatchFieldTitle = "title"
	SearchMatchFieldBody  = "body"
)

// searchExcerptWindowRunes は本文一致の抜粋で、ヒット位置の前後に残す rune 数。
// 日本語を含む本文を想定するため rune 単位（バイト単位ではない）。
const searchExcerptWindowRunes = 30

// SearchViewablePageResult は検索結果 1 件（ページ本体 + どこにヒットしたか）。
type SearchViewablePageResult struct {
	Page domain.Page
	// MatchField はヒットした場所（SearchMatchFieldTitle | SearchMatchFieldBody）。
	// 題名が一致していれば常に "title"（本文も一致していたとしても、利用者にとって
	// 分かりやすいのは題名一致であるほうなので、そちらを優先する）。
	MatchField string
	// Excerpt は MatchField が "body" のときだけ非空。ヒット周辺を rune 境界を壊さずに
	// 切り出した抜粋文字列（前後 searchExcerptWindowRunes 文字程度の窓）。
	Excerpt string
	// MatchStart / MatchLen は **Excerpt の中での** ヒット位置・長さ（rune 単位。
	// フロントが mark で囲むための材料）。MatchField が "title" のときは両方ゼロ。
	MatchStart int
	MatchLen   int
}

// SearchViewablePagesUseCase はワークスペース全体を題名 **または本文** で検索し、
// 閲覧できるページだけを返す（本文検索に対応）。
//
// ふるいは一覧（ListViewablePages）とまったく同じ domain.ResolvePageView。
// 検索だけ別の判定を持つと「一覧には出ないのに検索では出る」というずれ方をして、
// 伏せてあるページの実在が検索から漏れる。
//
// Limit は応答の件数。SQL 側は候補に上限を掛けない（可視でふるう前に切ると
// 本来見えるはずの一致を取りこぼす）。
type SearchViewablePagesUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewSearchViewablePagesUseCase(r repository.KnowledgeBasePermissionRepository) *SearchViewablePagesUseCase {
	return &SearchViewablePagesUseCase{repo: r}
}

type SearchViewablePagesInput struct {
	WorkspaceID string
	UserID      uint64
	// Query は題名 / 本文の部分一致（大文字小文字は区別しない）。空白だけは呼び出し側で弾く。
	Query string
	// Limit は返す最大件数。0 以下なら既定の 20。上限 50。
	Limit int
}

func (u *SearchViewablePagesUseCase) Execute(ctx context.Context, in SearchViewablePagesInput) ([]SearchViewablePageResult, error) {
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
	// 確保量は行数で決める。利用者由来の limit を確保量に使わない — 上で挟んでいても、
	// 確保だけ大きくする余地を入力に持たせない。
	results := make([]SearchViewablePageResult, 0, len(rows))
	for _, row := range rows {
		if !domain.ResolvePageView(row.Role, row.Page.Visibility, row.Page.CreatedByUserID == in.UserID) {
			continue
		}
		results = append(results, buildSearchViewablePageResult(row, query))
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

// buildSearchViewablePageResult は 1 行の検索候補（題名 + 本文）と query から、
// どこにヒットしたか（MatchField）と本文一致なら抜粋を組み立てる。
//
// 題名一致を優先する。SQL 側の候補（cand）は「題名一致 OR 本文一致」で絞っているため、
// 両方一致することもある — その場合は利用者にとって分かりやすい題名一致として返す
// （本文の抜粋よりも一致した題名そのものの方が判断材料として明確なため）。
func buildSearchViewablePageResult(row repository.PageSearchViewFact, query string) SearchViewablePageResult {
	if strings.Contains(strings.ToLower(row.Page.Title), strings.ToLower(query)) {
		return SearchViewablePageResult{Page: row.Page, MatchField: SearchMatchFieldTitle}
	}
	excerpt, start, length, found := computeSearchExcerpt(row.Body, query)
	if !found {
		// SQL 側は ILIKE で一致したはずだが、Go 側の判定（大文字小文字とレンダリング上の
		// 揺れ）とずれて見つからない場合の安全弁。抜粋なしの本文一致として返す —
		// 「一致した」という事実自体は SQL が保証しているので、ここで隠さない。
		return SearchViewablePageResult{Page: row.Page, MatchField: SearchMatchFieldBody}
	}
	return SearchViewablePageResult{
		Page:       row.Page,
		MatchField: SearchMatchFieldBody,
		Excerpt:    excerpt,
		MatchStart: start,
		MatchLen:   length,
	}
}

// computeSearchExcerpt は body の中から query に大文字小文字を無視して部分一致する最初の
// 位置を探し、その前後 searchExcerptWindowRunes rune の窓を切り出す。
//
// rune 単位で処理するのは、本文が日本語を含むため。byte 単位でスライスすると
// マルチバイト文字の途中で切れて壊れた文字列になり得る（rune 境界を壊さない）。
//
// 戻り値の start / length は **切り出した excerpt の中での** rune 位置・長さ
// （フロントが mark で囲む用）。見つからなければ found=false。
func computeSearchExcerpt(body, query string) (excerpt string, start, length int, found bool) {
	lowerBody := strings.ToLower(body)
	lowerQuery := strings.ToLower(query)
	if lowerQuery == "" || lowerBody == "" {
		return "", 0, 0, false
	}
	byteIdx := strings.Index(lowerBody, lowerQuery)
	if byteIdx < 0 {
		return "", 0, 0, false
	}
	// strings.Index はバイト位置を返す。以降の計算は rune 単位（マルチバイト文字を
	// 含む本文の境界を壊さない）なので、ここで一度だけ rune 位置へ変換する。
	idx := utf8.RuneCountInString(lowerBody[:byteIdx])
	queryLen := utf8.RuneCountInString(lowerQuery)
	bodyRunes := []rune(body)
	winStart := idx - searchExcerptWindowRunes
	if winStart < 0 {
		winStart = 0
	}
	winEnd := idx + queryLen + searchExcerptWindowRunes
	if winEnd > len(bodyRunes) {
		winEnd = len(bodyRunes)
	}
	return string(bodyRunes[winStart:winEnd]), idx - winStart, queryLen, true
}

// ListPageBacklinksUseCase は、対象ページを参照している（page_links.target_page_id =
// 対象ページ）ページのうち、閲覧できるものだけを返す（逆リンク）。
//
// 検索・逆リンクは kb パッケージ内の usecase として新設する（別パッケージにしない）。
// ページ本文に密結合した機能で、新しい権限軸を持たないため。ふるいは検索・一覧と同じ
// domain.ResolvePageView — 見えない参照元は存在も題名も一切出さない。
//
// 対象ページ自体を見られるかどうかの判定（CapabilityView）はここでは行わない。
// handler が requirePagePermission（CheckPagePermissionUseCase）で先に確かめる
// （他のページ名指し系エンドポイントと同じ形）。
//
// SQL 側には LIMIT を掛けない。ここで可視判定より先に絞ると、search と同じ理由で
// 本来見えるはずの参照元が取りこぼされる。代わりに、可視判定を終えた後の応答件数を
// listPageBacklinksMaxResults で打ち切る（無制限に参照されるページの応答が
// 際限なく膨らむのを防ぐ防御的な上限）。
type ListPageBacklinksUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListPageBacklinksUseCase(r repository.KnowledgeBasePermissionRepository) *ListPageBacklinksUseCase {
	return &ListPageBacklinksUseCase{repo: r}
}

// listPageBacklinksMaxResults は応答に含める逆リンク元ページの上限。
const listPageBacklinksMaxResults = 200

type ListPageBacklinksInput struct {
	WorkspaceID string
	UserID      uint64
	// PageID は逆リンクを求める対象（参照先）ページ。
	PageID string
}

func (u *ListPageBacklinksUseCase) Execute(ctx context.Context, in ListPageBacklinksInput) ([]domain.Page, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	if in.PageID == "" {
		return nil, errors.New("pageID is required")
	}
	rows, err := u.repo.ListPageLinkSourcePageViewFacts(ctx, in.WorkspaceID, in.UserID, in.PageID)
	if err != nil {
		return nil, err
	}
	pages := make([]domain.Page, 0, len(rows))
	for _, row := range rows {
		if !domain.ResolvePageView(row.Role, row.Page.Visibility, row.Page.CreatedByUserID == in.UserID) {
			continue
		}
		pages = append(pages, row.Page)
		if len(pages) >= listPageBacklinksMaxResults {
			break
		}
	}
	return pages, nil
}

// ListPagesReferencingTicketUseCase は ListPageBacklinksUseCase のチケット版
// （段 5）。対象チケットを本文の ticketRef ノードで埋め込んでいる
// （page_ticket_links.target_ticket_id = 対象チケット）ページのうち、閲覧できるものだけを
// 返す。判定・上限（listPageBacklinksMaxResults 流用）・SQL に LIMIT を掛けない理由は
// ListPageBacklinksUseCase と同一。
//
// 対象チケット自体を見られるかどうかの判定は handler（requireTicketPermission）が
// 先に行う。ticketID はここでは opaque な文字列として扱う — usecase/kb は usecase/ticket を
// import しない（サブパッケージ同士は import しない規約）ため、handler 層で
// usecase/ticket の権限判定と組み合わせて使う。
type ListPagesReferencingTicketUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListPagesReferencingTicketUseCase(r repository.KnowledgeBasePermissionRepository) *ListPagesReferencingTicketUseCase {
	return &ListPagesReferencingTicketUseCase{repo: r}
}

type ListPagesReferencingTicketInput struct {
	WorkspaceID string
	UserID      uint64
	TicketID    string
}

func (u *ListPagesReferencingTicketUseCase) Execute(ctx context.Context, in ListPagesReferencingTicketInput) ([]domain.Page, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	if in.TicketID == "" {
		return nil, errors.New("ticketID is required")
	}
	rows, err := u.repo.ListPageTicketLinkSourcePageViewFacts(ctx, in.WorkspaceID, in.UserID, in.TicketID)
	if err != nil {
		return nil, err
	}
	pages := make([]domain.Page, 0, len(rows))
	for _, row := range rows {
		if !domain.ResolvePageView(row.Role, row.Page.Visibility, row.Page.CreatedByUserID == in.UserID) {
			continue
		}
		pages = append(pages, row.Page)
		if len(pages) >= listPageBacklinksMaxResults {
			break
		}
	}
	return pages, nil
}

// ErrInvalidGrantRole は既知でない役割を指定したときに返す。
var ErrInvalidGrantRole = errors.New("invalid grant role")

// ErrInvalidCapability は既知でないケイパビリティを指定したときに返す。
var ErrInvalidCapability = errors.New("invalid capability")

// GrantWorkspaceRoleUseCase はワークスペース全体での既定の役割を主体に与える。
// 配下の全スペースに効くので、テナント全体の管理者はここで 1 行張れば足りる。
type GrantWorkspaceRoleUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewGrantWorkspaceRoleUseCase(r repository.KnowledgeBasePermissionRepository) *GrantWorkspaceRoleUseCase {
	return &GrantWorkspaceRoleUseCase{repo: r}
}

type GrantWorkspaceRoleInput struct {
	WorkspaceID string
	PrincipalID string
	Role        domain.GrantRole
	// ActorUserID は誰がこの役割を与えたか（段 6・監査）。
	ActorUserID uint64
}

func (u *GrantWorkspaceRoleUseCase) Execute(ctx context.Context, in GrantWorkspaceRoleInput) (*domain.WorkspaceGrant, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.PrincipalID == "" {
		return nil, errors.New("principalID is required")
	}
	if !in.Role.Valid() {
		return nil, ErrInvalidGrantRole
	}
	// 主体の実在とテナントの一致は DB の複合 FK でも守られるが、先に引いて
	// 「別ワークスペースの ID を渡した」を FK 違反ではなく not found として返す。
	if _, err := u.repo.FindPrincipal(ctx, in.WorkspaceID, in.PrincipalID); err != nil {
		return nil, err
	}
	return u.repo.UpsertWorkspaceGrant(ctx, in.WorkspaceID, in.PrincipalID, in.Role, in.ActorUserID)
}

// RevokeWorkspaceRoleUseCase はワークスペース全体での既定の役割を剥がす（冪等）。
type RevokeWorkspaceRoleUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewRevokeWorkspaceRoleUseCase(r repository.KnowledgeBasePermissionRepository) *RevokeWorkspaceRoleUseCase {
	return &RevokeWorkspaceRoleUseCase{repo: r}
}

type RevokeWorkspaceRoleInput struct {
	WorkspaceID string
	PrincipalID string
	// ActorUserID は誰がこの役割を剥がしたか（段 6・監査）。
	ActorUserID uint64
}

func (u *RevokeWorkspaceRoleUseCase) Execute(ctx context.Context, in RevokeWorkspaceRoleInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.PrincipalID == "" {
		return errors.New("principalID is required")
	}
	return u.repo.DeleteWorkspaceGrant(ctx, in.WorkspaceID, in.PrincipalID, in.ActorUserID)
}

// GrantSpaceRoleUseCase はスペースでの既定の役割を主体に与える。
type GrantSpaceRoleUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewGrantSpaceRoleUseCase(r repository.KnowledgeBasePermissionRepository) *GrantSpaceRoleUseCase {
	return &GrantSpaceRoleUseCase{repo: r}
}

type GrantSpaceRoleInput struct {
	WorkspaceID string
	SpaceID     string
	PrincipalID string
	Role        domain.GrantRole
}

func (u *GrantSpaceRoleUseCase) Execute(ctx context.Context, in GrantSpaceRoleInput) (*domain.SpaceGrant, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return nil, errors.New("spaceID is required")
	}
	if in.PrincipalID == "" {
		return nil, errors.New("principalID is required")
	}
	if !in.Role.Valid() {
		return nil, ErrInvalidGrantRole
	}
	if _, err := u.repo.FindPrincipal(ctx, in.WorkspaceID, in.PrincipalID); err != nil {
		return nil, err
	}
	return u.repo.UpsertSpaceGrant(ctx, in.WorkspaceID, in.SpaceID, in.PrincipalID, in.Role)
}

// RevokeSpaceRoleUseCase はスペースでの既定の役割を剥がす（冪等）。
type RevokeSpaceRoleUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewRevokeSpaceRoleUseCase(r repository.KnowledgeBasePermissionRepository) *RevokeSpaceRoleUseCase {
	return &RevokeSpaceRoleUseCase{repo: r}
}

type RevokeSpaceRoleInput struct {
	WorkspaceID string
	SpaceID     string
	PrincipalID string
}

func (u *RevokeSpaceRoleUseCase) Execute(ctx context.Context, in RevokeSpaceRoleInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return errors.New("spaceID is required")
	}
	if in.PrincipalID == "" {
		return errors.New("principalID is required")
	}
	return u.repo.DeleteSpaceGrant(ctx, in.WorkspaceID, in.SpaceID, in.PrincipalID)
}

// GrantPageRoleUseCase はページでの既定の役割を主体に与える。
//
// 既定の 3 段目（ワークスペース → スペース → ページ）で、このページとその子孫に効く。
// 合成は上の 2 段と同じで、複数の経路から届いた役割のうち最も強いものが実効になる。
//
// **これで誰かを弱めることはできない。** 上位で editor を得ている相手にここで viewer を
// 張っても editor のままで、下げたつもりが効かない。付与はどこまでも足し算だけで、
// 打ち消す層は持たない（domain.GrantRole.Rank と domain.PagePermissionFacts に規則と理由がある）。
// 狭めたい内容は private のスペースへ置く。
type GrantPageRoleUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewGrantPageRoleUseCase(r repository.KnowledgeBasePermissionRepository) *GrantPageRoleUseCase {
	return &GrantPageRoleUseCase{repo: r}
}

type GrantPageRoleInput struct {
	WorkspaceID string
	PageID      string
	PrincipalID string
	Role        domain.GrantRole
}

func (u *GrantPageRoleUseCase) Execute(ctx context.Context, in GrantPageRoleInput) (*domain.PageGrant, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.PageID == "" {
		return nil, errors.New("pageID is required")
	}
	if in.PrincipalID == "" {
		return nil, errors.New("principalID is required")
	}
	if !in.Role.Valid() {
		return nil, ErrInvalidGrantRole
	}
	if _, err := u.repo.FindPrincipal(ctx, in.WorkspaceID, in.PrincipalID); err != nil {
		return nil, err
	}
	return u.repo.UpsertPageGrant(ctx, in.WorkspaceID, in.PageID, in.PrincipalID, in.Role)
}

// RevokePageRoleUseCase はページでの既定の役割を剥がす（冪等）。
//
// 消えるのはこの段で足した分だけで、ワークスペース / スペース / 祖先のページから
// 届いている役割はそのまま残る。**「このページだけ見せない」は書けない** —
// 狭めたい内容は private のスペースへ置く。
//
// 「最後の admin」の検査は要らない。守っているのはワークスペースの admin が 0 人に
// なることで、ページの grant を全部消してもワークスペースの admin は配下の全ページに届く
// （RevokeSpaceRoleUseCase と同じ理由）。
type RevokePageRoleUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewRevokePageRoleUseCase(r repository.KnowledgeBasePermissionRepository) *RevokePageRoleUseCase {
	return &RevokePageRoleUseCase{repo: r}
}

type RevokePageRoleInput struct {
	WorkspaceID string
	PageID      string
	PrincipalID string
}

func (u *RevokePageRoleUseCase) Execute(ctx context.Context, in RevokePageRoleInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.PageID == "" {
		return errors.New("pageID is required")
	}
	if in.PrincipalID == "" {
		return errors.New("principalID is required")
	}
	return u.repo.DeletePageGrant(ctx, in.WorkspaceID, in.PageID, in.PrincipalID)
}

// ListPageGrantsUseCase はそのページ自身に張られた既定の役割の一覧を返す。
//
// **返るのは「このページを見られる人の一覧」ではない。** この段で足した行だけで、
// 上の段や祖先のページから届いている相手は含まれない。空で返ってきても
// 「誰も見られない」ではなく「この段では何も足していない」の意味になる。
// 呼び出し側（画面）はそれが分かる見せ方をすること。
type ListPageGrantsUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListPageGrantsUseCase(r repository.KnowledgeBasePermissionRepository) *ListPageGrantsUseCase {
	return &ListPageGrantsUseCase{repo: r}
}

type ListPageGrantsInput struct {
	WorkspaceID string
	PageID      string
}

func (u *ListPageGrantsUseCase) Execute(ctx context.Context, in ListPageGrantsInput) ([]domain.PageGrant, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.PageID == "" {
		return nil, errors.New("pageID is required")
	}
	return u.repo.ListPageGrants(ctx, in.WorkspaceID, in.PageID)
}

// ListGrantablePrincipalsUseCase は権限を張れる相手を表示名つきで返す。
//
// 返るのはワークスペース全体の主体で、ページでは絞らない。ページ単位の付与も
// 相手はワークスペースの主体だからで、ここで絞る意味が無い（絞ると
// 「同じ人に張れるはずなのに一覧に出ない」というずれが生まれる）。
//
// 呼べる範囲は handler 側の gate が決める。この一覧を使うのは「そのページの権限を
// 変えられる人」なので、認可もページ単位で掛ける。
type ListGrantablePrincipalsUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListGrantablePrincipalsUseCase(r repository.KnowledgeBasePermissionRepository) *ListGrantablePrincipalsUseCase {
	return &ListGrantablePrincipalsUseCase{repo: r}
}

type ListGrantablePrincipalsInput struct {
	WorkspaceID string
}

func (u *ListGrantablePrincipalsUseCase) Execute(
	ctx context.Context, in ListGrantablePrincipalsInput,
) ([]domain.GrantablePrincipal, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	return u.repo.ListGrantablePrincipals(ctx, in.WorkspaceID)
}

// ListWorkspaceMembersUseCase はワークスペースに属する人を表示名つきで返す。
//
// ListGrantablePrincipalsUseCase とは呼べる範囲が違う。あちらは権限を張る画面のための
// 一覧で、認可をページの管理権限で掛ける。こちらは担当の表示名と発言での名指しに使うので、
// 所属していれば読める（既定の役割は編集者で、管理権限は持たない）。
type ListWorkspaceMembersUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListWorkspaceMembersUseCase(r repository.KnowledgeBasePermissionRepository) *ListWorkspaceMembersUseCase {
	return &ListWorkspaceMembersUseCase{repo: r}
}

func (u *ListWorkspaceMembersUseCase) Execute(ctx context.Context, workspaceID string) ([]domain.WorkspaceMember, error) {
	if workspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	return u.repo.ListWorkspaceMembers(ctx, workspaceID)
}

// ListWorkspaceMembersForAdminUseCase はメンバー管理画面（段 7）向けの一覧を返す。
// ListWorkspaceMembersUseCase と違い、停止中のアカウントも含み、現在のワークスペース
// 全体の役割も一緒に返す。呼べるのは admin だけ（handler 側の判定を参照）。
type ListWorkspaceMembersForAdminUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListWorkspaceMembersForAdminUseCase(r repository.KnowledgeBasePermissionRepository) *ListWorkspaceMembersForAdminUseCase {
	return &ListWorkspaceMembersForAdminUseCase{repo: r}
}

func (u *ListWorkspaceMembersForAdminUseCase) Execute(ctx context.Context, workspaceID string) ([]domain.AdminWorkspaceMember, error) {
	if workspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	return u.repo.ListWorkspaceMembersForAdmin(ctx, workspaceID)
}

// ListSpaceMembersUseCase はそのスペースに届いている権限を人に解決して返す（段 9）。
// 可視判定（CanView）は handler 側が checkSpace で確かめてから呼ぶ（RenameSpace と同じ形）。
type ListSpaceMembersUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListSpaceMembersUseCase(r repository.KnowledgeBasePermissionRepository) *ListSpaceMembersUseCase {
	return &ListSpaceMembersUseCase{repo: r}
}

func (u *ListSpaceMembersUseCase) Execute(ctx context.Context, workspaceID, spaceID string) ([]domain.SpaceMember, error) {
	if workspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if spaceID == "" {
		return nil, errors.New("spaceID is required")
	}
	return u.repo.ListSpaceMembers(ctx, workspaceID, spaceID)
}

// ErrPrincipalKindMismatch は主体の種類が操作に合わないときに返す
// （グループでないものをグループとして扱おうとした等）。
var ErrPrincipalKindMismatch = errors.New("principal kind does not match the operation")

// kbGroupNameMaxLen は principals.name (varchar(200)) の上限。DB エラーの前に入口で弾く。
const kbGroupNameMaxLen = 200

// RemoveWorkspaceMemberUseCase はユーザーをワークスペースから外す（招待中なら取り消す）。
// principal があれば消え、その人に張られていた grant / グループ所属も FK の CASCADE で
// 消える（権限だけが残らない）。workspace_members は消さず left として記録に残す
// （repository.LeaveWorkspaceMembership 参照）。
type RemoveWorkspaceMemberUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewRemoveWorkspaceMemberUseCase(r repository.KnowledgeBasePermissionRepository) *RemoveWorkspaceMemberUseCase {
	return &RemoveWorkspaceMemberUseCase{repo: r}
}

type RemoveWorkspaceMemberInput struct {
	WorkspaceID string
	UserID      uint64
	// ActorUserID は誰がこの操作をしたか（段 6・監査）。UserID と同じなら本人の退会、
	// 違えば admin による除名として記録される（repository.LeaveWorkspaceMembership 参照）。
	ActorUserID uint64
}

func (u *RemoveWorkspaceMemberUseCase) Execute(ctx context.Context, in RemoveWorkspaceMemberInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.UserID == 0 {
		return errors.New("userID is required")
	}
	return u.repo.LeaveWorkspaceMembership(ctx, in.WorkspaceID, in.UserID, in.ActorUserID)
}

// ListMembershipEventsUseCase は所属・権限の変更履歴を新しい順で返す（段 6・監査）。
// 「なぜこの人が admin なのか」を後から説明できるようにするための読み取り専用の口。
type ListMembershipEventsUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListMembershipEventsUseCase(r repository.KnowledgeBasePermissionRepository) *ListMembershipEventsUseCase {
	return &ListMembershipEventsUseCase{repo: r}
}

func (u *ListMembershipEventsUseCase) Execute(ctx context.Context, workspaceID string) ([]domain.MembershipEvent, error) {
	if workspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	return u.repo.ListMembershipEvents(ctx, workspaceID)
}

// CreatePrincipalGroupUseCase は権限をまとめて張るためのグループを作る。
// 名前はワークスペース内で一意（同名が 2 つあると権限を張る先を人が選べない）。
type CreatePrincipalGroupUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewCreatePrincipalGroupUseCase(r repository.KnowledgeBasePermissionRepository) *CreatePrincipalGroupUseCase {
	return &CreatePrincipalGroupUseCase{repo: r}
}

type CreatePrincipalGroupInput struct {
	WorkspaceID string
	Name        string
}

func (u *CreatePrincipalGroupUseCase) Execute(ctx context.Context, in CreatePrincipalGroupInput) (*domain.Principal, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.Name == "" {
		return nil, errors.New("name is required")
	}
	if utf8.RuneCountInString(in.Name) > kbGroupNameMaxLen {
		return nil, errors.New("name is too long")
	}
	return u.repo.CreateGroupPrincipal(ctx, in.WorkspaceID, in.Name)
}

// AddGroupMemberUseCase はグループにユーザーを加える。
//
// 加える相手を主体 ID ではなくユーザー ID で受けるのは、グループの入れ子をこの入口から
// 作れないようにするため（DB 側も複合 FK で member を kind='user' に固定している）。
// 入れ子を許すと権限解決に再帰が要り、グループ同士の循環も防がなければならなくなる。
type AddGroupMemberUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewAddGroupMemberUseCase(r repository.KnowledgeBasePermissionRepository) *AddGroupMemberUseCase {
	return &AddGroupMemberUseCase{repo: r}
}

type AddGroupMemberInput struct {
	WorkspaceID string
	// GroupPrincipalID は kind='group' の主体。
	GroupPrincipalID string
	// MemberUserID は加えるユーザー。メンバーでなければ主体が無いのでエラーになる。
	MemberUserID uint64
}

func (u *AddGroupMemberUseCase) Execute(ctx context.Context, in AddGroupMemberInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.GroupPrincipalID == "" {
		return errors.New("groupPrincipalID is required")
	}
	if in.MemberUserID == 0 {
		return errors.New("memberUserID is required")
	}
	group, err := u.repo.FindPrincipal(ctx, in.WorkspaceID, in.GroupPrincipalID)
	if err != nil {
		return err
	}
	if group.Kind != domain.PrincipalKindGroup {
		return ErrPrincipalKindMismatch
	}
	member, err := u.repo.FindUserPrincipal(ctx, in.WorkspaceID, in.MemberUserID)
	if err != nil {
		return err
	}
	return u.repo.AddGroupMember(ctx, in.WorkspaceID, group.ID, member.ID)
}

// RemoveGroupMemberUseCase はグループからユーザーを外す（冪等）。
type RemoveGroupMemberUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewRemoveGroupMemberUseCase(r repository.KnowledgeBasePermissionRepository) *RemoveGroupMemberUseCase {
	return &RemoveGroupMemberUseCase{repo: r}
}

type RemoveGroupMemberInput struct {
	WorkspaceID      string
	GroupPrincipalID string
	MemberUserID     uint64
}

func (u *RemoveGroupMemberUseCase) Execute(ctx context.Context, in RemoveGroupMemberInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.GroupPrincipalID == "" {
		return errors.New("groupPrincipalID is required")
	}
	if in.MemberUserID == 0 {
		return errors.New("memberUserID is required")
	}
	member, err := u.repo.FindUserPrincipal(ctx, in.WorkspaceID, in.MemberUserID)
	if err != nil {
		if errors.Is(err, repository.ErrPrincipalNotFound) {
			return nil // 非メンバーはどのグループにも属していない（冪等）
		}
		return err
	}
	return u.repo.RemoveGroupMember(ctx, in.WorkspaceID, in.GroupPrincipalID, member.ID)
}

// EnsureSpaceEveryonePrincipalUseCase はスペースの「全員」を表す主体を用意する（冪等）。
// 「既定でチーム全員が編集できる」を 1 行の grant で表すための下ごしらえ。
type EnsureSpaceEveryonePrincipalUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewEnsureSpaceEveryonePrincipalUseCase(r repository.KnowledgeBasePermissionRepository) *EnsureSpaceEveryonePrincipalUseCase {
	return &EnsureSpaceEveryonePrincipalUseCase{repo: r}
}

type EnsureSpaceEveryonePrincipalInput struct {
	WorkspaceID string
	SpaceID     string
}

func (u *EnsureSpaceEveryonePrincipalUseCase) Execute(ctx context.Context, in EnsureSpaceEveryonePrincipalInput) (*domain.Principal, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return nil, errors.New("spaceID is required")
	}
	return u.repo.EnsureSpaceEveryonePrincipal(ctx, in.WorkspaceID, in.SpaceID)
}

// CanRemoveWorkspaceAdminUseCase は「この相手からワークスペースの admin を外しても、
// admin が 1 人以上残るか」に答える。権限を減らす操作の前に呼び、false なら断る。
//
// # なぜこの問いが要るのか（「最後の admin」を剥がせなくする理由）
//
// ナレッジの権限は principals / grants だけで閉じており、
// 「アプリの super_admin なら通る」という抜け道を意図的に持たない（domain/grant.go）。
// その裏返しとして、ワークスペースの admin が 0 人になった瞬間、そのワークスペースの
// 権限を変えられる人は API のどこにも居なくなる。スペースを増やすことも、
// 誰かに権限を戻すこともできず、復旧手段は DB を直接触ることだけになる。
//
// 逆に「最後の 1 人は自分を外せない」で詰まる場面は、先に別の誰かへ admin を渡せば
// 必ず解ける。取り返しがつかない側（0 人）を禁じ、手数が 1 つ増えるだけの側を許す。
//
// # 何を数えるか
//
// 数えるのは kind='user' の主体が持つ admin だけ。グループ宛ての admin を数に入れると、
// メンバーが 1 人も居ないグループが「最後の admin」として残り、結局誰も権限を
// 変えられないワークスペースが同じようにできてしまう（grant の行からはグループの
// 中身が分からない）。その分だけ判定は厳しくなるが、余計に断られるのは
// 「グループ経由の admin しか居ないのに、ユーザー宛ての admin を外そうとした」場合だけで、
// 誰か 1 人に admin を張れば必ず通る。安全側に外れる。
//
// # 競合について（この usecase では守れないこと）
//
// この確認と実際の書き換えは別のトランザクションなので、**ここだけでは競合を防げない。**
// admin 2 人をほぼ同時に外す 2 本の要求は、両方ともこの検査を通り抜けて両方成功し得る
// （実測: 2 本同時に流すと 60 回中 59 回 admin が 0 人になった）。
//
// 実際に 0 人を防いでいるのは repository 側で、判定と書き換えを同じトランザクションに入れ、
// admin の行を FOR UPDATE でロックしてから決める（persistence の withLastAdminGuard）。
// 競合で断られたときは repository.ErrLastWorkspaceAdmin が返り、handler はこの usecase が
// false を返したときと同じ 409 に落とす。
//
// ではなぜこの usecase を残すのか。日常の誤操作（1 人しか居ないと分かっている状態で外す）を
// **書き換えを 1 行も試みる前に**断れるからで、応答が競合の有無で揺れない。
// 「読んで確かめる口」と「書きながら守る歯止め」は役割が違い、後者だけで足りるわけではない。
type CanRemoveWorkspaceAdminUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewCanRemoveWorkspaceAdminUseCase(r repository.KnowledgeBasePermissionRepository) *CanRemoveWorkspaceAdminUseCase {
	return &CanRemoveWorkspaceAdminUseCase{repo: r}
}

// CanRemoveWorkspaceAdminInput は対象を主体 ID かユーザー ID のどちらかで指す。
// grant の取り消しは主体 ID を、メンバーの削除はユーザー ID を持っているため両方を受ける
// （メンバーを消すと principal ごと消え、その主体の grant も CASCADE で消えるので、
// 「grant を外す」と同じ影響がある）。
type CanRemoveWorkspaceAdminInput struct {
	WorkspaceID string
	// PrincipalID は admin を外す相手の主体。空なら UserID から引き直す。
	PrincipalID string
	// UserID は admin を外す相手をユーザーで指すときに使う。
	UserID uint64
}

func (u *CanRemoveWorkspaceAdminUseCase) Execute(ctx context.Context, in CanRemoveWorkspaceAdminInput) (bool, error) {
	if in.WorkspaceID == "" {
		return false, errors.New("workspaceID is required")
	}
	if in.PrincipalID == "" && in.UserID == 0 {
		return false, errors.New("principalID or userID is required")
	}

	target := in.PrincipalID
	if target == "" {
		principal, err := u.repo.FindUserPrincipal(ctx, in.WorkspaceID, in.UserID)
		if err != nil {
			if errors.Is(err, repository.ErrPrincipalNotFound) {
				// 非メンバーは grant を 1 つも持たない。外しても admin は減らない。
				return true, nil
			}
			return false, err
		}
		target = principal.ID
	}

	grants, err := u.repo.ListWorkspaceGrants(ctx, in.WorkspaceID)
	if err != nil {
		return false, err
	}
	targetIsAdmin := false
	others := make([]string, 0, len(grants))
	for _, g := range grants {
		if g.Role != domain.GrantRoleAdmin {
			continue
		}
		if g.PrincipalID == target {
			targetIsAdmin = true
			continue
		}
		others = append(others, g.PrincipalID)
	}
	if !targetIsAdmin {
		// 元から admin ではない相手なので、この操作で admin は 1 人も減らない。
		return true, nil
	}

	// 残る admin のうち、実際に人が入っていると確実に言えるもの（kind='user'）を探す。
	// 1 人でも見つかればそこで打ち切る（admin は多くないが、全件引く必要もない）。
	for _, principalID := range others {
		p, err := u.repo.FindPrincipal(ctx, in.WorkspaceID, principalID)
		if err != nil {
			if errors.Is(err, repository.ErrPrincipalNotFound) {
				// grant を読んでから主体を引くまでの間に消えた。数に入れない（安全側）。
				continue
			}
			return false, err
		}
		if p.Kind == domain.PrincipalKindUser {
			return true, nil
		}
	}
	return false, nil
}

// CheckSpacePermissionUseCase は「このユーザーはこのスペースで既定で何ができるか」に答える。
//
// ページを名指しできない操作（スペース直下へのページ作成）の入口で使う。
// **ページの可否をこれで決めてはいけない。** ここが集める事実にはページ付与
// （page_grants）が入っていないので、祖先のページで足された役割を取りこぼし、
// 必ず狭い側へ倒れる。ページには CheckPagePermissionUseCase を使う。
//
// 判定規則は domain.ResolveScopePermission にあり、ここには写経しない。
type CheckSpacePermissionUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewCheckSpacePermissionUseCase(r repository.KnowledgeBasePermissionRepository) *CheckSpacePermissionUseCase {
	return &CheckSpacePermissionUseCase{repo: r}
}

type CheckSpacePermissionInput struct {
	WorkspaceID string
	SpaceID     string
	UserID      uint64
}

func (u *CheckSpacePermissionUseCase) Execute(ctx context.Context, in CheckSpacePermissionInput) (*domain.ScopePermission, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return nil, repository.ErrSpaceNotFound
	}
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	facts, err := u.repo.SpacePermissionFactsForUser(ctx, in.WorkspaceID, in.SpaceID, in.UserID)
	if err != nil {
		return nil, err
	}
	perm := domain.ResolveScopePermission(*facts)
	return &perm, nil
}

// CheckWorkspacePermissionUseCase は「このユーザーはこのワークスペースで既定で何ができるか」に答える。
// どのスペースにも属さない操作（スペースの作成）の入口で使う。
type CheckWorkspacePermissionUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewCheckWorkspacePermissionUseCase(r repository.KnowledgeBasePermissionRepository) *CheckWorkspacePermissionUseCase {
	return &CheckWorkspacePermissionUseCase{repo: r}
}

type CheckWorkspacePermissionInput struct {
	WorkspaceID string
	UserID      uint64
}

func (u *CheckWorkspacePermissionUseCase) Execute(ctx context.Context, in CheckWorkspacePermissionInput) (*domain.ScopePermission, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	facts, err := u.repo.WorkspacePermissionFactsForUser(ctx, in.WorkspaceID, in.UserID)
	if err != nil {
		return nil, err
	}
	perm := domain.ResolveScopePermission(*facts)
	return &perm, nil
}

// ListViewableSpacesUseCase はワークスペース配下のスペースのうち、そのユーザーが
// 中身を閲覧できるものだけを返す。
//
// # なぜ権限でふるうのか
//
// スペースは「誰に何を見せるか」を分ける入れ物そのもの。人事のスペース・経営のスペースを
// 作って役割を絞る、という使い方が前提なので、**一覧が権限を無視すると入れ物の意味が消える**。
// 中身（ページ）が見えなくても、key と name が並べば「人事」「M&A 準備」といった名前から
// 何が進行中かが伝わる。名前そのものが情報になる以上、見せてよい相手を選ぶ必要がある。
//
// # 判定は domain、SQL は事実だけ
//
// repository が返すのは「そのスペースに届いている既定の役割の集合」まで。
// どう畳んで何を許すかは domain.ResolveScopePermission だけが持つ。ここに
// 「admin なら〜」を書き足すと、同じ役割の意味がスペース 1 つの解決（CheckSpacePermissionUseCase）
// と一覧で食い違い、「開けるのに一覧に出ない」「一覧に出るのに開けない」というずれ方をする。
//
// # ページ付与は見ていない
//
// 使うのは ScopePermission なので、ページに張った付与（page_grants）は一切見ない。
// これは正しい。スペースが見えるかは、そのスペース自体に届いている役割で決まるため。
// 逆に**この結果をページの可否に使ってはいけない**（必ず狭い側へ倒れる）。
type ListViewableSpacesUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListViewableSpacesUseCase(r repository.KnowledgeBasePermissionRepository) *ListViewableSpacesUseCase {
	return &ListViewableSpacesUseCase{repo: r}
}

type ListViewableSpacesInput struct {
	WorkspaceID string
	UserID      uint64
}

// Execute は閲覧できるスペースだけを返す（repository が返す順序＝ key 順を保つ）。
//
// 存在しないワークスペースでも空スライスを返す（エラーにしない）。実在を撃ち分けるのは
// URL の slug を解決する middleware の仕事で、そこで所属していないワークスペースと
// 存在しないワークスペースはどちらも 404 に畳まれている。ここで別の応答を作ると、
// せっかく畳んだ差がこの口だけで復活する。
func (u *ListViewableSpacesUseCase) Execute(ctx context.Context, in ListViewableSpacesInput) ([]domain.Space, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	rows, err := u.repo.ListWorkspaceSpaceScopeFacts(ctx, in.WorkspaceID, in.UserID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Space, 0, len(rows))
	for _, row := range rows {
		// ここが権限のふるい。repository は役割の届いていないスペースも返してくるので、
		// これを外すと閲覧権限の無いスペースがそのまま応答に載る。
		if !domain.ResolveScopePermission(row.Facts).CanView {
			continue
		}
		out = append(out, row.Space)
	}
	return out, nil
}

// ListMemberWorkspacesUseCase は自分が所属するワークスペースを返す。
// ナレッジのほかの経路と違い URL に slug を持たない（どの slug を開けるかを知るための口）。
type ListMemberWorkspacesUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListMemberWorkspacesUseCase(r repository.KnowledgeBasePermissionRepository) *ListMemberWorkspacesUseCase {
	return &ListMemberWorkspacesUseCase{repo: r}
}

type ListMemberWorkspacesInput struct {
	UserID uint64
}

func (u *ListMemberWorkspacesUseCase) Execute(ctx context.Context, in ListMemberWorkspacesInput) ([]domain.MemberWorkspace, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListMemberWorkspaces(ctx, in.UserID)
}
