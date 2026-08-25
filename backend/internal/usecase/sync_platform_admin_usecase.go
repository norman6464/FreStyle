package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// SyncPlatformAdminInput は同期対象（Cognito の subject）と、トークンから読んだ運営権限の事実。
type SyncPlatformAdminInput struct {
	CognitoSub string
	Claim      domain.PlatformAdminClaim
}

// SyncPlatformAdminUseCase は Cognito の admin グループ所属を users.is_platform_admin へ反映する。
// 付与（グループに居る）と剥奪（グループから外れた）の両方を扱うが、role_id には触らない。
//
// 剥奪はオフボーディングの要。これが無いと、Cognito のグループから外しても DB の
// super_admin が剥がれず、退任者が全顧客企業のデータへアクセスし続けられる。
type SyncPlatformAdminUseCase struct {
	users repository.UserRepository
}

// NewSyncPlatformAdminUseCase は SyncPlatformAdminUseCase を生成する。
func NewSyncPlatformAdminUseCase(users repository.UserRepository) *SyncPlatformAdminUseCase {
	return &SyncPlatformAdminUseCase{users: users}
}

// Execute は claim が運営権限を決めているときだけ DB を書き換える。
// changed は実際に値が変わったか。ユーザーが居ない、または既に同じ値なら (false, nil)。
//
// claim が存在しない（PlatformAdminClaimAbsent）ときは何もしない。groups claim は
// federated ユーザーのトークンに載らないことがあり、欠落を「グループに居ない」と
// 解釈すると正当な運営管理者を締め出すため。「知らない」は「否」ではない。
func (u *SyncPlatformAdminUseCase) Execute(
	ctx context.Context,
	in SyncPlatformAdminInput,
) (changed bool, err error) {
	if u.users == nil {
		return false, errors.New("user repository not configured")
	}
	if in.CognitoSub == "" {
		return false, errors.New("cognito sub is empty")
	}

	grant, decided := in.Claim.Decided()
	if !decided {
		return false, nil
	}

	existing, findErr := u.users.FindByCognitoSub(ctx, in.CognitoSub)
	if findErr != nil {
		// 「DB 障害」と「そのユーザーが居ない」は区別する。障害を無反応に畳むと、
		// 剥奪し損ねたことに誰も気付けない。
		return false, fmt.Errorf("find user by cognito sub: %w", findErr)
	}
	if existing == nil {
		return false, nil
	}
	if existing.IsPlatformAdmin == grant {
		return false, nil
	}

	if err := u.users.UpdatePlatformAdmin(ctx, existing.ID, grant); err != nil {
		return false, fmt.Errorf("update platform admin: %w", err)
	}
	return true, nil
}
