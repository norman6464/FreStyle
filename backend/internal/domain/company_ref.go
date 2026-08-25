package domain

// CompanyRef は「どの会社に属しているか」を表す値。users.company_id は NULL を取り得る
// （運営管理者のように会社へ属さない利用者がいる）ため、未所属を 0 のような番兵値へ潰さず
// 型として表現する。ゼロ値は未所属で、未所属はどの会社とも一致しない。
//
// 「未所属のときどうするか」は経路ごとに異なる（素通り / 許可 / 空一覧 / 拒否）ので、
// 受け取り側は CompanyID の第 2 戻り値か Matches を使い、必ず明示的に分岐する。
type CompanyRef struct {
	id         uint64
	affiliated bool
}

// NoCompany は未所属を表す CompanyRef を返す（CompanyRef のゼロ値と同じ）。
func NoCompany() CompanyRef { return CompanyRef{} }

// CompanyRefOf は指定した会社に所属していることを表す CompanyRef を返す。
func CompanyRefOf(companyID uint64) CompanyRef {
	return CompanyRef{id: companyID, affiliated: true}
}

// CompanyID は所属会社 ID と、会社に所属しているかを返す。未所属なら (0, false)。
func (r CompanyRef) CompanyID() (uint64, bool) { return r.id, r.affiliated }

// Matches は指定した会社と同一かを返す。未所属は「どの会社でもない」ため常に false。
func (r CompanyRef) Matches(companyID uint64) bool {
	return r.affiliated && r.id == companyID
}
