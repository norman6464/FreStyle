import { describe, expect, it } from 'vitest';
import {
  collectNoteAncestorIds,
  noteMoveActions,
  moveNotePageInTree,
  replaceNotePageInTree,
} from '../tree';
import type { NotePage, NotePageTreeNode } from '../../model/types';

/** page は木の形だけを見たいので、題名以外は既定値で埋める。 */
function page(id: string, title = id): NotePage {
  return {
    id,
    spaceId: 'space-1',
    title,
    createdByUserId: 1,
    createdAt: '2026-08-01T00:00:00Z',
    updatedAt: '2026-08-01T00:00:00Z',
  };
}

function node(id: string, children: NotePageTreeNode[] = [], hidden = false): NotePageTreeNode {
  return { page: page(id), children, hasHiddenChildren: hidden, parentArchived: false };
}

describe('collectNoteAncestorIds', () => {
  const tree = [node('a', [node('a1', [node('a1x')])]), node('b')];

  it('根からの祖先を順に返す', () => {
    expect(collectNoteAncestorIds(tree, 'a1x')).toEqual(['a', 'a1']);
  });

  it('根そのものなら空（祖先が居ない）', () => {
    expect(collectNoteAncestorIds(tree, 'b')).toEqual([]);
  });

  it('木に居なければ空', () => {
    expect(collectNoteAncestorIds(tree, 'unknown')).toEqual([]);
  });

  it('根を対象にしても、無関係な枝を祖先として返さない', () => {
    // 「見つからない」を空配列で表すと、探索が最初の枝を降りて失敗した時点で
    // 空配列＝成功と誤読し、その枝の ID を積んで返してしまう。
    expect(collectNoteAncestorIds(tree, 'a')).toEqual([]);
  });
});

describe('replaceNotePageInTree', () => {
  const tree = [node('a', [node('a1', [node('a1x')])]), node('b')];

  it('深い段のページも差し替える', () => {
    const renamed = { ...page('a1x'), title: '新しい名前' };

    const next = replaceNotePageInTree(tree, renamed);

    expect(next[0].children[0].children[0].page.title).toBe('新しい名前');
  });

  it('元の木は変えない', () => {
    replaceNotePageInTree(tree, { ...page('a1x'), title: '新しい名前' });

    expect(tree[0].children[0].children[0].page.title).toBe('a1x');
  });

  it('木の形は触らない', () => {
    // 兄弟順を決めるのはサーバー。手元で組み立てると必ずずれる。
    const next = replaceNotePageInTree(tree, { ...page('a'), title: 'A' });

    expect(next.map((n) => n.page.id)).toEqual(['a', 'b']);
    expect(next[0].children.map((n) => n.page.id)).toEqual(['a1']);
  });

  it('見つからなければ元の配列をそのまま返す', () => {
    // 新しい配列を作ると、参照で変化を見ている側が毎回描き直す。
    expect(replaceNotePageInTree(tree, page('unknown'))).toBe(tree);
  });

  it('関係ない枝は同じ参照のまま残す', () => {
    const next = replaceNotePageInTree(tree, { ...page('a1x'), title: '新しい名前' });

    expect(next[1]).toBe(tree[1]);
  });
});

describe('moveNotePageInTree', () => {
  /** ids は木の形を「親>子」の並びで表す（順序も含めて比べるため）。 */
  function shape(nodes: ReturnType<typeof moveNotePageInTree>): string[] {
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
    expect(shape(moveNotePageInTree(tree, 'c', { kind: 'before', pageId: 'b' }))).toEqual([
      'a', 'a>a1', 'a>a2', 'c', 'b',
    ]);
  });

  it('兄弟の直後に差し込む', () => {
    expect(shape(moveNotePageInTree(tree, 'b', { kind: 'after', pageId: 'c' }))).toEqual([
      'a', 'a>a1', 'a>a2', 'c', 'b',
    ]);
  });

  it('別の段の子として末尾に入る', () => {
    expect(shape(moveNotePageInTree(tree, 'b', { kind: 'into', pageId: 'a' }))).toEqual([
      'a', 'a>a1', 'a>a2', 'a>b', 'c',
    ]);
  });

  it('子孫ごと動く', () => {
    const nested = [node('a', [node('a1', [node('a1x')])]), node('b')];

    expect(shape(moveNotePageInTree(nested, 'a1', { kind: 'into', pageId: 'b' }))).toEqual([
      'a', 'b', 'b>a1', 'b>a1>a1x',
    ]);
  });

  it('元の木は変えない', () => {
    moveNotePageInTree(tree, 'b', { kind: 'into', pageId: 'a' });

    expect(shape(tree)).toEqual(['a', 'a>a1', 'a>a2', 'b', 'c']);
  });

  it('自分自身の中へは動かせない', () => {
    expect(moveNotePageInTree(tree, 'a', { kind: 'into', pageId: 'a' })).toBeNull();
  });

  it('自分の子孫の中へは動かせない', () => {
    // 動かすと木が根から切り離される。サーバーも同じ理由で断るが、
    // 画面が先に動いてから巻き戻るより、動かさないほうが分かりやすい。
    expect(moveNotePageInTree(tree, 'a', { kind: 'into', pageId: 'a1' })).toBeNull();
    expect(moveNotePageInTree(tree, 'a', { kind: 'after', pageId: 'a2' })).toBeNull();
  });

  it('木に無いページは動かせない', () => {
    expect(moveNotePageInTree(tree, 'unknown', { kind: 'into', pageId: 'a' })).toBeNull();
  });

  it.each(['before', 'after', 'into'] as const)(
    '落下先が木に無ければ何もしない（%s）',
    (kind) => {
      // 確かめずに進めると、取り除いたあと差し込む先が見つからず、
      // 動かしたページと子孫が木から消える。
      expect(moveNotePageInTree(tree, 'b', { kind, pageId: 'unknown' })).toBeNull();
    },
  );
});

describe('noteMoveActions', () => {
  const siblings = [node('a'), node('b'), node('c')];

  it('真ん中の行は 4 方向すべてへ動かせる', () => {
    expect(noteMoveActions(siblings, 1, 'parent')).toEqual({
      up: { kind: 'before', pageId: 'a' },
      down: { kind: 'after', pageId: 'c' },
      indent: { kind: 'into', pageId: 'a' },
      outdent: { kind: 'after', pageId: 'parent' },
    });
  });

  it('先頭は上へも内側へも動かせない', () => {
    // 内側へ入れる相手（ひとつ上の兄弟）がいない。
    const actions = noteMoveActions(siblings, 0, 'parent');

    expect(actions.up).toBeNull();
    expect(actions.indent).toBeNull();
    expect(actions.down).toEqual({ kind: 'after', pageId: 'b' });
  });

  it('末尾は下へ動かせない', () => {
    expect(noteMoveActions(siblings, 2, 'parent').down).toBeNull();
  });

  it('最上段は外側へ動かせない', () => {
    // 出る先の段が無い。
    expect(noteMoveActions(siblings, 1, null).outdent).toBeNull();
  });

  it('外側へは親の直後に出る', () => {
    // 親の子ではなく、親の兄弟になる。
    expect(noteMoveActions(siblings, 0, 'parent').outdent).toEqual({
      kind: 'after',
      pageId: 'parent',
    });
  });
});
