package domain

// ScopeFacts は入れ物（ワークスペース / スペース）1 つの実効権限を決めるのに必要な事実。
//
// ページを介さない判定に使う。「スペース直下にページを作る」「スペースを作る」のように
// 対象がまだ存在しない操作は、ページ 1 枚の権限では判断できず、かといって
// 「メンバーなら誰でもできる」で埋めるとあとから締められない穴になるため、
// 入れ物そのものに対する既定の役割で決める。
//
// PagePermissionFacts とは別の型にする。あちらは経路上の例外（page_restrictions）を
// 集めた事実を持ち、nil が「経路に制限が無い」＝ 既定どおりを意味する。
// スペースやワークスペースには例外の層がそもそも無いので、同じ型に載せると
// 「例外を見ていない」ことが「例外が無い」に化ける。集めていない事実は型に持たせない。
type ScopeFacts struct {
	// Roles は自分に効く主体（自分 / 所属グループ / スペース全員）が、その入れ物に届く
	// grant から得た役割すべて。重複・順序に意味は無い（1 つも無ければ空）。
	//
	// 「最も強いものを採る」という合成規則はここには入れず、
	// ResolveScopePermission（＝ StrongestGrantRole）だけが持つ。
	// 事実として役割の集合を渡し、畳み方は domain 側で決める。
	Roles []GrantRole
}

// ScopePermission は入れ物（ワークスペース / スペース）に対する実効権限。
//
// **ページ単位の例外（page_restrictions / page_allow_lists）は一切見ていない。**
// ページを名指しする操作の可否をこれで決めてはいけない（そのページで deny されていても
// 入れ物の既定が editor なら CanEdit が true になる）。ページには
// ResolvePagePermission を使い、こちらは対象がまだ無い操作
// （スペース直下への作成 / スペースの作成）にだけ使う。
type ScopePermission struct {
	// CanView は入れ物の中身を既定で閲覧できるか。
	CanView bool `json:"canView"`
	// CanEdit は入れ物の中身を既定で編集できるか（＝ 直下にページを作れるか）。
	CanEdit bool `json:"canEdit"`
	// CanManage は入れ物そのものの構成（配下のスペース / 権限）を変えられるか。
	CanManage bool `json:"canManage"`
}

// Allows は実効権限が指定のケイパビリティを満たすかを返す。
func (p ScopePermission) Allows(c Capability) bool {
	if c == CapabilityEdit {
		return p.CanEdit
	}
	return p.CanView
}

// ResolveScopePermission は集めた事実から入れ物 1 つの実効権限を決める。
//
// 既定の役割からケイパビリティへの写像は roleAllows を通す（ページ 1 枚の解決と同じ実装）。
// ここに if 文で「admin なら〜」を書き足すと、同じ役割の意味がページと入れ物で
// 食い違う余地ができる。
func ResolveScopePermission(f ScopeFacts) ScopePermission {
	role := StrongestGrantRole(f.Roles)
	return ScopePermission{
		CanView:   roleAllows(role, CapabilityView),
		CanEdit:   roleAllows(role, CapabilityEdit),
		CanManage: role != nil && role.CanManage(),
	}
}
