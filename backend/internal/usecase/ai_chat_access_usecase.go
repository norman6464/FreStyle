package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// AiChatEnabledForUserUseCase は「この user が AI チャットを使ってよいか」を判定する認可ポリシー。
// /auth/me の表示判定と、ai-chat 系エンドポイントの入口ゲート(middleware)の両方で使う。
//
// ルール: company_admin / super_admin は会社設定に関わらず常に true（自社設定確認のため自分で使う）。
// 会社未所属(company_id=null)も true。それ以外は自社の ai_chat_enabled_for_trainees に従う
// （会社行が見つからないときは安全側に倒さず既定 true）。
type AiChatEnabledForUserUseCase struct {
	companies repository.CompanyRepository
}

func NewAiChatEnabledForUserUseCase(c repository.CompanyRepository) *AiChatEnabledForUserUseCase {
	return &AiChatEnabledForUserUseCase{companies: c}
}

// Execute は user が AI チャットを使えるかを返す。actor は middleware が context に入れた *domain.User。
func (uc *AiChatEnabledForUserUseCase) Execute(ctx context.Context, user *domain.User) (bool, error) {
	if user == nil {
		return false, nil
	}
	if isCompanyAdminRole(user.Role) {
		return true, nil
	}
	if user.CompanyID == nil {
		return true, nil
	}
	// 個別上書き(ai_chat_enabled)が設定されていれば会社一括設定より優先する。
	if user.AiChatEnabled != nil {
		return *user.AiChatEnabled, nil
	}
	company, err := uc.companies.FindByID(ctx, *user.CompanyID)
	if err != nil {
		// 会社行が無い場合は既定 true（後方互換）。それ以外の DB エラーは伝搬する。
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, err
	}
	return company.AiChatEnabledForTrainees, nil
}
