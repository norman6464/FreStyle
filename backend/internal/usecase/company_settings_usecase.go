package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// 会社設定（trainee への AI 有効化）の操作で返すセンチネルエラー。handler が HTTP ステータスに分岐する。
var (
	// ErrCompanySettingsForbidden は company_admin / super_admin 以外が操作したとき。403。
	ErrCompanySettingsForbidden = errors.New("forbidden")
	// ErrCompanySettingsNoCompany は actor が会社未所属（company_id=null）のとき。400。
	ErrCompanySettingsNoCompany = errors.New("no_company")
)

func isCompanyAdminRole(role string) bool {
	return role == domain.RoleCompanyAdmin || role == domain.RoleSuperAdmin
}

// GetCompanyAiChatSettingUseCase は自社の AI 有効化フラグを取得する（company_admin / super_admin のみ）。
type GetCompanyAiChatSettingUseCase struct {
	companies repository.CompanyRepository
}

func NewGetCompanyAiChatSettingUseCase(c repository.CompanyRepository) *GetCompanyAiChatSettingUseCase {
	return &GetCompanyAiChatSettingUseCase{companies: c}
}

// Execute は actor の自社設定を返す。actor は middleware が context に入れた *domain.User。
func (uc *GetCompanyAiChatSettingUseCase) Execute(ctx context.Context, actor *domain.User) (bool, error) {
	if actor == nil || !isCompanyAdminRole(actor.Role) {
		return false, ErrCompanySettingsForbidden
	}
	if actor.CompanyID == nil {
		return false, ErrCompanySettingsNoCompany
	}
	company, err := uc.companies.FindByID(ctx, *actor.CompanyID)
	if err != nil {
		return false, err
	}
	return company.AiChatEnabledForTrainees, nil
}

// UpdateCompanyAiChatSettingUseCase は自社の AI 有効化フラグを更新する（company_admin / super_admin のみ）。
type UpdateCompanyAiChatSettingUseCase struct {
	companies repository.CompanyRepository
}

func NewUpdateCompanyAiChatSettingUseCase(c repository.CompanyRepository) *UpdateCompanyAiChatSettingUseCase {
	return &UpdateCompanyAiChatSettingUseCase{companies: c}
}

// Execute は actor の自社フラグを enabled に更新し、更新後の値を返す。
func (uc *UpdateCompanyAiChatSettingUseCase) Execute(ctx context.Context, actor *domain.User, enabled bool) (bool, error) {
	if actor == nil || !isCompanyAdminRole(actor.Role) {
		return false, ErrCompanySettingsForbidden
	}
	if actor.CompanyID == nil {
		return false, ErrCompanySettingsNoCompany
	}
	if err := uc.companies.UpdateAiChatEnabled(ctx, *actor.CompanyID, enabled); err != nil {
		return false, err
	}
	return enabled, nil
}
