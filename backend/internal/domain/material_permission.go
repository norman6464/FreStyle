package domain

import "time"

// CourseGrant はコースでの既定の権限。そのコースと配下の章に効く。
type CourseGrant struct {
	// WorkspaceID はテナント境界。principal との複合 FK に使う。
	WorkspaceID string `json:"workspaceId"`
	// CourseID は対象コース。
	CourseID uint64 `json:"courseId"`
	// PrincipalID は権限を与える相手。
	PrincipalID string `json:"principalId"`
	// Role は既定の役割。
	Role      GrantRole `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ChapterGrant は章 1 つだけに効く既定の権限（「この教材だけ編集してよい」）。
type ChapterGrant struct {
	WorkspaceID string `json:"workspaceId"`
	// ChapterID は対象の章（教材）。
	ChapterID   uint64    `json:"chapterId"`
	PrincipalID string    `json:"principalId"`
	Role        GrantRole `json:"role"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// MaterialFacts は教材（コース 1 つ、または章 1 つ）の実効権限を決める事実の集合。
//
// ノートの PagePermissionFacts と分けているのは、集める事実が違うため。
// 教材には例外（deny / 許可リスト）の層が無く、代わりに「公開済みか」がある。
// 同じ型に載せると、見ていない層の nil が「制限が無い」に化ける。
type MaterialFacts struct {
	// Member はそのワークスペースの一員か。所属は principals（kind='user'）の行が唯一の表現。
	Member bool
	// WorkspaceAdmin はワークスペース全体の admin か。
	//
	// **ワークスペースの grant が教材へ届くのはこの 1 つだけ。** editor / commenter / viewer は
	// ノートの木に対する既定で、教材には届かない。届かせると、ノートの editor である人が
	// 教材の編集権まで一度に得てしまう（本番に実際そういう人が居る）。
	//
	// admin だけを通すのは、付与された人が居なくなった教材の権限を誰も変えられなくなる
	// 事態を避けるため（ノート側で最後の admin を守っているのと同じ理由）。
	WorkspaceAdmin bool
	// Role はコース / 章から届いた既定の役割のうち最も強いもの。付与が 1 つも無ければ nil。
	Role *GrantRole
	// Published は公開済みか。下書きは編集できる人にしか見せない。
	Published bool
}

// MaterialPermission は教材 1 つに対する実効権限。
type MaterialPermission struct {
	// CanView は中身を読めるか。
	CanView bool `json:"canView"`
	// CanEdit は書き換えられるか。
	CanEdit bool `json:"canEdit"`
	// CanManage はその教材の権限そのものを変えられるか。
	CanManage bool `json:"canManage"`
}

// ResolveMaterialPermission は集めた事実から教材 1 つの実効権限を決める。
// 教材の権限規則はこの関数だけが持ち、呼び出し側（usecase / handler / SQL）へは写さない。
//
// # 読むことには付与を要求しない
//
// 公開済みの教材はワークスペースの一員なら誰でも読める。ここに付与を求めると、
// 研修を受ける人が教材を開くたびに権限を配って回ることになり、**学ぶための場が
// 成り立たない**。付与が意味を持つのは書き換えと、下書きを覗くところ。
func ResolveMaterialPermission(f MaterialFacts) MaterialPermission {
	role := f.Role
	if f.WorkspaceAdmin {
		// admin は最も強いので、ほかにどの役割が届いていても結果は変わらない。
		admin := GrantRoleAdmin
		role = &admin
	}
	// 所属していない相手には何もさせない。公開済みでも他テナントの教材は読めないし、
	// 権限を変えることもできない。
	//
	// いまは事実を集める側（SQL）が主体を辿るので、所属していなければ役割も届かない。
	// それでもここで閉じるのは、**規則の側で閉じておかないと集め方を変えたときに開く**ため。
	// 「所属している人にだけ効く」はこの型が持つ約束で、集め方の性質に頼らない。
	if !f.Member {
		return MaterialPermission{}
	}
	return MaterialPermission{
		CanView:   f.Published || roleAllows(role, CapabilityView),
		CanEdit:   roleAllows(role, CapabilityEdit),
		CanManage: role != nil && role.CanManage(),
	}
}
