package handler

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// チケット handler テスト用の in-memory fake（repository.TicketRepository）。
//
// kbFakePages / kbFakePerms と同じ考え方: 判定や採番の規則を fake が肩代わりしない
// （フィルタ・重複検査くらいはここで行うが、権限判定は一切持たない — permission_usecase.go が
// 使うのは kbFakePerms 側の SpacePermissionFactsForUser で、こちらとは別の依存）。
type ticketFakeRepo struct {
	statuses    map[string]*domain.TicketStatus
	types       map[string]*domain.TicketType
	tickets     map[string]*domain.Ticket
	assignments map[string]*domain.TicketAssignment   // ticketID -> assignment
	changes     map[string][]domain.TicketChangeGroup // ticketID -> グループ（追加順）
	pageLinks   map[string][]string                   // ticketID -> pageID
	ticketLinks map[string][]string                   // ticketID -> ticketID
	nextID      int
	// numbers はスペースごとの採番カウンタ（本番の ticket_counters の代わり）。
	numbers map[string]int64
}

func newTicketFakeRepo() *ticketFakeRepo {
	return &ticketFakeRepo{
		statuses:    map[string]*domain.TicketStatus{},
		types:       map[string]*domain.TicketType{},
		tickets:     map[string]*domain.Ticket{},
		assignments: map[string]*domain.TicketAssignment{},
		changes:     map[string][]domain.TicketChangeGroup{},
		pageLinks:   map[string][]string{},
		ticketLinks: map[string][]string{},
		numbers:     map[string]int64{},
	}
}

func (f *ticketFakeRepo) newID(prefix string) string {
	f.nextID++
	return fmt.Sprintf("%s-%04d", prefix, f.nextID)
}

// addStatus / addType はテストの下ごしらえ用（採番・一意性検査を経由しない直接投入）。
func (f *ticketFakeRepo) addStatus(s domain.TicketStatus) *domain.TicketStatus {
	stored := s
	f.statuses[s.ID] = &stored
	return &stored
}

func (f *ticketFakeRepo) addType(t domain.TicketType) *domain.TicketType {
	stored := t
	f.types[t.ID] = &stored
	return &stored
}

func (f *ticketFakeRepo) addTicket(t domain.Ticket) *domain.Ticket {
	if t.Doc == nil {
		t.Doc = []byte(`{"type":"doc","content":[]}`)
	}
	if t.Position == "" {
		t.Position = "a0"
	}
	stored := t
	f.tickets[t.ID] = &stored
	return &stored
}

// --- 状態 ---

func (f *ticketFakeRepo) HasActiveInitialTicketStatus(_ context.Context, workspaceID, spaceID string) (bool, error) {
	for _, s := range f.statuses {
		if s.WorkspaceID == workspaceID && s.SpaceID == spaceID && s.IsInitial && s.ArchivedAt == nil {
			return true, nil
		}
	}
	return false, nil
}

func (f *ticketFakeRepo) InsertTicketStatus(_ context.Context, s *domain.TicketStatus) error {
	for _, other := range f.statuses {
		if other.WorkspaceID == s.WorkspaceID && other.SpaceID == s.SpaceID && other.ArchivedAt == nil &&
			strings.EqualFold(other.Name, s.Name) {
			return repository.ErrTicketStatusNameTaken
		}
	}
	s.ID = f.newID("status")
	s.CreatedAt, s.UpdatedAt = time.Now(), time.Now()
	stored := *s
	f.statuses[s.ID] = &stored
	return nil
}

func (f *ticketFakeRepo) FindTicketStatus(_ context.Context, workspaceID, spaceID, statusID string) (*domain.TicketStatus, error) {
	s, ok := f.statuses[statusID]
	if !ok || s.WorkspaceID != workspaceID || s.SpaceID != spaceID {
		return nil, repository.ErrTicketStatusNotFound
	}
	cp := *s
	return &cp, nil
}

func (f *ticketFakeRepo) ListTicketStatuses(_ context.Context, workspaceID, spaceID string, includeArchived bool) ([]domain.TicketStatus, error) {
	var out []domain.TicketStatus
	for _, s := range f.statuses {
		if s.WorkspaceID != workspaceID || s.SpaceID != spaceID {
			continue
		}
		if s.ArchivedAt != nil && !includeArchived {
			continue
		}
		out = append(out, *s)
	}
	// 本番の SQL は ORDER BY "position"。呼び出し側（例: 並び替えの隣接探索）が
	// 順序に依存するので、map の走査順（毎プロセス起動でランダム化される）のまま返すと
	// -race の有無に関わらずテストが偶発的に落ちる（実測）。
	sort.Slice(out, func(i, j int) bool { return out[i].Position < out[j].Position })
	return out, nil
}

func (f *ticketFakeRepo) GetInitialTicketStatus(_ context.Context, workspaceID, spaceID string) (*domain.TicketStatus, error) {
	for _, s := range f.statuses {
		if s.WorkspaceID == workspaceID && s.SpaceID == spaceID && s.IsInitial && s.ArchivedAt == nil {
			cp := *s
			return &cp, nil
		}
	}
	return nil, repository.ErrTicketStatusNotFound
}

func (f *ticketFakeRepo) UpdateTicketStatus(_ context.Context, s *domain.TicketStatus) error {
	existing, ok := f.statuses[s.ID]
	if !ok || existing.WorkspaceID != s.WorkspaceID || existing.SpaceID != s.SpaceID {
		return repository.ErrTicketStatusNotFound
	}
	for _, other := range f.statuses {
		if other.ID != s.ID && other.WorkspaceID == s.WorkspaceID && other.SpaceID == s.SpaceID &&
			other.ArchivedAt == nil && strings.EqualFold(other.Name, s.Name) {
			return repository.ErrTicketStatusNameTaken
		}
	}
	existing.Name, existing.Category, existing.Color = s.Name, s.Category, s.Color
	existing.UpdatedAt = time.Now()
	*s = *existing
	return nil
}

func (f *ticketFakeRepo) SetTicketStatusInitial(_ context.Context, workspaceID, spaceID, statusID string) error {
	s, ok := f.statuses[statusID]
	if !ok || s.WorkspaceID != workspaceID || s.SpaceID != spaceID {
		return repository.ErrTicketStatusNotFound
	}
	for _, other := range f.statuses {
		if other.WorkspaceID == workspaceID && other.SpaceID == spaceID {
			other.IsInitial = false
		}
	}
	s.IsInitial = true
	return nil
}

func (f *ticketFakeRepo) ArchiveTicketStatus(_ context.Context, workspaceID, spaceID, statusID string) error {
	s, ok := f.statuses[statusID]
	if !ok || s.WorkspaceID != workspaceID || s.SpaceID != spaceID {
		return repository.ErrTicketStatusNotFound
	}
	now := time.Now()
	s.ArchivedAt = &now
	return nil
}

func (f *ticketFakeRepo) RestoreTicketStatus(_ context.Context, workspaceID, spaceID, statusID, position string) error {
	s, ok := f.statuses[statusID]
	if !ok || s.WorkspaceID != workspaceID || s.SpaceID != spaceID {
		return repository.ErrTicketStatusNotFound
	}
	for _, other := range f.statuses {
		if other.ID != statusID && other.WorkspaceID == workspaceID && other.SpaceID == spaceID &&
			other.ArchivedAt == nil && strings.EqualFold(other.Name, s.Name) {
			return repository.ErrTicketStatusNameTaken
		}
	}
	s.ArchivedAt = nil
	s.Position = position
	return nil
}

func (f *ticketFakeRepo) CountActiveTicketsByStatus(_ context.Context, workspaceID, spaceID, statusID string) (int64, error) {
	var n int64
	for _, t := range f.tickets {
		if t.WorkspaceID == workspaceID && t.SpaceID == spaceID && t.StatusID == statusID && t.ArchivedAt == nil {
			n++
		}
	}
	return n, nil
}

func (f *ticketFakeRepo) LastActiveTicketStatusPosition(_ context.Context, workspaceID, spaceID string) (string, error) {
	last := ""
	for _, s := range f.statuses {
		if s.WorkspaceID == workspaceID && s.SpaceID == spaceID && s.ArchivedAt == nil && s.Position > last {
			last = s.Position
		}
	}
	return last, nil
}

// --- 種別 ---

func (f *ticketFakeRepo) InsertTicketType(_ context.Context, t *domain.TicketType) error {
	for _, other := range f.types {
		if other.WorkspaceID == t.WorkspaceID && other.SpaceID == t.SpaceID && other.ArchivedAt == nil &&
			strings.EqualFold(other.Name, t.Name) {
			return repository.ErrTicketTypeNameTaken
		}
	}
	t.ID = f.newID("type")
	t.CreatedAt, t.UpdatedAt = time.Now(), time.Now()
	stored := *t
	f.types[t.ID] = &stored
	return nil
}

func (f *ticketFakeRepo) FindTicketType(_ context.Context, workspaceID, spaceID, typeID string) (*domain.TicketType, error) {
	t, ok := f.types[typeID]
	if !ok || t.WorkspaceID != workspaceID || t.SpaceID != spaceID {
		return nil, repository.ErrTicketTypeNotFound
	}
	cp := *t
	return &cp, nil
}

func (f *ticketFakeRepo) ListTicketTypes(_ context.Context, workspaceID, spaceID string, includeArchived bool) ([]domain.TicketType, error) {
	var out []domain.TicketType
	for _, t := range f.types {
		if t.WorkspaceID != workspaceID || t.SpaceID != spaceID {
			continue
		}
		if t.ArchivedAt != nil && !includeArchived {
			continue
		}
		out = append(out, *t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Position < out[j].Position })
	return out, nil
}

func (f *ticketFakeRepo) GetDefaultTicketType(_ context.Context, workspaceID, spaceID string) (*domain.TicketType, error) {
	for _, t := range f.types {
		if t.WorkspaceID == workspaceID && t.SpaceID == spaceID && t.IsDefault && t.ArchivedAt == nil {
			cp := *t
			return &cp, nil
		}
	}
	return nil, repository.ErrTicketTypeNotFound
}

func (f *ticketFakeRepo) UpdateTicketType(_ context.Context, t *domain.TicketType) error {
	existing, ok := f.types[t.ID]
	if !ok || existing.WorkspaceID != t.WorkspaceID || existing.SpaceID != t.SpaceID {
		return repository.ErrTicketTypeNotFound
	}
	for _, other := range f.types {
		if other.ID != t.ID && other.WorkspaceID == t.WorkspaceID && other.SpaceID == t.SpaceID &&
			other.ArchivedAt == nil && strings.EqualFold(other.Name, t.Name) {
			return repository.ErrTicketTypeNameTaken
		}
	}
	existing.Name, existing.Color, existing.HierarchyLevel = t.Name, t.Color, t.HierarchyLevel
	existing.UpdatedAt = time.Now()
	*t = *existing
	return nil
}

func (f *ticketFakeRepo) SetTicketTypeDefault(_ context.Context, workspaceID, spaceID, typeID string) error {
	t, ok := f.types[typeID]
	if !ok || t.WorkspaceID != workspaceID || t.SpaceID != spaceID {
		return repository.ErrTicketTypeNotFound
	}
	for _, other := range f.types {
		if other.WorkspaceID == workspaceID && other.SpaceID == spaceID {
			other.IsDefault = false
		}
	}
	t.IsDefault = true
	return nil
}

func (f *ticketFakeRepo) ArchiveTicketType(_ context.Context, workspaceID, spaceID, typeID string) error {
	t, ok := f.types[typeID]
	if !ok || t.WorkspaceID != workspaceID || t.SpaceID != spaceID {
		return repository.ErrTicketTypeNotFound
	}
	now := time.Now()
	t.ArchivedAt = &now
	return nil
}

func (f *ticketFakeRepo) RestoreTicketType(_ context.Context, workspaceID, spaceID, typeID, position string) error {
	t, ok := f.types[typeID]
	if !ok || t.WorkspaceID != workspaceID || t.SpaceID != spaceID {
		return repository.ErrTicketTypeNotFound
	}
	for _, other := range f.types {
		if other.ID != typeID && other.WorkspaceID == workspaceID && other.SpaceID == spaceID &&
			other.ArchivedAt == nil && strings.EqualFold(other.Name, t.Name) {
			return repository.ErrTicketTypeNameTaken
		}
	}
	t.ArchivedAt = nil
	t.Position = position
	return nil
}

func (f *ticketFakeRepo) CountActiveTicketsByType(_ context.Context, workspaceID, spaceID, typeID string) (int64, error) {
	var n int64
	for _, t := range f.tickets {
		if t.WorkspaceID == workspaceID && t.SpaceID == spaceID && t.TypeID == typeID && t.ArchivedAt == nil {
			n++
		}
	}
	return n, nil
}

func (f *ticketFakeRepo) LastActiveTicketTypePosition(_ context.Context, workspaceID, spaceID string) (string, error) {
	last := ""
	for _, t := range f.types {
		if t.WorkspaceID == workspaceID && t.SpaceID == spaceID && t.ArchivedAt == nil && t.Position > last {
			last = t.Position
		}
	}
	return last, nil
}

// --- チケット本体 ---

func (f *ticketFakeRepo) CreateTicket(_ context.Context, in repository.TicketCreateInput) (*domain.Ticket, error) {
	f.numbers[in.SpaceID]++
	t := &domain.Ticket{
		ID: f.newID("ticket"), WorkspaceID: in.WorkspaceID, SpaceID: in.SpaceID,
		Number: f.numbers[in.SpaceID], TypeID: in.TypeID, StatusID: in.StatusID, ParentID: in.ParentID,
		Title: in.Title, Doc: in.Doc, PlainText: in.PlainText, Priority: in.Priority,
		StartDate: in.StartDate, DueDate: in.DueDate, Position: in.Position,
		CreatedByUserID: in.CreatedByUserID, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	stored := *t
	f.tickets[t.ID] = &stored
	cp := *t
	return &cp, nil
}

func (f *ticketFakeRepo) FindTicket(_ context.Context, workspaceID, ticketID string) (*domain.Ticket, error) {
	t, ok := f.tickets[ticketID]
	if !ok || t.WorkspaceID != workspaceID {
		return nil, repository.ErrTicketNotFound
	}
	cp := *t
	return &cp, nil
}

func (f *ticketFakeRepo) ResolveTicketIDByKey(_ context.Context, workspaceID, spaceKey string, number int64) (string, error) {
	for _, t := range f.tickets {
		if t.WorkspaceID != workspaceID || t.Number != number {
			continue
		}
		// fake ではスペースの key をスペース ID と同一視する（addSpace が Key: spaceID で
		// 作っているため。ResolveTicketIDByKey は spaceKey → spaceID の解決を本番では
		// SQL の JOIN が担うが、fake はこの対応だけで足りる）。
		if t.SpaceID == spaceKey {
			return t.ID, nil
		}
	}
	return "", repository.ErrTicketNotFound
}

func (f *ticketFakeRepo) ListTickets(_ context.Context, in repository.ListTicketsInput) ([]domain.Ticket, error) {
	var out []domain.Ticket
	for _, t := range f.tickets {
		if t.WorkspaceID != in.WorkspaceID || t.SpaceID != in.SpaceID {
			continue
		}
		if t.ArchivedAt != nil && !in.IncludeArchived {
			continue
		}
		if in.StatusID != nil && t.StatusID != *in.StatusID {
			continue
		}
		if in.TypeID != nil && t.TypeID != *in.TypeID {
			continue
		}
		if in.AssigneePrincipalID != nil {
			a, ok := f.assignments[t.ID]
			if !ok || a.AssigneePrincipalID != *in.AssigneePrincipalID {
				continue
			}
		}
		out = append(out, *t)
	}
	// MoveTicketUseCase.placementPosition が「隣の兄弟」を position 順の隣接として
	// 探すため、本番の SQL（ORDER BY t."position"）と同じ順序で返す必要がある
	// （順不同のままだと並び替えが偶発的に不正な範囲を fracindex.Between へ渡し、
	// テストが -race の有無に関わらずランダムに失敗する。実測）。
	sort.Slice(out, func(i, j int) bool { return out[i].Position < out[j].Position })
	return out, nil
}

func (f *ticketFakeRepo) ListTicketChildren(_ context.Context, workspaceID, spaceID, parentID string) ([]domain.Ticket, error) {
	var out []domain.Ticket
	for _, t := range f.tickets {
		if t.WorkspaceID == workspaceID && t.SpaceID == spaceID && t.ParentID != nil && *t.ParentID == parentID {
			out = append(out, *t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Position < out[j].Position })
	return out, nil
}

func (f *ticketFakeRepo) UpdateTicket(_ context.Context, workspaceID, ticketID string, fields repository.TicketUpdateFields) (*domain.Ticket, error) {
	t, ok := f.tickets[ticketID]
	if !ok || t.WorkspaceID != workspaceID {
		return nil, repository.ErrTicketNotFound
	}
	if _, ok := f.types[fields.TypeID]; !ok {
		return nil, repository.ErrTicketNotFound
	}
	t.TypeID, t.ParentID, t.Title = fields.TypeID, fields.ParentID, fields.Title
	t.Doc, t.PlainText, t.Priority = fields.Doc, fields.PlainText, fields.Priority
	t.StartDate, t.DueDate = fields.StartDate, fields.DueDate
	t.UpdatedAt = time.Now()
	cp := *t
	return &cp, nil
}

func (f *ticketFakeRepo) ChangeTicketStatus(
	_ context.Context, workspaceID, ticketID, statusID string,
	closedAt *time.Time, resolution *domain.TicketResolution,
) (*domain.Ticket, error) {
	t, ok := f.tickets[ticketID]
	if !ok || t.WorkspaceID != workspaceID {
		return nil, repository.ErrTicketNotFound
	}
	t.StatusID, t.ClosedAt, t.Resolution = statusID, closedAt, resolution
	t.UpdatedAt = time.Now()
	cp := *t
	return &cp, nil
}

func (f *ticketFakeRepo) MoveTicket(_ context.Context, workspaceID, ticketID, position string) error {
	t, ok := f.tickets[ticketID]
	if !ok || t.WorkspaceID != workspaceID {
		return repository.ErrTicketNotFound
	}
	t.Position = position
	return nil
}

func (f *ticketFakeRepo) ArchiveTicket(_ context.Context, workspaceID, ticketID string) error {
	t, ok := f.tickets[ticketID]
	if !ok || t.WorkspaceID != workspaceID {
		return repository.ErrTicketNotFound
	}
	now := time.Now()
	t.ArchivedAt = &now
	return nil
}

func (f *ticketFakeRepo) RestoreTicket(_ context.Context, workspaceID, ticketID, position string) error {
	t, ok := f.tickets[ticketID]
	if !ok || t.WorkspaceID != workspaceID {
		return repository.ErrTicketNotFound
	}
	t.ArchivedAt = nil
	t.Position = position
	return nil
}

func (f *ticketFakeRepo) CountActiveTicketChildren(_ context.Context, workspaceID, ticketID string) (int64, error) {
	var n int64
	for _, t := range f.tickets {
		if t.WorkspaceID == workspaceID && t.ParentID != nil && *t.ParentID == ticketID && t.ArchivedAt == nil {
			n++
		}
	}
	return n, nil
}

func (f *ticketFakeRepo) ListTicketParentChain(_ context.Context, workspaceID, ticketID string) ([]domain.Ticket, error) {
	var chain []domain.Ticket
	cur, ok := f.tickets[ticketID]
	if !ok || cur.WorkspaceID != workspaceID {
		return nil, repository.ErrTicketNotFound
	}
	for cur.ParentID != nil {
		parent, ok := f.tickets[*cur.ParentID]
		if !ok {
			break
		}
		chain = append([]domain.Ticket{*parent}, chain...)
		cur = parent
	}
	return chain, nil
}

func (f *ticketFakeRepo) LastActiveTicketPosition(_ context.Context, workspaceID, spaceID string) (string, error) {
	last := ""
	for _, t := range f.tickets {
		if t.WorkspaceID == workspaceID && t.SpaceID == spaceID && t.ArchivedAt == nil && t.Position > last {
			last = t.Position
		}
	}
	return last, nil
}

func (f *ticketFakeRepo) FindActiveTicketPosition(_ context.Context, workspaceID, spaceID, ticketID string) (string, bool, error) {
	t, ok := f.tickets[ticketID]
	if !ok || t.WorkspaceID != workspaceID || t.SpaceID != spaceID || t.ArchivedAt != nil {
		return "", false, nil
	}
	return t.Position, true, nil
}

// --- 担当 ---

func (f *ticketFakeRepo) UpsertTicketAssignment(_ context.Context, a *domain.TicketAssignment) error {
	t, ok := f.tickets[a.TicketID]
	if !ok || t.WorkspaceID != a.WorkspaceID {
		return repository.ErrTicketNotFound
	}
	// fake における principal の実在確認: kbFakePerms が作る principal ID の形
	// （"principal-user-<workspaceID>-<userID>" 等）は fake ごとに閉じているため、ここでは
	// 「principalNotFound」を明示的に注入したテストケースだけを弾く単純化に留める。
	if a.AssigneePrincipalID == ticketFakeMissingPrincipalID {
		return repository.ErrTicketAssigneeNotFound
	}
	a.CreatedAt = time.Now()
	stored := *a
	f.assignments[a.TicketID] = &stored
	return nil
}

// ticketFakeMissingPrincipalID はテストが「実在しない担当」を表すのに使う予約値。
const ticketFakeMissingPrincipalID = "principal-missing"

func (f *ticketFakeRepo) DeleteTicketAssignment(_ context.Context, workspaceID, ticketID string) error {
	a, ok := f.assignments[ticketID]
	if !ok || a.WorkspaceID != workspaceID {
		return nil
	}
	delete(f.assignments, ticketID)
	return nil
}

func (f *ticketFakeRepo) FindTicketAssignment(_ context.Context, workspaceID, ticketID string) (*domain.TicketAssignment, error) {
	a, ok := f.assignments[ticketID]
	if !ok || a.WorkspaceID != workspaceID {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}

func (f *ticketFakeRepo) ListTicketsAssignedToPrincipal(_ context.Context, workspaceID, principalID string) ([]domain.Ticket, error) {
	var out []domain.Ticket
	for ticketID, a := range f.assignments {
		if a.WorkspaceID != workspaceID || a.AssigneePrincipalID != principalID {
			continue
		}
		if t, ok := f.tickets[ticketID]; ok {
			out = append(out, *t)
		}
	}
	return out, nil
}

// --- 変更履歴 ---

func (f *ticketFakeRepo) InsertTicketChangeGroup(_ context.Context, g *domain.TicketChangeGroup) error {
	g.ID = f.newID("change")
	g.CreatedAt = time.Now()
	for i := range g.Items {
		g.Items[i].ID = f.newID("item")
		g.Items[i].GroupID = g.ID
	}
	f.changes[g.TicketID] = append(f.changes[g.TicketID], *g)
	return nil
}

func (f *ticketFakeRepo) ListTicketChangeGroups(_ context.Context, workspaceID, ticketID string) ([]domain.TicketChangeGroup, error) {
	groups := f.changes[ticketID]
	out := make([]domain.TicketChangeGroup, 0, len(groups))
	for i := len(groups) - 1; i >= 0; i-- {
		if groups[i].WorkspaceID == workspaceID {
			out = append(out, groups[i])
		}
	}
	return out, nil
}

// --- 派生表 ---

func (f *ticketFakeRepo) ReplaceTicketPageLinks(_ context.Context, workspaceID, sourceTicketID string, targetPageIDs []string) error {
	f.pageLinks[sourceTicketID] = targetPageIDs
	return nil
}

func (f *ticketFakeRepo) ReplaceTicketTicketLinks(_ context.Context, workspaceID, sourceTicketID string, targetTicketIDs []string) error {
	f.ticketLinks[sourceTicketID] = targetTicketIDs
	return nil
}

func (f *ticketFakeRepo) ListTicketPageLinks(_ context.Context, workspaceID, sourceTicketID string) ([]domain.TicketPageLink, error) {
	var out []domain.TicketPageLink
	for _, pageID := range f.pageLinks[sourceTicketID] {
		out = append(out, domain.TicketPageLink{WorkspaceID: workspaceID, SourceTicketID: sourceTicketID, TargetPageID: pageID})
	}
	return out, nil
}

func (f *ticketFakeRepo) ListPagesReferencingTicket(_ context.Context, workspaceID, targetTicketID string) ([]domain.TicketPageLink, error) {
	return nil, nil
}

func (f *ticketFakeRepo) ListTicketTicketLinks(_ context.Context, workspaceID, sourceTicketID string) ([]domain.TicketTicketLink, error) {
	var out []domain.TicketTicketLink
	for _, targetID := range f.ticketLinks[sourceTicketID] {
		out = append(out, domain.TicketTicketLink{WorkspaceID: workspaceID, SourceTicketID: sourceTicketID, TargetTicketID: targetID})
	}
	return out, nil
}

func (f *ticketFakeRepo) ListTicketsReferencingTicket(_ context.Context, workspaceID, targetTicketID string) ([]domain.TicketTicketLink, error) {
	var out []domain.TicketTicketLink
	for sourceID, targets := range f.ticketLinks {
		for _, targetID := range targets {
			if targetID == targetTicketID {
				out = append(out, domain.TicketTicketLink{WorkspaceID: workspaceID, SourceTicketID: sourceID, TargetTicketID: targetTicketID})
			}
		}
	}
	return out, nil
}

var _ repository.TicketRepository = (*ticketFakeRepo)(nil)
