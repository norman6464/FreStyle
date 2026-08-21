package usecase_test

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/mock"
)

// usecase テストで共有する repository interface の testify/mock 実装。
// 1 interface = 1 定義に集約し、各テストファイルでの重複定義を禁止する
// (書き方の見本は send_ai_message_stream_usecase_test.go と同じ流儀)。

// --- mock: UserRepository ---

type mockUserRepo struct{ mock.Mock }

var _ repository.UserRepository = (*mockUserRepo)(nil)

func (m *mockUserRepo) FindByCognitoSub(ctx context.Context, sub string) (*domain.User, error) {
	args := m.Called(ctx, sub)
	u, _ := args.Get(0).(*domain.User)
	return u, args.Error(1)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uint64) (*domain.User, error) {
	args := m.Called(ctx, id)
	u, _ := args.Get(0).(*domain.User)
	return u, args.Error(1)
}

func (m *mockUserRepo) ListByRole(ctx context.Context, role domain.RoleName) ([]domain.User, error) {
	args := m.Called(ctx, role)
	rows, _ := args.Get(0).([]domain.User)
	return rows, args.Error(1)
}

func (m *mockUserRepo) ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.User, error) {
	args := m.Called(ctx, companyID)
	rows, _ := args.Get(0).([]domain.User)
	return rows, args.Error(1)
}

func (m *mockUserRepo) CreateWithOidcIdentity(ctx context.Context, u *domain.User, provider, subject string) error {
	return m.Called(ctx, u, provider, subject).Error(0)
}

func (m *mockUserRepo) EnsureOidcIdentity(ctx context.Context, userID uint64, provider, subject string) error {
	return m.Called(ctx, userID, provider, subject).Error(0)
}

func (m *mockUserRepo) FindActiveByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	u, _ := args.Get(0).(*domain.User)
	return u, args.Error(1)
}

func (m *mockUserRepo) CognitoSubjectByUserID(ctx context.Context, userID uint64) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *mockUserRepo) UpdateAiChatEnabled(ctx context.Context, userID uint64, enabled *bool) error {
	return m.Called(ctx, userID, enabled).Error(0)
}

func (m *mockUserRepo) UpdateActive(ctx context.Context, userID uint64, active bool) error {
	return m.Called(ctx, userID, active).Error(0)
}

func (m *mockUserRepo) SoftDelete(ctx context.Context, userID uint64) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockUserRepo) UpdateName(ctx context.Context, userID uint64, name string) error {
	return m.Called(ctx, userID, name).Error(0)
}

func (m *mockUserRepo) UpdateRole(ctx context.Context, userID uint64, role domain.RoleName) error {
	return m.Called(ctx, userID, role).Error(0)
}

func (m *mockUserRepo) UpdateCompanyID(ctx context.Context, userID, companyID uint64) error {
	return m.Called(ctx, userID, companyID).Error(0)
}

// --- mock: CourseRepository ---

type mockCourseRepo struct{ mock.Mock }

var _ repository.CourseRepository = (*mockCourseRepo)(nil)

func (m *mockCourseRepo) ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.Course, error) {
	args := m.Called(ctx, companyID, includeUnpublished)
	rows, _ := args.Get(0).([]domain.Course)
	return rows, args.Error(1)
}

func (m *mockCourseRepo) GetByID(ctx context.Context, id uint64) (*domain.Course, error) {
	args := m.Called(ctx, id)
	c, _ := args.Get(0).(*domain.Course)
	return c, args.Error(1)
}

func (m *mockCourseRepo) Create(ctx context.Context, c *domain.Course) error {
	return m.Called(ctx, c).Error(0)
}

func (m *mockCourseRepo) Update(ctx context.Context, c *domain.Course) error {
	return m.Called(ctx, c).Error(0)
}

func (m *mockCourseRepo) Delete(ctx context.Context, id uint64) error {
	return m.Called(ctx, id).Error(0)
}

// --- mock: TeachingMaterialRepository ---

type mockMaterialRepo struct{ mock.Mock }

var _ repository.TeachingMaterialRepository = (*mockMaterialRepo)(nil)

func (m *mockMaterialRepo) ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
	args := m.Called(ctx, companyID, includeUnpublished)
	rows, _ := args.Get(0).([]domain.TeachingMaterial)
	return rows, args.Error(1)
}

func (m *mockMaterialRepo) ListByCourse(ctx context.Context, courseID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
	args := m.Called(ctx, courseID, includeUnpublished)
	rows, _ := args.Get(0).([]domain.TeachingMaterial)
	return rows, args.Error(1)
}

func (m *mockMaterialRepo) GetByID(ctx context.Context, id uint64) (*domain.TeachingMaterial, error) {
	args := m.Called(ctx, id)
	tm, _ := args.Get(0).(*domain.TeachingMaterial)
	return tm, args.Error(1)
}

func (m *mockMaterialRepo) CountByCourseForCompany(ctx context.Context, companyID uint64, includeUnpublished bool) (map[uint64]int, error) {
	args := m.Called(ctx, companyID, includeUnpublished)
	counts, _ := args.Get(0).(map[uint64]int)
	return counts, args.Error(1)
}

func (m *mockMaterialRepo) Create(ctx context.Context, tm *domain.TeachingMaterial) error {
	return m.Called(ctx, tm).Error(0)
}

func (m *mockMaterialRepo) Update(ctx context.Context, tm *domain.TeachingMaterial) error {
	return m.Called(ctx, tm).Error(0)
}

func (m *mockMaterialRepo) Delete(ctx context.Context, id uint64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockMaterialRepo) DeleteByCourse(ctx context.Context, courseID uint64) error {
	return m.Called(ctx, courseID).Error(0)
}

// --- mock: LessonProgressRepository ---

type mockProgressRepo struct{ mock.Mock }

var _ repository.LessonProgressRepository = (*mockProgressRepo)(nil)

func (m *mockProgressRepo) MarkCompleted(ctx context.Context, userID, materialID, courseID uint64) (bool, error) {
	args := m.Called(ctx, userID, materialID, courseID)
	return args.Bool(0), args.Error(1)
}

func (m *mockProgressRepo) MarkIncomplete(ctx context.Context, userID, materialID uint64) error {
	return m.Called(ctx, userID, materialID).Error(0)
}

func (m *mockProgressRepo) ListByUser(ctx context.Context, userID uint64) ([]domain.UserLessonProgress, error) {
	args := m.Called(ctx, userID)
	rows, _ := args.Get(0).([]domain.UserLessonProgress)
	return rows, args.Error(1)
}

func (m *mockProgressRepo) CountCompletedByUserGroupedByCourse(ctx context.Context, userID uint64) (map[uint64]int, error) {
	args := m.Called(ctx, userID)
	counts, _ := args.Get(0).(map[uint64]int)
	return counts, args.Error(1)
}

// --- mock: UserChapterViewRepository ---

type mockChapterViewRepo struct{ mock.Mock }

var _ repository.UserChapterViewRepository = (*mockChapterViewRepo)(nil)

func (m *mockChapterViewRepo) UpsertView(ctx context.Context, userID, materialID, courseID uint64) error {
	return m.Called(ctx, userID, materialID, courseID).Error(0)
}

func (m *mockChapterViewRepo) ListRecentByUser(ctx context.Context, userID uint64, limit int) ([]domain.UserChapterView, error) {
	args := m.Called(ctx, userID, limit)
	rows, _ := args.Get(0).([]domain.UserChapterView)
	return rows, args.Error(1)
}

func (m *mockChapterViewRepo) GetLastViewedByUserAndCourse(ctx context.Context, userID, courseID uint64) (*domain.UserChapterView, error) {
	args := m.Called(ctx, userID, courseID)
	v, _ := args.Get(0).(*domain.UserChapterView)
	return v, args.Error(1)
}

// --- mock: AuditRepository ---

type mockAuditRepo struct{ mock.Mock }

var _ repository.AuditRepository = (*mockAuditRepo)(nil)

func (m *mockAuditRepo) Record(ctx context.Context, e *domain.AuditEvent) error {
	return m.Called(ctx, e).Error(0)
}

func (m *mockAuditRepo) ListRecent(ctx context.Context, limit int) ([]domain.AuditEvent, error) {
	args := m.Called(ctx, limit)
	rows, _ := args.Get(0).([]domain.AuditEvent)
	return rows, args.Error(1)
}

// --- mock: CompanyRepository ---

type mockCompanyRepo struct{ mock.Mock }

var _ repository.CompanyRepository = (*mockCompanyRepo)(nil)

func (m *mockCompanyRepo) ListAll(ctx context.Context) ([]domain.Company, error) {
	args := m.Called(ctx)
	rows, _ := args.Get(0).([]domain.Company)
	return rows, args.Error(1)
}

func (m *mockCompanyRepo) FindByID(ctx context.Context, id uint64) (*domain.Company, error) {
	args := m.Called(ctx, id)
	c, _ := args.Get(0).(*domain.Company)
	return c, args.Error(1)
}

func (m *mockCompanyRepo) UpdateAiChatEnabled(ctx context.Context, companyID uint64, enabled bool) error {
	return m.Called(ctx, companyID, enabled).Error(0)
}

func (m *mockCompanyRepo) UpdateActive(ctx context.Context, companyID uint64, active bool) error {
	return m.Called(ctx, companyID, active).Error(0)
}

// --- mock: CompanyApplicationRepository ---

type mockCompanyAppRepo struct{ mock.Mock }

var _ repository.CompanyApplicationRepository = (*mockCompanyAppRepo)(nil)

func (m *mockCompanyAppRepo) Create(ctx context.Context, app *domain.CompanyApplication) error {
	return m.Called(ctx, app).Error(0)
}

func (m *mockCompanyAppRepo) ListAll(ctx context.Context) ([]domain.CompanyApplication, error) {
	args := m.Called(ctx)
	rows, _ := args.Get(0).([]domain.CompanyApplication)
	return rows, args.Error(1)
}

func (m *mockCompanyAppRepo) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return m.Called(ctx, id, status).Error(0)
}

// --- mock: CompanyLearningActivitySummarizer ---

type mockLearningSummarizer struct{ mock.Mock }

var _ repository.CompanyLearningActivitySummarizer = (*mockLearningSummarizer)(nil)

func (m *mockLearningSummarizer) ListMemberActivities(ctx context.Context, companyID uint64, fromDate time.Time) ([]repository.MemberLearningActivity, error) {
	args := m.Called(ctx, companyID, fromDate)
	rows, _ := args.Get(0).([]repository.MemberLearningActivity)
	return rows, args.Error(1)
}

// --- mock: CompanyMemberCounter ---

type mockMemberCounter struct{ mock.Mock }

var _ repository.CompanyMemberCounter = (*mockMemberCounter)(nil)

func (m *mockMemberCounter) CountMembersByCompany(ctx context.Context) ([]repository.CompanyMemberCount, error) {
	args := m.Called(ctx)
	rows, _ := args.Get(0).([]repository.CompanyMemberCount)
	return rows, args.Error(1)
}

// --- mock: UserDailyActivityRepository ---

type mockDailyActivityRepo struct{ mock.Mock }

var _ repository.UserDailyActivityRepository = (*mockDailyActivityRepo)(nil)

func (m *mockDailyActivityRepo) Increment(ctx context.Context, userID uint64, date time.Time, inc repository.UserDailyActivityIncrement) error {
	return m.Called(ctx, userID, date, inc).Error(0)
}

func (m *mockDailyActivityRepo) ListByUser(ctx context.Context, userID uint64, from, to time.Time) ([]domain.UserDailyActivity, error) {
	args := m.Called(ctx, userID, from, to)
	rows, _ := args.Get(0).([]domain.UserDailyActivity)
	return rows, args.Error(1)
}
