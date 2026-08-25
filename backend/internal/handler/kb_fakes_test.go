package handler

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ナレッジ基盤 handler テスト用の fake repository。
//
// 権限は「事実」まで組み立てて domain.ResolvePagePermission に通す（判定結果を直接返さない）。
// fake が判定を肩代わりすると、本番と違う規則でテストが緑になってしまうため。

// kbFakePages は repository.KnowledgeBaseRepository の in-memory fake。
type kbFakePages struct {
	workspaces map[string]*domain.Workspace // slug -> workspace
	spaces     map[string]*domain.Space     // spaceID -> space
	pages      map[string]*domain.Page      // pageID -> page
	snapshots  map[string]string            // pageID -> doc
	nextID     int
	// failWith は次の書き込み系呼び出しを失敗させる（500 経路の確認用）。
	failWith error
}

var _ repository.KnowledgeBaseRepository = (*kbFakePages)(nil)

func newKbFakePages() *kbFakePages {
	return &kbFakePages{
		workspaces: map[string]*domain.Workspace{},
		spaces:     map[string]*domain.Space{},
		pages:      map[string]*domain.Page{},
		snapshots:  map[string]string{},
	}
}

func (f *kbFakePages) addWorkspace(id, slug string) {
	f.workspaces[slug] = &domain.Workspace{ID: id, Slug: slug, Name: slug}
}

func (f *kbFakePages) addSpace(workspaceID, spaceID string) {
	f.spaces[spaceID] = &domain.Space{ID: spaceID, WorkspaceID: workspaceID, Key: spaceID, Name: spaceID}
}

func (f *kbFakePages) addPage(p domain.Page) *domain.Page {
	if p.Position == "" {
		p.Position = "a0"
	}
	stored := p
	f.pages[p.ID] = &stored
	return &stored
}

func copyPage(p *domain.Page) *domain.Page {
	c := *p
	return &c
}

func (f *kbFakePages) FindWorkspaceByID(_ context.Context, workspaceID string) (*domain.Workspace, error) {
	for _, ws := range f.workspaces {
		if ws.ID == workspaceID {
			c := *ws
			return &c, nil
		}
	}
	return nil, repository.ErrWorkspaceNotFound
}

func (f *kbFakePages) FindWorkspaceBySlug(_ context.Context, slug string) (*domain.Workspace, error) {
	ws, ok := f.workspaces[slug]
	if !ok {
		return nil, repository.ErrWorkspaceNotFound
	}
	c := *ws
	return &c, nil
}

func (f *kbFakePages) FindSpace(_ context.Context, workspaceID, spaceID string) (*domain.Space, error) {
	s, ok := f.spaces[spaceID]
	if !ok || s.WorkspaceID != workspaceID {
		return nil, repository.ErrSpaceNotFound
	}
	c := *s
	return &c, nil
}

func (f *kbFakePages) FindPage(_ context.Context, workspaceID, pageID string) (*domain.Page, error) {
	p, ok := f.pages[pageID]
	if !ok || p.WorkspaceID != workspaceID {
		return nil, repository.ErrPageNotFound
	}
	return copyPage(p), nil
}

func (f *kbFakePages) ListActivePagesBySpace(_ context.Context, workspaceID, spaceID string) ([]domain.Page, error) {
	return f.activePages(workspaceID, spaceID), nil
}

// activePages は position 順に並んだ現役ページを返す（本番の ORDER BY "position" と同じ並び）。
func (f *kbFakePages) activePages(workspaceID, spaceID string) []domain.Page {
	out := make([]domain.Page, 0, len(f.pages))
	for _, p := range f.pages {
		if p.WorkspaceID == workspaceID && p.SpaceID == spaceID && p.ArchivedAt == nil {
			out = append(out, *p)
		}
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].Position < out[j-1].Position; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func (f *kbFakePages) LastActiveSiblingPosition(_ context.Context, _, _ string, _ *string) (string, error) {
	return "", nil
}

func (f *kbFakePages) HasActiveSiblingPosition(_ context.Context, _, _ string, _ *string, _, _ string) (bool, error) {
	return false, nil
}

func (f *kbFakePages) HasDescendant(_ context.Context, workspaceID, pageID, candidateID string) (bool, error) {
	cur, ok := f.pages[candidateID]
	for ok && cur.WorkspaceID == workspaceID {
		if cur.ID == pageID {
			return true, nil
		}
		if cur.ParentID == nil {
			break
		}
		cur, ok = f.pages[*cur.ParentID]
	}
	return false, nil
}

func (f *kbFakePages) CreatePage(_ context.Context, page *domain.Page) error {
	f.nextID++
	page.ID = "created-page-" + string(rune('a'+f.nextID-1))
	page.CreatedAt = time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	page.UpdatedAt = page.CreatedAt
	stored := *page
	f.pages[page.ID] = &stored
	return nil
}

func (f *kbFakePages) UpdatePageTitle(_ context.Context, workspaceID, pageID, title string) (*domain.Page, error) {
	if f.failWith != nil {
		return nil, f.failWith
	}
	p, ok := f.pages[pageID]
	if !ok || p.WorkspaceID != workspaceID {
		return nil, repository.ErrPageNotFound
	}
	p.Title = title
	return copyPage(p), nil
}

func (f *kbFakePages) MovePage(_ context.Context, workspaceID, pageID string, newParentID *string, newSpaceID, newPosition string) error {
	p, ok := f.pages[pageID]
	if !ok || p.WorkspaceID != workspaceID {
		return repository.ErrPageNotFound
	}
	p.ParentID = newParentID
	p.SpaceID = newSpaceID
	p.Position = newPosition
	return nil
}

func (f *kbFakePages) ArchivePageSubtree(ctx context.Context, workspaceID, pageID string) error {
	if f.failWith != nil {
		return f.failWith
	}
	at := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	for id, p := range f.pages {
		if p.WorkspaceID != workspaceID || p.ArchivedAt != nil {
			continue
		}
		desc, _ := f.HasDescendant(ctx, workspaceID, pageID, id)
		if desc {
			t := at
			p.ArchivedAt = &t
		}
	}
	return nil
}

func (f *kbFakePages) UnarchivePageSubtree(ctx context.Context, workspaceID, pageID string, archivedSince time.Time, newRootPosition *string) error {
	for id, p := range f.pages {
		if p.WorkspaceID != workspaceID || p.ArchivedAt == nil || !p.ArchivedAt.Equal(archivedSince) {
			continue
		}
		desc, _ := f.HasDescendant(ctx, workspaceID, pageID, id)
		if !desc {
			continue
		}
		p.ArchivedAt = nil
		if id == pageID && newRootPosition != nil {
			p.Position = *newRootPosition
		}
	}
	return nil
}

func (f *kbFakePages) ListBlocksByPage(_ context.Context, _, _ string) ([]domain.Block, error) {
	return []domain.Block{}, nil
}

func (f *kbFakePages) ReplacePageBlocks(_ context.Context, workspaceID, pageID string, _ []repository.BlockWrite, snapshotDoc string) error {
	p, ok := f.pages[pageID]
	if !ok || p.WorkspaceID != workspaceID {
		return repository.ErrPageNotFound
	}
	f.snapshots[pageID] = snapshotDoc
	return nil
}

func (f *kbFakePages) GetPageSnapshot(_ context.Context, workspaceID, pageID string) (*domain.PageSnapshot, error) {
	p, ok := f.pages[pageID]
	if !ok || p.WorkspaceID != workspaceID {
		return nil, repository.ErrPageSnapshotNotFound
	}
	doc, ok := f.snapshots[pageID]
	if !ok {
		return nil, repository.ErrPageSnapshotNotFound
	}
	return &domain.PageSnapshot{
		PageID:  pageID,
		Doc:     doc,
		BuiltAt: time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC),
	}, nil
}

// kbPermKey は「どのページを誰が」の組。
type kbPermKey struct {
	pageID string
	userID uint64
}

// kbFakePerms は repository.KnowledgeBasePermissionRepository の in-memory fake。
// 判定そのものはせず、domain.ResolvePagePermission がその答えを出すような「事実」を返す。
type kbFakePerms struct {
	pages *kbFakePages
	// members は workspaceID -> userID -> 所属。
	members map[string]map[uint64]bool
	// perPage は特定のページだけ別の実効権限にしたいときの上書き。
	perPage map[kbPermKey]domain.PagePermission
	// fallback は perPage に無いページの実効権限。
	fallback domain.PagePermission
	// membersErr は所属判定を失敗させる（middleware の 500 経路の確認用）。
	membersErr error
	// listFactsErr は一覧の事実収集を失敗させる（ツリー取得の 500 経路の確認用）。
	listFactsErr error
}

var _ repository.KnowledgeBasePermissionRepository = (*kbFakePerms)(nil)

func newKbFakePerms(pages *kbFakePages, fallback domain.PagePermission) *kbFakePerms {
	return &kbFakePerms{
		pages:    pages,
		members:  map[string]map[uint64]bool{},
		perPage:  map[kbPermKey]domain.PagePermission{},
		fallback: fallback,
	}
}

func (f *kbFakePerms) addMember(workspaceID string, userID uint64) {
	if f.members[workspaceID] == nil {
		f.members[workspaceID] = map[uint64]bool{}
	}
	f.members[workspaceID][userID] = true
}

func (f *kbFakePerms) setPagePermission(pageID string, userID uint64, perm domain.PagePermission) {
	f.perPage[kbPermKey{pageID: pageID, userID: userID}] = perm
}

// factsFor は望む実効権限になる事実を組み立てる。役割（既定）だけで表現し、
// 例外（page_restrictions）は使わない。ResolvePagePermission を必ず通す形にしてある。
func factsFor(perm domain.PagePermission) domain.PagePermissionFacts {
	facts := domain.PagePermissionFacts{Member: true}
	switch {
	case perm.CanEdit:
		role := domain.GrantRoleEditor
		facts.Role = &role
	case perm.CanView:
		role := domain.GrantRoleViewer
		facts.Role = &role
	}
	return facts
}

func (f *kbFakePerms) permFor(pageID string, userID uint64) domain.PagePermission {
	if perm, ok := f.perPage[kbPermKey{pageID: pageID, userID: userID}]; ok {
		return perm
	}
	return f.fallback
}

func (f *kbFakePerms) IsWorkspaceMember(_ context.Context, workspaceID string, userID uint64) (bool, error) {
	if f.membersErr != nil {
		return false, f.membersErr
	}
	return f.members[workspaceID][userID], nil
}

func (f *kbFakePerms) PagePermissionFactsForUser(ctx context.Context, workspaceID, pageID string, userID uint64) (*domain.PagePermissionFacts, error) {
	if _, err := f.pages.FindPage(ctx, workspaceID, pageID); err != nil {
		return nil, err
	}
	facts := factsFor(f.permFor(pageID, userID))
	return &facts, nil
}

func (f *kbFakePerms) ListSpacePageViewFacts(_ context.Context, workspaceID, spaceID string, userID uint64) ([]repository.PageViewFacts, error) {
	if f.listFactsErr != nil {
		return nil, f.listFactsErr
	}
	pages := f.pages.activePages(workspaceID, spaceID)
	out := make([]repository.PageViewFacts, 0, len(pages))
	for _, p := range pages {
		out = append(out, repository.PageViewFacts{Page: p, Facts: factsFor(f.permFor(p.ID, userID))})
	}
	return out, nil
}

// --- 以降はこのテストで使わない口（interface を満たすためのスタブ）。

func (f *kbFakePerms) EnsureUserPrincipal(context.Context, string, uint64) (*domain.Principal, error) {
	return nil, nil
}

func (f *kbFakePerms) EnsureSpaceEveryonePrincipal(context.Context, string, string) (*domain.Principal, error) {
	return nil, nil
}

func (f *kbFakePerms) CreateGroupPrincipal(context.Context, string, string) (*domain.Principal, error) {
	return nil, nil
}

func (f *kbFakePerms) FindPrincipal(context.Context, string, string) (*domain.Principal, error) {
	return nil, repository.ErrPrincipalNotFound
}

func (f *kbFakePerms) FindUserPrincipal(context.Context, string, uint64) (*domain.Principal, error) {
	return nil, repository.ErrPrincipalNotFound
}

func (f *kbFakePerms) DeletePrincipal(context.Context, string, string) error { return nil }

func (f *kbFakePerms) AddGroupMember(context.Context, string, string, string) error { return nil }

func (f *kbFakePerms) RemoveGroupMember(context.Context, string, string, string) error { return nil }

func (f *kbFakePerms) UpsertWorkspaceGrant(context.Context, string, string, domain.GrantRole) (*domain.WorkspaceGrant, error) {
	return nil, nil
}

func (f *kbFakePerms) DeleteWorkspaceGrant(context.Context, string, string) error { return nil }

func (f *kbFakePerms) ListWorkspaceGrants(context.Context, string) ([]domain.WorkspaceGrant, error) {
	return nil, nil
}

func (f *kbFakePerms) UpsertSpaceGrant(context.Context, string, string, string, domain.GrantRole) (*domain.SpaceGrant, error) {
	return nil, nil
}

func (f *kbFakePerms) DeleteSpaceGrant(context.Context, string, string, string) error { return nil }

func (f *kbFakePerms) ListSpaceGrants(context.Context, string, string) ([]domain.SpaceGrant, error) {
	return nil, nil
}

func (f *kbFakePerms) UpsertPageRestriction(context.Context, string, string, string, domain.Capability, domain.RestrictionMode) (*domain.PageRestriction, error) {
	return nil, nil
}

func (f *kbFakePerms) DeletePageRestriction(context.Context, string, string, string, domain.Capability) error {
	return nil
}

func (f *kbFakePerms) ListPageRestrictions(context.Context, string, string) ([]domain.PageRestriction, error) {
	return nil, nil
}

func (f *kbFakePerms) CreateShareLink(context.Context, repository.ShareLinkWrite) (*domain.ShareLink, error) {
	return nil, nil
}

func (f *kbFakePerms) RevokeShareLink(context.Context, string, string) error { return nil }

func (f *kbFakePerms) FindShareLinkByTokenHash(context.Context, []byte) (*domain.ShareLink, error) {
	return nil, repository.ErrShareLinkNotFound
}

func (f *kbFakePerms) ListPageShareLinks(context.Context, string, string) ([]domain.ShareLink, error) {
	return nil, nil
}

func (f *kbFakePerms) PagePermissionFactsForPrincipal(context.Context, string, string, string) (*domain.PagePermissionFacts, error) {
	return nil, repository.ErrPageNotFound
}
