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
func (f *fakeUsersForApp) Create(context.Context, *domain.User) error            { return nil }
func (f *fakeUsersForApp) UpdateName(context.Context, uint64, string) error      { return nil }
func (f *fakeUsersForApp) UpdateRole(context.Context, uint64, string) error      { return nil }
func (f *fakeUsersForApp) UpdateCompanyID(context.Context, uint64, uint64) error { return nil }
func (f *fakeUsersForApp) UpdateActive(context.Context, uint64, bool) error      { return nil }
func (f *fakeUsersForApp) SoftDelete(context.Context, uint64) error              { return nil }
func (f *fakeUsersForApp) MarkOnboarded(context.Context, uint64) error           { return nil }
func (f *fakeUsersForApp) ListByCompanyID(context.Context, uint64) ([]domain.User, error) {
	return nil, nil
}
func (f *fakeUsersForApp) UpdateAiChatEnabled(context.Context, uint64, *bool) error { return nil }

type recordingNotifRepo struct {
	created []domain.Notification
	// createManyCalls は「宛先の人数によらず 1 回で書き込む」ことを検証するための回数。
	createManyCalls int
	// createCalls は 1 件ずつ書き込む旧経路に戻っていないことの検証用。
	createCalls int
}

func (r *recordingNotifRepo) Create(_ context.Context, n *domain.Notification) error {
	r.createCalls++
	r.created = append(r.created, *n)
	return nil
}

func (r *recordingNotifRepo) CreateMany(_ context.Context, ns []domain.Notification) error {
	r.created = append(r.created, ns...)
	r.createManyCalls++
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

func Test_会社申請作成_運営管理者へ通知(t *testing.T) {
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

func Test_会社申請作成_バリデーション(t *testing.T) {
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

func Test_会社申請作成_通知失敗でも保存(t *testing.T) {
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

func (r *failingNotifRepo) CreateMany(context.Context, []domain.Notification) error {
	return errors.New("boom")
}

// 宛先が増えても DB への書き込みは 1 回に保つ（FRESTYLE-17）。
// 1 件ずつ書き込む実装だと管理者の人数に比例して往復が増え、申請処理が遅くなる。
func Test_会社申請作成_通知はまとめて1回で書き込む(t *testing.T) {
	makeAdmins := func(n int) []domain.User {
		out := make([]domain.User, 0, n)
		for i := 1; i <= n; i++ {
			out = append(out, domain.User{ID: uint64(i)})
		}
		return out
	}

	tests := []struct {
		name            string
		admins          []domain.User
		wantCreateMany  int // まとめ書き込みの回数
		wantNotifations int // 作られる通知の件数
	}{
		{name: "管理者が0人", admins: nil, wantCreateMany: 1, wantNotifations: 0},
		{name: "管理者が1人", admins: makeAdmins(1), wantCreateMany: 1, wantNotifations: 1},
		{name: "管理者が5人", admins: makeAdmins(5), wantCreateMany: 1, wantNotifations: 5},
		{name: "管理者が20人", admins: makeAdmins(20), wantCreateMany: 1, wantNotifations: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifs := &recordingNotifRepo{}
			uc := usecase.NewCreateCompanyApplicationUseCase(
				&fakeAppRepo{}, &fakeUsersForApp{admins: tt.admins}, notifs,
			)

			app, err := uc.Execute(context.Background(), usecase.CreateCompanyApplicationInput{
				CompanyName:   "Example Corp",
				ApplicantName: "山田太郎",
				Email:         "yamada@example.com",
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if app.Status != domain.CompanyApplicationStatusPending {
				t.Fatalf("status should be pending, got %q", app.Status)
			}

			// 人数によらず書き込みは 1 回。
			if notifs.createManyCalls != tt.wantCreateMany {
				t.Fatalf("まとめ書き込みの回数: want %d, got %d", tt.wantCreateMany, notifs.createManyCalls)
			}
			// 1 件ずつ書き込む旧経路に戻っていないこと。
			if notifs.createCalls != 0 {
				t.Fatalf("1 件ずつの Create は呼ばないこと: got %d", notifs.createCalls)
			}
			if len(notifs.created) != tt.wantNotifations {
				t.Fatalf("通知の件数: want %d, got %d", tt.wantNotifations, len(notifs.created))
			}
			// 宛先が取り違えられていないこと。
			for i, n := range notifs.created {
				if n.UserID != tt.admins[i].ID {
					t.Fatalf("宛先が一致しない: index=%d want=%d got=%d", i, tt.admins[i].ID, n.UserID)
				}
				if n.Type != domain.NotificationTypeCompanyApplication {
					t.Fatalf("種別が一致しない: %q", n.Type)
				}
			}
		})
	}
}
