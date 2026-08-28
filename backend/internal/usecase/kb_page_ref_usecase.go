package usecase

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

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
	// 参照があるなら、まず保存されている title を全部剥がす。保存側も剥がしているが
	// （StripPageRefTitles）、それは新しい保存にしか効かない — 剥がす前に保存された
	// doc が残っていれば、閲覧権限を失った読み手へ古い題名がそのまま返ってしまう。
	// 読み出し側でも剥がすことで、返る題名は必ず**この読み手の**可視判定を通った
	// 現在の値だけになる（事実の取得に失敗しても、剥がした doc を返す）。
	stripped := stripPageRefTitlesNode(root)

	render := func() (string, bool) {
		out, err := json.Marshal(root)
		if err != nil {
			return in.Doc, false
		}
		return string(out), true
	}

	rows, err := u.perms.ListWorkspacePageViewFactsByIDs(ctx, in.WorkspaceID, in.UserID, collector.ids)
	if err != nil {
		if stripped {
			if out, ok := render(); ok {
				return out, err
			}
		}
		return in.Doc, err
	}
	titles := make(map[string]string, len(rows))
	for _, row := range rows {
		// アーカイブ済みは題名に採らない（隠したページの現在の題名を本文へ映さない —
		// 検索が現役だけを対象にするのと同じ線引き）。パンくず側は逆に含める。
		if row.Page.ArchivedAt != nil {
			continue
		}
		if domain.ResolvePageView(row.Facts) {
			titles[row.Page.ID] = row.Page.Title
		}
	}
	rewritten := rewritePageRefTitles(root, titles)
	if !stripped && !rewritten {
		return in.Doc, nil
	}
	out, ok := render()
	if !ok {
		return in.Doc, nil
	}
	return out, nil
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
				if id, ok := attrs["pageId"].(string); ok {
					// UUID の正規形（小文字・ハイフン区切り）へ寄せてから数える。
					// repository も同じ正規化で照合するので、大文字や中括弧付きの
					// 表記で保存された参照もここで揃えないと、行は引けたのに
					// 題名の突き合わせだけが外れる。
					if canonical, ok := canonicalPageRefID(id); ok {
						if _, dup := c.seen[canonical]; !dup {
							c.seen[canonical] = struct{}{}
							c.ids = append(c.ids, canonical)
						}
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
					if canonical, cok := canonicalPageRefID(id); cok {
						if title, found := titles[canonical]; found && attrs["title"] != title {
							attrs["title"] = title
							changed = true
						}
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

// canonicalPageRefID は参照の pageId を UUID の正規形（小文字・ハイフン区切り）へ寄せる。
// UUID として読めない値は参照として扱わない（repository 側の落とし方と揃える）。
func canonicalPageRefID(id string) (string, bool) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return "", false
	}
	return parsed.String(), true
}

// AncestorRef はパンくず 1 段分（ページ ID と現在の題名）。
type AncestorRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ListViewableAncestorsUseCase はページの祖先のうち、読み手が閲覧できるものだけを
// 根から順に返す（パンくず用）。
//
// 見えない祖先は**行ごと出さない**（題名どころか実在も知らせない）。木の応答が
// 見えない親の配下を出さないのと同じ規則で、可視の判定も同じ事実
// （ListWorkspacePageViewFactsByIDs + domain.ResolvePageView）を通す。
// 結果、パンくずには穴があき得るが、それは木と同じ見え方 — パンくずだけ別の
// 判定を持つと「木には出ないのに道筋には出る」穴になる。
//
// **アーカイブ済みの祖先は含める**（閲覧できる限り）。アーカイブ済みのページは
// /p/{id} で開けるので、経路から抜くと「その段が無い」かのように場所を偽る。
type ListViewableAncestorsUseCase struct {
	pages repository.KnowledgeBaseRepository
	perms repository.KnowledgeBasePermissionRepository
}

func NewListViewableAncestorsUseCase(
	pages repository.KnowledgeBaseRepository,
	perms repository.KnowledgeBasePermissionRepository,
) *ListViewableAncestorsUseCase {
	return &ListViewableAncestorsUseCase{pages: pages, perms: perms}
}

type ListViewableAncestorsInput struct {
	WorkspaceID string
	UserID      uint64
	PageID      string
}

func (u *ListViewableAncestorsUseCase) Execute(ctx context.Context, in ListViewableAncestorsInput) ([]AncestorRef, error) {
	ids, err := u.pages.ListAncestorPageIDs(ctx, in.WorkspaceID, in.PageID)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []AncestorRef{}, nil
	}
	rows, err := u.perms.ListWorkspacePageViewFactsByIDs(ctx, in.WorkspaceID, in.UserID, ids)
	if err != nil {
		return nil, err
	}
	viewable := make(map[string]string, len(rows))
	for _, row := range rows {
		if domain.ResolvePageView(row.Facts) {
			viewable[row.Page.ID] = row.Page.Title
		}
	}
	// 並びは closure の順（根から）を保つ。facts の応答順は題名順なので使わない。
	out := make([]AncestorRef, 0, len(ids))
	for _, id := range ids {
		if title, ok := viewable[id]; ok {
			out = append(out, AncestorRef{ID: id, Title: title})
		}
	}
	return out, nil
}
