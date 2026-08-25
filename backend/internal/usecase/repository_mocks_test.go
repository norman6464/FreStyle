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

func (m *mockUserRepo) CreateFirstSuperAdminWithOidcIdentity(
	ctx context.Context, u *domain.User, provider, subject string,
) (bool, error) {
	args := m.Called(ctx, u, provider, subject)
	return args.Bool(0), args.Error(1)
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

func (m *mockMaterialRepo) UpdateDocWithRevision(ctx context.Context, id uint64, doc string, expectedRevision int) (*domain.TeachingMaterial, error) {
	args := m.Called(ctx, id, doc, expectedRevision)
	tm, _ := args.Get(0).(*domain.TeachingMaterial)
	return tm, args.Error(1)
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

// --- mock: RichDocumentRepository ---

type mockRichDocRepo struct{ mock.Mock }

var _ repository.RichDocumentRepository = (*mockRichDocRepo)(nil)

func (m *mockRichDocRepo) Create(ctx context.Context, doc *domain.RichDocument) error {
	return m.Called(ctx, doc).Error(0)
}

func (m *mockRichDocRepo) FindByID(ctx context.Context, id string) (*domain.RichDocument, error) {
	args := m.Called(ctx, id)
	d, _ := args.Get(0).(*domain.RichDocument)
	return d, args.Error(1)
}

func (m *mockRichDocRepo) UpdateWithRevision(ctx context.Context, doc *domain.RichDocument, expectedRevision int) error {
	return m.Called(ctx, doc, expectedRevision).Error(0)
}

func (m *mockRichDocRepo) SoftDelete(ctx context.Context, id string, ownerID uint64) error {
	return m.Called(ctx, id, ownerID).Error(0)
}

func (m *mockRichDocRepo) ListByOwner(ctx context.Context, ownerID uint64, kind domain.DocumentKind) ([]domain.RichDocument, error) {
	args := m.Called(ctx, ownerID, kind)
	rows, _ := args.Get(0).([]domain.RichDocument)
	return rows, args.Error(1)
}

// --- mock: KnowledgeBaseRepository ---

type mockKnowledgeBaseRepo struct{ mock.Mock }

var _ repository.KnowledgeBaseRepository = (*mockKnowledgeBaseRepo)(nil)

func (m *mockKnowledgeBaseRepo) FindWorkspaceByID(ctx context.Context, workspaceID string) (*domain.Workspace, error) {
	args := m.Called(ctx, workspaceID)
	w, _ := args.Get(0).(*domain.Workspace)
	return w, args.Error(1)
}

func (m *mockKnowledgeBaseRepo) FindWorkspaceBySlug(ctx context.Context, slug string) (*domain.Workspace, error) {
	args := m.Called(ctx, slug)
	w, _ := args.Get(0).(*domain.Workspace)
	return w, args.Error(1)
}

func (m *mockKnowledgeBaseRepo) FindSpace(ctx context.Context, workspaceID, spaceID string) (*domain.Space, error) {
	args := m.Called(ctx, workspaceID, spaceID)
	s, _ := args.Get(0).(*domain.Space)
	return s, args.Error(1)
}

func (m *mockKnowledgeBaseRepo) CreateSpace(ctx context.Context, space *domain.Space) error {
	return m.Called(ctx, space).Error(0)
}

func (m *mockKnowledgeBaseRepo) FindPage(ctx context.Context, workspaceID, pageID string) (*domain.Page, error) {
	args := m.Called(ctx, workspaceID, pageID)
	p, _ := args.Get(0).(*domain.Page)
	return p, args.Error(1)
}

func (m *mockKnowledgeBaseRepo) ListActivePagesBySpace(ctx context.Context, workspaceID, spaceID string) ([]domain.Page, error) {
	args := m.Called(ctx, workspaceID, spaceID)
	rows, _ := args.Get(0).([]domain.Page)
	return rows, args.Error(1)
}

func (m *mockKnowledgeBaseRepo) LastActiveSiblingPosition(ctx context.Context, workspaceID, spaceID string, parentID *string) (string, error) {
	args := m.Called(ctx, workspaceID, spaceID, parentID)
	return args.String(0), args.Error(1)
}

func (m *mockKnowledgeBaseRepo) HasActiveSiblingPosition(ctx context.Context, workspaceID, spaceID string, parentID *string, position, excludePageID string) (bool, error) {
	args := m.Called(ctx, workspaceID, spaceID, parentID, position, excludePageID)
	return args.Bool(0), args.Error(1)
}

func (m *mockKnowledgeBaseRepo) HasDescendant(ctx context.Context, workspaceID, pageID, candidateID string) (bool, error) {
	args := m.Called(ctx, workspaceID, pageID, candidateID)
	return args.Bool(0), args.Error(1)
}

func (m *mockKnowledgeBaseRepo) CreatePage(ctx context.Context, page *domain.Page) error {
	return m.Called(ctx, page).Error(0)
}

func (m *mockKnowledgeBaseRepo) UpdatePageTitle(ctx context.Context, workspaceID, pageID, title string) (*domain.Page, error) {
	args := m.Called(ctx, workspaceID, pageID, title)
	p, _ := args.Get(0).(*domain.Page)
	return p, args.Error(1)
}

func (m *mockKnowledgeBaseRepo) MovePage(ctx context.Context, workspaceID, pageID string, newParentID *string, newSpaceID, newPosition string) error {
	return m.Called(ctx, workspaceID, pageID, newParentID, newSpaceID, newPosition).Error(0)
}

func (m *mockKnowledgeBaseRepo) ArchivePageSubtree(ctx context.Context, workspaceID, pageID string) error {
	return m.Called(ctx, workspaceID, pageID).Error(0)
}

func (m *mockKnowledgeBaseRepo) UnarchivePageSubtree(ctx context.Context, workspaceID, pageID string, archivedSince time.Time, newRootPosition *string) error {
	return m.Called(ctx, workspaceID, pageID, archivedSince, newRootPosition).Error(0)
}

func (m *mockKnowledgeBaseRepo) ListBlocksByPage(ctx context.Context, workspaceID, pageID string) ([]domain.Block, error) {
	args := m.Called(ctx, workspaceID, pageID)
	rows, _ := args.Get(0).([]domain.Block)
	return rows, args.Error(1)
}

func (m *mockKnowledgeBaseRepo) ReplacePageBlocks(ctx context.Context, workspaceID, pageID string, blocks []repository.BlockWrite, snapshotDoc string) error {
	return m.Called(ctx, workspaceID, pageID, blocks, snapshotDoc).Error(0)
}

func (m *mockKnowledgeBaseRepo) GetPageSnapshot(ctx context.Context, workspaceID, pageID string) (*domain.PageSnapshot, error) {
	args := m.Called(ctx, workspaceID, pageID)
	s, _ := args.Get(0).(*domain.PageSnapshot)
	return s, args.Error(1)
}

// --- mock: KnowledgeBasePermissionRepository ---

type mockKBPermissionRepo struct{ mock.Mock }

var _ repository.KnowledgeBasePermissionRepository = (*mockKBPermissionRepo)(nil)

func (m *mockKBPermissionRepo) EnsureUserPrincipal(ctx context.Context, workspaceID string, userID uint64) (*domain.Principal, error) {
	args := m.Called(ctx, workspaceID, userID)
	p, _ := args.Get(0).(*domain.Principal)
	return p, args.Error(1)
}

func (m *mockKBPermissionRepo) EnsureSpaceEveryonePrincipal(ctx context.Context, workspaceID, spaceID string) (*domain.Principal, error) {
	args := m.Called(ctx, workspaceID, spaceID)
	p, _ := args.Get(0).(*domain.Principal)
	return p, args.Error(1)
}

func (m *mockKBPermissionRepo) ListMemberWorkspaces(ctx context.Context, userID uint64) ([]domain.Workspace, error) {
	args := m.Called(ctx, userID)
	rows, _ := args.Get(0).([]domain.Workspace)
	return rows, args.Error(1)
}

func (m *mockKBPermissionRepo) SpacePermissionFactsForUser(ctx context.Context, workspaceID, spaceID string, userID uint64) (*domain.ScopeFacts, error) {
	args := m.Called(ctx, workspaceID, spaceID, userID)
	f, _ := args.Get(0).(*domain.ScopeFacts)
	return f, args.Error(1)
}

func (m *mockKBPermissionRepo) WorkspacePermissionFactsForUser(ctx context.Context, workspaceID string, userID uint64) (*domain.ScopeFacts, error) {
	args := m.Called(ctx, workspaceID, userID)
	f, _ := args.Get(0).(*domain.ScopeFacts)
	return f, args.Error(1)
}

func (m *mockKBPermissionRepo) CreateGroupPrincipal(ctx context.Context, workspaceID, name string) (*domain.Principal, error) {
	args := m.Called(ctx, workspaceID, name)
	p, _ := args.Get(0).(*domain.Principal)
	return p, args.Error(1)
}

func (m *mockKBPermissionRepo) FindPrincipal(ctx context.Context, workspaceID, principalID string) (*domain.Principal, error) {
	args := m.Called(ctx, workspaceID, principalID)
	p, _ := args.Get(0).(*domain.Principal)
	return p, args.Error(1)
}

func (m *mockKBPermissionRepo) FindUserPrincipal(ctx context.Context, workspaceID string, userID uint64) (*domain.Principal, error) {
	args := m.Called(ctx, workspaceID, userID)
	p, _ := args.Get(0).(*domain.Principal)
	return p, args.Error(1)
}

func (m *mockKBPermissionRepo) DeletePrincipal(ctx context.Context, workspaceID, principalID string) error {
	return m.Called(ctx, workspaceID, principalID).Error(0)
}

func (m *mockKBPermissionRepo) IsWorkspaceMember(ctx context.Context, workspaceID string, userID uint64) (bool, error) {
	args := m.Called(ctx, workspaceID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *mockKBPermissionRepo) AddGroupMember(ctx context.Context, workspaceID, groupPrincipalID, memberPrincipalID string) error {
	return m.Called(ctx, workspaceID, groupPrincipalID, memberPrincipalID).Error(0)
}

func (m *mockKBPermissionRepo) RemoveGroupMember(ctx context.Context, workspaceID, groupPrincipalID, memberPrincipalID string) error {
	return m.Called(ctx, workspaceID, groupPrincipalID, memberPrincipalID).Error(0)
}

func (m *mockKBPermissionRepo) UpsertWorkspaceGrant(ctx context.Context, workspaceID, principalID string, role domain.GrantRole) (*domain.WorkspaceGrant, error) {
	args := m.Called(ctx, workspaceID, principalID, role)
	g, _ := args.Get(0).(*domain.WorkspaceGrant)
	return g, args.Error(1)
}

func (m *mockKBPermissionRepo) DeleteWorkspaceGrant(ctx context.Context, workspaceID, principalID string) error {
	return m.Called(ctx, workspaceID, principalID).Error(0)
}

func (m *mockKBPermissionRepo) ListWorkspaceGrants(ctx context.Context, workspaceID string) ([]domain.WorkspaceGrant, error) {
	args := m.Called(ctx, workspaceID)
	rows, _ := args.Get(0).([]domain.WorkspaceGrant)
	return rows, args.Error(1)
}

func (m *mockKBPermissionRepo) UpsertSpaceGrant(ctx context.Context, workspaceID, spaceID, principalID string, role domain.GrantRole) (*domain.SpaceGrant, error) {
	args := m.Called(ctx, workspaceID, spaceID, principalID, role)
	g, _ := args.Get(0).(*domain.SpaceGrant)
	return g, args.Error(1)
}

func (m *mockKBPermissionRepo) DeleteSpaceGrant(ctx context.Context, workspaceID, spaceID, principalID string) error {
	return m.Called(ctx, workspaceID, spaceID, principalID).Error(0)
}

func (m *mockKBPermissionRepo) ListSpaceGrants(ctx context.Context, workspaceID, spaceID string) ([]domain.SpaceGrant, error) {
	args := m.Called(ctx, workspaceID, spaceID)
	rows, _ := args.Get(0).([]domain.SpaceGrant)
	return rows, args.Error(1)
}

func (m *mockKBPermissionRepo) UpsertPageRestriction(ctx context.Context, workspaceID, pageID, principalID string, capability domain.Capability, mode domain.RestrictionMode) (*domain.PageRestriction, error) {
	args := m.Called(ctx, workspaceID, pageID, principalID, capability, mode)
	r, _ := args.Get(0).(*domain.PageRestriction)
	return r, args.Error(1)
}

func (m *mockKBPermissionRepo) DeletePageRestriction(ctx context.Context, workspaceID, pageID, principalID string, capability domain.Capability) error {
	return m.Called(ctx, workspaceID, pageID, principalID, capability).Error(0)
}

func (m *mockKBPermissionRepo) ListPageRestrictions(ctx context.Context, workspaceID, pageID string) ([]domain.PageRestriction, error) {
	args := m.Called(ctx, workspaceID, pageID)
	rows, _ := args.Get(0).([]domain.PageRestriction)
	return rows, args.Error(1)
}

func (m *mockKBPermissionRepo) ListPageAllowListCapabilities(ctx context.Context, workspaceID, pageID string) ([]domain.Capability, error) {
	args := m.Called(ctx, workspaceID, pageID)
	rows, _ := args.Get(0).([]domain.Capability)
	return rows, args.Error(1)
}

func (m *mockKBPermissionRepo) CreateShareLink(ctx context.Context, in repository.ShareLinkWrite) (*domain.ShareLink, error) {
	args := m.Called(ctx, in)
	l, _ := args.Get(0).(*domain.ShareLink)
	return l, args.Error(1)
}

func (m *mockKBPermissionRepo) RevokeShareLink(ctx context.Context, workspaceID, shareLinkID string) error {
	return m.Called(ctx, workspaceID, shareLinkID).Error(0)
}

func (m *mockKBPermissionRepo) FindShareLinkByTokenHash(ctx context.Context, tokenHash []byte) (*domain.ShareLink, error) {
	args := m.Called(ctx, tokenHash)
	l, _ := args.Get(0).(*domain.ShareLink)
	return l, args.Error(1)
}

func (m *mockKBPermissionRepo) ListPageShareLinks(ctx context.Context, workspaceID, pageID string) ([]domain.ShareLink, error) {
	args := m.Called(ctx, workspaceID, pageID)
	rows, _ := args.Get(0).([]domain.ShareLink)
	return rows, args.Error(1)
}

func (m *mockKBPermissionRepo) PagePermissionFactsForUser(ctx context.Context, workspaceID, pageID string, userID uint64) (*domain.PagePermissionFacts, error) {
	args := m.Called(ctx, workspaceID, pageID, userID)
	f, _ := args.Get(0).(*domain.PagePermissionFacts)
	return f, args.Error(1)
}

func (m *mockKBPermissionRepo) PagePermissionFactsForPrincipal(ctx context.Context, workspaceID, pageID, principalID string) (*domain.PagePermissionFacts, error) {
	args := m.Called(ctx, workspaceID, pageID, principalID)
	f, _ := args.Get(0).(*domain.PagePermissionFacts)
	return f, args.Error(1)
}

func (m *mockKBPermissionRepo) ListSpacePageViewFacts(ctx context.Context, workspaceID, spaceID string, userID uint64) ([]repository.PageWithViewFacts, error) {
	args := m.Called(ctx, workspaceID, spaceID, userID)
	rows, _ := args.Get(0).([]repository.PageWithViewFacts)
	return rows, args.Error(1)
}

func (m *mockKBPermissionRepo) ListSubtreePagePermissionFacts(ctx context.Context, workspaceID, pageID string, userID uint64) ([]repository.PageWithPermissionFacts, error) {
	args := m.Called(ctx, workspaceID, pageID, userID)
	rows, _ := args.Get(0).([]repository.PageWithPermissionFacts)
	return rows, args.Error(1)
}
