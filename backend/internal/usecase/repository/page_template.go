package repository

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ErrDuplicateTemplateName は作成しようとした雛形名が同じワークスペースで既に
// 使われているときに返す（uq_page_templates_workspace_name の一意制約違反の翻訳）。
var ErrDuplicateTemplateName = errors.New("page template name is already taken")

// PageTemplateRepository は page_templates テーブルへのアクセスを提供する。
//
// KnowledgeBaseRepository とは別の fat interface にする（PageVersionRepository と同じ役割
// 分担 — 雛形はページ本文とは別の生存期間を持つ独立した資産の塊で、KnowledgeBaseRepository に
// これ以上メソッドを足すと 1 interface が肥大しすぎる）。
type PageTemplateRepository interface {
	// Create は雛形を作成する。ID は呼び出し前（usecase 側）で採番済みの前提で、
	// 呼び出し後の tpl は DB で確定した行（created_at 等）で上書きされる。
	// name が同じワークスペースで使用済みなら ErrDuplicateTemplateName
	// （検査してから INSERT するまでの間に別の要求が同じ名前を取り得るため、
	// 一意制約を唯一の判定にする — CreateSpace の ErrSpaceKeyTaken と同じ考え方）。
	Create(ctx context.Context, tpl *domain.PageTemplate) error
	// List はワークスペースの雛形一覧を name 昇順で返す。
	//
	// spaceID が nil なら「space_id IS NULL の行（ワークスペース全体向け）」だけを返す。
	// 非 nil なら、それに加えて「space_id = *spaceID の行」も返す（ワークスペース全体向けの
	// 雛形は、どのスペースの一覧からも見える。呼び出し側は自分の spaceId を常に知っている
	// 文脈で呼ぶため、spaceID を渡さない = 未所属の文脈、という設計になる）。
	List(ctx context.Context, workspaceID string, spaceID *string) ([]domain.PageTemplate, error)
	// Get は雛形を 1 件引く（doc 込み）。無い・別ワークスペースなら domain.ErrPageTemplateNotFound。
	Get(ctx context.Context, workspaceID, templateID string) (*domain.PageTemplate, error)
	// Delete は雛形を削除する。対象が無ければ domain.ErrPageTemplateNotFound。
	Delete(ctx context.Context, workspaceID, templateID string) error
}
