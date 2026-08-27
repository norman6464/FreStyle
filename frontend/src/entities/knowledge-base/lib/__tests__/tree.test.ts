import { describe, expect, it } from 'vitest';
import {
  collectKbAncestorIds,
  flattenKbTree,
  moveKbPageInTree,
  replaceKbPageInTree,
} from '../tree';
import type { KbPage, KbPageTreeNode } from '../../model/types';

/** page は木の形だけを見たいので、題名以外は既定値で埋める。 */
function page(id: string, title = id): KbPage {
  return {
    id,
    spaceId: 'space-1',
    title,
    createdByUserId: 1,
    createdAt: '2026-08-01T00:00:00Z',
    updatedAt: '2026-08-01T00:00:00Z',
  };
}

function node(id: string, children: KbPageTreeNode[] = [], hidden = false): KbPageTreeNode {
  return { page: page(id), children, hasHiddenChildren: hidden };
}

/** ids は行の並びを「ページ ID」と「伏せたものが在る印」で表したもの（読み順そのもの）。 */
function ids(entries: ReturnType<typeof flattenKbTree>): string[] {
  return entries.map((e) => (e.kind === 'page' ? e.page.id : 'hidden'));
}

describe('flattenKbTree', () => {
  it('閉じている行の子孫は行にしない', () => {
    const tree = [node('a', [node('a1'), node('a2')]), node('b')];

    const entries = flattenKbTree(tree, new Set());

    expect(ids(entries)).toEqual(['a', 'b']);
    expect(entries[0]).toMatchObject({ kind: 'page', hasChildren: true, expanded: false });
  });

  it('開いている行の子は次の段として続く', () => {
    const tree = [node('a', [node('a1', [node('a1x')])]), node('b')];

    const entries = flattenKbTree(tree, new Set(['a', 'a1']));

    expect(ids(entries)).toEqual(['a', 'a1', 'a1x', 'b']);
    expect(entries.map((e) => e.depth)).toEqual([0, 1, 2, 0]);
  });

  it('祖先が閉じていれば、子が開いた印を持っていても行にしない', () => {
    // 開いた状態は集合で持つので、親を閉じても子の印は残る。それでも描いてはいけない。
    const tree = [node('a', [node('a1', [node('a1x')])])];

    expect(ids(flattenKbTree(tree, new Set(['a1'])))).toEqual(['a']);
  });

  it('伏せたものが在る印は、その段の子の最後に出す', () => {
    // 平らにした配列では親の要素が子より先に来るので、ページの行に混ぜると
    // 「表示できないページがあります」が子より前に出てしまう。
    const tree = [node('a', [node('a1'), node('a2')], true)];

    expect(ids(flattenKbTree(tree, new Set(['a'])))).toEqual(['a', 'a1', 'a2', 'hidden']);
  });

  it('閉じている段では伏せた印を出さない', () => {
    // 見える子も出していないのに伏せた分だけ出すと、閉じているのに何か書いてある行になる。
    const tree = [node('a', [node('a1')], true)];

    expect(ids(flattenKbTree(tree, new Set()))).toEqual(['a']);
  });

  it('見える子が 1 枚も無い段では、閉じていても伏せた印を出す', () => {
    // 開けない段なので、出さないと永久に伝わらない。
    const tree = [node('a', [], true)];

    const entries = flattenKbTree(tree, new Set());

    expect(ids(entries)).toEqual(['a', 'hidden']);
    expect(entries[0]).toMatchObject({ kind: 'page', hasChildren: false, expanded: false });
  });

  it('伏せた印の行は、子と同じ段に置く', () => {
    const tree = [node('a', [node('a1')], true)];

    const entries = flattenKbTree(tree, new Set(['a']));

    expect(entries.map((e) => e.depth)).toEqual([0, 1, 1]);
  });

  it('兄弟の順は入力のまま保つ', () => {
    // 並びの正は backend の position（分数インデックスの辞書順）。ここで並べ替えない。
    const tree = [node('z'), node('a'), node('m')];

    expect(ids(flattenKbTree(tree, new Set()))).toEqual(['z', 'a', 'm']);
  });
});

describe('collectKbAncestorIds', () => {
  const tree = [node('a', [node('a1', [node('a1x')])]), node('b')];

  it('根からの祖先を順に返す', () => {
    expect(collectKbAncestorIds(tree, 'a1x')).toEqual(['a', 'a1']);
  });

  it('根そのものなら空（祖先が居ない）', () => {
    expect(collectKbAncestorIds(tree, 'b')).toEqual([]);
  });

  it('木に居なければ空', () => {
    expect(collectKbAncestorIds(tree, 'unknown')).toEqual([]);
  });

  it('根を対象にしても、無関係な枝を祖先として返さない', () => {
    // 「見つからない」を空配列で表すと、探索が最初の枝を降りて失敗した時点で
    // 空配列＝成功と誤読し、その枝の ID を積んで返してしまう。
    expect(collectKbAncestorIds(tree, 'a')).toEqual([]);
  });
});

describe('replaceKbPageInTree', () => {
  const tree = [node('a', [node('a1', [node('a1x')])]), node('b')];

  it('深い段のページも差し替える', () => {
    const renamed = { ...page('a1x'), title: '新しい名前' };

    const next = replaceKbPageInTree(tree, renamed);

    expect(next[0].children[0].children[0].page.title).toBe('新しい名前');
  });

  it('元の木は変えない', () => {
    replaceKbPageInTree(tree, { ...page('a1x'), title: '新しい名前' });

    expect(tree[0].children[0].children[0].page.title).toBe('a1x');
  });

  it('木の形は触らない', () => {
    // 兄弟順を決めるのはサーバー。手元で組み立てると必ずずれる。
    const next = replaceKbPageInTree(tree, { ...page('a'), title: 'A' });

    expect(next.map((n) => n.page.id)).toEqual(['a', 'b']);
    expect(next[0].children.map((n) => n.page.id)).toEqual(['a1']);
  });

  it('見つからなければ元の配列をそのまま返す', () => {
    // 新しい配列を作ると、参照で変化を見ている側が毎回描き直す。
    expect(replaceKbPageInTree(tree, page('unknown'))).toBe(tree);
  });

  it('関係ない枝は同じ参照のまま残す', () => {
    const next = replaceKbPageInTree(tree, { ...page('a1x'), title: '新しい名前' });

    expect(next[1]).toBe(tree[1]);
  });
});

describe('moveKbPageInTree', () => {
  /** ids は木の形を「親>子」の並びで表す（順序も含めて比べるため）。 */
  function shape(nodes: ReturnType<typeof moveKbPageInTree>): string[] {
    if (!nodes) return [];
    const out: string[] = [];
    const walk = (list: typeof nodes, prefix: string) => {
      for (const node of list) {
        out.push(prefix + node.page.id);
        walk(node.children, prefix + node.page.id + '>');
      }
    };
    walk(nodes, '');
    return out;
  }

  const tree = [node('a', [node('a1'), node('a2')]), node('b'), node('c')];

  it('兄弟の手前に差し込む', () => {
    expect(shape(moveKbPageInTree(tree, 'c', { kind: 'before', pageId: 'b' }))).toEqual([
      'a', 'a>a1', 'a>a2', 'c', 'b',
    ]);
  });

  it('兄弟の直後に差し込む', () => {
    expect(shape(moveKbPageInTree(tree, 'b', { kind: 'after', pageId: 'c' }))).toEqual([
      'a', 'a>a1', 'a>a2', 'c', 'b',
    ]);
  });

  it('別の段の子として末尾に入る', () => {
    expect(shape(moveKbPageInTree(tree, 'b', { kind: 'into', pageId: 'a' }))).toEqual([
      'a', 'a>a1', 'a>a2', 'a>b', 'c',
    ]);
  });

  it('子孫ごと動く', () => {
    const nested = [node('a', [node('a1', [node('a1x')])]), node('b')];

    expect(shape(moveKbPageInTree(nested, 'a1', { kind: 'into', pageId: 'b' }))).toEqual([
      'a', 'b', 'b>a1', 'b>a1>a1x',
    ]);
  });

  it('元の木は変えない', () => {
    moveKbPageInTree(tree, 'b', { kind: 'into', pageId: 'a' });

    expect(shape(tree)).toEqual(['a', 'a>a1', 'a>a2', 'b', 'c']);
  });

  it('自分自身の中へは動かせない', () => {
    expect(moveKbPageInTree(tree, 'a', { kind: 'into', pageId: 'a' })).toBeNull();
  });

  it('自分の子孫の中へは動かせない', () => {
    // 動かすと木が根から切り離される。サーバーも同じ理由で断るが、
    // 画面が先に動いてから巻き戻るより、動かさないほうが分かりやすい。
    expect(moveKbPageInTree(tree, 'a', { kind: 'into', pageId: 'a1' })).toBeNull();
    expect(moveKbPageInTree(tree, 'a', { kind: 'after', pageId: 'a2' })).toBeNull();
  });

  it('木に無いページは動かせない', () => {
    expect(moveKbPageInTree(tree, 'unknown', { kind: 'into', pageId: 'a' })).toBeNull();
  });
});
