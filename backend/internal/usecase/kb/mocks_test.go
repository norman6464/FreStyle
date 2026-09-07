package kb_test

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

func (m *mockUserRepo) ListByWorkspaceID(ctx context.Context, workspaceID string) ([]domain.User, error) {
	args := m.Called(ctx, workspaceID)
	rows, _ := args.Get(0).([]domain.User)
	return rows, args.Error(1)
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	return m.Called(ctx, u).Error(0)
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

func (m *mockUserRepo) UpdateActive(ctx context.Context, userID uint64, active bool) error {
	return m.Called(ctx, userID, active).Error(0)
}

func (m *mockUserRepo) SoftDelete(ctx context.Context, userID uint64) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockUserRepo) UpdateName(ctx context.Context, userID uint64, name string) error {
	return m.Called(ctx, userID, name).Error(0)
}

func (m *mockUserRepo) UpdateWorkspaceID(ctx context.Context, userID uint64, workspaceID *string) error {
	return m.Called(ctx, userID, workspaceID).Error(0)
}

// --- mock: KnowledgeBaseRepository ---

type mockKnowledgeBaseRepo struct{ mock.Mock }

var _ repository.KnowledgeBaseRepository = (*mockKnowledgeBaseRepo)(nil)

func (m *mockKnowledgeBaseRepo) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

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

func (m *mockKnowledgeBaseRepo) FindPersonalWorkspaceByOwner(ctx context.Context, userID uint64) (*domain.Workspace, error) {
	args := m.Called(ctx, userID)
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

func (m *mockKnowledgeBaseRepo) UpdateSpaceName(ctx context.Context, workspaceID, spaceID, name string) error {
	return m.Called(ctx, workspaceID, spaceID, name).Error(0)
}

func (m *mockKnowledgeBaseRepo) DeletePageSubtree(ctx context.Context, workspaceID, pageID string) error {
	return m.Called(ctx, workspaceID, pageID).Error(0)
}

func (m *mockKnowledgeBaseRepo) ListAncestorPageIDs(ctx context.Context, workspaceID, pageID string) ([]string, error) {
	args := m.Called(ctx, workspaceID, pageID)
	ids, _ := args.Get(0).([]string)
	return ids, args.Error(1)
}

func (m *mockKnowledgeBaseRepo) FindPageByIDAcrossWorkspaces(ctx context.Context, pageID string) (*domain.Page, error) {
	args := m.Called(ctx, pageID)
	p, _ := args.Get(0).(*domain.Page)
	return p, args.Error(1)
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

func (m *mockKnowledgeBaseRepo) SiblingPositionsAround(
	ctx context.Context, workspaceID, spaceID string, parentID *string, anchorPageID, movingPageID string,
) (bool, string, string, string, error) {
	args := m.Called(ctx, workspaceID, spaceID, parentID, anchorPageID, movingPageID)
	return args.Bool(0), args.String(1), args.String(2), args.String(3), args.Error(4)
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

func (m *mockKnowledgeBaseRepo) UpdatePageIcon(ctx context.Context, workspaceID, pageID string, icon *domain.PageIcon) (*domain.Page, error) {
	args := m.Called(ctx, workspaceID, pageID, icon)
	p, _ := args.Get(0).(*domain.Page)
	return p, args.Error(1)
}

func (m *mockKnowledgeBaseRepo) UpdatePageCover(ctx context.Context, workspaceID, pageID string, cover *domain.PageCover) (*domain.Page, error) {
	args := m.Called(ctx, workspaceID, pageID, cover)
	p, _ := args.Get(0).(*domain.Page)
	return p, args.Error(1)
}

func (m *mockKnowledgeBaseRepo) PageReferencesImageKey(ctx context.Context, workspaceID, pageID, key string) (bool, error) {
	args := m.Called(ctx, workspaceID, pageID, key)
	return args.Bool(0), args.Error(1)
}

func (m *mockKnowledgeBaseRepo) TouchPageLastEditedBy(ctx context.Context, workspaceID, pageID string, userID uint64) error {
	return m.Called(ctx, workspaceID, pageID, userID).Error(0)
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

func (m *mockKBPermissionRepo) ListMemberWorkspaces(ctx context.Context, userID uint64) ([]domain.MemberWorkspace, error) {
	args := m.Called(ctx, userID)
	rows, _ := args.Get(0).([]domain.MemberWorkspace)
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

func (m *mockKBPermissionRepo) ListGrantablePrincipals(ctx context.Context, workspaceID string) ([]domain.GrantablePrincipal, error) {
	args := m.Called(ctx, workspaceID)
	p, _ := args.Get(0).([]domain.GrantablePrincipal)
	return p, args.Error(1)
}

func (m *mockKBPermissionRepo) UpsertPageGrant(ctx context.Context, workspaceID, pageID, principalID string, role domain.GrantRole) (*domain.PageGrant, error) {
	args := m.Called(ctx, workspaceID, pageID, principalID, role)
	g, _ := args.Get(0).(*domain.PageGrant)
	return g, args.Error(1)
}

func (m *mockKBPermissionRepo) DeletePageGrant(ctx context.Context, workspaceID, pageID, principalID string) error {
	args := m.Called(ctx, workspaceID, pageID, principalID)
	return args.Error(0)
}

func (m *mockKBPermissionRepo) ListPageGrants(ctx context.Context, workspaceID, pageID string) ([]domain.PageGrant, error) {
	args := m.Called(ctx, workspaceID, pageID)
	g, _ := args.Get(0).([]domain.PageGrant)
	return g, args.Error(1)
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

func (m *mockKBPermissionRepo) SearchWorkspacePageViewFacts(ctx context.Context, workspaceID string, userID uint64, query string) ([]repository.PageWithViewFacts, error) {
	args := m.Called(ctx, workspaceID, userID, query)
	rows, _ := args.Get(0).([]repository.PageWithViewFacts)
	return rows, args.Error(1)
}

func (m *mockKBPermissionRepo) ListWorkspacePageViewFactsByIDs(ctx context.Context, workspaceID string, userID uint64, pageIDs []string) ([]repository.PageWithViewFacts, error) {
	args := m.Called(ctx, workspaceID, userID, pageIDs)
	rows, _ := args.Get(0).([]repository.PageWithViewFacts)
	return rows, args.Error(1)
}

func (m *mockKBPermissionRepo) GrantWorkspaceRoleIfAbsent(ctx context.Context, workspaceID, principalID string, role domain.GrantRole) error {
	return m.Called(ctx, workspaceID, principalID, role).Error(0)
}

func (m *mockKBPermissionRepo) ListSpacePageViewFacts(ctx context.Context, workspaceID, spaceID string, userID uint64, archived bool) ([]repository.PageWithViewFacts, error) {
	args := m.Called(ctx, workspaceID, spaceID, userID, archived)
	rows, _ := args.Get(0).([]repository.PageWithViewFacts)
	return rows, args.Error(1)
}

func (m *mockKBPermissionRepo) ListWorkspaceSpaceScopeFacts(ctx context.Context, workspaceID string, userID uint64) ([]repository.SpaceWithScopeFacts, error) {
	args := m.Called(ctx, workspaceID, userID)
	rows, _ := args.Get(0).([]repository.SpaceWithScopeFacts)
	return rows, args.Error(1)
}

func (m *mockKBPermissionRepo) ListSubtreePagePermissionFacts(ctx context.Context, workspaceID, pageID string, userID uint64) ([]repository.PageWithPermissionFacts, error) {
	args := m.Called(ctx, workspaceID, pageID, userID)
	rows, _ := args.Get(0).([]repository.PageWithPermissionFacts)
	return rows, args.Error(1)
}

type mockShareLinkRepo struct{ mock.Mock }

var _ repository.ShareLinkRepository = (*mockShareLinkRepo)(nil)

func (m *mockShareLinkRepo) Create(ctx context.Context, in repository.ShareLinkWrite) (*domain.ShareLink, error) {
	args := m.Called(ctx, in)
	l, _ := args.Get(0).(*domain.ShareLink)
	return l, args.Error(1)
}

func (m *mockShareLinkRepo) Revoke(ctx context.Context, workspaceID, shareLinkID string) error {
	return m.Called(ctx, workspaceID, shareLinkID).Error(0)
}

func (m *mockShareLinkRepo) FindByTokenHash(ctx context.Context, tokenHash []byte) (*domain.ShareLink, error) {
	args := m.Called(ctx, tokenHash)
	l, _ := args.Get(0).(*domain.ShareLink)
	return l, args.Error(1)
}

func (m *mockShareLinkRepo) ListByPage(ctx context.Context, workspaceID, pageID string) ([]domain.ShareLink, error) {
	args := m.Called(ctx, workspaceID, pageID)
	rows, _ := args.Get(0).([]domain.ShareLink)
	return rows, args.Error(1)
}

// --- fake: TxManager ---

// txMarkerKey は fakeTxManager が DoInTx の中で ctx に埋め込む印。
// 本物の *sql.Tx を持たないテストで「repository がトランザクションの中で呼ばれたか」を
// mock.MatchedBy(inTx) で確かめられるようにするためだけの値。
type txMarkerKeyType struct{}

var txMarkerKey = txMarkerKeyType{}

// inTx は mock.MatchedBy に渡す述語。ctx に txMarkerKey が乗っていれば
// fakeTxManager.DoInTx の fn の中（＝トランザクションの中）から呼ばれたことを意味する。
func inTx(ctx context.Context) bool {
	v, _ := ctx.Value(txMarkerKey).(bool)
	return v
}

// fakeTxManager は repository.TxManager のテスト用実装。実 DB もトランザクションも
// 介さず fn(ctx) をそのまま呼ぶが、呼び出し回数と「tx の中で呼んだ」印だけは残す。
type fakeTxManager struct {
	calls int
}

var _ repository.TxManager = (*fakeTxManager)(nil)

func (f *fakeTxManager) DoInTx(ctx context.Context, fn func(context.Context) error) error {
	f.calls++
	return fn(context.WithValue(ctx, txMarkerKey, true))
}

// --- mock: KbImagePresigner ---

type mockKbImagePresigner struct{ mock.Mock }

var _ repository.KbImagePresigner = (*mockKbImagePresigner)(nil)

func (m *mockKbImagePresigner) PresignUpload(ctx context.Context, key, contentType string, size int64) (string, int, error) {
	args := m.Called(ctx, key, contentType, size)
	return args.String(0), args.Int(1), args.Error(2)
}

func (m *mockKbImagePresigner) PresignDownload(ctx context.Context, key string) (string, int, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Int(1), args.Error(2)
}

// --- mock: PageVersionRepository ---

type mockPageVersionRepo struct{ mock.Mock }

var _ repository.PageVersionRepository = (*mockPageVersionRepo)(nil)

func (m *mockPageVersionRepo) LockPage(ctx context.Context, workspaceID, pageID string) error {
	args := m.Called(ctx, workspaceID, pageID)
	return args.Error(0)
}

func (m *mockPageVersionRepo) CreateVersionIfDue(
	ctx context.Context, workspaceID, pageID, doc string, authorUserID uint64, note *string, force bool,
) (bool, *domain.PageVersion, error) {
	args := m.Called(ctx, workspaceID, pageID, doc, authorUserID, note, force)
	var v *domain.PageVersion
	if args.Get(1) != nil {
		v = args.Get(1).(*domain.PageVersion)
	}
	return args.Bool(0), v, args.Error(2)
}

func (m *mockPageVersionRepo) ListVersions(ctx context.Context, workspaceID, pageID string) ([]domain.PageVersion, error) {
	args := m.Called(ctx, workspaceID, pageID)
	var v []domain.PageVersion
	if args.Get(0) != nil {
		v = args.Get(0).([]domain.PageVersion)
	}
	return v, args.Error(1)
}

func (m *mockPageVersionRepo) GetVersion(ctx context.Context, workspaceID, pageID string, seq int64) (*domain.PageVersion, error) {
	args := m.Called(ctx, workspaceID, pageID, seq)
	var v *domain.PageVersion
	if args.Get(0) != nil {
		v = args.Get(0).(*domain.PageVersion)
	}
	return v, args.Error(1)
}
