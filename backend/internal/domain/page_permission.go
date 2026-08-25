package domain

// RestrictionFacts はあるケイパビリティについて、対象ページ自身から根までの経路上に
// 張られた PageRestriction を集計した結果。
//
// 「段」とは、経路上でその制限を 1 行以上持つページ 1 枚のこと（page_paths の depth で
// 近さを測る。0 が対象ページ自身）。deny と allow で見る範囲が違う。
//
//   - deny は経路全体で効く。祖先で外された人は、その下のどのページでも外れたまま。
//   - allow（＝ 限定公開の許可リスト）は「allow 行を持つ最も近い段」だけで決める。
//
// deny を「最も近い段」だけで見てはいけない。deny 行しか無い段が最近段になると
// 「deny だけの段は既定に戻す」という規則が働き、より遠い祖先の許可リストごと
// 無効化されてしまう。つまり「この人だけ外す」という無害な操作 1 行で、
// 祖先の限定公開が第三者に開く。
type RestrictionFacts struct {
	// DeniedAnywhere は経路上（対象ページ自身を含む）のどこかに
	// 「自分（が属する主体のいずれか）を deny する行」があるか。
	DeniedAnywhere bool
	// HasAllowList は経路上に allow 行を持つ段があるか（相手が自分かどうかを問わない）。
	// allow が 1 つでもあれば、その段以下は「載っている者だけ」の限定公開になっている。
	HasAllowList bool
	// AllowedAtNearest は「allow 行を持つ最も近い段」に自分を allow する行があるか。
	// HasAllowList が false のときは意味を持たない。
	//
	// より近い許可リストがより遠い許可リストを上書きするのは意図した挙動
	// （root が [alice]・child が [alice, bob] なら bob は child 以下だけ見える）。
	// 「この枝だけもう少し広く共有する」を書けるようにするためで、
	// deny と違って足し合わせない。
	AllowedAtNearest bool
}

// PagePermissionFacts は 1 ページの実効権限を決めるのに必要な事実の集合。
// repository が 1 回のクエリで集め、ResolvePagePermission が組み合わせて答えを出す。
//
// 事実の収集（SQL）と規則の適用（この型のメソッド）を分けているのは、
// 優先規則を DB に写経させないため。ページ一覧のように 1 回のクエリで多数のページを
// 扱う経路でも、SQL が返すのは事実だけで、規則は同じ 1 つの関数を通る。
type PagePermissionFacts struct {
	// Member はそのユーザーがワークスペースのメンバーか（kind='user' の Principal があるか）。
	// 所属は principals が唯一の表現で、専用のメンバーシップ表は持たない。
	// 共有リンク経由（ログインしていない来訪者）では false。
	Member bool
	// Role は既定の役割。ページが属するスペースの grant とワークスペースの grant、
	// および複数の主体（自分 / 所属グループ / スペース全員）から得た役割のうち
	// 最も強いものが入る（GrantRole.Rank 参照）。grant が 1 つも無ければ nil。
	//
	// 「grant が無い」を GrantRole("") のような値で表さずポインタにしているのは、
	// 未設定と最弱の役割を型で区別するため。
	Role *GrantRole
	// ShareLinkCapability は共有リンク経由のアクセスのときだけ非 nil で、そのリンクの既定。
	// Role とは同時に使わない（呼び出し側がどちらの主体として解決するかを決める）。
	ShareLinkCapability *Capability
	// View は閲覧についての経路上の制限の集計。制限が経路のどこにも無ければ nil。
	View *RestrictionFacts
	// Edit は編集についての経路上の制限の集計。制限が経路のどこにも無ければ nil。
	Edit *RestrictionFacts
}

// PagePermission は 1 ページに対する実効権限。
type PagePermission struct {
	// CanView はページを閲覧できるか。
	CanView bool `json:"canView"`
	// CanEdit はページを編集できるか。CanView が false のとき必ず false。
	CanEdit bool `json:"canEdit"`
}

// defaultAllows は例外がまったく無いときに許されるか（既定）を返す。
func (f PagePermissionFacts) defaultAllows(c Capability) bool {
	// 共有リンク経由は grant を持たない（ログインしていない相手なので所属が無い）。
	// リンク自身が持つケイパビリティが既定になる。
	if f.ShareLinkCapability != nil {
		if c == CapabilityEdit {
			return *f.ShareLinkCapability == CapabilityEdit
		}
		return true
	}
	if f.Role == nil {
		return false
	}
	if c == CapabilityEdit {
		return f.Role.CanEdit()
	}
	return f.Role.CanView()
}

// resolveCapability は 1 つのケイパビリティについて既定と例外を突き合わせる。
//
// 優先規則（この順で決まる。上ほど強い）:
//
//  1. 対象ページ自身または祖先のどこかに自分を deny する行がある → 不許可。
//     deny は allow に勝ち、経路全体で効く。複数のグループに属していて片方で外され
//     片方で許されている場合も、安全側（不許可）へ倒す。
//  2. 経路上に allow 行を持つ段がある → そのうち最も近い段に自分が載っているかで決める。
//     載っていれば許可、載っていなければ既定が admin でも不許可（限定公開）。
//     祖先の許可リストは、その下に allow の無い段がいくつ挟まっても効き続ける。
//  3. allow 行を持つ段が無い（deny 行だけ、または制限そのものが経路に無い）→ 既定に戻る。
//     deny だけの段は「その人だけ外す」であり、ほかの人の既定は変えない。
func resolveCapability(defaultAllowed bool, exc *RestrictionFacts) bool {
	if exc == nil {
		return defaultAllowed
	}
	switch {
	case exc.DeniedAnywhere:
		return false
	case exc.HasAllowList:
		return exc.AllowedAtNearest
	default:
		return defaultAllowed
	}
}

// ResolvePagePermission は集めた事実から 1 ページの実効権限を決める。
// ナレッジ基盤の権限規則はこの関数だけが持ち、呼び出し側（usecase / handler / SQL）へは写さない。
func ResolvePagePermission(f PagePermissionFacts) PagePermission {
	canView := resolveCapability(f.defaultAllows(CapabilityView), f.View)
	// 編集は閲覧を含む。閲覧できないページを編集できる状態は、UI でも監査でも説明できず、
	// 「view だけ deny した」つもりが編集経路から中身を読めてしまう穴になる。
	canEdit := canView && resolveCapability(f.defaultAllows(CapabilityEdit), f.Edit)
	return PagePermission{CanView: canView, CanEdit: canEdit}
}

// Allows は実効権限が指定のケイパビリティを満たすかを返す。
func (p PagePermission) Allows(c Capability) bool {
	if c == CapabilityEdit {
		return p.CanEdit
	}
	return p.CanView
}
