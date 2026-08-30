package domain

// WorkspaceRef は「どのワークスペースに属しているか」を表す値。所属を持たない利用者もいる
// ため、未所属を空文字のような番兵値へ潰さず型として表現する。ゼロ値は未所属で、
// 未所属はどのワークスペースとも一致しない。
//
// 「未所属のときどうするか」は経路ごとに異なる（素通り / 許可 / 空一覧 / 拒否）ので、
// 受け取り側は WorkspaceID の第 2 戻り値か Matches を使い、必ず明示的に分岐する。
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
