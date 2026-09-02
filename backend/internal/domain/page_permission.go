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
	// HasAllowList は経路上に許可リスト制の段があるか。
	// その段以下は「許可リストに載っている者だけ」の限定公開になっている。
	//
	// これは「allow 行が 1 行でもあるか」ではない。allow 行は主体と一緒に消えるので、
	// 行の有無で数えると許可リストの誰かを削除しただけで限定公開が解ける。
	// 段であること自体は主体を参照しない印（page_allow_lists）が持ち、
	// 許可リストが空になった段は「誰も載っていない」＝ 不許可として扱う。
	HasAllowList bool
	// AllowedAtNearest は「最も近い許可リスト制の段」に自分を allow する行があるか。
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

// PageViewFacts は「閲覧できるか」だけを決める事実。
//
// ページ一覧のように閲覧の列しか集めない経路はこの型を使い、PagePermissionFacts を
// 流用しない。あちらは編集の事実も持つ型で、集めていないときの nil が
// 「経路に制限が無い」＝ 既定どおりという意味になる。閲覧しか集めていない事実を
// あの型に載せると、編集の例外を 1 つも見ないまま CanEdit が既定（editor なら true）で
// 返り、同じページの可否が経路によって食い違う。
// 「集めていない」と「制限が無い」を型で分ける。
type PageViewFacts struct {
	// Role は既定の役割。意味は PagePermissionFacts.Role と同じ。
	Role *GrantRole
	// View は閲覧についての経路上の制限の集計。制限が経路のどこにも無ければ nil。
	View *RestrictionFacts
}

// ResolvePageView は集めた事実から閲覧できるかを決める。
// 例外の突き合わせは ResolvePagePermission と同じ 1 つの実装（resolveCapability）を通る。
func ResolvePageView(f PageViewFacts) bool {
	return resolveCapability(roleAllows(f.Role, CapabilityView), f.View)
}

// PagePermission は 1 ページに対する実効権限。
type PagePermission struct {
	// CanView はページを閲覧できるか。
	CanView bool `json:"canView"`
	// CanEdit はページを編集できるか。CanView が false のとき必ず false。
	CanEdit bool `json:"canEdit"`
	// CanManage はそのページの権限（grant / 例外 / 共有リンク）を変えられるか。
	//
	// **ほかの 2 つと違い、経路上の例外を見ない。** 見るのは届いている既定の役割だけで、
	// 自分を deny したページでも true のままになる。そうしないと、管理者が自分を
	// 締め出した瞬間にその例外を自分で戻せなくなる（閉じ込めを解く手段が消える）。
	//
	// 例外の層が admin を表せないことも理由の 1 つ。Capability は view / edit しか無く、
	// 「この人だけ管理者から外す」は書けないので、見るべき例外がそもそも存在しない。
	CanManage bool `json:"canManage"`
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
	return roleAllows(f.Role, c)
}

// roleAllows は既定の役割が指定のケイパビリティを許すかを返す（grant が無ければ何もできない）。
func roleAllows(role *GrantRole, c Capability) bool {
	if role == nil {
		return false
	}
	if c == CapabilityEdit {
		return role.CanEdit()
	}
	return role.CanView()
}

// resolveCapability は 1 つのケイパビリティについて既定と例外を突き合わせる。
//
// 優先規則（この順で決まる。上ほど強い）:
//
//  1. 対象ページ自身または祖先のどこかに自分を deny する行がある → 不許可。
//     deny は allow に勝ち、経路全体で効く。複数のグループに属していて片方で外され
//     片方で許されている場合も、安全側（不許可）へ倒す。
//  2. 経路上に許可リスト制の段がある → そのうち最も近い段に自分が載っているかで決める。
//     載っていれば許可、載っていなければ既定が admin でも不許可（限定公開）。
//     祖先の許可リストは、その下に許可リスト制でない段がいくつ挟まっても効き続ける。
//  3. 許可リスト制の段が無い（deny 行だけ、または制限そのものが経路に無い）→ 既定に戻る。
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
// ノートの権限規則はこの関数だけが持ち、呼び出し側（usecase / handler / SQL）へは写さない。
func ResolvePagePermission(f PagePermissionFacts) PagePermission {
	canView := resolveCapability(f.defaultAllows(CapabilityView), f.View)
	// 編集は閲覧を含む。閲覧できないページを編集できる状態は、UI でも監査でも説明できず、
	// 「view だけ deny した」つもりが編集経路から中身を読めてしまう穴になる。
	canEdit := canView && resolveCapability(f.defaultAllows(CapabilityEdit), f.Edit)
	// 管理は例外を通さない（CanManage の doc に理由がある）。共有リンクの来訪者は
	// 役割を持たないので、ここは必ず false になる。
	canManage := f.Role != nil && f.Role.CanManage()
	return PagePermission{CanView: canView, CanEdit: canEdit, CanManage: canManage}
}

// Allows は実効権限が指定のケイパビリティを満たすかを返す。
func (p PagePermission) Allows(c Capability) bool {
	if c == CapabilityEdit {
		return p.CanEdit
	}
	return p.CanView
}
