package usecase

import "github.com/norman6464/FreStyle/backend/internal/domain"

// PageTreeNode はページツリーの 1 ノード。
type PageTreeNode struct {
	Page     domain.Page     `json:"page"`
	Children []*PageTreeNode `json:"children"`
}

// PageTreeOrphanPolicy は「親が一覧に含まれていないページ」の扱いを決める。
type PageTreeOrphanPolicy int

const (
	// PageTreeOrphanAsRoot は親が一覧に無いページを根として見せる。
	// 一覧がスペースの現役ページ全件（ふるいにかけていない）で、親が欠けているのは
	// アーカイブ運用の不変条件が崩れた行に限られる場面で使う。データを隠さないことを優先する。
	PageTreeOrphanAsRoot PageTreeOrphanPolicy = iota
	// PageTreeOrphanHidden は親が一覧に無いページを、その子孫ごとツリーから落とす。
	// 権限でふるいにかけた一覧から木を組むときはこちらを使う。
	PageTreeOrphanHidden
)

// BuildPageTree は position 順に並んだページの平坦な一覧を木に組み立てる。
//
// 一覧が権限でふるいにかけられている場合、親が落ちたページを根に昇格させてはならない
// （PageTreeOrphanHidden）。昇格させると「見えないはずの親の下に何かがある」ことが
// ツリーの形から読み取れてしまい、隠した親のタイトルは伏せたまま配下の存在だけが漏れる。
// 見えない親の配下はツリーに現れず、直リンクで開いたときに個別の権限で判断される
// （祖先を隠したら配下も一覧から消えるのは、この種のツールで一般的な振る舞いでもある）。
//
// 入力の並び（同じ親を持つページ同士の相対順）はそのまま兄弟順として保たれる。
func BuildPageTree(pages []domain.Page, policy PageTreeOrphanPolicy) []*PageTreeNode {
	nodes := make(map[string]*PageTreeNode, len(pages))
	for _, p := range pages {
		nodes[p.ID] = &PageTreeNode{Page: p, Children: make([]*PageTreeNode, 0)}
	}
	roots := make([]*PageTreeNode, 0)
	for _, p := range pages {
		node := nodes[p.ID]
		if p.ParentID == nil {
			roots = append(roots, node)
			continue
		}
		if parent, ok := nodes[*p.ParentID]; ok {
			parent.Children = append(parent.Children, node)
			continue
		}
		if policy == PageTreeOrphanAsRoot {
			roots = append(roots, node)
		}
		// PageTreeOrphanHidden: どこにも繋がない。根から辿れないので子孫ごと結果に出ない。
	}
	return roots
}
