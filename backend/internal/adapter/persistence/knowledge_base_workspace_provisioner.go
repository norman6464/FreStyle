package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/norman6464/frestyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// workspaceProvisioner は [repository.WorkspaceProvisioner] の実装。
// ナレッジは GORM を通さないので、sqlc 生成コード + 素の *sql.DB で書く。
type workspaceProvisioner struct {
	baseRepository
}

func NewWorkspaceProvisioner(db *sql.DB) repository.WorkspaceProvisioner {
	return &workspaceProvisioner{baseRepository{db: db}}
}

// runInTx は 1 つのトランザクションを開き、その中でだけ有効な Queries を fn に渡す。
// ctx に既に外側の DoInTx が開いたトランザクションがあれば、新規に開始せずそれへ相乗りする
// （二重に BeginTx するとデッドロックの原因になる。commit/rollback は外側だけが持つ）。
func (p *workspaceProvisioner) runInTx(ctx context.Context, fn func(qtx *sqlcgen.Queries) error) error {
	if tx, ok := getTx(ctx); ok {
		return fn(sqlcgen.New(tx))
	}
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // Commit 済みなら no-op
	if err := fn(sqlcgen.New(tx)); err != nil {
		return err
	}
	return tx.Commit()
}

func (p *workspaceProvisioner) ProvisionWorkspace(
	ctx context.Context, in repository.WorkspaceProvisionInput,
) (*domain.Workspace, error) {
	// principals.user_id は bigint。素の int64(in.OwnerUserID) は math.MaxInt64 超で負数へ
	// 巻き戻り、無関係な user_id で主体を作り得るため、範囲外なら（users にも存在し得ない）
	// トランザクション前にエラーで止める（nil を返すと作成できたと誤認される）。
	ownerID, ok := toInt64ID(in.OwnerUserID)
	if !ok {
		return nil, outOfRangeIDError("user_id", in.OwnerUserID)
	}
	var personalOwnerID sql.NullInt64
	if in.PersonalOwnerUserID != nil {
		pid, ok := toInt64ID(*in.PersonalOwnerUserID)
		if !ok {
			return nil, outOfRangeIDError("personal_owner_user_id", *in.PersonalOwnerUserID)
		}
		personalOwnerID = sql.NullInt64{Int64: pid, Valid: true}
	}
	wsID, err := kbNewID()
	if err != nil {
		return nil, err
	}
	principalID, err := kbNewID()
	if err != nil {
		return nil, err
	}

	var created domain.Workspace
	err = p.runInTx(ctx, func(qtx *sqlcgen.Queries) error {
		ws, err := qtx.InsertWorkspace(ctx, sqlcgen.InsertWorkspaceParams{
			ID:                  wsID,
			Slug:                in.Slug,
			Name:                in.Name,
			PersonalOwnerUserID: personalOwnerID,
		})
		if err != nil {
			// slug（uq_workspaces_slug）と personal_owner_user_id（uq_workspaces_personal_owner）
			// は共に TOCTOU で競合し得るので一意制約を唯一の判定にする。どちらが競合したかで
			// 意味が違う（slug は使用済み、personal_owner は「もう作られていた」で失敗ではない）
			// ため制約名で振り分ける。
			if name, ok := uniqueViolationConstraint(err); ok {
				switch name {
				case "uq_workspaces_personal_owner":
					return repository.ErrPersonalWorkspaceAlreadyExists
				default:
					return repository.ErrWorkspaceSlugTaken
				}
			}
			return err
		}
		// 作成者を workspace_members の active な所属として記録する。自作ワークスペースは
		// 招待の手順を踏む理由が無いので invited を経由せず直接 active にする。
		if err := qtx.InsertActiveWorkspaceMember(ctx, sqlcgen.InsertActiveWorkspaceMemberParams{
			WorkspaceID: wsID,
			UserID:      ownerID,
		}); err != nil {
			return err
		}
		// 作成者をこのワークスペースの主体にする。principals の行があること自体が所属なので、
		// この 1 行が無いと非メンバー扱いになり（middleware が全経路で所属を確認する）、
		// 作成者自身がワークスペースに入れなくなる。
		if _, err := qtx.InsertPrincipal(ctx, sqlcgen.InsertPrincipalParams{
			ID:          principalID,
			WorkspaceID: wsID,
			Kind:        string(domain.PrincipalKindUser),
			UserID:      sql.NullInt64{Int64: ownerID, Valid: true},
		}); err != nil {
			return err
		}
		// 所属だけでは何も見えない（役割が 1 つも無ければ実効権限は空）。作成者が自分の
		// ワークスペースを設定できるよう admin を張る。ここまでが 1 トランザクション。
		if _, err := qtx.UpsertWorkspaceGrant(ctx, sqlcgen.UpsertWorkspaceGrantParams{
			WorkspaceID: wsID,
			PrincipalID: principalID,
			Role:        string(domain.GrantRoleAdmin),
		}); err != nil {
			return err
		}
		// 監査（段 6）。招待の手順を踏まない唯一の所属経路なので、作成者本人が actor になる。
		adminLabel := string(domain.GrantRoleAdmin)
		if err := recordMembershipEvent(
			ctx, qtx, wsID, ownerID, ownerID, domain.MembershipEventMemberAdded, nil, &adminLabel,
		); err != nil {
			return err
		}
		created = toDomainWorkspace(ws)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// ProvisionPrivateSpace はプライベートスペースと作成者への space_grant(admin) を
// 1 トランザクションで作る（分けない理由は port の doc）。
func (p *workspaceProvisioner) ProvisionPrivateSpace(
	ctx context.Context, in repository.PrivateSpaceProvisionInput,
) (*domain.Space, error) {
	wsID, ok := kbParseID(in.WorkspaceID)
	if !ok {
		return nil, repository.ErrWorkspaceNotFound
	}
	// principals.user_id は bigint。範囲外はどの主体にも一致し得ないので先に止める
	// （ProvisionWorkspace と同じ判断）。
	creatorID, ok := toInt64ID(in.CreatorUserID)
	if !ok {
		return nil, outOfRangeIDError("user_id", in.CreatorUserID)
	}
	spaceID, err := kbNewID()
	if err != nil {
		return nil, err
	}

	var created domain.Space
	err = p.runInTx(ctx, func(qtx *sqlcgen.Queries) error {
		// 作成者の主体（＝ 所属そのもの）。無ければ非メンバーで、作らせない。
		principal, err := qtx.GetUserPrincipal(ctx, sqlcgen.GetUserPrincipalParams{
			WorkspaceID: wsID,
			UserID:      sql.NullInt64{Int64: creatorID, Valid: true},
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return repository.ErrPrincipalNotFound
			}
			return err
		}
		row, err := qtx.InsertSpace(ctx, sqlcgen.InsertSpaceParams{
			ID:          spaceID,
			WorkspaceID: wsID,
			Key:         in.Key,
			Name:        in.Name,
			Visibility:  string(domain.SpaceVisibilityPrivate),
		})
		if err != nil {
			// key の重複は一意制約を唯一の判定にする（CreateSpace と同じ）。
			if isUniqueViolation(err) {
				return repository.ErrSpaceKeyTaken
			}
			if isForeignKeyViolation(err) {
				return repository.ErrWorkspaceNotFound
			}
			return err
		}
		// ワークスペース既定が届かないスペースなので、この grant が作成者の唯一の入口。
		if _, err := qtx.UpsertSpaceGrant(ctx, sqlcgen.UpsertSpaceGrantParams{
			WorkspaceID: wsID,
			SpaceID:     spaceID,
			PrincipalID: principal.ID,
			Role:        string(domain.GrantRoleAdmin),
		}); err != nil {
			return err
		}
		created = toDomainSpace(row)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}
