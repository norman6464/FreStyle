package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ErrMemberNotInActorCompany は対象 user が actor の会社に属していないときのエラー（403 相当）。
var ErrMemberNotInActorCompany = errors.New("member is not in the actor's company")

// UpdateMemberAiAccessUseCase は company_admin が自社従業員の AI 利用可否を個別に上書きする。
// 別会社の user は更新できない（テナント境界）。enabled=nil で会社設定に従う状態に戻す。
type UpdateMemberAiAccessUseCase struct {
	users repository.UserRepository
}

func NewUpdateMemberAiAccessUseCase(u repository.UserRepository) *UpdateMemberAiAccessUseCase {
	return &UpdateMemberAiAccessUseCase{users: u}
}

// Execute は actor の自社の targetUserID の AI 利用可否を enabled に更新する。
func (uc *UpdateMemberAiAccessUseCase) Execute(ctx context.Context, actor *domain.User, targetUserID uint64, enabled *bool) error {
	if actor == nil || actor.CompanyID == nil {
		return ErrMemberNotInActorCompany
	}
	target, err := uc.users.FindByID(ctx, targetUserID)
	if err != nil {
		return err
	}
	if target == nil || target.CompanyID == nil || *target.CompanyID != *actor.CompanyID {
		return ErrMemberNotInActorCompany
	}
	return uc.users.UpdateAiChatEnabled(ctx, targetUserID, enabled)
}
