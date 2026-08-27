import { describe, expect, it } from 'vitest';
import { collectKbAncestorIds, flattenKbTree } from '../tree';
import type { KbPage, KbPageTreeNode } from '../../model/types';

/** page は木の形だけを見たいので、題名以外は既定値で埋める。 */
function page(id: string, title = id): KbPage {
  return {
    id,
    spaceId: 'space-1',
    position: 'a0',
    title,
    createdByUserId: 1,
    createdAt: '2026-08-01T00:00:00Z',
    updatedAt: '2026-08-01T00:00:00Z',
  };
}

function node(id: string, children: KbPageTreeNode[] = [], hidden = 0): KbPageTreeNode {
  return { page: page(id), children, hiddenChildCount: hidden };
}

describe('flattenKbTree', () => {
  it('閉じている行の子孫は行にしない', () => {
    const tree = [node('a', [node('a1'), node('a2')]), node('b')];

    const rows = flattenKbTree(tree, new Set());

    expect(rows.map((r) => r.page.id)).toEqual(['a', 'b']);
    expect(rows[0].hasChildren).toBe(true);
    expect(rows[0].expanded).toBe(false);
  });

  it('開いている行の子は次の段として続く', () => {
    const tree = [node('a', [node('a1', [node('a1x')])]), node('b')];

    const rows = flattenKbTree(tree, new Set(['a', 'a1']));

    expect(rows.map((r) => r.page.id)).toEqual(['a', 'a1', 'a1x', 'b']);
    expect(rows.map((r) => r.depth)).toEqual([0, 1, 2, 0]);
  });

  it('祖先が閉じていれば、子が開いた印を持っていても行にしない', () => {
    // 開いた状態は集合で持つので、親を閉じても子の印は残る。それでも描いてはいけない。
    const tree = [node('a', [node('a1', [node('a1x')])])];

    const rows = flattenKbTree(tree, new Set(['a1']));

    expect(rows.map((r) => r.page.id)).toEqual(['a']);
  });

  it('伏せた子しか居ない行は開けない扱いにする', () => {
    // 開いても何も出ない行に開閉の三角を出すと、押しても反応しない行になる。
    const tree = [node('a', [], 3)];

    const rows = flattenKbTree(tree, new Set(['a']));

    expect(rows[0].hasChildren).toBe(false);
    expect(rows[0].expanded).toBe(false);
    expect(rows[0].hiddenChildCount).toBe(3);
  });

  it('兄弟の順は入力のまま保つ', () => {
    // 並びの正は backend の position（分数インデックスの辞書順）。ここで並べ替えない。
    const tree = [node('z'), node('a'), node('m')];

    expect(flattenKbTree(tree, new Set()).map((r) => r.page.id)).toEqual(['z', 'a', 'm']);
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
