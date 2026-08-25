package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// knowledgeBasePermissionRepository は [repository.KnowledgeBasePermissionRepository] の実装。
// ナレッジ基盤は GORM を通さない方針のため、クエリはすべて sqlc 生成コード + 素の *sql.DB で書く。
type knowledgeBasePermissionRepository struct {
	db *sql.DB
	q  *sqlcgen.Queries
}

// NewKnowledgeBasePermissionRepository はナレッジ基盤の権限 repository を組み立てる。
func NewKnowledgeBasePermissionRepository(db *sql.DB) repository.KnowledgeBasePermissionRepository {
	return &knowledgeBasePermissionRepository{db: db, q: sqlcgen.New(db)}
}

func toDomainPrincipal(row sqlcgen.Principal) domain.Principal {
	p := domain.Principal{
		ID:          row.ID.String(),
		WorkspaceID: row.WorkspaceID.String(),
		Kind:        domain.PrincipalKind(row.Kind),
		Name:        row.Name,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.UserID.Valid {
		id := uint64(row.UserID.Int64)
		p.UserID = &id
	}
	if row.SpaceID.Valid {
		id := row.SpaceID.UUID.String()
		p.SpaceID = &id
	}
	if row.PageID.Valid {
		id := row.PageID.UUID.String()
		p.PageID = &id
	}
	return p
}

func toDomainWorkspaceGrant(row sqlcgen.WorkspaceGrant) domain.WorkspaceGrant {
	return domain.WorkspaceGrant{
		WorkspaceID: row.WorkspaceID.String(),
		PrincipalID: row.PrincipalID.String(),
		Role:        domain.GrantRole(row.Role),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toDomainSpaceGrant(row sqlcgen.SpaceGrant) domain.SpaceGrant {
	return domain.SpaceGrant{
		WorkspaceID: row.WorkspaceID.String(),
		SpaceID:     row.SpaceID.String(),
		PrincipalID: row.PrincipalID.String(),
		Role:        domain.GrantRole(row.Role),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toDomainPageRestriction(row sqlcgen.PageRestriction) domain.PageRestriction {
	return domain.PageRestriction{
		WorkspaceID: row.WorkspaceID.String(),
		PageID:      row.PageID.String(),
		PrincipalID: row.PrincipalID.String(),
		Capability:  domain.Capability(row.Capability),
		Mode:        domain.RestrictionMode(row.Mode),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toDomainShareLink(row sqlcgen.ShareLink) domain.ShareLink {
	l := domain.ShareLink{
		ID:              row.ID.String(),
		WorkspaceID:     row.WorkspaceID.String(),
		PageID:          row.PageID.String(),
		PrincipalID:     row.PrincipalID.String(),
		Capability:      domain.Capability(row.Capability),
		TokenHash:       row.TokenHash,
		CreatedByUserID: uint64(row.CreatedByUserID),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.PasswordHash.Valid {
		h := row.PasswordHash.String
		l.PasswordHash = &h
	}
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		l.ExpiresAt = &t
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		l.RevokedAt = &t
	}
	return l
}

// restrictionFacts は「経路上に制限が 1 行も無い」を nil で返す
// （domain 側が nil を既定へのフォールバックと解釈する）。
func restrictionFacts(restricted, deniedAnywhere, hasAllowList, allowedAtNearest bool) *domain.RestrictionFacts {
	if !restricted {
		return nil
	}
	return &domain.RestrictionFacts{
		DeniedAnywhere:   deniedAnywhere,
		HasAllowList:     hasAllowList,
		AllowedAtNearest: allowedAtNearest,
	}
}

func (r *knowledgeBasePermissionRepository) EnsureUserPrincipal(ctx context.Context, workspaceID string, userID uint64) (*domain.Principal, error) {
	wsID, ok := kbParseID(workspaceID)
	if !ok {
		return nil, repository.ErrWorkspaceNotFound
	}
	// 先に引いてから作る。ユーザーの主体は (workspace_id, user_id) の部分 UNIQUE で 1 つに限られ、
	// 競合したら INSERT が一意制約で落ちるので、その場合はもう一度引き直して既存を返す。
	row, err := r.q.GetUserPrincipal(ctx, sqlcgen.GetUserPrincipalParams{
		WorkspaceID: wsID,
		UserID:      sql.NullInt64{Int64: int64(userID), Valid: true},
	})
	if err == nil {
		p := toDomainPrincipal(row)
		return &p, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	id, err := kbNewID()
	if err != nil {
		return nil, err
	}
	created, err := r.q.InsertPrincipal(ctx, sqlcgen.InsertPrincipalParams{
		ID:          id,
		WorkspaceID: wsID,
		Kind:        string(domain.PrincipalKindUser),
		UserID:      sql.NullInt64{Int64: int64(userID), Valid: true},
	})
	if err != nil {
		// 同時に同じユーザーを追加したときは一意制約で落ちる。既存を返して冪等にする。
		existing, getErr := r.q.GetUserPrincipal(ctx, sqlcgen.GetUserPrincipalParams{
			WorkspaceID: wsID,
			UserID:      sql.NullInt64{Int64: int64(userID), Valid: true},
		})
		if getErr != nil {
			return nil, err
		}
		p := toDomainPrincipal(existing)
		return &p, nil
	}
	p := toDomainPrincipal(created)
	return &p, nil
}

func (r *knowledgeBasePermissionRepository) EnsureSpaceEveryonePrincipal(ctx context.Context, workspaceID, spaceID string) (*domain.Principal, error) {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	if !ok || !ok2 {
		return nil, repository.ErrSpaceNotFound
	}
	row, err := r.q.GetSpaceEveryonePrincipal(ctx, sqlcgen.GetSpaceEveryonePrincipalParams{
		WorkspaceID: wsID,
		SpaceID:     uuid.NullUUID{UUID: spID, Valid: true},
	})
	if err == nil {
		p := toDomainPrincipal(row)
		return &p, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	id, err := kbNewID()
	if err != nil {
		return nil, err
	}
	created, err := r.q.InsertPrincipal(ctx, sqlcgen.InsertPrincipalParams{
		ID:          id,
		WorkspaceID: wsID,
		Kind:        string(domain.PrincipalKindSpaceAll),
		SpaceID:     uuid.NullUUID{UUID: spID, Valid: true},
	})
	if err != nil {
		existing, getErr := r.q.GetSpaceEveryonePrincipal(ctx, sqlcgen.GetSpaceEveryonePrincipalParams{
			WorkspaceID: wsID,
			SpaceID:     uuid.NullUUID{UUID: spID, Valid: true},
		})
		if getErr != nil {
			return nil, err
		}
		p := toDomainPrincipal(existing)
		return &p, nil
	}
	p := toDomainPrincipal(created)
	return &p, nil
}

func (r *knowledgeBasePermissionRepository) CreateGroupPrincipal(ctx context.Context, workspaceID, name string) (*domain.Principal, error) {
	wsID, ok := kbParseID(workspaceID)
	if !ok {
		return nil, repository.ErrWorkspaceNotFound
	}
	id, err := kbNewID()
	if err != nil {
		return nil, err
	}
	row, err := r.q.InsertPrincipal(ctx, sqlcgen.InsertPrincipalParams{
		ID:          id,
		WorkspaceID: wsID,
		Kind:        string(domain.PrincipalKindGroup),
		Name:        name,
	})
	if err != nil {
		return nil, err
	}
	p := toDomainPrincipal(row)
	return &p, nil
}

func (r *knowledgeBasePermissionRepository) FindPrincipal(ctx context.Context, workspaceID, principalID string) (*domain.Principal, error) {
	wsID, ok := kbParseID(workspaceID)
	prID, ok2 := kbParseID(principalID)
	if !ok || !ok2 {
		return nil, repository.ErrPrincipalNotFound
	}
	row, err := r.q.GetPrincipal(ctx, sqlcgen.GetPrincipalParams{WorkspaceID: wsID, ID: prID})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrPrincipalNotFound
	}
	if err != nil {
		return nil, err
	}
	p := toDomainPrincipal(row)
	return &p, nil
}

func (r *knowledgeBasePermissionRepository) FindUserPrincipal(ctx context.Context, workspaceID string, userID uint64) (*domain.Principal, error) {
	wsID, ok := kbParseID(workspaceID)
	if !ok {
		return nil, repository.ErrPrincipalNotFound
	}
	row, err := r.q.GetUserPrincipal(ctx, sqlcgen.GetUserPrincipalParams{
		WorkspaceID: wsID,
		UserID:      sql.NullInt64{Int64: int64(userID), Valid: true},
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrPrincipalNotFound
	}
	if err != nil {
		return nil, err
	}
	p := toDomainPrincipal(row)
	return &p, nil
}

func (r *knowledgeBasePermissionRepository) DeletePrincipal(ctx context.Context, workspaceID, principalID string) error {
	wsID, ok := kbParseID(workspaceID)
	prID, ok2 := kbParseID(principalID)
	if !ok || !ok2 {
		return repository.ErrPrincipalNotFound
	}
	n, err := r.q.DeletePrincipal(ctx, sqlcgen.DeletePrincipalParams{WorkspaceID: wsID, ID: prID})
	if err != nil {
		return err
	}
	if n == 0 {
		return repository.ErrPrincipalNotFound
	}
	return nil
}

func (r *knowledgeBasePermissionRepository) IsWorkspaceMember(ctx context.Context, workspaceID string, userID uint64) (bool, error) {
	wsID, ok := kbParseID(workspaceID)
	if !ok {
		return false, nil
	}
	return r.q.IsWorkspaceMember(ctx, sqlcgen.IsWorkspaceMemberParams{
		WorkspaceID: wsID,
		UserID:      sql.NullInt64{Int64: int64(userID), Valid: true},
	})
}

func (r *knowledgeBasePermissionRepository) AddGroupMember(ctx context.Context, workspaceID, groupPrincipalID, memberPrincipalID string) error {
	wsID, ok := kbParseID(workspaceID)
	gID, ok2 := kbParseID(groupPrincipalID)
	mID, ok3 := kbParseID(memberPrincipalID)
	if !ok || !ok2 || !ok3 {
		return repository.ErrPrincipalNotFound
	}
	return r.q.InsertPrincipalMember(ctx, sqlcgen.InsertPrincipalMemberParams{
		WorkspaceID:       wsID,
		GroupPrincipalID:  gID,
		MemberPrincipalID: mID,
	})
}

func (r *knowledgeBasePermissionRepository) RemoveGroupMember(ctx context.Context, workspaceID, groupPrincipalID, memberPrincipalID string) error {
	wsID, ok := kbParseID(workspaceID)
	gID, ok2 := kbParseID(groupPrincipalID)
	mID, ok3 := kbParseID(memberPrincipalID)
	if !ok || !ok2 || !ok3 {
		return nil // 存在し得ない ID = 既に外れている
	}
	_, err := r.q.DeletePrincipalMember(ctx, sqlcgen.DeletePrincipalMemberParams{
		WorkspaceID:       wsID,
		GroupPrincipalID:  gID,
		MemberPrincipalID: mID,
	})
	return err
}

func (r *knowledgeBasePermissionRepository) UpsertWorkspaceGrant(ctx context.Context, workspaceID, principalID string, role domain.GrantRole) (*domain.WorkspaceGrant, error) {
	wsID, ok := kbParseID(workspaceID)
	prID, ok2 := kbParseID(principalID)
	if !ok || !ok2 {
		return nil, repository.ErrPrincipalNotFound
	}
	row, err := r.q.UpsertWorkspaceGrant(ctx, sqlcgen.UpsertWorkspaceGrantParams{
		WorkspaceID: wsID,
		PrincipalID: prID,
		Role:        string(role),
	})
	if err != nil {
		return nil, err
	}
	g := toDomainWorkspaceGrant(row)
	return &g, nil
}

func (r *knowledgeBasePermissionRepository) DeleteWorkspaceGrant(ctx context.Context, workspaceID, principalID string) error {
	wsID, ok := kbParseID(workspaceID)
	prID, ok2 := kbParseID(principalID)
	if !ok || !ok2 {
		return nil
	}
	_, err := r.q.DeleteWorkspaceGrant(ctx, sqlcgen.DeleteWorkspaceGrantParams{
		WorkspaceID: wsID,
		PrincipalID: prID,
	})
	return err
}

func (r *knowledgeBasePermissionRepository) ListWorkspaceGrants(ctx context.Context, workspaceID string) ([]domain.WorkspaceGrant, error) {
	wsID, ok := kbParseID(workspaceID)
	if !ok {
		return []domain.WorkspaceGrant{}, nil
	}
	rows, err := r.q.ListWorkspaceGrants(ctx, wsID)
	if err != nil {
		return nil, err
	}
	grants := make([]domain.WorkspaceGrant, 0, len(rows))
	for _, row := range rows {
		grants = append(grants, toDomainWorkspaceGrant(row))
	}
	return grants, nil
}

func (r *knowledgeBasePermissionRepository) UpsertSpaceGrant(ctx context.Context, workspaceID, spaceID, principalID string, role domain.GrantRole) (*domain.SpaceGrant, error) {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	prID, ok3 := kbParseID(principalID)
	if !ok || !ok2 || !ok3 {
		return nil, repository.ErrPrincipalNotFound
	}
	row, err := r.q.UpsertSpaceGrant(ctx, sqlcgen.UpsertSpaceGrantParams{
		WorkspaceID: wsID,
		SpaceID:     spID,
		PrincipalID: prID,
		Role:        string(role),
	})
	if err != nil {
		return nil, err
	}
	g := toDomainSpaceGrant(row)
	return &g, nil
}

func (r *knowledgeBasePermissionRepository) DeleteSpaceGrant(ctx context.Context, workspaceID, spaceID, principalID string) error {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	prID, ok3 := kbParseID(principalID)
	if !ok || !ok2 || !ok3 {
		return nil
	}
	_, err := r.q.DeleteSpaceGrant(ctx, sqlcgen.DeleteSpaceGrantParams{
		WorkspaceID: wsID,
		SpaceID:     spID,
		PrincipalID: prID,
	})
	return err
}

func (r *knowledgeBasePermissionRepository) ListSpaceGrants(ctx context.Context, workspaceID, spaceID string) ([]domain.SpaceGrant, error) {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	if !ok || !ok2 {
		return []domain.SpaceGrant{}, nil
	}
	rows, err := r.q.ListSpaceGrants(ctx, sqlcgen.ListSpaceGrantsParams{WorkspaceID: wsID, SpaceID: spID})
	if err != nil {
		return nil, err
	}
	grants := make([]domain.SpaceGrant, 0, len(rows))
	for _, row := range rows {
		grants = append(grants, toDomainSpaceGrant(row))
	}
	return grants, nil
}

func (r *knowledgeBasePermissionRepository) UpsertPageRestriction(ctx context.Context, workspaceID, pageID, principalID string, capability domain.Capability, mode domain.RestrictionMode) (*domain.PageRestriction, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	prID, ok3 := kbParseID(principalID)
	if !ok || !ok2 || !ok3 {
		return nil, repository.ErrPageNotFound
	}

	// 例外 1 行と「その段が許可リスト制か」の印は必ず同じトランザクションで揃える。
	// 別々に書くと、片方だけ入った瞬間にページが開く / 閉じる中間状態ができる。
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }() // Commit 済みなら no-op
	qtx := r.q.WithTx(tx)

	prevMode, err := qtx.GetPageRestrictionMode(ctx, sqlcgen.GetPageRestrictionModeParams{
		WorkspaceID: wsID,
		PageID:      pgID,
		PrincipalID: prID,
		Capability:  string(capability),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	row, err := qtx.UpsertPageRestriction(ctx, sqlcgen.UpsertPageRestrictionParams{
		WorkspaceID: wsID,
		PageID:      pgID,
		PrincipalID: prID,
		Capability:  string(capability),
		Mode:        string(mode),
	})
	if err != nil {
		return nil, err
	}
	switch {
	case mode == domain.RestrictionModeAllow:
		if err := qtx.MarkPageAllowList(ctx, sqlcgen.MarkPageAllowListParams{
			WorkspaceID: wsID, PageID: pgID, Capability: string(capability),
		}); err != nil {
			return nil, err
		}
	case prevMode == string(domain.RestrictionModeAllow):
		// allow を deny へ書き換えた。その段の最後の allow だったなら印も畳む。
		if err := qtx.UnmarkPageAllowListIfEmpty(ctx, sqlcgen.UnmarkPageAllowListIfEmptyParams{
			WorkspaceID: wsID, PageID: pgID, Capability: string(capability),
		}); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	res := toDomainPageRestriction(row)
	return &res, nil
}

func (r *knowledgeBasePermissionRepository) DeletePageRestriction(ctx context.Context, workspaceID, pageID, principalID string, capability domain.Capability) error {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	prID, ok3 := kbParseID(principalID)
	if !ok || !ok2 || !ok3 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // Commit 済みなら no-op
	qtx := r.q.WithTx(tx)

	// 消したのが allow 行だったときだけ印を畳む。deny 行の解除で畳むと、
	// 無関係な 1 行の解除で限定公開が解けることになる。
	prevMode, err := qtx.GetPageRestrictionMode(ctx, sqlcgen.GetPageRestrictionModeParams{
		WorkspaceID: wsID,
		PageID:      pgID,
		PrincipalID: prID,
		Capability:  string(capability),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // 元から無い（冪等）
		}
		return err
	}
	if _, err := qtx.DeletePageRestriction(ctx, sqlcgen.DeletePageRestrictionParams{
		WorkspaceID: wsID,
		PageID:      pgID,
		PrincipalID: prID,
		Capability:  string(capability),
	}); err != nil {
		return err
	}
	if prevMode == string(domain.RestrictionModeAllow) {
		if err := qtx.UnmarkPageAllowListIfEmpty(ctx, sqlcgen.UnmarkPageAllowListIfEmptyParams{
			WorkspaceID: wsID, PageID: pgID, Capability: string(capability),
		}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *knowledgeBasePermissionRepository) ListPageAllowListCapabilities(ctx context.Context, workspaceID, pageID string) ([]domain.Capability, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return []domain.Capability{}, nil
	}
	rows, err := r.q.ListPageAllowLists(ctx, sqlcgen.ListPageAllowListsParams{WorkspaceID: wsID, PageID: pgID})
	if err != nil {
		return nil, err
	}
	list := make([]domain.Capability, 0, len(rows))
	for _, c := range rows {
		list = append(list, domain.Capability(c))
	}
	return list, nil
}

func (r *knowledgeBasePermissionRepository) ListPageRestrictions(ctx context.Context, workspaceID, pageID string) ([]domain.PageRestriction, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return []domain.PageRestriction{}, nil
	}
	rows, err := r.q.ListPageRestrictions(ctx, sqlcgen.ListPageRestrictionsParams{WorkspaceID: wsID, PageID: pgID})
	if err != nil {
		return nil, err
	}
	list := make([]domain.PageRestriction, 0, len(rows))
	for _, row := range rows {
		list = append(list, toDomainPageRestriction(row))
	}
	return list, nil
}

func (r *knowledgeBasePermissionRepository) CreateShareLink(ctx context.Context, in repository.ShareLinkWrite) (*domain.ShareLink, error) {
	wsID, ok := kbParseID(in.WorkspaceID)
	pgID, ok2 := kbParseID(in.PageID)
	if !ok || !ok2 {
		return nil, repository.ErrPageNotFound
	}
	principalID, err := kbNewID()
	if err != nil {
		return nil, err
	}
	linkID, err := kbNewID()
	if err != nil {
		return nil, err
	}

	// 主体とリンクは 1 対 1 で、どちらかだけが残ると失効も解決もできない行になる。
	// 同じトランザクションで両方作る。
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }() // Commit 済みなら no-op

	qtx := r.q.WithTx(tx)
	if _, err := qtx.InsertPrincipal(ctx, sqlcgen.InsertPrincipalParams{
		ID:          principalID,
		WorkspaceID: wsID,
		Kind:        string(domain.PrincipalKindShareLink),
		PageID:      uuid.NullUUID{UUID: pgID, Valid: true},
	}); err != nil {
		return nil, err
	}
	row, err := qtx.InsertShareLink(ctx, sqlcgen.InsertShareLinkParams{
		ID:              linkID,
		WorkspaceID:     wsID,
		PageID:          pgID,
		PrincipalID:     principalID,
		Capability:      string(in.Capability),
		TokenHash:       in.TokenHash,
		PasswordHash:    nullString(in.PasswordHash),
		ExpiresAt:       nullTime(in.ExpiresAt),
		CreatedByUserID: int64(in.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	link := toDomainShareLink(row)
	return &link, nil
}

func (r *knowledgeBasePermissionRepository) RevokeShareLink(ctx context.Context, workspaceID, shareLinkID string) error {
	wsID, ok := kbParseID(workspaceID)
	lnID, ok2 := kbParseID(shareLinkID)
	if !ok || !ok2 {
		return repository.ErrShareLinkNotFound
	}
	n, err := r.q.RevokeShareLink(ctx, sqlcgen.RevokeShareLinkParams{WorkspaceID: wsID, ID: lnID})
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	// 0 件は「既に失効済み」と「そもそも無い」の両方があり得る。区別して返す
	// （失効の再実行は冪等に成功させ、存在しない ID は not found にする）。
	if _, err := r.q.GetShareLink(ctx, sqlcgen.GetShareLinkParams{WorkspaceID: wsID, ID: lnID}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.ErrShareLinkNotFound
		}
		return err
	}
	return nil
}

func (r *knowledgeBasePermissionRepository) FindShareLinkByTokenHash(ctx context.Context, tokenHash []byte) (*domain.ShareLink, error) {
	row, err := r.q.GetShareLinkByTokenHash(ctx, tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrShareLinkNotFound
	}
	if err != nil {
		return nil, err
	}
	link := toDomainShareLink(row)
	return &link, nil
}

func (r *knowledgeBasePermissionRepository) ListPageShareLinks(ctx context.Context, workspaceID, pageID string) ([]domain.ShareLink, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return []domain.ShareLink{}, nil
	}
	rows, err := r.q.ListPageShareLinks(ctx, sqlcgen.ListPageShareLinksParams{WorkspaceID: wsID, PageID: pgID})
	if err != nil {
		return nil, err
	}
	links := make([]domain.ShareLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, toDomainShareLink(row))
	}
	return links, nil
}

func (r *knowledgeBasePermissionRepository) PagePermissionFactsForUser(ctx context.Context, workspaceID, pageID string, userID uint64) (*domain.PagePermissionFacts, error) {
	return r.pagePermissionFacts(ctx, workspaceID, pageID,
		sql.NullInt64{Int64: int64(userID), Valid: true}, uuid.NullUUID{})
}

func (r *knowledgeBasePermissionRepository) PagePermissionFactsForPrincipal(ctx context.Context, workspaceID, pageID, principalID string) (*domain.PagePermissionFacts, error) {
	prID, ok := kbParseID(principalID)
	if !ok {
		return nil, repository.ErrPrincipalNotFound
	}
	return r.pagePermissionFacts(ctx, workspaceID, pageID,
		sql.NullInt64{}, uuid.NullUUID{UUID: prID, Valid: true})
}

// pagePermissionFacts はユーザーとしての解決と共有リンクの来訪者としての解決の実体。
// どちらも同じ 1 本のクエリを通す（主体の種類で解決の道筋が分かれないようにするため）。
func (r *knowledgeBasePermissionRepository) pagePermissionFacts(
	ctx context.Context, workspaceID, pageID string,
	userID sql.NullInt64, principalID uuid.NullUUID,
) (*domain.PagePermissionFacts, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return nil, repository.ErrPageNotFound
	}
	row, err := r.q.ResolvePagePermissionFacts(ctx, sqlcgen.ResolvePagePermissionFactsParams{
		WorkspaceID: wsID,
		PageID:      pgID,
		UserID:      userID,
		PrincipalID: principalID,
	})
	if err != nil {
		return nil, err
	}
	if !row.PageExists {
		return nil, repository.ErrPageNotFound
	}
	return &domain.PagePermissionFacts{
		Member: row.IsMember,
		Role:   domain.GrantRoleByRank(int(row.GrantRank)),
		View:   restrictionFacts(row.ViewRestricted, row.ViewDeniedAnywhere, row.ViewHasAllowList, row.ViewAllowedAtNearest),
		Edit:   restrictionFacts(row.EditRestricted, row.EditDeniedAnywhere, row.EditHasAllowList, row.EditAllowedAtNearest),
	}, nil
}

func (r *knowledgeBasePermissionRepository) ListSpacePageViewFacts(ctx context.Context, workspaceID, spaceID string, userID uint64) ([]repository.PageWithViewFacts, error) {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	if !ok || !ok2 {
		return []repository.PageWithViewFacts{}, nil
	}
	rows, err := r.q.ListSpacePageViewFacts(ctx, sqlcgen.ListSpacePageViewFactsParams{
		WorkspaceID: wsID,
		SpaceID:     spID,
		UserID:      sql.NullInt64{Int64: int64(userID), Valid: true},
	})
	if err != nil {
		return nil, err
	}
	out := make([]repository.PageWithViewFacts, 0, len(rows))
	for _, row := range rows {
		page := toDomainPage(sqlcgen.Page{
			ID:              row.ID,
			WorkspaceID:     row.WorkspaceID,
			SpaceID:         row.SpaceID,
			ParentID:        row.ParentID,
			Position:        row.Position,
			Title:           row.Title,
			CreatedByUserID: row.CreatedByUserID,
			ArchivedAt:      row.ArchivedAt,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
		out = append(out, repository.PageWithViewFacts{
			Page: page,
			Facts: domain.PageViewFacts{
				Role: domain.GrantRoleByRank(int(row.GrantRank)),
				View: restrictionFacts(row.ViewRestricted, row.ViewDeniedAnywhere, row.ViewHasAllowList, row.ViewAllowedAtNearest),
			},
		})
	}
	return out, nil
}

func (r *knowledgeBasePermissionRepository) ListSubtreePagePermissionFacts(ctx context.Context, workspaceID, pageID string, userID uint64) ([]repository.PageWithPermissionFacts, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return []repository.PageWithPermissionFacts{}, nil
	}
	rows, err := r.q.ListSubtreePagePermissionFacts(ctx, sqlcgen.ListSubtreePagePermissionFactsParams{
		WorkspaceID: wsID,
		PageID:      pgID,
		UserID:      sql.NullInt64{Int64: int64(userID), Valid: true},
	})
	if err != nil {
		return nil, err
	}
	out := make([]repository.PageWithPermissionFacts, 0, len(rows))
	for _, row := range rows {
		out = append(out, repository.PageWithPermissionFacts{
			PageID: row.PageID.String(),
			Facts: domain.PagePermissionFacts{
				Member: row.IsMember,
				Role:   domain.GrantRoleByRank(int(row.GrantRank)),
				View:   restrictionFacts(row.ViewRestricted, row.ViewDeniedAnywhere, row.ViewHasAllowList, row.ViewAllowedAtNearest),
				Edit:   restrictionFacts(row.EditRestricted, row.EditDeniedAnywhere, row.EditHasAllowList, row.EditAllowedAtNearest),
			},
		})
	}
	return out, nil
}

// nullString は *string を sql.NullString へ変換する。
func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// nullTime は *time.Time を sql.NullTime へ変換する。
func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
