package repository

import (
	"context"
	"errors"

	"github.com/norman6464/frestyle/backend/internal/domain"
)

// ErrDuplicateTemplateName は雛形名が同じワークスペースで既に使われているときに返す
// （uq_page_templates_workspace_name の一意制約違反の翻訳）。
var ErrDuplicateTemplateName = errors.New("page template name is already taken")

// PageTemplateRepository は page_templates テーブルへのアクセスを提供する。
// KnowledgeBaseRepository とは別の fat interface にする（雛形はページ本文とは別の
// 生存期間を持つ独立した資産の塊で、これ以上メソッドを足すと肥大しすぎる）。
type PageTemplateRepository interface {
	// Create は雛形を作成する。ID は usecase 側で採番済みの前提で、呼び出し後の tpl は
	// DB 確定行で上書きされる。name が使用済みなら ErrDuplicateTemplateName（検査と INSERT の
	// 間の競合を避けるため、一意制約を唯一の判定にする）。
	Create(ctx context.Context, tpl *domain.PageTemplate) error
	// List はワークスペースの雛形一覧を name 昇順で返す。spaceID が nil なら
	// ワークスペース全体向け（space_id IS NULL）だけ、非 nil ならそれに加えて
	// そのスペース向けも返す（全体向けの雛形はどのスペースの一覧からも見える）。
	List(ctx context.Context, workspaceID string, spaceID *string) ([]domain.PageTemplate, error)
	// Get は雛形を 1 件引く（doc 込み）。無い・別ワークスペースなら domain.ErrPageTemplateNotFound。
	Get(ctx context.Context, workspaceID, templateID string) (*domain.PageTemplate, error)
	// Delete は雛形を削除する。対象が無ければ domain.ErrPageTemplateNotFound。
	Delete(ctx context.Context, workspaceID, templateID string) error
}
