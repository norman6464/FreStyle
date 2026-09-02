package domain

import "time"

// PrincipalKind は principals.kind に入る主体の種類。
//
// ユーザー・グループ・スペース全員・公開リンクは「権限を与える相手」という点で同じなので
// 1 つの表にまとめる。grant / restriction 側は主体の種類を知らずに 1 本の FK で参照でき、
// 権限を解く SQL が主体の種類だけ分岐することもない。
type PrincipalKind string

const (
	// PrincipalKindUser は 1 人のユーザー（users.id）。ワークスペース所属はこの行の有無で表す。
	PrincipalKindUser PrincipalKind = "user"
	// PrincipalKindGroup は名前を持つユーザーの束。所属は principal_members が持つ。
	PrincipalKindGroup PrincipalKind = "group"
	// PrincipalKindSpaceAll はそのスペースの全メンバー。「既定でチーム全員が編集できる」を
	// 1 行の grant で表すために使う。
	PrincipalKindSpaceAll PrincipalKind = "space_all"
	// PrincipalKindShareLink は公開リンクからの来訪者（ログインしていない相手）。
	// リンクの実体（期限・パスワード・権限）は ShareLink が持つ。
	PrincipalKindShareLink PrincipalKind = "share_link"
)

// ValidPrincipalKinds は principals.kind に保存を許す値の一覧。
var ValidPrincipalKinds = []PrincipalKind{
	PrincipalKindUser,
	PrincipalKindGroup,
	PrincipalKindSpaceAll,
	PrincipalKindShareLink,
}

// Valid は既知の主体種別かを返す（保存前の検証に使う）。
func (k PrincipalKind) Valid() bool {
	for _, v := range ValidPrincipalKinds {
		if v == k {
			return true
		}
	}
	return false
}

// Principal は権限を与える相手（主体）。
//
// 使うフィールドは Kind で決まり、DB 側も CHECK で「その kind のときだけ非 NULL」を強制する
// （kind='user' なら UserID だけ、kind='space_all' なら SpaceID だけ、
// kind='group' なら Name だけが埋まる）。任意の key/value に逃がす形（EAV）は取らない。
type Principal struct {
	ID string `json:"id"`
	// WorkspaceID はテナント境界。grant / restriction からの複合 FK の参照先にもなり、
	// 別ワークスペースの主体へ権限を張ることを DB が弾く。
	WorkspaceID string `json:"workspaceId"`
	// Kind は主体の種類。
	Kind PrincipalKind `json:"kind"`
	// UserID は Kind が user のときの users.id。それ以外は nil。
	UserID *uint64 `json:"userId,omitempty"`
	// SpaceID は Kind が space_all のときの対象スペース。それ以外は nil。
	SpaceID *string `json:"spaceId,omitempty"`
	// PageID は Kind が share_link のときの対象ページ。それ以外は nil。
	// 主体を「それが意味を持つ入れ物」に結び付けることで、ページが消えたら主体も消える。
	PageID *string `json:"pageId,omitempty"`
	// Name は Kind が group のときの表示名。それ以外は空文字
	// （ユーザー名は users、スペース名は spaces が正本で、ここへは写さない）。
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// GrantablePrincipal は権限を張れる相手 1 件と、その表示名。
//
// Principal と分けているのは、名前の出どころが kind ごとに違うため。
// principals.name が埋まるのは group だけで、ユーザー名は users、スペース名は spaces が
// 正本になる。Principal に名前を後入れすると「どの kind なら埋まっているか」が
// 呼び出し側に漏れるので、突き合わせ済みの型を別に持つ。
type GrantablePrincipal struct {
	ID   string        `json:"id"`
	Kind PrincipalKind `json:"kind"`
	// Name は表示名。引けなかった場合は空文字（行は落とさない — 選べない相手が
	// 一覧から黙って消えると、消せない権限が画面に残る）。
	Name string `json:"name"`
}
