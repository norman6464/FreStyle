package domain

// RestrictionFacts はあるケイパビリティについて「最も近い制限の段」を集計した結果。
//
// 「段」とは、対象ページ自身とその祖先のうち、そのケイパビリティの PageRestriction を
// 1 行以上持つ最も近いページ（page_paths の depth が最小のもの）のこと。
// 判定に使うのはその 1 段だけで、さらに遠い祖先の制限は見ない
// （近い例外が遠い例外を上書きするのが「継承 + 例外」の意味であり、
// 遠い祖先まで足し合わせると、子ページに例外を張っても親の設定から逃げられなくなる）。
type RestrictionFacts struct {
	// Denied はその段に「自分（が属する主体のいずれか）を deny する行」があるか。
	Denied bool
	// Allowed はその段に「自分を allow する行」があるか。
	Allowed bool
	// HasAllowList はその段に allow 行が 1 つでもあるか（相手が自分かどうかを問わない）。
	// allow が 1 つでもあれば、その段は「載っている者だけ」の限定公開になっている。
	HasAllowList bool
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
	// View は閲覧についての「最も近い制限の段」。制限がどの祖先にも無ければ nil。
	View *RestrictionFacts
	// Edit は編集についての「最も近い制限の段」。制限がどの祖先にも無ければ nil。
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
//  1. 最も近い段に自分を deny する行がある → 不許可。deny は allow に勝つ。
//     同じ段で allow と deny の両方に当たるのは、複数のグループに属していて片方で外され
//     片方で許されている場合。安全側（不許可）へ倒す。
//  2. 最も近い段に自分を allow する行がある → 許可。
//  3. どちらにも当たらず、その段に allow 行が 1 つでもある → 不許可。
//     allow 行の存在は「載っている者だけ」の限定公開を意味し、載っていない自分は対象外。
//  4. どちらにも当たらず、その段が deny 行だけでできている → 既定に戻る。
//     deny だけの段は「その人だけ外す」であり、ほかの人の既定は変えない。
//  5. そもそも制限のある祖先が無い → 既定に従う。
func resolveCapability(defaultAllowed bool, exc *RestrictionFacts) bool {
	if exc == nil {
		return defaultAllowed
	}
	switch {
	case exc.Denied:
		return false
	case exc.Allowed:
		return true
	case exc.HasAllowList:
		return false
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
