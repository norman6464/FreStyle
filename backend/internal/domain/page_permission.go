package domain

// PagePermissionFacts は 1 ページの実効権限を決めるのに必要な事実の集合。
// repository が 1 回のクエリで集め、ResolvePagePermission が組み合わせて答えを出す。
//
// 事実の収集（SQL）と規則の適用（この型のメソッド）を分けているのは、
// 優先規則を DB に写経させないため。ページ一覧のように 1 回のクエリで多数のページを
// 扱う経路でも、SQL が返すのは事実だけで、規則は同じ 1 つの関数を通る。
//
// # 打ち消す層は持たない
//
// 権限は 3 段の付与（workspace / space / page）を足し合わせ、届いた中で最も強い役割で
// 決まる。下の段が上の段を弱めることはなく、「親は共有、この子だけ隠す」は書けない。
// 狭めたい内容は private のスペースへ置く。
//
// この形にしているのは、打ち消しを許すと「なぜこの人に見える／見えないのか」が
// 経路をさかのぼらないと答えられなくなるため。同じ設計を採った製品は、後から
// 経路のどの段で許され拒まれたかを一覧する専用の検査機能を用意する羽目になっている。
type PagePermissionFacts struct {
	// Member はそのユーザーがワークスペースのメンバーか（kind='user' の Principal があるか）。
	// 所属は principals が唯一の表現で、専用のメンバーシップ表は持たない。
	// 共有リンク経由（ログインしていない来訪者）では false。
	Member bool
	// Role は届いた中で最も強い役割。ワークスペース / スペース / ページの 3 段の grant と、
	// 複数の主体（自分 / 所属グループ / スペース全員）から得た役割のうち最も強いものが入る
	// （GrantRole.Rank 参照）。grant が 1 つも無ければ nil。
	//
	// 「grant が無い」を GrantRole("") のような値で表さずポインタにしているのは、
	// 未設定と最弱の役割を型で区別するため。
	Role *GrantRole
	// ShareLinkCapability は共有リンク経由のアクセスのときだけ非 nil で、そのリンクの既定。
	// Role とは同時に使わない（呼び出し側がどちらの主体として解決するかを決める）。
	//
	// 共有リンクは広げる方向にしか働かない。ログインしていない相手へ「見せる」を足すだけで、
	// すでに見えている人から取り上げることはない。
	ShareLinkCapability *Capability
}

// PagePermission は 1 ページに対する実効権限。
type PagePermission struct {
	// CanView はページを閲覧できるか。
	CanView bool `json:"canView"`
	// CanEdit はページを編集できるか。CanView が false のとき必ず false。
	CanEdit bool `json:"canEdit"`
	// CanManage はそのページの権限（grant / 共有リンク）を変えられるか。
	CanManage bool `json:"canManage"`
}

// defaultAllows は届いた既定が指定のケイパビリティを許すかを返す。
func (f PagePermissionFacts) defaultAllows(c Capability) bool {
	// 共有リンク経由は grant を持たない（ログインしていない相手なので所属が無い）。
	// リンク自身が持つケイパビリティが既定になる。
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

// ResolvePageView は集めた事実から閲覧できるかを決める。
// ページ一覧のように閲覧の列しか集めない経路が使う。
func ResolvePageView(role *GrantRole) bool {
	return roleAllows(role, CapabilityView)
}

// ResolvePagePermission は集めた事実から 1 ページの実効権限を決める。
// ノートの権限規則はこの関数だけが持ち、呼び出し側（usecase / handler / SQL）へは写さない。
func ResolvePagePermission(f PagePermissionFacts) PagePermission {
	// 所属していない相手には何もさせない。
	//
	// いまは事実を集める側（SQL）が主体を辿るので、所属していなければ役割も届かない。
	// それでもここで閉じるのは、**規則の側で閉じておかないと集め方を変えたときに開く**ため。
	// 「所属している人にだけ効く」はこの型が持つ約束で、集め方の性質に頼らない
	// （ResolveMaterialPermission が同じ理由で同じことをしている）。
	//
	// 共有リンクの来訪者はログインしていないので Member は false だが、そちらは
	// 所属ではなくリンク自身のケイパビリティで決まる。だから Member を見るのは
	// 「リンク経由ではないとき」に限る。
	if f.ShareLinkCapability == nil && !f.Member {
		return PagePermission{}
	}
	canView := f.defaultAllows(CapabilityView)
	// 編集は閲覧を含む。閲覧できないページを編集できる状態は、UI でも監査でも説明できない。
	// いまの役割の並び（GrantRole.Rank）では編集できる者は必ず閲覧もできるので、この
	// 掛け合わせで結果が変わることはない。役割を増やしたときに崩れないよう残してある。
	canEdit := canView && f.defaultAllows(CapabilityEdit)
	// 権限そのものを変えられるのは、届いている役割が admin のときだけ。
	//
	// **共有リンク経由では必ず false にする。** 来訪者はログインしていないので、ここが
	// true になると「URL を知っているだけの人が、誰に何を見せるかを決められる」ことになる。
	//
	// 「リンクの主体には役割が届かないはず」に頼ってはいけない。付与の口は主体の実在しか
	// 確かめず種類を見ないので、リンクの主体へ admin を張ることが API から実際にできる
	// （リンクの主体 ID は一覧の応答に載っている）。閲覧と編集はリンク自身のケイパビリティで
	// 頭打ちになるが、管理だけは defaultAllows を通らないのでそこだけ抜けていた。
	canManage := f.ShareLinkCapability == nil && f.Role != nil && f.Role.CanManage()
	return PagePermission{CanView: canView, CanEdit: canEdit, CanManage: canManage}
}

// Allows は実効権限が指定のケイパビリティを満たすかを返す。
func (p PagePermission) Allows(c Capability) bool {
	if c == CapabilityEdit {
		return p.CanEdit
	}
	return p.CanView
}
