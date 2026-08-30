package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// knowledgeBasePermissionRepository は [repository.KnowledgeBasePermissionRepository] の実装。
// ノートは GORM を通さない方針のため、クエリはすべて sqlc 生成コード + 素の *sql.DB で書く。
//
// # ユーザー ID の境界（uint64 → bigint）について
//
// domain のユーザー ID は uint64、DB の principals.user_id / share_links.created_by_user_id は
// bigint（＝ 符号付き 64bit・int64）。この境界を int64(userID) と素で書いてはいけない。
// Go の変換はビット列をそのまま読み替えるだけなので、userID が math.MaxInt64 を超えると
// 最上位ビットが符号ビットとして解釈され、値が負数へ巻き戻る
// （例: 1<<63 = 9223372036854775808 → -9223372036854775808）。
// 巻き戻った値は「たまたま別の行に一致し得る値」であって、元の入力とは無関係な行を指す。
// この API はユーザー ID を URL のパスから uint64 として受ける（handler の kbUserIDParam は
// strconv.ParseUint(..., 10, 64) なので 2^63 以上も通る）ため、利用者が指定した値が
// そのままここへ届く。変換は必ず toInt64ID（ids.go）を通し、範囲外は下の規則で扱う。
//
// 範囲外（> math.MaxInt64）の userID が意味するもの: users.id は bigint なので、
// その値を持つユーザーは**存在し得ない**。したがって扱いは読み書きで分かれる。
//
//   - 書き込み: エラーを返す。1 行も書けていないのに nil を返すと、呼び出し側が
//     「作成・更新できた」と誤認する。
//   - 読み取り（権限の判定・一覧）: 「該当なし」を返す。クエリを投げても 0 行になる入力なので、
//     0 行のときとまったく同じ値を返すのが正しい。**必ず拒否側（deny）に倒す**こと。
//     ここで「許可」側の値（役割あり・メンバーである・閲覧できる）を返すと、
//     存在しないユーザー ID を名乗るだけで権限が湧く＝権限昇格になる。
type knowledgeBasePermissionRepository struct {
	db *sql.DB
	q  *sqlcgen.Queries
}

// NewKnowledgeBasePermissionRepository はノートの権限 repository を組み立てる。
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
	// bigint に収まらない userID は users のどの行の id にもなり得ない ＝ そんなユーザーは居ない。
	// これは下の FK 違反（実在しないユーザー ID を渡された場合）とまったく同じ状況なので、
	// 同じ ErrUserNotFound を返して呼び出し側の分岐を増やさない。
	// nil を返してはいけない — 主体を 1 行も作れていないのに「作れた」と誤認される。
	uid, uok := toInt64ID(userID)
	if !uok {
		return nil, repository.ErrUserNotFound
	}
	// 先に引いてから作る。ユーザーの主体は (workspace_id, user_id) の部分 UNIQUE で 1 つに限られ、
	// 競合したら INSERT が一意制約で落ちるので、その場合はもう一度引き直して既存を返す。
	row, err := r.q.GetUserPrincipal(ctx, sqlcgen.GetUserPrincipalParams{
		WorkspaceID: wsID,
		UserID:      sql.NullInt64{Int64: uid, Valid: true},
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
		UserID:      sql.NullInt64{Int64: uid, Valid: true},
	})
	if err != nil {
		// 実在しないユーザー ID を渡された場合は users への FK で落ちる。入力の誤りなので
		// 制約違反のまま上へ流さず、「そのユーザーは居ない」として返す（500 にしない）。
		if isForeignKeyViolation(err) {
			return nil, repository.ErrUserNotFound
		}
		// 同時に同じユーザーを追加したときは一意制約で落ちる。既存を返して冪等にする。
		existing, getErr := r.q.GetUserPrincipal(ctx, sqlcgen.GetUserPrincipalParams{
			WorkspaceID: wsID,
			UserID:      sql.NullInt64{Int64: uid, Valid: true},
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
		// 名前はワークスペース内で一意（uq_principals_group_name）。検査してから INSERT する
		// までの間に別の要求が同じ名前を取り得るので、一意制約を唯一の判定にする
		// （ワークスペース作成の slug と同じ考え方）。
		if isUniqueViolation(err) {
			return nil, repository.ErrPrincipalGroupNameTaken
		}
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
	// bigint に収まらない userID はどの principals の行にも一致しない。クエリを投げれば
	// 0 行 ＝ sql.ErrNoRows になる入力なので、その分岐（下の ErrPrincipalNotFound）と
	// 同じ値を返す。呼び出し側から見た意味は「非メンバー」で、拒否側に倒れている。
	uid, uok := toInt64ID(userID)
	if !uok {
		return nil, repository.ErrPrincipalNotFound
	}
	row, err := r.q.GetUserPrincipal(ctx, sqlcgen.GetUserPrincipalParams{
		WorkspaceID: wsID,
		UserID:      sql.NullInt64{Int64: uid, Valid: true},
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

// DeletePrincipal は主体を 1 件消す。
//
// 主体を消すと、その主体に張られていた grant も FK の CASCADE で消える。つまりこれは
// 「ワークスペースの admin を 1 人減らし得る操作」でもあるので、grant の取り消しと
// まったく同じ検査を、同じトランザクションの中で通す（withLastAdminGuard の doc を参照）。
func (r *knowledgeBasePermissionRepository) DeletePrincipal(ctx context.Context, workspaceID, principalID string) error {
	wsID, ok := kbParseID(workspaceID)
	prID, ok2 := kbParseID(principalID)
	if !ok || !ok2 {
		return repository.ErrPrincipalNotFound
	}
	return r.withLastAdminGuard(ctx, wsID, prID, func(qtx *sqlcgen.Queries) error {
		n, err := qtx.DeletePrincipal(ctx, sqlcgen.DeletePrincipalParams{WorkspaceID: wsID, ID: prID})
		if err != nil {
			return err
		}
		if n == 0 {
			return repository.ErrPrincipalNotFound
		}
		return nil
	})
}

func (r *knowledgeBasePermissionRepository) IsWorkspaceMember(ctx context.Context, workspaceID string, userID uint64) (bool, error) {
	wsID, ok := kbParseID(workspaceID)
	if !ok {
		return false, nil
	}
	// bigint に収まらない userID は principals のどの行にも一致しない。SQL は EXISTS を
	// 返すので 0 行 ＝ false。同じ false を返す（上の ID 不正の分岐と同じ値）。
	// true 側へ倒すと、存在しないユーザー ID を名乗るだけでワークスペースの中身が
	// 見える口が開く（所属は principals の行が唯一の表現なので、ここが所属判定の本体）。
	uid, uok := toInt64ID(userID)
	if !uok {
		return false, nil
	}
	return r.q.IsWorkspaceMember(ctx, sqlcgen.IsWorkspaceMemberParams{
		WorkspaceID: wsID,
		UserID:      sql.NullInt64{Int64: uid, Valid: true},
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

// RemoveGroupMember はグループから 1 人外す。
//
// ここは 0 行削除を成功のままにする（not-found にしない）。求められているのは
// 「その主体がこのグループに載っていない状態」であって、今回の呼び出しで実際に 1 行
// 消えたかどうかではない。元から載っていない主体を外す要求は、その事後条件を既に
// 満たしているので冪等に成功で良い。Update 系と違い、成功を返しても「保存したはずの値が
// 保存されていない」という取り違えは起きない。
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

// UpsertWorkspaceGrant はワークスペース全体の既定の役割を 1 行に揃える。
//
// admin を**与える**向きは admin を減らさないので検査も行ロックも要らない（素で書く）。
// admin から他の役割へ**落とす**向きは、行が消えないだけで「admin を外す」操作そのものなので、
// 取り消し・メンバー削除とまったく同じ検査を同じトランザクションで通す。
// GrantWorkspaceRoleIfAbsent は既定の役割を無いときだけ与える（詳細は port のコメント）。
func (r *knowledgeBasePermissionRepository) GrantWorkspaceRoleIfAbsent(ctx context.Context, workspaceID, principalID string, role domain.GrantRole) error {
	wsID, ok := kbParseID(workspaceID)
	prID, ok2 := kbParseID(principalID)
	if !ok || !ok2 {
		return repository.ErrPrincipalNotFound
	}
	return r.q.InsertWorkspaceGrantIfAbsent(ctx, sqlcgen.InsertWorkspaceGrantIfAbsentParams{
		WorkspaceID: wsID,
		PrincipalID: prID,
		Role:        string(role),
	})
}

func (r *knowledgeBasePermissionRepository) UpsertWorkspaceGrant(ctx context.Context, workspaceID, principalID string, role domain.GrantRole) (*domain.WorkspaceGrant, error) {
	wsID, ok := kbParseID(workspaceID)
	prID, ok2 := kbParseID(principalID)
	if !ok || !ok2 {
		return nil, repository.ErrPrincipalNotFound
	}
	params := sqlcgen.UpsertWorkspaceGrantParams{
		WorkspaceID: wsID,
		PrincipalID: prID,
		Role:        string(role),
	}
	if role == domain.GrantRoleAdmin {
		row, err := r.q.UpsertWorkspaceGrant(ctx, params)
		if err != nil {
			return nil, err
		}
		g := toDomainWorkspaceGrant(row)
		return &g, nil
	}
	var g domain.WorkspaceGrant
	if err := r.withLastAdminGuard(ctx, wsID, prID, func(qtx *sqlcgen.Queries) error {
		row, err := qtx.UpsertWorkspaceGrant(ctx, params)
		if err != nil {
			return err
		}
		g = toDomainWorkspaceGrant(row)
		return nil
	}); err != nil {
		return nil, err
	}
	return &g, nil
}

// DeleteWorkspaceGrant はワークスペース権限を 1 件取り消す。
// RemoveGroupMember と同じ理由で 0 行削除は成功のまま（「権限が無い状態」が事後条件で、
// 元から無ければ既に満たされている）。取り消しの再実行を 404 にしない。
//
// ただし「ユーザーの admin が 0 人になる取り消し」だけは冪等では済まないので、
// withLastAdminGuard を通して同じトランザクションの中で断る。
func (r *knowledgeBasePermissionRepository) DeleteWorkspaceGrant(ctx context.Context, workspaceID, principalID string) error {
	wsID, ok := kbParseID(workspaceID)
	prID, ok2 := kbParseID(principalID)
	if !ok || !ok2 {
		return nil
	}
	return r.withLastAdminGuard(ctx, wsID, prID, func(qtx *sqlcgen.Queries) error {
		_, err := qtx.DeleteWorkspaceGrant(ctx, sqlcgen.DeleteWorkspaceGrantParams{
			WorkspaceID: wsID,
			PrincipalID: prID,
		})
		return err
	})
}

// withLastAdminGuard は「この主体から admin を外す」操作を 1 トランザクションで包み、
// ユーザーの admin が 0 人になる場合は repository.ErrLastWorkspaceAdmin を返して
// mutate を一度も呼ばない。
//
// # なぜここまでするのか
//
// ワークスペースの admin が 0 人になると、そのワークスペースの権限を変えられる人は
// API のどこにも居なくなる。ノートは「アプリの super_admin なら通る」という
// 抜け道を意図的に持たないため、**元 admin を含めて誰も復旧できない**（DB を直接
// 触るしかない）。逆に「最後の 1 人は自分を外せない」で詰まる場面は、先に別の誰かへ
// admin を渡せば必ず解ける。取り返しがつかない側だけを禁じる。
//
// # なぜ検査を手前（usecase）に置くだけでは足りないのか
//
// 手前の CanRemoveWorkspaceAdminUseCase は読み取りだけで、そのあとの書き換えは別の
// トランザクションになる。admin 2 人をほぼ同時に外す 2 本の要求は、両方ともその検査を
// 通り抜けて両方とも成功し得る（実測: 2 本同時に流すと 60 回中 59 回 admin が 0 人になった）。
//
// # なぜ「検査を DELETE の EXISTS へ畳んで単一文にする」だけでは足りないのか
//
// PostgreSQL の既定は READ COMMITTED で、EXISTS の副問い合わせは行をロックしない。
// 2 つのトランザクションが互いの admin 行を「まだ在る」と見たまま、それぞれ相手を
// 消せてしまう。実測でもこの形は明示トランザクションを重ねると admin 0 人を再現した。
// LockWorkspaceAdminGrantsForRemoval が FOR UPDATE で admin 行をロックし、そのロックを
// 書き換えと同じトランザクションが握り続けることで初めて塞がる。
func (r *knowledgeBasePermissionRepository) withLastAdminGuard(
	ctx context.Context,
	workspaceID, principalID uuid.UUID,
	mutate func(qtx *sqlcgen.Queries) error,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // Commit 済みなら no-op
	qtx := r.q.WithTx(tx)

	guard, err := qtx.LockWorkspaceAdminGrantsForRemoval(ctx, sqlcgen.LockWorkspaceAdminGrantsForRemovalParams{
		WorkspaceID: workspaceID,
		PrincipalID: principalID,
	})
	if err != nil {
		return err
	}
	// 元から admin ではない相手なら、この操作で admin は 1 人も減らない。
	if guard.TargetIsAdmin && !guard.OtherUserAdminRemains {
		return repository.ErrLastWorkspaceAdmin
	}
	if err := mutate(qtx); err != nil {
		return err
	}
	return tx.Commit()
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

// DeleteSpaceGrant はスペース権限を 1 件取り消す。
// DeleteWorkspaceGrant と同じ理由で 0 行削除は成功のまま（取り消しは冪等）。
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

// DeletePageRestriction はページの例外（allow / deny 1 行）を取り消す。
//
// ここも 0 行削除は成功のまま。そもそも直前の GetPageRestrictionMode が sql.ErrNoRows なら
// 「元から無い（冪等）」として早期 return しており、DELETE まで来て 0 行になるのは
// その 2 文のあいだに他の管理操作が同じ行を消した競合だけ。求められている事後条件
// （その例外が無い状態）はどちらにせよ満たされているので、取り消しは冪等に成功させる。
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
	// share_links.created_by_user_id は bigint で、users への FK も張っている。
	// 範囲外の発行者 ID では 1 行も書けないので、書き込みに入る前にエラーで止める
	// （nil を返すと呼び出し側がリンクを発行できたと誤認する）。
	createdBy, cok := toInt64ID(in.CreatedByUserID)
	if !cok {
		return nil, outOfRangeIDError("created_by_user_id", in.CreatedByUserID)
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
		CreatedByUserID: createdBy,
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
	// bigint に収まらない userID は principals のどの行にも一致しない。クエリ側で言えば
	// me / mine の CTE が空になる状態で、そのとき SQL が返す自分についての事実は
	// is_member=false / grant_rank=0 / denied_anywhere=false / allowed_at_nearest=false。
	// つまり Member=false / Role=nil で、非メンバーが得るものと同じ。
	//
	// ゼロ値で返すので View / Edit は nil（経路に制限が無い）になる。ページ側に制限が
	// 張られていれば実際のクエリは非 nil を返すが、そこは違っていて構わない。
	// domain.ResolvePagePermission に通したときの答えがどちらも同じだからで、
	// 既定が roleAllows(nil) = false である以上、resolveCapability は
	// 例外が nil でも（deny / 許可リストで）非 nil でも false を返す。
	// CanEdit は CanView を含むのでさらに閉じる。拒否側へ倒れることが確実に決まる。
	//
	// ページの実在は確かめない（確かめる術がクエリしか無く、その入力がここでは作れない）。
	// 存在しないページを名指しされたときの応答が 404 ではなく 403 になるが、
	// どちらも拒否で、ページの実在を漏らす向きでもない。
	uid, uok := toInt64ID(userID)
	if !uok {
		return &domain.PagePermissionFacts{}, nil
	}
	return r.pagePermissionFacts(ctx, workspaceID, pageID,
		sql.NullInt64{Int64: uid, Valid: true}, uuid.NullUUID{})
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

func (r *knowledgeBasePermissionRepository) ListSpacePageViewFacts(ctx context.Context, workspaceID, spaceID string, userID uint64, archived bool) ([]repository.PageWithViewFacts, error) {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	if !ok || !ok2 {
		return []repository.PageWithViewFacts{}, nil
	}
	// bigint に収まらない userID はどの主体にも一致しない。この一覧は「見えるページ」を
	// 組み立てる材料なので、0 件（空スライス）が該当なしの答え。上の ID 不正の分岐と同じ値。
	//
	// ここで「行を返しつつ Role だけ nil」のような中途半端な値を作らないのは、
	// 呼び出し側（ListViewablePagesUseCase）が domain.ResolvePageView でふるいに掛ける前に
	// ページの中身（タイトル等）を受け取ってしまうため。空で返せば 1 枚も漏れない。
	uid, uok := toInt64ID(userID)
	if !uok {
		return []repository.PageWithViewFacts{}, nil
	}
	rows, err := r.q.ListSpacePageViewFacts(ctx, sqlcgen.ListSpacePageViewFactsParams{
		WorkspaceID: wsID,
		SpaceID:     spID,
		UserID:      sql.NullInt64{Int64: uid, Valid: true},
		Archived:    archived,
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
			ParentArchived: row.ParentArchived,
		})
	}
	return out, nil
}

// kbEscapeLike は LIKE / ILIKE の特殊文字（% _ \）をエスケープする。
// LIKE の既定のエスケープ文字はバックスラッシュなので ESCAPE 句は書かない。
// 生のまま渡すと「%」1 文字で全件一致になり、候補の天井まで無関係な行が埋まる。
func kbEscapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func (r *knowledgeBasePermissionRepository) SearchWorkspacePageViewFacts(ctx context.Context, workspaceID string, userID uint64, query string) ([]repository.PageWithViewFacts, error) {
	wsID, ok := kbParseID(workspaceID)
	if !ok {
		return []repository.PageWithViewFacts{}, nil
	}
	// bigint に収まらない userID はどの主体にも一致しない（ListSpacePageViewFacts と同じ扱い）。
	uid, uok := toInt64ID(userID)
	if !uok {
		return []repository.PageWithViewFacts{}, nil
	}
	rows, err := r.q.SearchWorkspacePageViewFacts(ctx, sqlcgen.SearchWorkspacePageViewFactsParams{
		WorkspaceID: wsID,
		UserID:      sql.NullInt64{Int64: uid, Valid: true},
		Needle:      kbEscapeLike(query),
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
			// ParentArchived は集めない（検索は現役だけが対象）。既定の false のまま。
		})
	}
	return out, nil
}

func (r *knowledgeBasePermissionRepository) ListWorkspacePageViewFactsByIDs(
	ctx context.Context, workspaceID string, userID uint64, pageIDs []string,
) ([]repository.PageWithViewFacts, error) {
	wsID, ok := kbParseID(workspaceID)
	if !ok {
		return []repository.PageWithViewFacts{}, nil
	}
	uid, uok := toInt64ID(userID)
	if !uok {
		return []repository.PageWithViewFacts{}, nil
	}
	// UUID として読めない ID はここで落とす。SQL 側の ::uuid が失敗すると
	// クエリ全体が落ち、壊れた参照 1 つでページの読み出しが死ぬため。
	valid := make([]string, 0, len(pageIDs))
	for _, id := range pageIDs {
		if pid, pok := kbParseID(id); pok {
			valid = append(valid, pid.String())
		}
	}
	if len(valid) == 0 {
		return []repository.PageWithViewFacts{}, nil
	}
	encoded, err := json.Marshal(valid)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListWorkspacePageViewFactsByIDs(ctx, sqlcgen.ListWorkspacePageViewFactsByIDsParams{
		WorkspaceID: wsID,
		UserID:      sql.NullInt64{Int64: uid, Valid: true},
		PageIds:     encoded,
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
			// ParentArchived は集めない（検索と同じく現役だけが対象）。既定の false のまま。
		})
	}
	return out, nil
}

func (r *knowledgeBasePermissionRepository) ListMemberWorkspaces(ctx context.Context, userID uint64) ([]domain.MemberWorkspace, error) {
	// ここは唯一テナントを跨いで読むメソッドで、絞り込みは user_id だけが行う。
	// つまり userID の取り違えがそのままテナント境界の越境になるので、
	// 巻き戻った値で問い合わせることは絶対に避ける。
	//
	// bigint に収まらない userID は principals のどの行にも一致しない ＝ 所属ゼロ。
	// クエリが 0 行を返したときと同じ空スライスを返す（下のループが作る値と同じ）。
	uid, uok := toInt64ID(userID)
	if !uok {
		return []domain.MemberWorkspace{}, nil
	}
	rows, err := r.q.ListMemberWorkspaces(ctx, sql.NullInt64{Int64: uid, Valid: true})
	if err != nil {
		return nil, err
	}
	out := make([]domain.MemberWorkspace, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.MemberWorkspace{
			Workspace: toDomainWorkspace(sqlcgen.Workspace{
				ID:                  row.ID,
				Slug:                row.Slug,
				Name:                row.Name,
				IsActive:            row.IsActive,
				PersonalOwnerUserID: row.PersonalOwnerUserID,
				CreatedAt:           row.CreatedAt,
				UpdatedAt:           row.UpdatedAt,
			}),
			CanManage: row.IsAdmin,
		})
	}
	return out, nil
}

func (r *knowledgeBasePermissionRepository) PageSpaceScopeFactsForUser(
	ctx context.Context, workspaceID, pageID string, userID uint64,
) (*repository.PageScopeFacts, error) {
	empty := &repository.PageScopeFacts{Facts: domain.ScopeFacts{Roles: toGrantRoles(nil)}}

	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		// UUID ですらない文字列はどのページにも一致しない。**エラーではなく空を返す**
		// （見つからなかったときと同じ扱い）。ここだけ別のエラーにすると、呼び出し側が
		// 撃ち分けなくても応答の作られ方が変わる余地が残る。
		return empty, nil
	}
	// bigint に収まらない userID はどの主体にも一致しない ＝ 役割を 1 つも持たない。
	// これも空で返す（0 行が返ったときと同じ）。
	uid, uok := toInt64ID(userID)
	if !uok {
		return empty, nil
	}

	// **問い合わせは 1 回だけ。** ページの実在確認とスペースの実在確認と役割の収集を
	// 別々に行うと、落ちる段によって DB の往復が変わり、同じ 404 でも返るまでの時間から
	// ページ ID の実在が読める（ListPageSpaceScopeGrantRoles の doc に経緯がある）。
	rows, err := r.q.ListPageSpaceScopeGrantRoles(ctx, sqlcgen.ListPageSpaceScopeGrantRolesParams{
		WorkspaceID: wsID,
		PageID:      pgID,
		UserID:      sql.NullInt64{Int64: uid, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		// ページが無い / そのスペースに役割が 1 つも届いていない のどちらか。
		// **区別しない。** どちらも呼び出し側では拒否に落ちる。
		return empty, nil
	}

	roles := make([]string, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, row.Role)
	}
	return &repository.PageScopeFacts{
		// 全行が同じページから引いた space_id なので、どれを採っても同じ。
		SpaceID: rows[0].SpaceID.String(),
		Facts:   domain.ScopeFacts{Roles: toGrantRoles(roles)},
	}, nil
}

func (r *knowledgeBasePermissionRepository) SpacePermissionFactsForUser(
	ctx context.Context, workspaceID, spaceID string, userID uint64,
) (*domain.ScopeFacts, error) {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	if !ok || !ok2 {
		return nil, repository.ErrSpaceNotFound
	}
	// 役割を集める前にスペースの実在（と同じワークスペースに属すること）を確かめる。
	// workspace_grants は配下の全スペースに届くので、確かめずに役割だけを集めると
	// 存在しないスペースや他テナントのスペースに対しても「自分のワークスペースでの役割」が
	// そのまま返り、口そのものが緩い側へ倒れる。
	if _, err := r.q.GetSpace(ctx, sqlcgen.GetSpaceParams{WorkspaceID: wsID, ID: spID}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrSpaceNotFound
		}
		return nil, err
	}
	// bigint に収まらない userID はどの主体にも一致しない ＝ 役割を 1 つも持たない。
	// 役割の問い合わせが 0 行を返したときと同じ値（空の Roles）を返す。
	// domain.ResolveScopePermission は StrongestGrantRole([]) = nil → roleAllows(nil) = false
	// なので CanView / CanEdit / CanManage がすべて false になり、拒否側へ倒れる。
	//
	// この検査を「スペースの実在を確かめたあと」に置いているのは、ErrSpaceNotFound を
	// 返す条件がスペース側の事情だけで決まるようにするため。userID が範囲外かどうかで
	// 存在しないスペースへの応答が変わると、同じ URL の答えが呼び出し方で揺れる。
	uid, uok := toInt64ID(userID)
	if !uok {
		return &domain.ScopeFacts{Roles: toGrantRoles(nil)}, nil
	}
	roles, err := r.q.ListSpaceScopeGrantRoles(ctx, sqlcgen.ListSpaceScopeGrantRolesParams{
		WorkspaceID: wsID,
		UserID:      sql.NullInt64{Int64: uid, Valid: true},
		SpaceID:     uuid.NullUUID{UUID: spID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &domain.ScopeFacts{Roles: toGrantRoles(roles)}, nil
}

func (r *knowledgeBasePermissionRepository) WorkspacePermissionFactsForUser(
	ctx context.Context, workspaceID string, userID uint64,
) (*domain.ScopeFacts, error) {
	wsID, ok := kbParseID(workspaceID)
	if !ok {
		return nil, repository.ErrWorkspaceNotFound
	}
	// bigint に収まらない userID はどの主体にも一致しない ＝ 役割を 1 つも持たない。
	// SpacePermissionFactsForUser と同じく、0 行のときと同じ空の Roles を返す。
	// これが admin を含む役割へ倒れると、ワークスペースの権限設定そのものを
	// 書き換えられる（CanManage が true になる）ので、必ず空側に倒す。
	uid, uok := toInt64ID(userID)
	if !uok {
		return &domain.ScopeFacts{Roles: toGrantRoles(nil)}, nil
	}
	roles, err := r.q.ListWorkspaceScopeGrantRoles(ctx, sqlcgen.ListWorkspaceScopeGrantRolesParams{
		WorkspaceID: wsID,
		UserID:      sql.NullInt64{Int64: uid, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &domain.ScopeFacts{Roles: toGrantRoles(roles)}, nil
}

func (r *knowledgeBasePermissionRepository) ListWorkspaceSpaceScopeFacts(
	ctx context.Context, workspaceID string, userID uint64,
) ([]repository.SpaceWithScopeFacts, error) {
	wsID, ok := kbParseID(workspaceID)
	if !ok {
		// 解釈できないワークスペース ID は 1 行にも一致しない。0 件と同じ空スライスを返す
		// （nil を返さないのは JSON で null にしないため。cmd/slicelint が検査している）。
		return []repository.SpaceWithScopeFacts{}, nil
	}
	// bigint に収まらない userID はどの主体にも一致しない ＝ 役割を 1 つも持たない。
	// ここは一覧の材料なので、空スライス（＝ 1 件も見えない）が拒否側の答えになる。
	//
	// 「全スペースを Roles 空で返す」ではなく空にするのは、見えないスペースの key / name を
	// 呼び出し側へ渡さないため（ListSubtreePagePermissionFacts と同じ判断）。
	uid, uok := toInt64ID(userID)
	if !uok {
		return []repository.SpaceWithScopeFacts{}, nil
	}
	rows, err := r.q.ListWorkspaceSpaceScopeFacts(ctx, sqlcgen.ListWorkspaceSpaceScopeFactsParams{
		WorkspaceID: wsID,
		UserID:      sql.NullInt64{Int64: uid, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	// クエリは (スペース × 届いている役割) の直積を返すので、スペース単位に畳み直す。
	// 役割が 1 つも無いスペースは LEFT JOIN の右が NULL の 1 行として来る（＝ Roles は空のまま）。
	//
	// どれを採るかの規則はここでは決めない（domain.StrongestGrantRole の仕事）。
	// ここでやるのは行を集めることだけ。
	out := make([]repository.SpaceWithScopeFacts, 0, len(rows))
	indexBySpace := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		i, seen := indexBySpace[row.ID]
		if !seen {
			i = len(out)
			indexBySpace[row.ID] = i
			out = append(out, repository.SpaceWithScopeFacts{
				Space: toDomainSpace(sqlcgen.Space{
					ID:          row.ID,
					WorkspaceID: row.WorkspaceID,
					Key:         row.Key,
					Name:        row.Name,
					Visibility:  row.Visibility,
					CreatedAt:   row.CreatedAt,
					UpdatedAt:   row.UpdatedAt,
				}),
				Facts: domain.ScopeFacts{Roles: []domain.GrantRole{}},
			})
		}
		if row.Role.Valid {
			out[i].Facts.Roles = append(out[i].Facts.Roles, domain.GrantRole(row.Role.String))
		}
	}
	return out, nil
}

// toGrantRoles は SQL が返した役割の文字列を domain の型へ移すだけの変換。
// どれを採るかの規則（最も強いものを採る）はここでは決めず、domain.StrongestGrantRole に任せる。
func toGrantRoles(rows []string) []domain.GrantRole {
	roles := make([]domain.GrantRole, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, domain.GrantRole(row))
	}
	return roles
}

func (r *knowledgeBasePermissionRepository) ListSubtreePagePermissionFacts(ctx context.Context, workspaceID, pageID string, userID uint64) ([]repository.PageWithPermissionFacts, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return []repository.PageWithPermissionFacts{}, nil
	}
	// bigint に収まらない userID はどの主体にも一致しない。上の ID 不正の分岐と同じ
	// 空スライスを返す。呼び出し側（CanEditPageSubtreeUseCase）は 0 行を
	// 「許可には倒さない」と決めて false を返すので、ここも拒否側で一致する。
	//
	// 空ではなく「全ページを Role=nil で返す」ようにしてはいけない。そちらでも判定自体は
	// false になるが、見えないページの ID を呼び出し側へ渡すことになる。
	uid, uok := toInt64ID(userID)
	if !uok {
		return []repository.PageWithPermissionFacts{}, nil
	}
	rows, err := r.q.ListSubtreePagePermissionFacts(ctx, sqlcgen.ListSubtreePagePermissionFactsParams{
		WorkspaceID: wsID,
		PageID:      pgID,
		UserID:      sql.NullInt64{Int64: uid, Valid: true},
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
