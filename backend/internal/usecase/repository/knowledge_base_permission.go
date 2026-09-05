package repository

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ErrPrincipalNotFound は対象の主体が存在しない（または別ワークスペースのもの）ときに返す。
var ErrPrincipalNotFound = errors.New("principal not found")

// ErrUserNotFound は主体を作ろうとしたユーザーが users に存在しないときに返す。
//
// principals.user_id は users への FK なので、実在しないユーザー ID で
// EnsureUserPrincipal を呼ぶと制約違反になる。それをそのまま上へ流すと
// 「ユーザー ID を間違えた」という入力の誤りが 500 になり、呼び出し側は
// DB 障害と区別できない（再試行すべきだと誤解する）。
var ErrUserNotFound = errors.New("user not found")

// ErrLastWorkspaceAdmin は「ユーザーの admin が 1 人も残らなくなる操作」を断ったときに返す。
//
// ノートの権限は principals / grants だけで閉じており、
// 「アプリの super_admin なら通る」という抜け道を意図的に持たない（domain/grant.go）。
// その裏返しとして、ワークスペースの admin が 0 人になった瞬間、そのワークスペースの
// 権限を変えられる人は API のどこにも居なくなる。**元 admin を含めて誰も復旧できず、
// DB を直接触るしか手が無い。** 逆に「最後の 1 人は自分を外せない」で詰まる場面は、
// 先に別の誰かへ admin を渡せば必ず解ける。取り返しがつかない側を禁じる。
//
// このセンチネルは repository（＝ 実際に行を書き換える層）が返す。手前の usecase
// （CanRemoveWorkspaceAdminUseCase）も同じ判定を持つが、あちらは操作の前に読むだけなので
// 競合を防げない。最後の砦は書き込みと同じトランザクションで判定するこちら側にある。
var ErrLastWorkspaceAdmin = errors.New("last workspace admin cannot be removed")

// ErrPrincipalGroupNameTaken はグループ名が同じワークスペースで使用済みのときに返す。
// 名前はワークスペース内で一意（uq_principals_group_name）で、同名が 2 つあると
// 権限を張る先を人が選べなくなる。
var ErrPrincipalGroupNameTaken = errors.New("principal group name is already taken")

// PageWithViewFacts は 1 ページと、そのページを閲覧できるかを決める事実の組。
// ListSpacePageViewFacts が返す（ふるい落としは domain.ResolvePageView が行う）。
//
// 事実が役割 1 つなのは、権限が打ち消しを持たないため。届いた中で最も強い役割だけで
// 閲覧可否が決まり、経路のどこで得たかは結果に影響しない。
type PageWithViewFacts struct {
	Page domain.Page
	// Role は届いた中で最も強い役割。grant が 1 つも無ければ nil。
	Role *domain.GrantRole
	// ParentArchived は親がアーカイブ済みか（親を持たない行は false）。
	//
	// これは**事実**で、判断ではない。「復帰できるか」の規則は UnarchivePageUseCase が
	// 持っている（親がアーカイブ中なら断る）。ここで canRestore のような名前にすると、
	// 同じ規則が 2 箇所に置かれて必ずずれる。
	ParentArchived bool
}

// PageWithPermissionFacts は 1 ページの ID と、その実効権限を決める事実の組。
// ListSubtreePagePermissionFacts が返す（判定は domain.ResolvePagePermission が行う）。
//
// 閲覧専用の PageWithViewFacts と分かれているのは、こちらが所属（Member）も集めるため。
// ページ本体を持たないのは、この型を使う経路（サブツリー一括操作の入口検査）が
// 可否だけを必要とし、見えないページの中身を呼び出し側へ渡す必要が無いため。
type PageWithPermissionFacts struct {
	PageID string
	Facts  domain.PagePermissionFacts
}

// SpaceWithScopeFacts は 1 スペースと、その入れ物に対する実効権限を決める事実の組。
// ListWorkspaceSpaceScopeFacts が返す（判定は domain.ResolveScopePermission が行う）。
//
// ページの事実（PageWithViewFacts / PageWithPermissionFacts）と型を分けているのは、
// 集めた事実が違うため。ここにあるのはワークスペース / スペースの grants で届いた
// 役割だけで、ページ付与（page_grants）は含まない。同じ型に載せると
// 「ページ付与を見ていない」ことが「ページ付与が無い」に化ける。
//
// スペース本体を持つのは、これが一覧の材料そのものだから（呼び出し側は key / name を返す）。
// 事実の側で見えないスペースをふるい落とすのは呼び出し側の責務で、
// この型が返ってきた時点ではまだ「見せてよいスペース」に絞られていない。
type SpaceWithScopeFacts struct {
	Space domain.Space
	Facts domain.ScopeFacts
}

// KnowledgeBasePermissionRepository はノートの権限モデル（principals /
// principal_members / workspace_grants / space_grants / page_grants）への
// アクセスを提供する（share_links は [ShareLinkRepository] が持つ）。
//
// KnowledgeBaseRepository（ページとブロック）と分けているのは、境界が違うため。
// あちらをひとつの fat interface にまとめている理由は「ページ作成 = pages + page_paths」
// のように複数テーブルを 1 トランザクションで書く操作が中心だからで、権限の書き込みは
// それらと同じトランザクションに入らない（権限を張る操作とページを書く操作は別の要求）。
// 同じ interface に足すと、ページだけを扱う実装や fake が権限のメソッドまで
// 実装しなければならなくなり、境界の意味も薄れる。
//
// 読み取りは pages / page_paths をまたぐ（実効権限の解決に closure が要る）が、
// 境界を決めるのは書き込みのトランザクション単位なので問題にしない。
type KnowledgeBasePermissionRepository interface {
	// EnsureUserPrincipal はユーザーの主体を作る（既にあればそれを返す）。
	// この行があること自体がワークスペース所属を意味する。
	EnsureUserPrincipal(ctx context.Context, workspaceID string, userID uint64) (*domain.Principal, error)
	// EnsureSpaceEveryonePrincipal はスペースの「全員」を表す主体を作る（既にあればそれを返す）。
	EnsureSpaceEveryonePrincipal(ctx context.Context, workspaceID, spaceID string) (*domain.Principal, error)
	// CreateGroupPrincipal はグループの主体を作る。名前はワークスペース内で一意。
	CreateGroupPrincipal(ctx context.Context, workspaceID, name string) (*domain.Principal, error)
	// FindPrincipal は主体を 1 件引く。無い・別ワークスペースなら ErrPrincipalNotFound。
	FindPrincipal(ctx context.Context, workspaceID, principalID string) (*domain.Principal, error)
	// FindUserPrincipal はユーザーの主体を引く。無ければ ErrPrincipalNotFound（= 非メンバー）。
	FindUserPrincipal(ctx context.Context, workspaceID string, userID uint64) (*domain.Principal, error)
	// DeletePrincipal は主体を消す。紐づく grant / グループ所属も
	// FK の CASCADE で消える。対象が無ければ ErrPrincipalNotFound。
	//
	// grant も CASCADE で消えるので、これはワークスペースの admin を減らし得る操作でもある。
	// ユーザーの admin が 0 人になるなら ErrLastWorkspaceAdmin を返して何も消さない
	// （判定・ロック・削除はすべて同じトランザクション）。
	DeletePrincipal(ctx context.Context, workspaceID, principalID string) error
	// IsWorkspaceMember はユーザーがワークスペースのメンバーかを返す。
	IsWorkspaceMember(ctx context.Context, workspaceID string, userID uint64) (bool, error)
	// ListMemberWorkspaces はそのユーザーが所属するワークスペースと、そこでの CanManage
	// （DeleteWorkspace が要求する admin 権限と同じ）を返す（slug 順）。
	// 所属は principals（kind='user'）の行が唯一の表現なので、その JOIN がそのまま答えになる。
	// ノートで唯一テナントを跨いで読むメソッド（どのテナントに入れるかを答える口）で、
	// 絞り込みは user_id だけが行う。
	ListMemberWorkspaces(ctx context.Context, userID uint64) ([]domain.MemberWorkspace, error)

	// AddGroupMember はグループに主体を所属させる（冪等）。member 側は kind='user' でなければ
	// DB の複合 FK が弾く（グループの入れ子を作らせない）。
	AddGroupMember(ctx context.Context, workspaceID, groupPrincipalID, memberPrincipalID string) error
	// RemoveGroupMember はグループから主体を外す。存在しなければ何もしない（冪等）。
	RemoveGroupMember(ctx context.Context, workspaceID, groupPrincipalID, memberPrincipalID string) error

	// UpsertWorkspaceGrant はワークスペース全体での既定の役割を与える（同じ主体には 1 行だけ）。
	// admin から他の役割へ落とす向きは「admin を外す」操作なので、それでユーザーの admin が
	// 0 人になるなら ErrLastWorkspaceAdmin を返して何も書かない（判定は書き込みと同じトランザクション）。
	UpsertWorkspaceGrant(ctx context.Context, workspaceID, principalID string, role domain.GrantRole) (*domain.WorkspaceGrant, error)
	// GrantWorkspaceRoleIfAbsent は既定の役割を**無いときだけ**与える（既存の行は触らない）。
	// メンバー追加の既定 editor 用。上書きの Upsert を使うと、冪等な追加のやり直しで
	// admin が editor に落ちる（最後の admin なら保護の検査に当たって追加自体が失敗する）。
	GrantWorkspaceRoleIfAbsent(ctx context.Context, workspaceID, principalID string, role domain.GrantRole) error
	// DeleteWorkspaceGrant はワークスペース全体での既定の役割を剥がす（冪等）。
	// これでユーザーの admin が 0 人になるなら ErrLastWorkspaceAdmin を返して何も書かない。
	DeleteWorkspaceGrant(ctx context.Context, workspaceID, principalID string) error
	// ListWorkspaceGrants はワークスペースの grant 一覧を返す。
	ListWorkspaceGrants(ctx context.Context, workspaceID string) ([]domain.WorkspaceGrant, error)

	// UpsertSpaceGrant はスペースでの既定の役割を与える（同じ主体には 1 行だけ）。
	UpsertSpaceGrant(ctx context.Context, workspaceID, spaceID, principalID string, role domain.GrantRole) (*domain.SpaceGrant, error)
	// DeleteSpaceGrant はスペースでの既定の役割を剥がす（冪等）。
	DeleteSpaceGrant(ctx context.Context, workspaceID, spaceID, principalID string) error
	// ListSpaceGrants はスペースの grant 一覧を返す。
	ListSpaceGrants(ctx context.Context, workspaceID, spaceID string) ([]domain.SpaceGrant, error)

	// UpsertPageGrant はページでの既定の役割を与える（同じ主体には 1 行だけ）。
	//
	// workspace / space に続く 3 段目で、このページとその子孫に効く。合成は他の 2 段と
	// 同じ「最も強いものを採る」なので、**これで誰かを弱めることはできない**
	// （上位で editor を得ている相手にここで viewer を張っても editor のまま）。
	// **弱める手段はどの層にも無い。** 狭めたい内容は private のスペースへ置く。
	UpsertPageGrant(ctx context.Context, workspaceID, pageID, principalID string, role domain.GrantRole) (*domain.PageGrant, error)
	// DeletePageGrant はページでの既定の役割を剥がす（冪等）。
	// 上位の段で得ている役割はそのまま残る（消えるのはこの段で足した分だけ）。
	DeletePageGrant(ctx context.Context, workspaceID, pageID, principalID string) error
	// ListGrantablePrincipals は権限を張れる相手を表示名つきで返す（kind → 名前 → id 順）。
	//
	// share_link は含まない。あれはリンクを踏んだ来訪者を表す主体で、リンクの発行時に
	// 自動で作られる。人が選んで役割を与える相手ではない。
	//
	// 名前が引けなかった行も落とさず、Name を空文字にして返す。一覧から黙って消すと、
	// その主体に張った権限が画面に出たまま選べない（取り消せない行）になる。
	ListGrantablePrincipals(ctx context.Context, workspaceID string) ([]domain.GrantablePrincipal, error)
	// ListPageGrants はそのページ自身に張られた grant の一覧を返す（継承分は含まない）。
	//
	// **これは「このページを見られる人の一覧」ではない。** 返るのはこの段で足した行だけで、
	// ワークスペース / スペースの grant で届いている相手も、祖先のページに張られた
	// grant で届いている相手も含まれない。空で返ってきても「誰も見られない」ではなく
	// 「この段では何も足していない」の意味（ListPageRestrictions と同じ見方）。
	ListPageGrants(ctx context.Context, workspaceID, pageID string) ([]domain.PageGrant, error)

	// PagePermissionFactsForUser はログイン済みユーザーとして、1 ページの実効権限を決める
	// 事実を 1 回のクエリで集める。判定は domain.ResolvePagePermission が行う。
	// ページが無い・別ワークスペースなら ErrPageNotFound。
	PagePermissionFactsForUser(ctx context.Context, workspaceID, pageID string, userID uint64) (*domain.PagePermissionFacts, error)
	// PagePermissionFactsForPrincipal は共有リンクの来訪者（kind='share_link' の主体）として
	// 同じ事実を集める。既定（リンクの capability）は呼び出し側が facts に載せる。
	PagePermissionFactsForPrincipal(ctx context.Context, workspaceID, pageID, principalID string) (*domain.PagePermissionFacts, error)
	// ListSpacePageViewFacts はスペース配下のページ全件と、その閲覧の事実を
	// archived で現役／アーカイブ済みを切り替える（false で現役）。アーカイブ用に
	// 別のクエリを持たないのは、権限の事実を組み立てる部分を写経しないため
	// （同じ判断が 2 箇所にあると必ずずれる）。
	// 1 回のクエリで返す（ページごとに問い合わせない）。編集の事実は集めないので、
	// 編集可否をここから出さないこと（返す型がそれを表している）。
	ListSpacePageViewFacts(ctx context.Context, workspaceID, spaceID string, userID uint64, archived bool) ([]PageWithViewFacts, error)
	// SearchWorkspacePageViewFacts はワークスペース全体から題名が部分一致する現役ページを
	// 候補にし、その閲覧の事実を返す（サイドバーの題名検索用）。判定は呼び出し側が
	// domain.ResolvePageView で行う。query はエスケープ前の生の文字列を渡す
	// （% _ \ のエスケープは実装が行う — 呼び出し側に SQL の都合を漏らさない）。
	// ParentArchived は常に false（検索は現役だけを対象にするため集めない）。
	SearchWorkspacePageViewFacts(ctx context.Context, workspaceID string, userID uint64, query string) ([]PageWithViewFacts, error)
	// ListWorkspacePageViewFactsByIDs は指定 ID 群のページの閲覧の事実を返す
	// （ページ参照の題名解決とパンくずが使う）。事実の見方は検索と同一で、判定は
	// 呼び出し側が domain.ResolvePageView で行う。UUID として読めない ID・
	// 他ワークスペースの ID は行にならない（エラーにしない — 壊れた参照で
	// ページ全体の読み出しを落とさない）。**アーカイブ済みも行として返す**
	// （Page.ArchivedAt に載る）。除外するかは用途で違うため呼び出し側が決める —
	// 題名解決は除外し、パンくずは含める（経路から抜くと場所を偽る）。
	// ParentArchived は集めない（この口の用途では使わない）。常に false。
	ListWorkspacePageViewFactsByIDs(ctx context.Context, workspaceID string, userID uint64, pageIDs []string) ([]PageWithViewFacts, error)
	// SpacePermissionFactsForUser はページを介さず、スペース 1 つの実効権限を決める事実を集める。
	// 判定は domain.ResolveScopePermission が行う。スペースが無い・別ワークスペースなら
	// ErrSpaceNotFound。
	//
	// 返すのは「その入れ物に届いている役割の集合」だけで、どれを採るかの規則は持たない。
	// ページ付与（page_grants）も見ない。したがってこの口の答えを
	// 「そのスペースのあるページを編集してよいか」に使ってはいけない
	// （祖先のページに張られた付与を取りこぼし、必ず狭い側へ倒れる）。使ってよいのは
	// 対象がまだ存在しない操作（スペース直下へのページ作成）だけ。
	//
	// スペースの実在をここで確かめるのは、確かめないと fail-open になるため。
	// workspace_grants は配下の全スペースに届くので、別ワークスペースのスペース ID を
	// 渡されても「自分のワークスペースでの役割」がそのまま返り、他テナントのスペースに対して
	// editor と答えてしまう。
	SpacePermissionFactsForUser(ctx context.Context, workspaceID, spaceID string, userID uint64) (*domain.ScopeFacts, error)

	// WorkspacePermissionFactsForUser はワークスペースそのものに対する実効権限を決める事実を集める。
	// スペースを作る操作のように、どのスペースにも属さない判定に使う。
	//
	// 実在を確かめないのは、ワークスペースが無ければ grant も 1 行も無く、
	// 役割 0 個 ＝ 何もできない（fail-closed）に自然と倒れるため。
	// 呼び出し側は middleware が slug から解決したワークスペースを渡す。
	WorkspacePermissionFactsForUser(ctx context.Context, workspaceID string, userID uint64) (*domain.ScopeFacts, error)
	// ListWorkspaceSpaceScopeFacts はワークスペース配下のスペース全件と、それぞれで
	// 呼び出し元に届いている役割（事実）を 1 回のクエリで返す。判定は
	// domain.ResolveScopePermission が行う。
	//
	// **返り値はまだ「見せてよいスペース」ではない。** 役割が 1 つも届いていないスペースも
	// Roles が空のまま含まれる。ふるい落とすのは呼び出し側（ListViewableSpacesUseCase）で、
	// ここで絞らないのは判定規則を domain の 1 箇所に閉じるため。
	//
	// スペースごとに SpacePermissionFactsForUser を呼ぶ（N+1）ことはしない。
	// サイドバーはワークスペースを開くたびにこの一覧を引くので、スペース数だけ往復すると
	// そのまま画面の待ち時間になる。
	//
	// ワークスペースの実在は確かめない。無ければスペースが 1 件も無く、空スライスに倒れる
	// （WorkspacePermissionFactsForUser が実在を確かめないのと同じ理由）。
	ListWorkspaceSpaceScopeFacts(ctx context.Context, workspaceID string, userID uint64) ([]SpaceWithScopeFacts, error)
	// ListSubtreePagePermissionFacts はサブツリー（対象ページ自身 + 全子孫）の各ページと、
	// その実効権限を決める事実を 1 回のクエリで返す。アーカイブ済みのページも含む。
	// ページとその子孫をまとめて書き換える操作が、根 1 枚の権限だけで通らないようにするための口。
	// ページが無い・別ワークスペースなら空スライス（呼び出し側が先に根の権限を確かめている）。
	ListSubtreePagePermissionFacts(ctx context.Context, workspaceID, pageID string, userID uint64) ([]PageWithPermissionFacts, error)
}
