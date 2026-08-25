package domain

import "time"

// Capability はページに対してできることの単位。page_restrictions と share_links が使う。
//
// コメント（SpaceRole.CanComment）はここには入れない。コメント機能そのものが段 4 で、
// 例外を張る対象がまだ存在しないため。必要になった時点で値を 1 つ足す
// （DB 側の CHECK と ValidCapabilities の両方を同時に増やすこと）。
type Capability string

const (
	// CapabilityView はページを閲覧できること。
	CapabilityView Capability = "view"
	// CapabilityEdit はページを編集できること。編集できる者は必ず閲覧もできる
	// （ResolvePagePermission が edit に view を含める）。
	CapabilityEdit Capability = "edit"
)

// ValidCapabilities は保存を許すケイパビリティの一覧。
var ValidCapabilities = []Capability{CapabilityView, CapabilityEdit}

// Valid は既知のケイパビリティかを返す（保存前の検証に使う）。
func (c Capability) Valid() bool {
	for _, v := range ValidCapabilities {
		if v == c {
			return true
		}
	}
	return false
}

// RestrictionMode は page_restrictions.mode に入る例外の向き。
type RestrictionMode string

const (
	// RestrictionModeAllow は「この主体には許す」。同じ段に allow 行が 1 つでもあると
	// 「載っていない者は入れない」（限定公開）に切り替わる。
	RestrictionModeAllow RestrictionMode = "allow"
	// RestrictionModeDeny は「この主体だけ外す」。deny だけの段では、名指しされていない者は
	// 既定（space_grants）のままになる。
	RestrictionModeDeny RestrictionMode = "deny"
)

// ValidRestrictionModes は保存を許す例外の向きの一覧。
var ValidRestrictionModes = []RestrictionMode{RestrictionModeAllow, RestrictionModeDeny}

// Valid は既知の向きかを返す（保存前の検証に使う）。
func (m RestrictionMode) Valid() bool {
	for _, v := range ValidRestrictionModes {
		if v == m {
			return true
		}
	}
	return false
}

// PageRestriction はそのページ以下だけ既定を上書きする例外。行を持つのは例外だけで、
// 例外の無いページには 1 行も存在しない（全ページへ ACL を展開しない）。
//
// 同じ (ページ, 主体, ケイパビリティ) に allow と deny の 2 行は作れない
// （DB の PK が (workspace_id, page_id, principal_id, capability)）。
//
// Workspace と同じくナレッジ基盤の型なので GORM を通さない。
type PageRestriction struct {
	// WorkspaceID はテナント境界。page / principal との複合 FK に使う。
	WorkspaceID string `json:"workspaceId"`
	// PageID は例外を張るページ。このページとその子孫が対象になる。
	PageID string `json:"pageId"`
	// PrincipalID は例外の対象となる主体。
	PrincipalID string `json:"principalId"`
	// Capability はどのケイパビリティについての例外か。
	Capability Capability `json:"capability"`
	// Mode は例外の向き。
	Mode      RestrictionMode `json:"mode"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}
