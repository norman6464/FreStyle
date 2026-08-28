import type { NotePage, NotePageTreeNode } from '../model/types';

/**
 * searchAncestors は「見つからなかった」を null、「見つかった」を根からの祖先 ID で返す。
 *
 * 空配列を「見つからなかった」に使えないのが要点。**対象が根そのものだったとき**の
 * 正しい答えも空配列（祖先が居ない）なので、両者を区別できる型でないと
 * 「根のページを開くと、無関係な枝が全部開く」といった壊れ方をする。
 */
function searchAncestors(nodes: NotePageTreeNode[], pageId: string): string[] | null {
  for (const node of nodes) {
    if (node.page.id === pageId) return [];
    const below = searchAncestors(node.children, pageId);
    if (below !== null) return [node.page.id, ...below];
  }
  return null;
}

/**
 * collectNoteAncestorIds は指定したページに至る**祖先の ID**を根から順に返す。
 * 木に居なければ空配列。
 *
 * 現在位置のページを開いたとき、その祖先を自動で開くために使う。返すのは祖先だけで、
 * 当のページ自身は含まない（自分を開く必要は無く、含めると葉が開いた扱いになる）。
 */
export function collectNoteAncestorIds(nodes: NotePageTreeNode[], pageId: string): string[] {
  return searchAncestors(nodes, pageId) ?? [];
}

/**
 * replaceNotePageInTree は木の中の 1 ページを差し替えた**新しい木**を返す（元は変えない）。
 *
 * 名前を変えたあとに使う。木ごと取り直すのでも正しいが、取り直すと一瞬空になり、
 * 開いていた段も畳まれて見える。サーバーが返した新しいページで 1 枚だけ差し替えれば、
 * 表示が飛ばない。
 *
 * 差し替えるのは**そのページの値だけ**で、木の形（親子・兄弟順）は触らない。
 * 形が変わる操作（作成・移動）でこれを使わないこと — 兄弟順はサーバーが決めるので、
 * 手元で組み立てると必ずずれる。
 */
export function replaceNotePageInTree(nodes: NotePageTreeNode[], page: NotePage): NotePageTreeNode[] {
  let changed = false;
  const next = nodes.map((node) => {
    if (node.page.id === page.id) {
      changed = true;
      return { ...node, page };
    }
    const children = replaceNotePageInTree(node.children, page);
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
export type NoteDropTarget =
  /** その行の手前に、同じ親の兄弟として置く。 */
  | { kind: 'before'; pageId: string }
  /** その行の直後に、同じ親の兄弟として置く。 */
  | { kind: 'after'; pageId: string }
  /** その行の子として、末尾に置く。 */
  | { kind: 'into'; pageId: string };

/** findNode は木から 1 ノードとその親の ID を探す。 */
function findNode(
  nodes: NotePageTreeNode[],
  pageId: string,
  parentId: string | null = null,
): { node: NotePageTreeNode; parentId: string | null } | null {
  for (const node of nodes) {
    if (node.page.id === pageId) return { node, parentId };
    const found = findNode(node.children, pageId, node.page.id);
    if (found) return found;
  }
  return null;
}

/** removeNode は木からそのページ（と子孫）を取り除いた新しい木を返す。 */
function removeNode(nodes: NotePageTreeNode[], pageId: string): NotePageTreeNode[] {
  return nodes
    .filter((node) => node.page.id !== pageId)
    .map((node) => ({ ...node, children: removeNode(node.children, pageId) }));
}

/** insertNode は target の指す場所へ node を差し込んだ新しい木を返す。 */
function insertNode(
  nodes: NotePageTreeNode[],
  node: NotePageTreeNode,
  target: NoteDropTarget,
): NotePageTreeNode[] {
  const out: NotePageTreeNode[] = [];
  for (const current of nodes) {
    if (target.kind === 'before' && current.page.id === target.pageId) out.push(node);
    if (current.page.id === target.pageId && target.kind === 'into') {
      out.push({ ...current, children: [...insertNode(current.children, node, target), node] });
      continue;
    }
    out.push({ ...current, children: insertNode(current.children, node, target) });
    if (target.kind === 'after' && current.page.id === target.pageId) out.push(node);
  }
  return out;
}

/**
 * moveNotePageInTree は落とした先へページを動かした**新しい木**を返す（元は変えない）。
 * 動かせない指定なら null を返す。
 *
 * 画面を先に動かすためだけの計算で、**正しい並びを決めるのはサーバー**。
 * ここで作る木はサーバーの返事が来るまでの見た目でしかなく、失敗したら丸ごと捨てる。
 *
 * 自分自身や自分の子孫の中へは動かせない（木が根から切り離される）。サーバーも同じ理由で
 * 断るが、画面が先に動いてから巻き戻るより、動かさないほうが分かりやすい。
 */
export function moveNotePageInTree(
  nodes: NotePageTreeNode[],
  pageId: string,
  target: NoteDropTarget,
): NotePageTreeNode[] | null {
  if (pageId === target.pageId) return null;
  const found = findNode(nodes, pageId);
  if (!found) return null;
  // 落下先が木に無ければ何もしない。**確かめずに進めると、取り除いたあと差し込む先が
  // 見つからず、動かしたページと子孫が木から消える**（画面から丸ごと居なくなる）。
  if (!findNode(nodes, target.pageId)) return null;
  // 自分の子孫が落下先なら、動かすと木が根から切り離される。
  if (findNode(found.node.children, target.pageId)) return null;
  return insertNode(removeNode(nodes, pageId), found.node, target);
}

/**
 * 行のメニューから呼べる 4 つの動かし方。動かせない向きは null。
 *
 * ドラッグと同じ「隣のページの ID」で表すので、送り先の API も同じ。
 * キーボードのためだけに別の経路を作らない（作ると失敗の扱いも二重になる）。
 */
export interface NoteMoveActions {
  /** ひとつ上の兄弟の手前へ。先頭なら null。 */
  up: NoteDropTarget | null;
  /** ひとつ下の兄弟の直後へ。末尾なら null。 */
  down: NoteDropTarget | null;
  /** ひとつ上の兄弟の子へ。先頭なら null（受け入れる相手がいない）。 */
  indent: NoteDropTarget | null;
  /** 親の直後へ（ひとつ外側の段に出る）。すでに最上段なら null。 */
  outdent: NoteDropTarget | null;
}

/**
 * noteMoveActions は、その行から動かせる 4 つの向きを求める。
 *
 * 兄弟は**画面に出ている並び**をそのまま使う。伏せられている兄弟は数に入っていないが、
 * 「見えている隣の隣へ」という利用者の意図はそれで正しく表せる（実際のキーの計算は
 * サーバーが伏せた兄弟も含めて行うので、隙間に落ちることはない）。
 */
export function noteMoveActions(
  siblings: NotePageTreeNode[],
  index: number,
  parentId: string | null,
): NoteMoveActions {
  const previous = siblings[index - 1];
  const next = siblings[index + 1];
  return {
    up: previous ? { kind: 'before', pageId: previous.page.id } : null,
    down: next ? { kind: 'after', pageId: next.page.id } : null,
    indent: previous ? { kind: 'into', pageId: previous.page.id } : null,
    outdent: parentId ? { kind: 'after', pageId: parentId } : null,
  };
}
