package handler

import (
	"context"
	"errors"
	"sort"
	"strconv"
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
	// moveErr は MovePage を指定のエラーで失敗させる。移動でしか起きないセンチネル
	// （スペース全員宛ての例外が失効する移動）を handler 越しに見るために分けてある。
	moveErr error
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
	if f.moveErr != nil {
		return f.moveErr
	}
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

// ancestorsOf は対象ページから根までの経路を近い順に返す（本番の page_paths と同じ並びで、
// 先頭が対象ページ自身 = depth 0）。返る添字がそのまま depth になる。
func (f *kbFakePages) ancestorsOf(workspaceID, pageID string) []string {
	path := make([]string, 0, 4)
	seen := map[string]bool{}
	for cur, ok := f.pages[pageID], true; ok && cur != nil; {
		if cur.WorkspaceID != workspaceID || seen[cur.ID] {
			break
		}
		seen[cur.ID] = true
		path = append(path, cur.ID)
		if cur.ParentID == nil {
			break
		}
		cur, ok = f.pages[*cur.ParentID]
	}
	return path
}

// kbPermKey は「どのページを誰が」の組。
type kbPermKey struct {
	pageID string
	userID uint64
}

// kbRestrictionKey は page_restrictions の主キー（ワークスペースはページから決まる）。
type kbRestrictionKey struct {
	pageID      string
	principalID string
	capability  domain.Capability
}

// kbAllowListKey は page_allow_lists の主キー。主体を含まないのが要点で、
// 許可リストに載っていた主体が消えても印だけは残る。
type kbAllowListKey struct {
	pageID     string
	capability domain.Capability
}

// errKbFakeNotModeled はこの fake が再現していない口を呼ばれたときのエラー。
//
// nil を返して黙って成功させない。ページ handler の経路にはまだ権限管理・共有リンクの
// エンドポイントが無く、ここが呼ばれるのは配線が変わったときだけなので、
// そのときは 500 として目に見えるようにする。
var errKbFakeNotModeled = errors.New("kb fake: この口は再現していない")

// kbFakePerms は repository.KnowledgeBasePermissionRepository の in-memory fake。
//
// 判定そのものはせず、domain.ResolvePagePermission がその答えを出すような「事実」を返す。
// 例外（page_restrictions）と「その段が許可リスト制である」という印（page_allow_lists）は
// 本番の repository と同じ規則で持ち回る:
//
//   - allow を張ると印が立つ
//   - 最後の allow を消す / allow を deny へ書き換えて allow が 0 行になると印が畳まれる
//   - deny 行を消しても印は畳まれない
//   - 主体を消すと allow 行は道連れに消えるが、印は残る（＝ 閉じる側に倒れる）
//
// 印を「allow 行があるか」で代用すると、主体を 1 つ消しただけで限定公開が解けるという
// 本番には無い挙動になり、その穴を踏むテストが緑のまま通ってしまう。
type kbFakePerms struct {
	pages *kbFakePages
	// principals は principalID -> 主体。ワークスペース所属は kind='user' の行の有無で表す
	// （本番と同じく、メンバーシップ専用の表は持たない）。
	principals map[string]*domain.Principal
	// groupMembers は groupPrincipalID -> memberPrincipalID の集合。
	groupMembers map[string]map[string]bool
	// restrictions は page_restrictions の行。
	restrictions map[kbRestrictionKey]domain.RestrictionMode
	// allowLists は page_allow_lists の印。
	allowLists map[kbAllowListKey]bool
	nextID     int
	// perPage は特定のページだけ別の既定にしたいときの上書き。
	perPage map[kbPermKey]domain.PagePermission
	// fallback は perPage に無いページの既定。
	fallback domain.PagePermission
	// membersErr は所属判定を失敗させる（middleware の 500 経路の確認用）。
	membersErr error
	// listFactsErr は一覧の事実収集を失敗させる（ツリー取得の 500 経路の確認用）。
	listFactsErr error
	// subtreeFactsErr はサブツリーの事実収集を失敗させる（アーカイブの 500 経路の確認用）。
	subtreeFactsErr error
}

var _ repository.KnowledgeBasePermissionRepository = (*kbFakePerms)(nil)

func newKbFakePerms(pages *kbFakePages, fallback domain.PagePermission) *kbFakePerms {
	return &kbFakePerms{
		pages:        pages,
		principals:   map[string]*domain.Principal{},
		groupMembers: map[string]map[string]bool{},
		restrictions: map[kbRestrictionKey]domain.RestrictionMode{},
		allowLists:   map[kbAllowListKey]bool{},
		perPage:      map[kbPermKey]domain.PagePermission{},
		fallback:     fallback,
	}
}

// addMember はユーザーをワークスペースに所属させる（kind='user' の主体を 1 行作る）。
func (f *kbFakePerms) addMember(workspaceID string, userID uint64) {
	_, _ = f.EnsureUserPrincipal(context.Background(), workspaceID, userID)
}

// setPagePermission はそのページでの既定（役割）を差し替える。
// 例外（page_restrictions）ではないので、子孫には伝わらない。
func (f *kbFakePerms) setPagePermission(pageID string, userID uint64, perm domain.PagePermission) {
	f.perPage[kbPermKey{pageID: pageID, userID: userID}] = perm
}

// roleFor は望む実効権限になる既定の役割を返す（例外を使わずに表現する）。
func roleFor(perm domain.PagePermission) *domain.GrantRole {
	switch {
	case perm.CanEdit:
		role := domain.GrantRoleEditor
		return &role
	case perm.CanView:
		role := domain.GrantRoleViewer
		return &role
	default:
		return nil
	}
}

func (f *kbFakePerms) permFor(pageID string, userID uint64) domain.PagePermission {
	if perm, ok := f.perPage[kbPermKey{pageID: pageID, userID: userID}]; ok {
		return perm
	}
	return f.fallback
}

// newPrincipal は主体を 1 行作って複製を返す。
func (f *kbFakePerms) newPrincipal(p domain.Principal) *domain.Principal {
	f.nextID++
	p.ID = "principal-" + strconv.Itoa(f.nextID)
	stored := p
	f.principals[p.ID] = &stored
	c := stored
	return &c
}

// userPrincipal はそのユーザーの主体を引く（無ければ nil = 非メンバー）。
func (f *kbFakePerms) userPrincipal(workspaceID string, userID uint64) *domain.Principal {
	for _, p := range f.principals {
		if p.WorkspaceID == workspaceID && p.Kind == domain.PrincipalKindUser && p.UserID != nil && *p.UserID == userID {
			return p
		}
	}
	return nil
}

// mine は「そのユーザーとして扱われる主体」の集合（本人 / 所属グループ / そのスペースの全員）。
// 本番の権限クエリの mine CTE と同じ組み立て方で、非メンバーには空集合を返す。
func (f *kbFakePerms) mine(workspaceID, spaceID string, userID uint64) map[string]bool {
	self := f.userPrincipal(workspaceID, userID)
	if self == nil {
		return map[string]bool{}
	}
	out := map[string]bool{self.ID: true}
	for groupID, members := range f.groupMembers {
		if members[self.ID] {
			out[groupID] = true
		}
	}
	for _, p := range f.principals {
		if p.WorkspaceID == workspaceID && p.Kind == domain.PrincipalKindSpaceAll && p.SpaceID != nil && *p.SpaceID == spaceID {
			out[p.ID] = true
		}
	}
	return out
}

// restrictionFacts は経路上の例外と印を「deny は経路全体・許可リストは最も近い段」に畳む。
// 経路に 1 行も無ければ nil（domain 側が既定へのフォールバックと解釈する）。
func (f *kbFakePerms) restrictionFacts(
	workspaceID, pageID string, capability domain.Capability, mine map[string]bool,
) *domain.RestrictionFacts {
	restricted := false
	denied := false
	nearest := -1
	for depth, ancestorID := range f.pages.ancestorsOf(workspaceID, pageID) {
		if f.allowLists[kbAllowListKey{pageID: ancestorID, capability: capability}] {
			restricted = true
			if nearest < 0 {
				nearest = depth
			}
		}
		for key, mode := range f.restrictions {
			if key.pageID != ancestorID || key.capability != capability {
				continue
			}
			// 自分宛てでない行も「経路に制限がある」ことは示す（本番の view_restricted と同じ）。
			restricted = true
			if mode == domain.RestrictionModeDeny && mine[key.principalID] {
				denied = true
			}
		}
	}
	if !restricted {
		return nil
	}
	allowedAtNearest := false
	if nearest >= 0 {
		nearestPageID := f.pages.ancestorsOf(workspaceID, pageID)[nearest]
		for key, mode := range f.restrictions {
			if key.pageID == nearestPageID && key.capability == capability &&
				mode == domain.RestrictionModeAllow && mine[key.principalID] {
				allowedAtNearest = true
			}
		}
	}
	return &domain.RestrictionFacts{
		DeniedAnywhere:   denied,
		HasAllowList:     nearest >= 0,
		AllowedAtNearest: allowedAtNearest,
	}
}

func (f *kbFakePerms) IsWorkspaceMember(_ context.Context, workspaceID string, userID uint64) (bool, error) {
	if f.membersErr != nil {
		return false, f.membersErr
	}
	return f.userPrincipal(workspaceID, userID) != nil, nil
}

func (f *kbFakePerms) PagePermissionFactsForUser(ctx context.Context, workspaceID, pageID string, userID uint64) (*domain.PagePermissionFacts, error) {
	page, err := f.pages.FindPage(ctx, workspaceID, pageID)
	if err != nil {
		return nil, err
	}
	mine := f.mine(workspaceID, page.SpaceID, userID)
	return &domain.PagePermissionFacts{
		Member: f.userPrincipal(workspaceID, userID) != nil,
		Role:   roleFor(f.permFor(pageID, userID)),
		View:   f.restrictionFacts(workspaceID, pageID, domain.CapabilityView, mine),
		Edit:   f.restrictionFacts(workspaceID, pageID, domain.CapabilityEdit, mine),
	}, nil
}

func (f *kbFakePerms) ListSpacePageViewFacts(_ context.Context, workspaceID, spaceID string, userID uint64) ([]repository.PageWithViewFacts, error) {
	if f.listFactsErr != nil {
		return nil, f.listFactsErr
	}
	mine := f.mine(workspaceID, spaceID, userID)
	pages := f.pages.activePages(workspaceID, spaceID)
	out := make([]repository.PageWithViewFacts, 0, len(pages))
	for _, p := range pages {
		out = append(out, repository.PageWithViewFacts{
			Page: p,
			Facts: domain.PageViewFacts{
				Role: roleFor(f.permFor(p.ID, userID)),
				View: f.restrictionFacts(workspaceID, p.ID, domain.CapabilityView, mine),
			},
		})
	}
	return out, nil
}

// ListSubtreePagePermissionFacts は対象ページ自身と全子孫の事実を返す（アーカイブ済みも含む）。
// 本番のクエリと同じく、主体の集合は根のスペースで決める（サブツリーは 1 スペースに収まる）。
func (f *kbFakePerms) ListSubtreePagePermissionFacts(
	ctx context.Context, workspaceID, pageID string, userID uint64,
) ([]repository.PageWithPermissionFacts, error) {
	if f.subtreeFactsErr != nil {
		return nil, f.subtreeFactsErr
	}
	root, ok := f.pages.pages[pageID]
	if !ok || root.WorkspaceID != workspaceID {
		// closure が 1 行も無い状態と同じ（呼び出し側はここを許可に倒さない）。
		return []repository.PageWithPermissionFacts{}, nil
	}
	mine := f.mine(workspaceID, root.SpaceID, userID)
	member := f.userPrincipal(workspaceID, userID) != nil
	ids := make([]string, 0, len(f.pages.pages))
	for id, p := range f.pages.pages {
		if p.WorkspaceID != workspaceID {
			continue
		}
		if desc, _ := f.pages.HasDescendant(ctx, workspaceID, pageID, id); desc {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	out := make([]repository.PageWithPermissionFacts, 0, len(ids))
	for _, id := range ids {
		out = append(out, repository.PageWithPermissionFacts{
			PageID: id,
			Facts: domain.PagePermissionFacts{
				Member: member,
				Role:   roleFor(f.permFor(id, userID)),
				View:   f.restrictionFacts(workspaceID, id, domain.CapabilityView, mine),
				Edit:   f.restrictionFacts(workspaceID, id, domain.CapabilityEdit, mine),
			},
		})
	}
	return out, nil
}

func (f *kbFakePerms) EnsureUserPrincipal(_ context.Context, workspaceID string, userID uint64) (*domain.Principal, error) {
	if p := f.userPrincipal(workspaceID, userID); p != nil {
		c := *p
		return &c, nil
	}
	uid := userID
	return f.newPrincipal(domain.Principal{
		WorkspaceID: workspaceID, Kind: domain.PrincipalKindUser, UserID: &uid,
	}), nil
}

func (f *kbFakePerms) EnsureSpaceEveryonePrincipal(_ context.Context, workspaceID, spaceID string) (*domain.Principal, error) {
	for _, p := range f.principals {
		if p.WorkspaceID == workspaceID && p.Kind == domain.PrincipalKindSpaceAll && p.SpaceID != nil && *p.SpaceID == spaceID {
			c := *p
			return &c, nil
		}
	}
	sid := spaceID
	return f.newPrincipal(domain.Principal{
		WorkspaceID: workspaceID, Kind: domain.PrincipalKindSpaceAll, SpaceID: &sid,
	}), nil
}

func (f *kbFakePerms) CreateGroupPrincipal(_ context.Context, workspaceID, name string) (*domain.Principal, error) {
	return f.newPrincipal(domain.Principal{
		WorkspaceID: workspaceID, Kind: domain.PrincipalKindGroup, Name: name,
	}), nil
}

func (f *kbFakePerms) FindPrincipal(_ context.Context, workspaceID, principalID string) (*domain.Principal, error) {
	p, ok := f.principals[principalID]
	if !ok || p.WorkspaceID != workspaceID {
		return nil, repository.ErrPrincipalNotFound
	}
	c := *p
	return &c, nil
}

func (f *kbFakePerms) FindUserPrincipal(_ context.Context, workspaceID string, userID uint64) (*domain.Principal, error) {
	p := f.userPrincipal(workspaceID, userID)
	if p == nil {
		return nil, repository.ErrPrincipalNotFound
	}
	c := *p
	return &c, nil
}

// DeletePrincipal は主体と、それに紐づく例外・グループ所属を消す（本番の FK CASCADE と同じ）。
// 許可リスト制の印には触れない。載っていた主体が全員消えた段は「誰も載っていない許可リスト」
// として残り、閉じたままになる。
func (f *kbFakePerms) DeletePrincipal(_ context.Context, workspaceID, principalID string) error {
	p, ok := f.principals[principalID]
	if !ok || p.WorkspaceID != workspaceID {
		return repository.ErrPrincipalNotFound
	}
	delete(f.principals, principalID)
	delete(f.groupMembers, principalID)
	for _, members := range f.groupMembers {
		delete(members, principalID)
	}
	for key := range f.restrictions {
		if key.principalID == principalID {
			delete(f.restrictions, key)
		}
	}
	return nil
}

func (f *kbFakePerms) AddGroupMember(_ context.Context, _, groupPrincipalID, memberPrincipalID string) error {
	if f.groupMembers[groupPrincipalID] == nil {
		f.groupMembers[groupPrincipalID] = map[string]bool{}
	}
	f.groupMembers[groupPrincipalID][memberPrincipalID] = true
	return nil
}

func (f *kbFakePerms) RemoveGroupMember(_ context.Context, _, groupPrincipalID, memberPrincipalID string) error {
	delete(f.groupMembers[groupPrincipalID], memberPrincipalID)
	return nil
}

// UpsertPageRestriction は例外 1 行と許可リスト制の印を同時に揃える（本番と同じ 1 トランザクション）。
func (f *kbFakePerms) UpsertPageRestriction(
	_ context.Context, workspaceID, pageID, principalID string,
	capability domain.Capability, mode domain.RestrictionMode,
) (*domain.PageRestriction, error) {
	page, ok := f.pages.pages[pageID]
	if !ok || page.WorkspaceID != workspaceID {
		return nil, repository.ErrPageNotFound
	}
	key := kbRestrictionKey{pageID: pageID, principalID: principalID, capability: capability}
	prev, existed := f.restrictions[key]
	f.restrictions[key] = mode
	switch {
	case mode == domain.RestrictionModeAllow:
		f.allowLists[kbAllowListKey{pageID: pageID, capability: capability}] = true
	case existed && prev == domain.RestrictionModeAllow:
		// allow を deny へ書き換えた。その段の最後の allow だったなら印も畳む。
		f.unmarkAllowListIfEmpty(pageID, capability)
	}
	return &domain.PageRestriction{
		WorkspaceID: workspaceID, PageID: pageID, PrincipalID: principalID,
		Capability: capability, Mode: mode,
	}, nil
}

// DeletePageRestriction は例外を解除する。消したのが allow 行のときだけ印を畳む。
func (f *kbFakePerms) DeletePageRestriction(
	_ context.Context, workspaceID, pageID, principalID string, capability domain.Capability,
) error {
	page, ok := f.pages.pages[pageID]
	if !ok || page.WorkspaceID != workspaceID {
		return nil
	}
	key := kbRestrictionKey{pageID: pageID, principalID: principalID, capability: capability}
	prev, existed := f.restrictions[key]
	if !existed {
		return nil // 元から無い（冪等）
	}
	delete(f.restrictions, key)
	if prev == domain.RestrictionModeAllow {
		f.unmarkAllowListIfEmpty(pageID, capability)
	}
	return nil
}

// unmarkAllowListIfEmpty は allow 行が 1 行も残っていない段の印を畳む。
// 呼ぶのは「allow 行を明示的に減らした」直後だけ（本番の UnmarkPageAllowListIfEmpty と同じ）。
func (f *kbFakePerms) unmarkAllowListIfEmpty(pageID string, capability domain.Capability) {
	for key, mode := range f.restrictions {
		if key.pageID == pageID && key.capability == capability && mode == domain.RestrictionModeAllow {
			return
		}
	}
	delete(f.allowLists, kbAllowListKey{pageID: pageID, capability: capability})
}

func (f *kbFakePerms) ListPageRestrictions(_ context.Context, workspaceID, pageID string) ([]domain.PageRestriction, error) {
	page, ok := f.pages.pages[pageID]
	if !ok || page.WorkspaceID != workspaceID {
		return []domain.PageRestriction{}, nil
	}
	out := make([]domain.PageRestriction, 0, len(f.restrictions))
	for key, mode := range f.restrictions {
		if key.pageID != pageID {
			continue
		}
		out = append(out, domain.PageRestriction{
			WorkspaceID: workspaceID, PageID: pageID, PrincipalID: key.principalID,
			Capability: key.capability, Mode: mode,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PrincipalID != out[j].PrincipalID {
			return out[i].PrincipalID < out[j].PrincipalID
		}
		return out[i].Capability < out[j].Capability
	})
	return out, nil
}

// ListPageAllowListCapabilities はそのページ自身が許可リスト制になっているケイパビリティを返す。
// 載っていた主体が消えて allow 行が 0 行になった段はここにしか現れない。
func (f *kbFakePerms) ListPageAllowListCapabilities(_ context.Context, workspaceID, pageID string) ([]domain.Capability, error) {
	page, ok := f.pages.pages[pageID]
	if !ok || page.WorkspaceID != workspaceID {
		return []domain.Capability{}, nil
	}
	out := make([]domain.Capability, 0, len(f.allowLists))
	for key := range f.allowLists {
		if key.pageID == pageID {
			out = append(out, key.capability)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

// --- 以降はこの段の handler が通らない口。呼ばれたら黙って成功させずエラーにする
// （権限管理・共有リンクのエンドポイントはまだ無く、通り始めたら配線の変化として気づきたい）。

func (f *kbFakePerms) UpsertWorkspaceGrant(context.Context, string, string, domain.GrantRole) (*domain.WorkspaceGrant, error) {
	return nil, errKbFakeNotModeled
}

func (f *kbFakePerms) DeleteWorkspaceGrant(context.Context, string, string) error {
	return errKbFakeNotModeled
}

func (f *kbFakePerms) ListWorkspaceGrants(context.Context, string) ([]domain.WorkspaceGrant, error) {
	return nil, errKbFakeNotModeled
}

func (f *kbFakePerms) UpsertSpaceGrant(context.Context, string, string, string, domain.GrantRole) (*domain.SpaceGrant, error) {
	return nil, errKbFakeNotModeled
}

func (f *kbFakePerms) DeleteSpaceGrant(context.Context, string, string, string) error {
	return errKbFakeNotModeled
}

func (f *kbFakePerms) ListSpaceGrants(context.Context, string, string) ([]domain.SpaceGrant, error) {
	return nil, errKbFakeNotModeled
}

func (f *kbFakePerms) CreateShareLink(context.Context, repository.ShareLinkWrite) (*domain.ShareLink, error) {
	return nil, errKbFakeNotModeled
}

func (f *kbFakePerms) RevokeShareLink(context.Context, string, string) error {
	return errKbFakeNotModeled
}

func (f *kbFakePerms) FindShareLinkByTokenHash(context.Context, []byte) (*domain.ShareLink, error) {
	return nil, repository.ErrShareLinkNotFound
}

func (f *kbFakePerms) ListPageShareLinks(context.Context, string, string) ([]domain.ShareLink, error) {
	return nil, errKbFakeNotModeled
}

func (f *kbFakePerms) PagePermissionFactsForPrincipal(context.Context, string, string, string) (*domain.PagePermissionFacts, error) {
	return nil, errKbFakeNotModeled
}
