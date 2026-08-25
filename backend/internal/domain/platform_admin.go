package domain

import "slices"

// CognitoAdminGroupName は運営権限（プラットフォーム全体の管理者）を表す Cognito グループ名。
// case-sensitive。グループ名を知るのはこの 1 箇所だけにする。
const CognitoAdminGroupName = "admin"

// PlatformAdminClaim は cognito:groups claim から読み取れる「運営権限の事実」を表す。
//
// 3 値なのは、claim が「無い」ことと「あって admin を含まない」ことを混同させないため。
// federated（Google 連携等）のトークンには groups claim 自体が載らないことがあり、
// 欠落を「グループに居ない」と解釈すると正当な運営管理者を締め出す。
type PlatformAdminClaim int

const (
	// PlatformAdminClaimAbsent は claim キー自体が無い状態。権限について何も分からないので、
	// 付与も剥奪もしない。
	PlatformAdminClaimAbsent PlatformAdminClaim = iota
	// PlatformAdminClaimGranted は claim があり admin グループを含む状態。
	PlatformAdminClaimGranted
	// PlatformAdminClaimRevoked は claim があり admin グループを含まない状態。
	// グループから外れた（オフボーディング）ことを意味する。
	PlatformAdminClaimRevoked
)

// PlatformAdminFromGroups は cognito:groups claim を PlatformAdminClaim へ畳む。
// present は claim キーが JSON に存在したか（空配列との区別が要る）。
func PlatformAdminFromGroups(present bool, groups []string) PlatformAdminClaim {
	if !present {
		return PlatformAdminClaimAbsent
	}
	if slices.Contains(groups, CognitoAdminGroupName) {
		return PlatformAdminClaimGranted
	}
	return PlatformAdminClaimRevoked
}

// Decided は claim が運営権限を決めているかと、その値を返す。
// decided=false（claim 欠落）のときは grant を見てはならない。
func (c PlatformAdminClaim) Decided() (grant bool, decided bool) {
	switch c {
	case PlatformAdminClaimGranted:
		return true, true
	case PlatformAdminClaimRevoked:
		return false, true
	default:
		return false, false
	}
}

// ResolveEffectiveRole は「保存された役割」と「運営権限が今も在るか」から、
// この主体に通してよい実効役割を返す。認可はすべてこの結果だけを見る。
//
// 役割の実効値を決める規則はこの関数 1 つが持つ。判定を各所へ写経しない
// （role を見ている箇所は 10 箇所以上あり、1 つ書き忘れれば穴になる）。
//
// users.role_id は下げない。下げ先が決まらないため（元が trainee だったのか
// company_admin だったのかを DB は覚えていない）。代わりに運営権限の在否を
// users.is_platform_admin が別に持ち、失効した super_admin は最小権限へ倒す。
func ResolveEffectiveRole(stored RoleName, isPlatformAdmin bool) RoleName {
	if stored == RoleSuperAdmin && !isPlatformAdmin {
		return RoleTrainee
	}
	return stored
}
