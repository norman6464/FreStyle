package domain

import "time"

// TeachingMaterial はコースを構成する「章」。必ず 1 つの Course に所属する（course 1 : N chapter）。
// テーブルは course_chapters（FRESTYLE-184 で teaching_materials から改名）。
//
// 本文はリッチテキスト（tiptap の ProseMirror JSON）の doc(jsonb) が正本。
// コース内の並び順は sort_order 列（同値時 ID 昇順）。
type TeachingMaterial struct {
	ID              uint64 `json:"id"`
	CompanyID       uint64 `json:"companyId"`
	CourseID        uint64 `json:"courseId"`
	CreatedByUserID uint64 `json:"createdByUserId"`
	Title           string `json:"title"`
	// Doc はリッチテキスト本文（tiptap JSON）。未移行の章は NULL。
	// JSON 出力は handler 側で json.RawMessage として制御するため json:"-"。
	Doc *string `json:"-"`
	// Revision は doc 更新の楽観ロック用。doc を更新するたびに +1（不一致は 409）。
	Revision int `json:"revision"`
	// SchemaVersion は doc のエディタスキーマ版。読込時アップキャストの目印（現行 1）。
	SchemaVersion int  `json:"schemaVersion"`
	OrderInCourse int  `json:"orderInCourse"`
	IsPublished   bool `json:"isPublished"`
	// WorkspaceID は所属ワークスペースへの参照。CompanyID から dual-write されるため
	// 通常は必ず埋まっているが、Course と同じ理由で NULL を許容する型にする（FRESTYLE-403）。
	WorkspaceID *string   `json:"workspaceId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// WorkspaceRef は所属ワークスペースへの参照を返す。未設定(workspace_id = NULL)は NoWorkspace。
func (m TeachingMaterial) WorkspaceRef() WorkspaceRef {
	if m.WorkspaceID == nil {
		return NoWorkspace()
	}
	return WorkspaceRefOf(*m.WorkspaceID)
}
