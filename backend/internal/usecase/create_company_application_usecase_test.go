package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type fakeAppRepo struct {
	created *domain.CompanyApplication
	err     error
}

func (r *fakeAppRepo) Create(_ context.Context, app *domain.CompanyApplication) error {
	if r.err != nil {
		return r.err
	}
	app.ID = 1
	r.created = app
	return nil
}

func (r *fakeAppRepo) ListAll(context.Context) ([]domain.CompanyApplication, error) { return nil, nil }
func (r *fakeAppRepo) UpdateStatus(context.Context, uint64, string) error           { return nil }

// fakeUsersForApp は ListByRole で super_admin を返す最小 UserRepository。
type fakeUsersForApp struct{ admins []domain.User }

func (f *fakeUsersForApp) FindByCognitoSub(context.Context, string) (*domain.User, error) {
	return nil, nil
}

func (f *fakeUsersForApp) FindByID(context.Context, uint64) (*domain.User, error) { return nil, nil }

func (f *fakeUsersForApp) ListByRole(_ context.Context, role string) ([]domain.User, error) {
	if role == domain.RoleSuperAdmin {
		return f.admins, nil
	}
	return nil, nil
}
func (f *fakeUsersForApp) Create(context.Context, *domain.User) error              { return nil }
func (f *fakeUsersForApp) UpdateDisplayName(context.Context, uint64, string) error { return nil }
func (f *fakeUsersForApp) UpdateRole(context.Context, uint64, string) error        { return nil }
func (f *fakeUsersForApp) UpdateCompanyID(context.Context, uint64, uint64) error   { return nil }
func (f *fakeUsersForApp) UpdateActive(context.Context, uint64, bool) error        { return nil }
func (f *fakeUsersForApp) SoftDelete(context.Context, uint64) error                { return nil }
func (f *fakeUsersForApp) MarkOnboarded(context.Context, uint64) error             { return nil }
func (f *fakeUsersForApp) ListByCompanyID(context.Context, uint64) ([]domain.User, error) {
	return nil, nil
}
func (f *fakeUsersForApp) UpdateAiChatEnabled(context.Context, uint64, *bool) error { return nil }

type recordingNotifRepo struct{ created []domain.Notification }

func (r *recordingNotifRepo) Create(_ context.Context, n *domain.Notification) error {
	r.created = append(r.created, *n)
	return nil
}

func (r *recordingNotifRepo) ListByUserID(context.Context, uint64) ([]domain.Notification, error) {
	return nil, nil
}
func (r *recordingNotifRepo) MarkRead(context.Context, uint64, uint64) error { return nil }
func (r *recordingNotifRepo) MarkAllRead(context.Context, uint64) error      { return nil }
func (r *recordingNotifRepo) CountUnread(context.Context, uint64) (int64, error) {
	return 0, nil
}

func TestCreateCompanyApplication_NotifiesSuperAdmins(t *testing.T) {
	apps := &fakeAppRepo{}
	users := &fakeUsersForApp{admins: []domain.User{{ID: 10}, {ID: 11}}}
	notifs := &recordingNotifRepo{}
	uc := usecase.NewCreateCompanyApplicationUseCase(apps, users, notifs)

	app, err := uc.Execute(context.Background(), usecase.CreateCompanyApplicationInput{
		CompanyName:   "Example Corp",
		ApplicantName: "山田太郎",
		Email:         "yamada@example.com",
		Message:       "利用を検討しています",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.Status != domain.CompanyApplicationStatusPending {
		t.Fatalf("status should be pending, got %q", app.Status)
	}
	if len(notifs.created) != 2 {
		t.Fatalf("expected 2 super_admin notifications, got %d", len(notifs.created))
	}
	if notifs.created[0].Type != domain.NotificationTypeCompanyApplication {
		t.Fatalf("notification type mismatch: %q", notifs.created[0].Type)
	}
}

func TestCreateCompanyApplication_Validation(t *testing.T) {
	uc := usecase.NewCreateCompanyApplicationUseCase(&fakeAppRepo{}, &fakeUsersForApp{}, &recordingNotifRepo{})
	cases := []usecase.CreateCompanyApplicationInput{
		{CompanyName: "", ApplicantName: "a", Email: "a@b.com"},     // company 欠落
		{CompanyName: "c", ApplicantName: "", Email: "a@b.com"},     // name 欠落
		{CompanyName: "c", ApplicantName: "a", Email: ""},           // email 欠落
		{CompanyName: "c", ApplicantName: "a", Email: "no-at-sign"}, // email 形式不正
	}
	for i, in := range cases {
		if _, err := uc.Execute(context.Background(), in); !errors.Is(err, usecase.ErrCompanyApplicationInvalid) {
			t.Errorf("case %d: expected ErrCompanyApplicationInvalid, got %v", i, err)
		}
	}
}

func TestCreateCompanyApplication_SavesEvenIfNotifyFails(t *testing.T) {
	// 通知作成に失敗しても申請保存は成功扱い（best-effort）。
	apps := &fakeAppRepo{}
	users := &fakeUsersForApp{admins: []domain.User{{ID: 10}}}
	uc := usecase.NewCreateCompanyApplicationUseCase(apps, users, &failingNotifRepo{})
	if _, err := uc.Execute(context.Background(), usecase.CreateCompanyApplicationInput{
		CompanyName: "c", ApplicantName: "a", Email: "a@b.com",
	}); err != nil {
		t.Fatalf("application should be created despite notify failure, got %v", err)
	}
	if apps.created == nil {
		t.Fatal("application was not saved")
	}
}

type failingNotifRepo struct{ recordingNotifRepo }

func (r *failingNotifRepo) Create(context.Context, *domain.Notification) error {
	return errors.New("boom")
}
