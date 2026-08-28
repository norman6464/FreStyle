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
// アーカイブ済み・他ワークスペースの参照には題名を入れない（保存側が題名を持たない —
// stripPageRefTitles を参照 — ので、表示は「ページ」の代替文字に落ちる）。
//
// 解決は本文を壊さない: 読めない doc は元のまま返す（題名は表示の飾りで、本文が
// 開けることの方が重い）。事実の取得失敗も元の doc を返すが、error は呼び出し側へ
// 返す — 解決が恒常的に死んでいることに気づけるよう、握り潰す判断は handler が行う。
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

func (u *ResolvePageRefTitlesUseCase) Execute(ctx context.Context, in ResolvePageRefTitlesInput) (string, error) {
	var root any
	if err := json.Unmarshal([]byte(in.Doc), &root); err != nil {
		// 読めない doc は「解決できない」ではなく「解決の対象が無い」。エラーにしない
		// （保存経路が別途 400 で弾いており、ここで返しても呼び出し側にできることが無い）。
		return in.Doc, nil //nolint:nilerr // 意図した劣化: 本文の読み出しを題名の都合で止めない
	}
	collector := newPageRefCollector()
	collector.collect(root)
	if len(collector.ids) == 0 {
		return in.Doc, nil
	}
	rows, err := u.perms.ListWorkspacePageViewFactsByIDs(ctx, in.WorkspaceID, in.UserID, collector.ids)
	if err != nil {
		return in.Doc, err
	}
	titles := make(map[string]string, len(rows))
	for _, row := range rows {
		if domain.ResolvePageView(row.Facts) {
			titles[row.Page.ID] = row.Page.Title
		}
	}
	if len(titles) == 0 {
		return in.Doc, nil
	}
	if !rewritePageRefTitles(root, titles) {
		return in.Doc, nil
	}
	out, err := json.Marshal(root)
	if err != nil {
		return in.Doc, nil //nolint:nilerr // 意図した劣化: 元の doc を返せば表示は成立する
	}
	return string(out), nil
}

// StripPageRefTitles は保存前の doc からページ参照の title を取り除く。
//
// title は**読み手ごとに**読み出し時へ解決する派生値で、保存してはいけない。
// 保存すると、閲覧できる編集者の画面で解決された現在の題名が、その人の通常の
// 保存 1 回で本文へ焼き込まれ、閲覧できない読み手にもそのまま返ってしまう
// （読み出し時の可視判定を素通りする抜け道になる）。
//
// 参照が無い・読めない doc は元のまま返す（読めない doc は保存経路の検証が弾く）。
func StripPageRefTitles(doc string) string {
	var root any
	if err := json.Unmarshal([]byte(doc), &root); err != nil {
		return doc
	}
	if !stripPageRefTitlesNode(root) {
		return doc
	}
	out, err := json.Marshal(root)
	if err != nil {
		return doc
	}
	return string(out)
}

func stripPageRefTitlesNode(node any) bool {
	changed := false
	switch v := node.(type) {
	case map[string]any:
		if v["type"] == kbPageRefNodeType {
			if attrs, ok := v["attrs"].(map[string]any); ok {
				if _, has := attrs["title"]; has && attrs["title"] != nil {
					attrs["title"] = nil
					changed = true
				}
			}
		}
		if stripPageRefTitlesNode(v["content"]) {
			changed = true
		}
	case []any:
		for _, child := range v {
			if stripPageRefTitlesNode(child) {
				changed = true
			}
		}
	}
	return changed
}

// pageRefCollector は doc を歩いて pageRef の pageId を文書順・重複なしで集める。
// 重複の判定は set（O(1)）で行い、天井（kbPageRefMaxResolve）に達したら**収集自体を
// 打ち切る** — 線形走査の重複判定や収集後の切り詰めだと、参照を大量に並べた本文
// 1 つで読み出しのたびに CPU を燃やせてしまう。
//
// 辿るのは content 配列だけ（ProseMirror のノードの子はそこにしか居ない）。
// map の range で全キーを辿ると順序が実行ごとに変わり、天井を切る位置が不定になる。
type pageRefCollector struct {
	ids  []string
	seen map[string]struct{}
}

func newPageRefCollector() *pageRefCollector {
	return &pageRefCollector{seen: map[string]struct{}{}}
}

func (c *pageRefCollector) collect(node any) {
	if len(c.ids) >= kbPageRefMaxResolve {
		return
	}
	switch v := node.(type) {
	case map[string]any:
		if v["type"] == kbPageRefNodeType {
			if attrs, ok := v["attrs"].(map[string]any); ok {
				if id, ok := attrs["pageId"].(string); ok && id != "" {
					if _, dup := c.seen[id]; !dup {
						c.seen[id] = struct{}{}
						c.ids = append(c.ids, id)
					}
				}
			}
		}
		c.collect(v["content"])
	case []any:
		for _, child := range v {
			c.collect(child)
		}
	}
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
