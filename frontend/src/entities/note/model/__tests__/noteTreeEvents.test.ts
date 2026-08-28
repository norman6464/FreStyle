import { describe, it, expect, vi } from 'vitest';
import { emitNoteTreeEvent, subscribeNoteTreeEvents } from '../noteTreeEvents';
import type { NotePage } from '../types';

const page: NotePage = {
  id: 'p1',
  spaceId: 's1',
  title: '設計メモ',
  createdByUserId: 1,
  createdAt: '2026-08-01T00:00:00Z',
  updatedAt: '2026-08-01T00:00:00Z',
};

describe('noteTreeEvents', () => {
  it('購読者へイベントが届き、解除後は届かない', () => {
    const listener = vi.fn();
    const unsubscribe = subscribeNoteTreeEvents(listener);

    emitNoteTreeEvent({ type: 'page-created', page });
    expect(listener).toHaveBeenCalledWith({ type: 'page-created', page });

    unsubscribe();
    emitNoteTreeEvent({ type: 'page-renamed', page });
    expect(listener).toHaveBeenCalledTimes(1);
  });

  it('購読者がいなくても emit は失敗しない', () => {
    expect(() => emitNoteTreeEvent({ type: 'page-created', page })).not.toThrow();
  });

  it('emit の最中にある購読者が解除しても、残りの購読者へ届く', () => {
    // 走査中の Set 変更で他の購読者が飛ばされないこと（写しを回している証拠）。
    const second = vi.fn();
    const first = vi.fn(() => unsubscribeFirst());
    const unsubscribeFirst = subscribeNoteTreeEvents(first);
    const unsubscribeSecond = subscribeNoteTreeEvents(second);

    emitNoteTreeEvent({ type: 'page-created', page });

    expect(first).toHaveBeenCalledTimes(1);
    expect(second).toHaveBeenCalledTimes(1);
    unsubscribeSecond();
  });
});
