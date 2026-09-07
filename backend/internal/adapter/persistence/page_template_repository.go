package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// pageTemplateRepository は [repository.PageTemplateRepository] の実装。page_templates は
// KnowledgeBaseRepository と同じくスキーマの正本が schema.hcl で GORM を通さない方針のため、
// クエリはすべて sqlc 生成コード + 素の *sql.DB で書く。
type pageTemplateRepository struct {
	baseRepository
}

// NewPageTemplateRepository は雛形の repository を組み立てる。
func NewPageTemplateRepository(db *sql.DB) repository.PageTemplateRepository {
	return &pageTemplateRepository{baseRepository{db: db}}
}

// queries は ctx に乗っているトランザクション（あれば）に束縛した sqlc の Queries を作る。
func (r *pageTemplateRepository) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(r.dbtx(ctx))
}

func toDomainPageTemplate(row sqlcgen.PageTemplate) domain.PageTemplate {
	tpl := domain.PageTemplate{
		ID:              row.ID.String(),
		WorkspaceID:     row.WorkspaceID.String(),
		Name:            row.Name,
		Doc:             string(row.Doc),
		CreatedByUserID: uint64(row.CreatedByUserID),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.SpaceID.Valid {
		id := row.SpaceID.UUID.String()
		tpl.SpaceID = &id
	}
	// icon は飾り（見た目）であって、壊れていても雛形本体の読み出しを止める理由にはならない。
	// json.Unmarshal に失敗したら nil に倒す（toDomainPage の icon 扱いと同じ判断）。
	if row.Icon != nil {
		var icon domain.PageIcon
		if err := json.Unmarshal(*row.Icon, &icon); err == nil {
			tpl.Icon = &icon
		}
	}
	return tpl
}

func (r *pageTemplateRepository) Create(ctx context.Context, tpl *domain.PageTemplate) error {
	wsID, ok := kbParseID(tpl.WorkspaceID)
	if !ok {
		return repository.ErrWorkspaceNotFound
	}
	spaceID, ok := kbNullID(tpl.SpaceID)
	if !ok {
		return repository.ErrSpaceNotFound
	}
	id, err := kbNewID()
	if err != nil {
		return err
	}
	createdBy, ok := toInt64ID(tpl.CreatedByUserID)
	if !ok {
		return outOfRangeIDError("created_by_user_id", tpl.CreatedByUserID)
	}
	var icon *json.RawMessage
	if tpl.Icon != nil {
		// 正規形（domain.PageIcon を Marshal し直したもの）だけを書く（UpdatePageIcon と同じ理由 —
		// 呼び出し側から渡された値をそのまま書かず、フィールド順・空白の揺れを DB に持ち込まない）。
		encoded, err := json.Marshal(tpl.Icon)
		if err != nil {
			return err
		}
		msg := json.RawMessage(encoded)
		icon = &msg
	}
	row, err := r.queries(ctx).InsertPageTemplate(ctx, sqlcgen.InsertPageTemplateParams{
		ID:              id,
		WorkspaceID:     wsID,
		SpaceID:         spaceID,
		Name:            tpl.Name,
		Icon:            icon,
		Doc:             json.RawMessage(tpl.Doc),
		CreatedByUserID: createdBy,
	})
	if err != nil {
		// name の重複（uq_page_templates_workspace_name）は入口の検証では防げない
		// （検査してから INSERT するまでの間に別の要求が同じ名前を取り得る）。
		// 一意制約を唯一の判定にして、業務上の衝突として返す（CreateSpace の
		// ErrSpaceKeyTaken と同じ考え方）。
		if isUniqueViolation(err) {
			return repository.ErrDuplicateTemplateName
		}
		return err
	}
	*tpl = toDomainPageTemplate(row)
	return nil
}

func (r *pageTemplateRepository) List(ctx context.Context, workspaceID string, spaceID *string) ([]domain.PageTemplate, error) {
	wsID, ok := kbParseID(workspaceID)
	if !ok {
		return []domain.PageTemplate{}, nil
	}
	spID, ok := kbNullID(spaceID)
	if !ok {
		// 不正な spaceID は「該当なし」として空を返す（存在し得ない ID を DB エラーにしない）。
		return []domain.PageTemplate{}, nil
	}
	rows, err := r.queries(ctx).ListPageTemplates(ctx, sqlcgen.ListPageTemplatesParams{
		WorkspaceID: wsID,
		SpaceID:     spID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.PageTemplate, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainPageTemplate(row))
	}
	return out, nil
}

func (r *pageTemplateRepository) Get(ctx context.Context, workspaceID, templateID string) (*domain.PageTemplate, error) {
	wsID, ok := kbParseID(workspaceID)
	tplID, ok2 := kbParseID(templateID)
	if !ok || !ok2 {
		return nil, domain.ErrPageTemplateNotFound
	}
	row, err := r.queries(ctx).GetPageTemplate(ctx, sqlcgen.GetPageTemplateParams{WorkspaceID: wsID, ID: tplID})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPageTemplateNotFound
	}
	if err != nil {
		return nil, err
	}
	tpl := toDomainPageTemplate(row)
	return &tpl, nil
}

func (r *pageTemplateRepository) Delete(ctx context.Context, workspaceID, templateID string) error {
	wsID, ok := kbParseID(workspaceID)
	tplID, ok2 := kbParseID(templateID)
	if !ok || !ok2 {
		return domain.ErrPageTemplateNotFound
	}
	affected, err := r.queries(ctx).DeletePageTemplate(ctx, sqlcgen.DeletePageTemplateParams{WorkspaceID: wsID, ID: tplID})
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrPageTemplateNotFound
	}
	return nil
}
