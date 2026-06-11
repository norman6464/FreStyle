package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// settingsCompanyRepo は CompanyRepository の最小スタブ（state 検証・古典派）。
type settingsCompanyRepo struct {
	company    *domain.Company
	findErr    error
	lastEnable *bool
}

func (s *settingsCompanyRepo) ListAll(context.Context) ([]domain.Company, error) { return nil, nil }
func (s *settingsCompanyRepo) FindByID(context.Context, uint64) (*domain.Company, error) {
	return s.company, s.findErr
}

func (s *settingsCompanyRepo) UpdateActive(context.Context, uint64, bool) error { return nil }

func (s *settingsCompanyRepo) UpdateAiChatEnabled(_ context.Context, _ uint64, enabled bool) error {
	s.lastEnable = &enabled
	return nil
}

func u64p(v uint64) *uint64 { return &v }

func TestGetCompanyAiChatSetting(t *testing.T) {
	t.Run("trainee は forbidden", func(t *testing.T) {
		uc := usecase.NewGetCompanyAiChatSettingUseCase(&settingsCompanyRepo{})
		_, err := uc.Execute(context.Background(), &domain.User{Role: domain.RoleTrainee, CompanyID: u64p(1)})
		require.ErrorIs(t, err, usecase.ErrCompanySettingsForbidden)
	})

	t.Run("会社未所属の admin は no_company", func(t *testing.T) {
		uc := usecase.NewGetCompanyAiChatSettingUseCase(&settingsCompanyRepo{})
		_, err := uc.Execute(context.Background(), &domain.User{Role: domain.RoleCompanyAdmin, CompanyID: nil})
		require.ErrorIs(t, err, usecase.ErrCompanySettingsNoCompany)
	})

	t.Run("admin は自社フラグを取得", func(t *testing.T) {
		repo := &settingsCompanyRepo{company: &domain.Company{ID: 1, AiChatEnabledForTrainees: false}}
		uc := usecase.NewGetCompanyAiChatSettingUseCase(repo)
		got, err := uc.Execute(context.Background(), &domain.User{Role: domain.RoleCompanyAdmin, CompanyID: u64p(1)})
		require.NoError(t, err)
		assert.False(t, got)
	})
}

func TestUpdateCompanyAiChatSetting(t *testing.T) {
	t.Run("非 admin は forbidden", func(t *testing.T) {
		uc := usecase.NewUpdateCompanyAiChatSettingUseCase(&settingsCompanyRepo{})
		_, err := uc.Execute(context.Background(), &domain.User{Role: domain.RoleTrainee, CompanyID: u64p(1)}, false)
		require.ErrorIs(t, err, usecase.ErrCompanySettingsForbidden)
	})

	t.Run("admin はフラグを更新する", func(t *testing.T) {
		repo := &settingsCompanyRepo{}
		uc := usecase.NewUpdateCompanyAiChatSettingUseCase(repo)
		got, err := uc.Execute(context.Background(), &domain.User{Role: domain.RoleSuperAdmin, CompanyID: u64p(7)}, false)
		require.NoError(t, err)
		assert.False(t, got)
		require.NotNil(t, repo.lastEnable)
		assert.False(t, *repo.lastEnable)
	})
}

func TestAiChatEnabledForUser(t *testing.T) {
	t.Run("admin は会社設定に関わらず常に true", func(t *testing.T) {
		repo := &settingsCompanyRepo{company: &domain.Company{AiChatEnabledForTrainees: false}}
		uc := usecase.NewAiChatEnabledForUserUseCase(repo)
		got, err := uc.Execute(context.Background(), &domain.User{Role: domain.RoleCompanyAdmin, CompanyID: u64p(1)})
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("会社未所属の trainee は true", func(t *testing.T) {
		uc := usecase.NewAiChatEnabledForUserUseCase(&settingsCompanyRepo{})
		got, err := uc.Execute(context.Background(), &domain.User{Role: domain.RoleTrainee, CompanyID: nil})
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("trainee は自社フラグに従う(false)", func(t *testing.T) {
		repo := &settingsCompanyRepo{company: &domain.Company{AiChatEnabledForTrainees: false}}
		uc := usecase.NewAiChatEnabledForUserUseCase(repo)
		got, err := uc.Execute(context.Background(), &domain.User{Role: domain.RoleTrainee, CompanyID: u64p(1)})
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("会社行が無い(RecordNotFound)なら既定 true", func(t *testing.T) {
		repo := &settingsCompanyRepo{findErr: gorm.ErrRecordNotFound}
		uc := usecase.NewAiChatEnabledForUserUseCase(repo)
		got, err := uc.Execute(context.Background(), &domain.User{Role: domain.RoleTrainee, CompanyID: u64p(99)})
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("個別 OFF は会社設定(true)より優先される", func(t *testing.T) {
		repo := &settingsCompanyRepo{company: &domain.Company{AiChatEnabledForTrainees: true}}
		uc := usecase.NewAiChatEnabledForUserUseCase(repo)
		got, err := uc.Execute(context.Background(), &domain.User{Role: domain.RoleTrainee, CompanyID: u64p(1), AiChatEnabled: ptrBool(false)})
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("個別 ON は会社設定(false)より優先される", func(t *testing.T) {
		repo := &settingsCompanyRepo{company: &domain.Company{AiChatEnabledForTrainees: false}}
		uc := usecase.NewAiChatEnabledForUserUseCase(repo)
		got, err := uc.Execute(context.Background(), &domain.User{Role: domain.RoleTrainee, CompanyID: u64p(1), AiChatEnabled: ptrBool(true)})
		require.NoError(t, err)
		assert.True(t, got)
	})
}
