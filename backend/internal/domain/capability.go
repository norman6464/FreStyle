package domain

// Capability はページに対してできることの単位。share_links が使う。
//
// コメント（SpaceRole.CanComment）はここには入れない。コメント機能そのものが段 4 で、
// 共有リンクに渡す既定としてまだ意味を持たないため。必要になった時点で値を 1 つ足す
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
