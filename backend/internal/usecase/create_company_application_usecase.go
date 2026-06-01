package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// CreateCompanyApplicationUseCase は公開フォームから来た企業申請を保存し、
// 全 super_admin に「申請が届いた」通知を作成する。
type CreateCompanyApplicationUseCase struct {
	apps          repository.CompanyApplicationRepository
	users         repository.UserRepository
	notifications repository.NotificationRepository
}

func NewCreateCompanyApplicationUseCase(
	apps repository.CompanyApplicationRepository,
	users repository.UserRepository,
	notifications repository.NotificationRepository,
) *CreateCompanyApplicationUseCase {
	return &CreateCompanyApplicationUseCase{apps: apps, users: users, notifications: notifications}
}

type CreateCompanyApplicationInput struct {
	CompanyName   string
	ApplicantName string
	Email         string
	Message       string
}

// ErrCompanyApplicationInvalid は必須項目欠落 / 形式不正のとき返る。
var ErrCompanyApplicationInvalid = errors.New("company application: invalid input")

// 入力長の上限（DB カラム長 + スパム肥大化対策）。
const (
	maxCompanyNameLen   = 200
	maxApplicantNameLen = 120
	maxEmailLen         = 255
	maxMessageLen       = 2000
)

func (u *CreateCompanyApplicationUseCase) Execute(ctx context.Context, in CreateCompanyApplicationInput) (*domain.CompanyApplication, error) {
	company := strings.TrimSpace(in.CompanyName)
	name := strings.TrimSpace(in.ApplicantName)
	email := strings.TrimSpace(in.Email)
	message := strings.TrimSpace(in.Message)

	if company == "" || name == "" || email == "" {
		return nil, fmt.Errorf("%w: companyName, applicantName, email are required", ErrCompanyApplicationInvalid)
	}
	if !strings.Contains(email, "@") || strings.Contains(email, " ") {
		return nil, fmt.Errorf("%w: invalid email", ErrCompanyApplicationInvalid)
	}
	if len(company) > maxCompanyNameLen || len(name) > maxApplicantNameLen ||
		len(email) > maxEmailLen || len(message) > maxMessageLen {
		return nil, fmt.Errorf("%w: field too long", ErrCompanyApplicationInvalid)
	}

	app := &domain.CompanyApplication{
		CompanyName:   company,
		ApplicantName: name,
		Email:         email,
		Message:       message,
		Status:        domain.CompanyApplicationStatusPending,
	}
	if err := u.apps.Create(ctx, app); err != nil {
		return nil, fmt.Errorf("create company application: %w", err)
	}

	// 申請保存は成功扱いとし、通知作成失敗はログのみ（best-effort）。
	u.notifySuperAdmins(ctx, app)
	return app, nil
}

func (u *CreateCompanyApplicationUseCase) notifySuperAdmins(ctx context.Context, app *domain.CompanyApplication) {
	if u.users == nil || u.notifications == nil {
		return
	}
	admins, err := u.users.ListByRole(ctx, domain.RoleSuperAdmin)
	if err != nil {
		log.Printf("company-application: list super_admins failed: %v", err)
		return
	}
	title := "新しい企業申請が届きました"
	body := fmt.Sprintf("%s（%s / %s）から利用申請がありました。", app.CompanyName, app.ApplicantName, app.Email)
	for _, admin := range admins {
		n := &domain.Notification{
			UserID: admin.ID,
			Type:   domain.NotificationTypeCompanyApplication,
			Title:  title,
			Body:   body,
		}
		if err := u.notifications.Create(ctx, n); err != nil {
			log.Printf("company-application: notify super_admin %d failed: %v", admin.ID, err)
		}
	}
}
