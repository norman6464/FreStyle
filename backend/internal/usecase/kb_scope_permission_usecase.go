package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// CheckPageSpacePermissionUseCase は「このユーザーは、**このページが属するスペース**で
// 既定で何ができるか」に答える。ページを名指しする権限操作の入口で使う。
//
// CheckSpacePermissionUseCase との違いは、スペース ID を引数で受けずページから引くこと。
// 手間を省くためではなく、**問い合わせの回数を一定にする**ため。以前は
// 「ページを引く → スペースの実在を確かめる → 役割を集める」の 3 段で、どれも同じ 404 を
// 返すのに落ちる段によって DB の往復が 0 / 1 / 3 回に分かれ、応答のバイト列を揃えても
// 返るまでの時間から「そのページ ID が実在するか」が読めた。
//
// **ページ単位の例外は見ていない**（CheckSpacePermissionUseCase と同じ制限）。
// ページの閲覧・編集の可否をこれで決めてはいけない。権限そのものを変える操作の入口
// （admin かどうか）にだけ使う — 例外はページの閲覧・編集にしか効かず、admin かどうかは
// スペースの既定だけで決まるため。
//
// 判定規則は domain.ResolveScopePermission にあり、ここには写経しない。
type CheckPageSpacePermissionUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewCheckPageSpacePermissionUseCase(r repository.KnowledgeBasePermissionRepository) *CheckPageSpacePermissionUseCase {
	return &CheckPageSpacePermissionUseCase{repo: r}
}

type CheckPageSpacePermissionInput struct {
	WorkspaceID string
	PageID      string
	UserID      uint64
}

// CheckPageSpacePermissionOutput は判定結果と、ページが属するスペースの ID。
//
// SpaceID は通ったときだけ意味を持つ。**通らなかったときは空文字**で、それが
// 「ページが無い」なのか「役割が無い」なのかは呼び出し側からは分からない（分かってはいけない）。
type CheckPageSpacePermissionOutput struct {
	SpaceID    string
	Permission domain.ScopePermission
}

func (u *CheckPageSpacePermissionUseCase) Execute(
	ctx context.Context, in CheckPageSpacePermissionInput,
) (*CheckPageSpacePermissionOutput, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	// PageID が空でもエラーにせず repository へ渡す。空文字は UUID として解釈できないので
	// 見つからなかったときと同じ空が返り、拒否へ落ちる。ここで別のエラーにすると、
	// 空文字だけ応答の作られ方が変わる。
	facts, err := u.repo.PageSpaceScopeFactsForUser(ctx, in.WorkspaceID, in.PageID, in.UserID)
	if err != nil {
		return nil, err
	}
	return &CheckPageSpacePermissionOutput{
		SpaceID:    facts.SpaceID,
		Permission: domain.ResolveScopePermission(facts.Facts),
	}, nil
}

// CheckSpacePermissionUseCase は「このユーザーはこのスペースで既定で何ができるか」に答える。
//
// ページを名指しできない操作（スペース直下へのページ作成）の入口で使う。
// **ページの可否をこれで決めてはいけない。** スペースにはページ単位の例外
// （page_restrictions）の層が無く、集めた事実にも入っていないので、あるページで
// deny されていても CanEdit が true のまま返る。ページには
// CheckPagePermissionUseCase（domain.ResolvePagePermission）を使う。
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
// # ページ単位の例外は見ていない
//
// 使うのは ScopePermission なので、ページに張った例外（page_restrictions）は一切見ない。
// これは正しい。スペースの中の 1 枚が deny されていても、そのスペース自体が見えないことには
// ならないため。逆に**この結果をページの可否に使ってはいけない**（必ず緩い側へ倒れる）。
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
// ナレッジ基盤のほかの経路と違い URL に slug を持たない（どの slug を開けるかを知るための口）。
type ListMemberWorkspacesUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListMemberWorkspacesUseCase(r repository.KnowledgeBasePermissionRepository) *ListMemberWorkspacesUseCase {
	return &ListMemberWorkspacesUseCase{repo: r}
}

type ListMemberWorkspacesInput struct {
	UserID uint64
}

func (u *ListMemberWorkspacesUseCase) Execute(ctx context.Context, in ListMemberWorkspacesInput) ([]domain.Workspace, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListMemberWorkspaces(ctx, in.UserID)
}
