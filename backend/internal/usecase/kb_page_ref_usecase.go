package usecase

import (
	"context"
	"encoding/json"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// kbPageRefNodeType は本文中の「ページ参照」インラインノードの type 名。
// エディタ側のスキーマ（PageRef）と一致させる。参照は attrs.pageId でページを指し、
// attrs.title は**表示のための派生値**で正本ではない — 正本は pages.title で、
// この usecase が読み出しのたびに引き直す。
const kbPageRefNodeType = "pageRef"

// kbPageRefMaxResolve は 1 ドキュメントで題名を解決する参照数の天井。
// 超えた分は保存されている文字のまま出す（読めなくなるわけではない）。
// 本文に数百の参照が並ぶのは異常系で、そこに可視判定のコストを払わない。
const kbPageRefMaxResolve = 100

// ResolvePageRefTitlesUseCase は本文（ProseMirror doc）中のページ参照の題名を、
// 読み手にとっての「いまの題名」へ差し替える。
//
// 差し替えるのは**読み手が閲覧できる現役ページ**の参照だけ。閲覧できない・存在しない・
// アーカイブ済み・他ワークスペースの参照は保存されている文字のまま — 著者が書けた
// 情報以上を読み手へ渡さない（可視判定は木・検索と同じ事実 + domain.ResolvePageView）。
//
// 解決はいかなる失敗でも本文を壊さない（読めない doc・事実の取得失敗は元の doc を
// そのまま返す）。題名は表示の飾りで、本文が開けることの方が重い。
type ResolvePageRefTitlesUseCase struct {
	perms repository.KnowledgeBasePermissionRepository
}

func NewResolvePageRefTitlesUseCase(r repository.KnowledgeBasePermissionRepository) *ResolvePageRefTitlesUseCase {
	return &ResolvePageRefTitlesUseCase{perms: r}
}

type ResolvePageRefTitlesInput struct {
	WorkspaceID string
	UserID      uint64
	// Doc は ProseMirror ドキュメント（JSON 文字列）。
	Doc string
}

func (u *ResolvePageRefTitlesUseCase) Execute(ctx context.Context, in ResolvePageRefTitlesInput) string {
	var root any
	if err := json.Unmarshal([]byte(in.Doc), &root); err != nil {
		return in.Doc
	}
	ids := collectPageRefIDs(root, nil)
	if len(ids) == 0 {
		return in.Doc
	}
	if len(ids) > kbPageRefMaxResolve {
		ids = ids[:kbPageRefMaxResolve]
	}
	rows, err := u.perms.ListWorkspacePageViewFactsByIDs(ctx, in.WorkspaceID, in.UserID, ids)
	if err != nil {
		return in.Doc
	}
	titles := make(map[string]string, len(rows))
	for _, row := range rows {
		if domain.ResolvePageView(row.Facts) {
			titles[row.Page.ID] = row.Page.Title
		}
	}
	if len(titles) == 0 {
		return in.Doc
	}
	if !rewritePageRefTitles(root, titles) {
		return in.Doc
	}
	out, err := json.Marshal(root)
	if err != nil {
		return in.Doc
	}
	return string(out)
}

// collectPageRefIDs は doc を歩いて pageRef の pageId を文書順・重複なしで集める。
// 辿るのは content 配列だけ（ProseMirror のノードの子はそこにしか居ない）。
// map の range で全キーを辿ると順序が実行ごとに変わり、解決数の天井を切る位置が
// 不定になる（＝どの参照が解決されるかが揺れる）。
func collectPageRefIDs(node any, ids []string) []string {
	switch v := node.(type) {
	case map[string]any:
		if v["type"] == kbPageRefNodeType {
			if attrs, ok := v["attrs"].(map[string]any); ok {
				if id, ok := attrs["pageId"].(string); ok && id != "" && !containsString(ids, id) {
					ids = append(ids, id)
				}
			}
		}
		ids = collectPageRefIDs(v["content"], ids)
	case []any:
		for _, child := range v {
			ids = collectPageRefIDs(child, ids)
		}
	}
	return ids
}

// rewritePageRefTitles は解決できた参照の title を書き換える。1 つでも書き換えたら true。
func rewritePageRefTitles(node any, titles map[string]string) bool {
	changed := false
	switch v := node.(type) {
	case map[string]any:
		if v["type"] == kbPageRefNodeType {
			if attrs, ok := v["attrs"].(map[string]any); ok {
				if id, ok := attrs["pageId"].(string); ok {
					if title, found := titles[id]; found && attrs["title"] != title {
						attrs["title"] = title
						changed = true
					}
				}
			}
		}
		if rewritePageRefTitles(v["content"], titles) {
			changed = true
		}
	case []any:
		for _, child := range v {
			if rewritePageRefTitles(child, titles) {
				changed = true
			}
		}
	}
	return changed
}

func containsString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
