import { describe, it, expect } from 'vitest';
import { buildCommentTree } from '../buildCommentTree';
import type { TicketComment } from '@/entities/ticket';

function comment(over: Partial<TicketComment> & { id: string }): TicketComment {
  return {
    parentCommentId: null,
    author: { userId: 1, name: '田中 太郎' },
    body: [{ kind: 'text', text: over.id }],
    edited: false,
    reactions: [],
    createdAt: `2026-09-10T00:00:0${over.id.length}Z`,
    updatedAt: '',
    ...over,
  };
}

describe('buildCommentTree', () => {
  it('空配列は空配列', () => {
    expect(buildCommentTree([])).toEqual([]);
  });

  it('幹だけを幹の列にする', () => {
    const a = comment({ id: 'a' });
    const b = comment({ id: 'b' });
    const threads = buildCommentTree([a, b]);
    expect(threads).toEqual([
      { root: a, replies: [] },
      { root: b, replies: [] },
    ]);
  });

  it('幹への直接の返信は宛先なしで返信列に入る', () => {
    const root = comment({ id: 'root' });
    const reply = comment({ id: 'r1', parentCommentId: 'root' });
    const threads = buildCommentTree([root, reply]);
    expect(threads).toEqual([{ root, replies: [{ comment: reply, replyToId: null }] }]);
  });

  it('返信への返信（孫）は幹の返信列へ平坦化され、直接の宛先を持つ', () => {
    const root = comment({ id: 'root' });
    const mid = comment({ id: 'mid', parentCommentId: 'root' });
    const grandchild = comment({ id: 'gc', parentCommentId: 'mid' });
    const threads = buildCommentTree([root, mid, grandchild]);
    expect(threads).toEqual([
      {
        root,
        replies: [
          { comment: mid, replyToId: null },
          { comment: grandchild, replyToId: 'mid' },
        ],
      },
    ]);
  });

  it('曾孫まで平坦化しても常に直接の親を宛先にする', () => {
    const root = comment({ id: 'root' });
    const a = comment({ id: 'a', parentCommentId: 'root' });
    const b = comment({ id: 'b', parentCommentId: 'a' });
    const c = comment({ id: 'c', parentCommentId: 'b' });
    const threads = buildCommentTree([root, a, b, c]);
    expect(threads[0].replies.map((r) => [r.comment.id, r.replyToId])).toEqual([
      ['a', null],
      ['b', 'a'],
      ['c', 'b'],
    ]);
  });

  it('親だけ削除された孤児は自分自身を幹に格上げする', () => {
    // 'missing' は comments 配列に存在しない（削除済み）。
    const orphan = comment({ id: 'orphan', parentCommentId: 'missing' });
    const threads = buildCommentTree([orphan]);
    expect(threads).toEqual([{ root: orphan, replies: [] }]);
  });

  it('孤児の子孫は孤児自身へ平坦化される（子孫まで個別に幹へ昇格しない）', () => {
    const orphan = comment({ id: 'orphan', parentCommentId: 'missing' });
    const child = comment({ id: 'child', parentCommentId: 'orphan' });
    const grandchild = comment({ id: 'gc', parentCommentId: 'child' });
    const threads = buildCommentTree([orphan, child, grandchild]);
    expect(threads).toEqual([
      {
        root: orphan,
        replies: [
          { comment: child, replyToId: null },
          { comment: grandchild, replyToId: 'child' },
        ],
      },
    ]);
  });

  it('循環（parentCommentId は作成時に固定で既存の発言しか指せないため実データでは作れないはずの形）でも無限ループにならず有限で終わる', () => {
    const a = comment({ id: 'a', parentCommentId: 'b' });
    const b = comment({ id: 'b', parentCommentId: 'a' });
    // 2 者が互いを見て別々に「自分が幹」と判断するため、この形だけは重複して現れうる
    // （どちらも失われはしない）。実データでは作れない形なので、これ以上は追わない。
    const threads = buildCommentTree([a, b]);
    expect(threads.length).toBeGreaterThan(0);
    expect(threads.length).toBeLessThanOrEqual(2);
  });

  it('複数の幹と返信が混在しても、それぞれの幹の下に正しく振り分けられる', () => {
    const root1 = comment({ id: 'root1' });
    const root2 = comment({ id: 'root2' });
    const reply1 = comment({ id: 'reply1', parentCommentId: 'root1' });
    const reply2 = comment({ id: 'reply2', parentCommentId: 'root2' });
    const threads = buildCommentTree([root1, root2, reply1, reply2]);
    expect(threads).toEqual([
      { root: root1, replies: [{ comment: reply1, replyToId: null }] },
      { root: root2, replies: [{ comment: reply2, replyToId: null }] },
    ]);
  });

  it('幹の並びは配列に現れた順（backend の作成日時昇順）を保つ', () => {
    const root1 = comment({ id: 'root1' });
    const root2 = comment({ id: 'root2' });
    const threads = buildCommentTree([root1, root2]);
    expect(threads.map((t) => t.root.id)).toEqual(['root1', 'root2']);
  });
});
