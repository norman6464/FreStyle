package domain

// WorkspaceRef は「どのワークスペースに属しているか」を表す値。users.workspace_id は NULL を
// 取り得る（運営管理者のように所属を持たない利用者がいる）ため、未所属を空文字のような番兵値へ
// 潰さず型として表現する。ゼロ値は未所属で、未所属はどのワークスペースとも一致しない。
//
// CompanyRef と対称の API を持つ。テナントの正本を companies から workspaces へ寄せる移行の
// 一環（FRESTYLE-355 段4）で、company_id を直接読んでいた読み取り経路のうち、対象データが
// users テーブル自身であるものから順に、この型を経由する形へ切り替えている。
type WorkspaceRef struct {
	id         string
	affiliated bool
}

// NoWorkspace は未所属を表す WorkspaceRef を返す（WorkspaceRef のゼロ値と同じ）。
func NoWorkspace() WorkspaceRef { return WorkspaceRef{} }

// WorkspaceRefOf は指定したワークスペースに所属していることを表す WorkspaceRef を返す。
func WorkspaceRefOf(workspaceID string) WorkspaceRef {
	return WorkspaceRef{id: workspaceID, affiliated: true}
}

// WorkspaceID は所属ワークスペース ID と、所属しているかを返す。未所属なら ("", false)。
func (r WorkspaceRef) WorkspaceID() (string, bool) { return r.id, r.affiliated }

// Matches は指定したワークスペースと同一かを返す。未所属は「どのワークスペースでもない」ため常に false。
func (r WorkspaceRef) Matches(workspaceID string) bool {
	return r.affiliated && r.id == workspaceID
}
