package persistence

import (
	"context"
	"database/sql"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// workspaceProvisioner は [repository.WorkspaceProvisioner] の実装。
// ナレッジ基盤は GORM を通さないので、sqlc 生成コード + 素の *sql.DB で書く。
type workspaceProvisioner struct {
	db *sql.DB
	q  *sqlcgen.Queries
}

// NewWorkspaceProvisioner はワークスペース作成の port を組み立てる。
func NewWorkspaceProvisioner(db *sql.DB) repository.WorkspaceProvisioner {
	return &workspaceProvisioner{db: db, q: sqlcgen.New(db)}
}

func (p *workspaceProvisioner) ProvisionWorkspace(
	ctx context.Context, in repository.WorkspaceProvisionInput,
) (*domain.Workspace, error) {
	wsID, err := kbNewID()
	if err != nil {
		return nil, err
	}
	principalID, err := kbNewID()
	if err != nil {
		return nil, err
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }() // Commit 済みなら no-op

	qtx := p.q.WithTx(tx)
	ws, err := qtx.InsertWorkspace(ctx, sqlcgen.InsertWorkspaceParams{
		ID:   wsID,
		Slug: in.Slug,
		Name: in.Name,
	})
	if err != nil {
		// slug はグローバルに一意（uq_workspaces_slug）。検査してから INSERT するまでの間に
		// 別の要求が同じ slug を取り得るので、一意制約を唯一の判定にする。
		if isUniqueViolation(err) {
			return nil, repository.ErrWorkspaceSlugTaken
		}
		return nil, err
	}
	// 作成者をこのワークスペースの主体にする。principals の行があること自体が所属なので、
	// この 1 行が入らないとワークスペースは誰も入れないまま残る（middleware が全経路で
	// 所属を確かめ、非メンバーには 404 を返すため、作成者にも見えなくなる）。
	if _, err := qtx.InsertPrincipal(ctx, sqlcgen.InsertPrincipalParams{
		ID:          principalID,
		WorkspaceID: wsID,
		Kind:        string(domain.PrincipalKindUser),
		UserID:      sql.NullInt64{Int64: int64(in.OwnerUserID), Valid: true},
	}); err != nil {
		return nil, err
	}
	// 所属だけでは何も見えない（役割が 1 つも無ければ実効権限は空）。作成者が自分の
	// ワークスペースを設定できるよう admin を張る。ここまでが 1 トランザクション。
	if _, err := qtx.UpsertWorkspaceGrant(ctx, sqlcgen.UpsertWorkspaceGrantParams{
		WorkspaceID: wsID,
		PrincipalID: principalID,
		Role:        string(domain.GrantRoleAdmin),
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	created := toDomainWorkspace(ws)
	return &created, nil
}
