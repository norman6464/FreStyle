package domain

import "time"

// GrantRole は grant（既定の権限）で与える役割。workspace_grants と space_grants が共有する。
//
// 既存アプリの RoleName（super_admin / company_admin / trainee）とは別物で、統合もしない。
// あちらは「アプリ全体で何ができるか」を 1 人 1 つ持つグローバルなロール、こちらは
// 「この入れ物（ワークスペース / スペース）で何ができるか」を入れ物ごとに持つ。
// ノートの権限は principals / grants だけで閉じており、
// 「特権ロールなら全部見える」という抜け道は解決ロジックに持たせない。
type GrantRole string

const (
	// GrantRoleAdmin は管理者。閲覧・編集に加えて権限そのものを変えられる。
	GrantRoleAdmin GrantRole = "admin"
	// GrantRoleEditor は閲覧と編集ができる。
	GrantRoleEditor GrantRole = "editor"
	// GrantRoleCommenter は閲覧とコメントができる（コメント機能自体は段 4）。
	GrantRoleCommenter GrantRole = "commenter"
	// GrantRoleViewer は閲覧だけができる。
	GrantRoleViewer GrantRole = "viewer"
)

// ValidGrantRoles は grant に保存を許す役割の一覧（強い順）。
var ValidGrantRoles = []GrantRole{
	GrantRoleAdmin,
	GrantRoleEditor,
	GrantRoleCommenter,
	GrantRoleViewer,
}

// Valid は既知の役割かを返す（保存前の検証に使う）。
func (r GrantRole) Valid() bool {
	for _, v := range ValidGrantRoles {
		if v == r {
			return true
		}
	}
	return false
}

// Rank は役割の強さ（大きいほど強い）。未知の値は 0（何もできない）。
//
// 1 人は複数の経路で役割を得る（自分の主体 / 所属グループ / スペース全員 / ワークスペースの grant）。
// そのとき採るのは**最も強いもの**で、これが唯一の合成規則。理由は 2 つ:
//
//   - 順序に依存しない。弱い方を採る規則にすると、grant を張った順や主体の並び順で
//     結果が変わり、「なぜ見えないのか」を説明できなくなる
//   - ワークスペース管理者がスペース単位の grant で降格されない。降格できてしまうと
//     テナント全体の管理者という概念が成り立たない（スペースを 1 つ作って viewer を
//     張るだけで管理者を締め出せる）
//
// **弱める手段はどの層にも無い。** 付与は足し算だけで、下の段が上の段を打ち消すことはない。
// 「親は共有、この子だけ隠す」は書けず、狭めたい内容は private のスペースへ置く
// （理由は PagePermissionFacts の doc）。
func (r GrantRole) Rank() int {
	switch r {
	case GrantRoleAdmin:
		return 4
	case GrantRoleEditor:
		return 3
	case GrantRoleCommenter:
		return 2
	case GrantRoleViewer:
		return 1
	default:
		return 0
	}
}

// GrantRoleByRank は Rank の逆写像。既知の強さなら対応する役割を、そうでなければ nil を返す。
//
// 「最も強い役割」の判定を DB 側で行うとき、役割そのもの（text）を返すと
// 「grant が 1 つも無い」が NULL になり、生成コードの型付けが崩れる。強さ（整数）で受け取り、
// 0（該当なし）を nil に、それ以外をこの関数で役割へ戻す。変換は persistence 層で完結し、
// 整数としての強さが usecase 以上へ漏れることはない。
func GrantRoleByRank(rank int) *GrantRole {
	for _, r := range ValidGrantRoles {
		if r.Rank() == rank {
			role := r
			return &role
		}
	}
	return nil
}

// StrongestGrantRole は複数の経路で得た役割のうち最も強いものを返す（1 つも無ければ nil）。
// Rank のコメントにある「採るのは最も強いもの」という合成規則を Go 側で適用する唯一の関数。
//
// 事実（どの役割を持っているか）を集めるのは SQL、畳み方を決めるのはここ、と分けている。
// 規則を SQL へ写すと、片方だけ直したときに「1 ページを開くと編集できるのに、
// 同じスペースの直下にページを作れない」といった経路ごとのずれになる。
// 未知の値（Rank が 0）は役割として数えない。
//
// ページ 1 枚 / ページ一覧の経路だけは別で、SQL が役割の集合ではなく強さ（整数）を返す
// （ResolvePagePermissionFacts のコメント参照。ページごとに集約するため、
// 役割を集合のまま返すと 1 リクエストで多数のページを扱えない）。強さから役割へは
// GrantRoleByRank で戻し、両経路の答えが一致することは結合テストで固定する。
func StrongestGrantRole(roles []GrantRole) *GrantRole {
	var strongest *GrantRole
	for _, r := range roles {
		if r.Rank() == 0 {
			continue
		}
		if strongest == nil || r.Rank() > strongest.Rank() {
			role := r
			strongest = &role
		}
	}
	return strongest
}

// CanView は既定でページを閲覧できる役割かを返す。
func (r GrantRole) CanView() bool { return r.Rank() >= GrantRoleViewer.Rank() }

// CanComment は既定でコメントできる役割かを返す（コメント機能自体は段 4）。
func (r GrantRole) CanComment() bool { return r.Rank() >= GrantRoleCommenter.Rank() }

// CanEdit は既定でページを編集できる役割かを返す。
func (r GrantRole) CanEdit() bool { return r.Rank() >= GrantRoleEditor.Rank() }

// CanManage は権限そのもの（grant / 共有リンク）を変えられる役割かを返す。
func (r GrantRole) CanManage() bool { return r.Rank() >= GrantRoleAdmin.Rank() }

// WorkspaceGrant はワークスペース全体での既定の権限。配下の全スペースに効く。
//
// スペース単位の grant だけでは「テナント全体の管理者」を表すのにスペースの数だけ
// grant を張って回ることになり、スペースが増えるたびに漏れる。入れ物の階層が
// ワークスペース ⊃ スペース である以上、既定も 2 段で持つ。
type WorkspaceGrant struct {
	// WorkspaceID は対象ワークスペース。
	WorkspaceID string `json:"workspaceId"`
	// PrincipalID は権限を与える相手。
	PrincipalID string `json:"principalId"`
	// Role は既定の役割。
	Role      GrantRole `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PageGrant はページ以下での既定の権限。workspace / space に続く 3 段目で、経路を
// さかのぼって効き、最も強いものが実効になる（合成は他の 2 段と同じ）。
type PageGrant struct {
	// WorkspaceID はテナント境界。principal との複合 FK に使う。
	WorkspaceID string `json:"workspaceId"`
	// PageID は対象ページ。この付与はこのページとその子孫に効く。
	PageID string `json:"pageId"`
	// PrincipalID は権限を与える相手。
	PrincipalID string `json:"principalId"`
	// Role は既定の役割。
	Role      GrantRole `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// SpaceGrant はスペースでの既定の権限（誰が何をできるか）。
// 同じ主体がひとつのスペースで持つ役割は 1 つだけ（DB の PK が (workspace_id, space_id, principal_id)）。
type SpaceGrant struct {
	// WorkspaceID はテナント境界。principal との複合 FK に使う。
	WorkspaceID string `json:"workspaceId"`
	// SpaceID は対象スペース。
	SpaceID string `json:"spaceId"`
	// PrincipalID は権限を与える相手。
	PrincipalID string `json:"principalId"`
	// Role は既定の役割。
	Role      GrantRole `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
