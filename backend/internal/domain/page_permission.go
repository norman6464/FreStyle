package domain

// PagePermissionFacts は 1 ページの実効権限を決めるのに必要な事実の集合。
// repository が 1 回のクエリで集め、ResolvePagePermission が組み合わせて答えを出す。
// 事実の収集（SQL）と規則の適用（この型のメソッド）を分けるのは、優先規則を DB に写経させないため。
//
// 打ち消す層は持たない（唯一の例外: ページの visibility='private'）。権限は 3 段の付与
// （workspace / space / page）を足し合わせ、届いた中で最も強い役割で決まる。下の段が上の段を
// 弱めることはなく、「親は共有、この子だけ隠す」は書けない（狭めたいなら private のスペースへ
// 置く）。打ち消しを許すと「なぜこの人に見える／見えないのか」が経路をさかのぼらないと
// 答えられなくなる。
//
// ページ単位の Visibility=private だけは意図した唯一の例外（段 13）。「作成者以外には一切
// 見せない」という個人の下書き向けの要求で、grants を増やす方向の話ではないため、この 1 つに
// 限って明示的に打ち消す（ResolvePagePermission 冒頭の早期リターン）。共有ボタンで他人に
// page_grants を足しても、visibility が private のままなら効かない。
type PagePermissionFacts struct {
	// Member はそのユーザーがワークスペースのメンバーか。所属は principals が唯一の表現で、
	// 専用のメンバーシップ表は持たない。共有リンク経由（未ログイン）では false。
	Member bool
	// Role は届いた中で最も強い役割（workspace / space / page の 3 段 × 複数主体のうち最強、
	// GrantRole.Rank 参照）。grant が無ければ nil。ポインタにしているのは「grant 無し」と
	// 最弱の役割を型で区別するため。
	Role *GrantRole
	// ShareLinkCapability は共有リンク経由のときだけ非 nil。Role とは同時に使わない。
	// 共有リンクは広げる方向にしか働かない（未ログインの相手へ「見せる」を足すだけ）。
	ShareLinkCapability *Capability
	// Visibility はページの公開範囲。ゼロ値（""）は PageVisibilitySpace 扱い。
	// 'private' のときだけ IsOwner を見る。
	Visibility PageVisibility
	// IsOwner はこの facts を解決した相手がページの作成者か。共有リンク経由では常に false。
	IsOwner bool
}

// PagePermission は 1 ページに対する実効権限。
type PagePermission struct {
	CanView bool `json:"canView"`
	// CanEdit は CanView が false のとき必ず false。
	CanEdit bool `json:"canEdit"`
	// CanManage はそのページの権限（grant / 共有リンク）を変えられるか。
	CanManage bool `json:"canManage"`
	// CanComment は閲覧できて役割が commenter 以上のとき true。共有リンク経由では常に false。
	CanComment bool `json:"canComment"`
}

// defaultAllows は届いた既定が指定のケイパビリティを許すかを返す。
func (f PagePermissionFacts) defaultAllows(c Capability) bool {
	// 共有リンク経由は grant を持たない。リンク自身のケイパビリティが既定になる。
	if f.ShareLinkCapability != nil {
		if c == CapabilityEdit {
			return *f.ShareLinkCapability == CapabilityEdit
		}
		return true
	}
	return roleAllows(f.Role, c)
}

// roleAllows は役割が指定のケイパビリティを許すかを返す（grant が無ければ何もできない）。
func roleAllows(role *GrantRole, c Capability) bool {
	if role == nil {
		return false
	}
	if c == CapabilityEdit {
		return role.CanEdit()
	}
	return role.CanView()
}

// pageViewableGivenVisibility は visibility='private' の早期リターンを ResolvePageView /
// ResolvePagePermission で共有するヘルパー。private でなければ常に true。
func pageViewableGivenVisibility(visibility PageVisibility, isOwner bool) bool {
	return visibility != PageVisibilityPrivate || isOwner
}

// ResolvePageView は集めた事実から閲覧できるかを決める。ページ一覧のように閲覧の可否だけを
// 集める経路が使う。private かつ非オーナーなら role の強さに関わらず false
// （ResolvePagePermission と同じ、唯一の打ち消し例外）。
func ResolvePageView(role *GrantRole, visibility PageVisibility, isOwner bool) bool {
	if !pageViewableGivenVisibility(visibility, isOwner) {
		return false
	}
	return roleAllows(role, CapabilityView)
}

// ResolvePagePermission は集めた事実から 1 ページの実効権限を決める。
// ナレッジの権限規則はこの関数だけが持ち、呼び出し側（usecase / handler / SQL）へは写さない。
func ResolvePagePermission(f PagePermissionFacts) PagePermission {
	// visibility='private' は唯一の打ち消し例外（型の doc 参照）。作成者本人でなければ
	// grants・共有リンクどちらでも何も許さない。他の判定より先に閉じる — 下のケイパビリティ
	// ごとの判定は defaultAllows/canView を経由しない独自の道もあり、後から AND するのでは
	// 足りない場所が出るため。
	if !pageViewableGivenVisibility(f.Visibility, f.IsOwner) {
		return PagePermission{}
	}
	// 所属していない相手には何もさせない。事実を集める側（SQL）が主体を辿るため所属していなければ
	// 役割も届かないはずだが、規則の側でも閉じておく — 集め方を変えたときにここが開かないため。
	// 共有リンクの来訪者は未ログインで Member=false だが、そちらは所属ではなくリンク自身の
	// ケイパビリティで決まるので、Member を見るのは「リンク経由でないとき」に限る。
	if f.ShareLinkCapability == nil && !f.Member {
		return PagePermission{}
	}
	canView := f.defaultAllows(CapabilityView)
	// 編集は閲覧を含む。いまの役割の並び（GrantRole.Rank）では崩れないが、役割を増やしたときの
	// 安全のため残す。
	canEdit := canView && f.defaultAllows(CapabilityEdit)
	// 権限そのものを変えられるのは役割が admin のときだけ。**共有リンク経由では必ず false。**
	// 付与の口は主体の種類を見ずに実在しか確かめないため、リンクの主体へ admin を張ることが
	// API から実際にでき（リンクの主体 ID は一覧の応答に載る）、defaultAllows を通らないこの
	// 判定だけが抜け穴になっていた。
	canManage := f.ShareLinkCapability == nil && f.Role != nil && f.Role.CanManage()
	// コメントも defaultAllows を通らない別軸の判定。**共有リンク経由では必ず false**
	// （共有リンクの来訪者はコメント不可という設計。domain.Capability に 'comment' が無い理由と同じ）。
	canComment := canView && f.ShareLinkCapability == nil && f.Role != nil && f.Role.CanComment()
	return PagePermission{CanView: canView, CanEdit: canEdit, CanManage: canManage, CanComment: canComment}
}

// Allows は実効権限が指定のケイパビリティを満たすかを返す。
func (p PagePermission) Allows(c Capability) bool {
	if c == CapabilityEdit {
		return p.CanEdit
	}
	return p.CanView
}
