import type { KbPage, KbPageTreeNode } from '../model/types';

/**
 * サイドバーに 1 行として描くための、木を平らにしたもの。
 *
 * 木のまま再帰コンポーネントで描くこともできるが、平らな配列にしておくと
 * キーボード操作（上下移動）と仮想スクロールがそのまま素直に書ける。木の形は depth が持つ。
 */
export interface KbTreeRow {
  page: KbPage;
  /** 0 がスペース直下。1 段下がるごとに 1 増える。 */
  depth: number;
  /** 見える子を持つか。**伏せた子しか居ない場合は false**（開いても何も出ないため）。 */
  hasChildren: boolean;
  /** この行の直下で伏せた件数。0 なら何も示さない。 */
  hiddenChildCount: number;
  /** いま開いているか。子を持たない行では常に false。 */
  expanded: boolean;
}

/**
 * flattenKbTree は木を、いま開いている段だけの平らな行の並びに変換する。
 *
 * 閉じている行の子孫は結果に入らない（描かないものは行にしない）。
 * 兄弟順は backend が position 順で返したものをそのまま保つ。ここで並べ替えないこと
 * （分数インデックスの辞書順が正で、フロントで再現すると必ずずれる）。
 */
export function flattenKbTree(
  nodes: KbPageTreeNode[],
  expandedIds: ReadonlySet<string>,
  depth = 0,
): KbTreeRow[] {
  const rows: KbTreeRow[] = [];
  for (const node of nodes) {
    const hasChildren = node.children.length > 0;
    const expanded = hasChildren && expandedIds.has(node.page.id);
    rows.push({
      page: node.page,
      depth,
      hasChildren,
      hiddenChildCount: node.hiddenChildCount,
      expanded,
    });
    if (expanded) {
      rows.push(...flattenKbTree(node.children, expandedIds, depth + 1));
    }
  }
  return rows;
}

/**
 * searchAncestors は「見つからなかった」を null、「見つかった」を根からの祖先 ID で返す。
 *
 * 空配列を「見つからなかった」に使えないのが要点。**対象が根そのものだったとき**の
 * 正しい答えも空配列（祖先が居ない）なので、両者を区別できる型でないと
 * 「根のページを開くと、無関係な枝が全部開く」といった壊れ方をする。
 */
function searchAncestors(nodes: KbPageTreeNode[], pageId: string): string[] | null {
  for (const node of nodes) {
    if (node.page.id === pageId) return [];
    const below = searchAncestors(node.children, pageId);
    if (below !== null) return [node.page.id, ...below];
  }
  return null;
}

/**
 * collectKbAncestorIds は指定したページに至る**祖先の ID**を根から順に返す。
 * 木に居なければ空配列。
 *
 * 現在位置のページを開いたとき、その祖先を自動で開くために使う。返すのは祖先だけで、
 * 当のページ自身は含まない（自分を開く必要は無く、含めると葉が開いた扱いになる）。
 */
export function collectKbAncestorIds(nodes: KbPageTreeNode[], pageId: string): string[] {
  return searchAncestors(nodes, pageId) ?? [];
}
