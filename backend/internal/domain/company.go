package domain

import "time"

type Company struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	// IsActive は会社アカウントの有効/無効。false（無効）にすると、その会社の全ユーザーが
	// ログイン/利用不可になる（middleware で弾く）。super_admin が会社一覧から切り替える。
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	// WorkspaceID は会社に対応するワークスペース（1:1）。テナントの正本は workspace_id 側で、
	// companies は「会社という実体」を表す表として残る。両者を繋ぐ唯一の列がこれ。
	// 起動時バックフィルが未到達の会社は NULL になり得る。
	WorkspaceID *string `json:"workspaceId,omitempty"`
}

// WorkspaceRef は対応ワークスペースへの参照を返す。未紐付け(workspace_id = NULL)は NoWorkspace。
func (c Company) WorkspaceRef() WorkspaceRef {
	if c.WorkspaceID == nil {
		return NoWorkspace()
	}
	return WorkspaceRefOf(*c.WorkspaceID)
}
