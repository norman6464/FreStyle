import type { KbPage, KbPageTreeNode } from '../model/types';

/**
 * MAX_TREE_NODES は木を歩く関数（探索・組み替え）に共通の反復回数の上限。
 *
 * ページ階層の深さ・件数はどちらも今のところ backend 側の上限を持たない
 * （段数の上限は別チケットで backend 側へ足す予定）。上限が無いまま再帰・反復で
 * 木全体を辿ると、極端に深い・多いツリーで JS のコールスタックを使い切るか、
 * 反復が終わらないまま画面を固まらせる。通常のページ階層（数十〜数百件）より
 * ずっと大きく取り、実用上は一切影響しない値にしてある。
 *
 * 上限に達したら「見つからなかった」「これ以上は組み替えない」側へ倒す
 * （見えている範囲を壊すより、深い場所の操作を諦める方が安全）。
 */
const MAX_TREE_NODES = 100_000;

/**
 * MAX_TREE_DEPTH は木を組み替える再帰（replaceKbPageInTree / removeNode / insertNode）の
 * 段数の上限。この 3 つは「子孫を組み立て直してから親を作る」形（親の children に子の
 * 組み立て結果を詰める）なので、探索専用の searchAncestors / findNode と違って
 * 明示スタックへは素直に書き換えられない（組み立て中の親を指すフレームをスタック側に
 * 持たせる必要があり、書き換えの複雑さの割に得られる保証が変わらない）。段数に上限を課す方式
 * （linkSafety.ts / stableBlockId.ts と同じ考え方）で、再帰の深さそのものを頭打ちにする。
 * 上限を超えた先は組み替えをやめ、その部分木をそのまま返す。
 */
const MAX_TREE_DEPTH = 300;

/**
 * AncestorLink は「今の祖先の列」を、末尾（直近の親）から根へ向かって辿れる連結リストで持つ。
 *
 * 配列で持つと、探索中の各階層で `[...ancestors, id]` のコピーが要り、1 本鎖（枝分かれの
 * 無い深い連なり）では深さ d の探索が合計 O(d²) の要素コピーになる（浅い木では気付かない
 * コストだが、深さに上限を課してもなお通す想定の入力では効いてくる）。連結リストなら
 * 1 階層進むたびの費用が O(1) で済み、実際に見つかったときに 1 度だけ配列へ組み直す。
 */
type AncestorLink = { id: string; parent: AncestorLink | null };

function ancestorLinkToArray(link: AncestorLink | null): string[] {
  const out: string[] = [];
  for (let l = link; l !== null; l = l.parent) out.push(l.id);
  out.reverse();
  return out;
}

/**
 * searchAncestors は「見つからなかった」を null、「見つかった」を根からの祖先 ID で返す。
 *
 * 空配列を「見つからなかった」に使えないのが要点。**対象が根そのものだったとき**の
 * 正しい答えも空配列（祖先が居ない）なので、両者を区別できる型でないと
 * 「根のページを開くと、無関係な枝が全部開く」といった壊れ方をする。
 *
 * 明示スタックによる反復の先行順（pre-order）探索。元の再帰は「同じ階層の次の兄弟へ移る前に
 * 必ずその子孫を先に見終える」順で辿るため、スタックには常に「今の枝の子孫」を
 * 「残りの兄弟」より上に積む（子を逆順で積むことで、pop したときに元の左から右の順になる）。
 */
function searchAncestors(nodes: KbPageTreeNode[], pageId: string): string[] | null {
  const stack: { node: KbPageTreeNode; parent: AncestorLink | null }[] = [];
  for (let i = nodes.length - 1; i >= 0; i -= 1) stack.push({ node: nodes[i], parent: null });

  for (let visited = 0; stack.length > 0; visited += 1) {
    if (visited >= MAX_TREE_NODES) return null;
    const { node, parent } = stack.pop()!;
    if (node.page.id === pageId) return ancestorLinkToArray(parent);
    const childParent: AncestorLink = { id: node.page.id, parent };
    for (let i = node.children.length - 1; i >= 0; i -= 1) {
      stack.push({ node: node.children[i], parent: childParent });
    }
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

/**
 * replaceKbPageInTree は木の中の 1 ページを差し替えた**新しい木**を返す（元は変えない）。
 *
 * 改名・アイコンの変更のように、ページ自身の値だけが変わったあとに使う。木ごと取り直す
 * のでも正しいが、取り直すと一瞬空になり、開いていた段も畳まれて見える。サーバーが返した
 * 新しいページで 1 枚だけ差し替えれば、表示が飛ばない。
 *
 * 差し替えるのは**そのページの値だけ**で、木の形（親子・兄弟順）は触らない。
 * 形が変わる操作（作成・移動）でこれを使わないこと — 兄弟順はサーバーが決めるので、
 * 手元で組み立てると必ずずれる。
 */
export function replaceKbPageInTree(nodes: KbPageTreeNode[], page: KbPage, depth = 0): KbPageTreeNode[] {
  if (depth >= MAX_TREE_DEPTH) return nodes;
  let changed = false;
  const next = nodes.map((node) => {
    if (node.page.id === page.id) {
      changed = true;
      return { ...node, page };
    }
    const children = replaceKbPageInTree(node.children, page, depth + 1);
    if (children !== node.children) {
      changed = true;
      return { ...node, children };
    }
    return node;
  });
  // 見つからなければ元の配列をそのまま返す。新しい配列を作ると、
  // 参照で変化を見ている側（React）が毎回描き直す。
  return changed ? next : nodes;
}

/**
 * ドラッグで落とした先。**「どの行の、どこに」**の 2 つだけで表す。
 *
 * 並び順のキーを持たないので、位置は必ず隣のページの ID で表す
 * （キーは応答に入っていない。整数部が兄弟の通し番号になるため、飛びから伏せた枚数が読める）。
 */
export type KbDropTarget =
  /** その行の手前に、同じ親の兄弟として置く。 */
  | { kind: 'before'; pageId: string }
  /** その行の直後に、同じ親の兄弟として置く。 */
  | { kind: 'after'; pageId: string }
  /** その行の子として、末尾に置く。 */
  | { kind: 'into'; pageId: string };

/**
 * findNode は木から 1 ノードとその親の ID を探す。
 * searchAncestors と同じ、明示スタックによる先行順探索。
 */
function findNode(
  nodes: KbPageTreeNode[],
  pageId: string,
): { node: KbPageTreeNode; parentId: string | null } | null {
  const stack: { node: KbPageTreeNode; parentId: string | null }[] = [];
  for (let i = nodes.length - 1; i >= 0; i -= 1) stack.push({ node: nodes[i], parentId: null });

  for (let visited = 0; stack.length > 0; visited += 1) {
    if (visited >= MAX_TREE_NODES) return null;
    const { node, parentId } = stack.pop()!;
    if (node.page.id === pageId) return { node, parentId };
    for (let i = node.children.length - 1; i >= 0; i -= 1) {
      stack.push({ node: node.children[i], parentId: node.page.id });
    }
  }
  return null;
}

/** removeNode は木からそのページ（と子孫）を取り除いた新しい木を返す。 */
function removeNode(nodes: KbPageTreeNode[], pageId: string, depth = 0): KbPageTreeNode[] {
  if (depth >= MAX_TREE_DEPTH) return nodes;
  return nodes
    .filter((node) => node.page.id !== pageId)
    .map((node) => ({ ...node, children: removeNode(node.children, pageId, depth + 1) }));
}

/** insertNode は target の指す場所へ node を差し込んだ新しい木を返す。 */
function insertNode(
  nodes: KbPageTreeNode[],
  node: KbPageTreeNode,
  target: KbDropTarget,
  depth = 0,
): KbPageTreeNode[] {
  if (depth >= MAX_TREE_DEPTH) return nodes;
  const out: KbPageTreeNode[] = [];
  for (const current of nodes) {
    if (target.kind === 'before' && current.page.id === target.pageId) out.push(node);
    if (current.page.id === target.pageId && target.kind === 'into') {
      out.push({ ...current, children: [...insertNode(current.children, node, target, depth + 1), node] });
      continue;
    }
    out.push({ ...current, children: insertNode(current.children, node, target, depth + 1) });
    if (target.kind === 'after' && current.page.id === target.pageId) out.push(node);
  }
  return out;
}

/**
 * moveKbPageInTree は落とした先へページを動かした**新しい木**を返す（元は変えない）。
 * 動かせない指定なら null を返す。
 *
 * 画面を先に動かすためだけの計算で、**正しい並びを決めるのはサーバー**。
 * ここで作る木はサーバーの返事が来るまでの見た目でしかなく、失敗したら丸ごと捨てる。
 *
 * 自分自身や自分の子孫の中へは動かせない（木が根から切り離される）。サーバーも同じ理由で
 * 断るが、画面が先に動いてから巻き戻るより、動かさないほうが分かりやすい。
 *
 * 動かす本人・落下先のどちらかが MAX_TREE_DEPTH より深ければ動かさない。findNode
 * （明示スタック・段数の上限を持たない）はそこまで深くても見つけてしまうが、
 * その後の removeNode/insertNode（段数に上限を持つ再帰）はその深さへ辿り着けず
 * 部分木を素通りする。確かめずに進めると、findNode は見つけたのに removeNode は
 * 取り除けず、結果として同じページが 2 か所（元の深い位置と、新しい浅い位置）に
 * 重複して現れる壊れ方をする。
 */
export function moveKbPageInTree(
  nodes: KbPageTreeNode[],
  pageId: string,
  target: KbDropTarget,
): KbPageTreeNode[] | null {
  if (pageId === target.pageId) return null;
  const found = findNode(nodes, pageId);
  if (!found) return null;
  // 落下先が木に無ければ何もしない。**確かめずに進めると、取り除いたあと差し込む先が
  // 見つからず、動かしたページと子孫が木から消える**（画面から丸ごと居なくなる）。
  if (!findNode(nodes, target.pageId)) return null;
  // 自分の子孫が落下先なら、動かすと木が根から切り離される。
  if (findNode(found.node.children, target.pageId)) return null;
  const movedDepth = searchAncestors(nodes, pageId)?.length ?? Infinity;
  const targetDepth = searchAncestors(nodes, target.pageId)?.length ?? Infinity;
  if (movedDepth >= MAX_TREE_DEPTH || targetDepth >= MAX_TREE_DEPTH) return null;
  return insertNode(removeNode(nodes, pageId), found.node, target);
}

/**
 * 行のメニューから呼べる 4 つの動かし方。動かせない向きは null。
 *
 * ドラッグと同じ「隣のページの ID」で表すので、送り先の API も同じ。
 * キーボードのためだけに別の経路を作らない（作ると失敗の扱いも二重になる）。
 */
export interface KbMoveActions {
  /** ひとつ上の兄弟の手前へ。先頭なら null。 */
  up: KbDropTarget | null;
  /** ひとつ下の兄弟の直後へ。末尾なら null。 */
  down: KbDropTarget | null;
  /** ひとつ上の兄弟の子へ。先頭なら null（受け入れる相手がいない）。 */
  indent: KbDropTarget | null;
  /** 親の直後へ（ひとつ外側の段に出る）。すでに最上段なら null。 */
  outdent: KbDropTarget | null;
}

/**
 * kbMoveActions は、その行から動かせる 4 つの向きを求める。
 *
 * 兄弟は**画面に出ている並び**をそのまま使う。伏せられている兄弟は数に入っていないが、
 * 「見えている隣の隣へ」という利用者の意図はそれで正しく表せる（実際のキーの計算は
 * サーバーが伏せた兄弟も含めて行うので、隙間に落ちることはない）。
 */
export function kbMoveActions(
  siblings: KbPageTreeNode[],
  index: number,
  parentId: string | null,
): KbMoveActions {
  const previous = siblings[index - 1];
  const next = siblings[index + 1];
  return {
    up: previous ? { kind: 'before', pageId: previous.page.id } : null,
    down: next ? { kind: 'after', pageId: next.page.id } : null,
    indent: previous ? { kind: 'into', pageId: previous.page.id } : null,
    outdent: parentId ? { kind: 'after', pageId: parentId } : null,
  };
}
