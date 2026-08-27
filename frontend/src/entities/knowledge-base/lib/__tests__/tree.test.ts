import { describe, expect, it } from 'vitest';
import { collectKbAncestorIds, flattenKbTree } from '../tree';
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
