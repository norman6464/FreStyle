import { describe, it, expect } from 'vitest';
import { summarizeReactions } from '../summarizeReactions';

describe('summarizeReactions', () => {
  it('空なら空配列', () => {
    expect(summarizeReactions([], 1)).toEqual([]);
  });

  it('絵文字ごとに件数を畳み、初出順で並べる', () => {
    const reactions = [
      { userId: 1, emoji: '👍' },
      { userId: 2, emoji: '🎉' },
      { userId: 3, emoji: '👍' },
    ];
    expect(summarizeReactions(reactions, 99)).toEqual([
      { emoji: '👍', count: 2, mine: false },
      { emoji: '🎉', count: 1, mine: false },
    ]);
  });

  it('自分が押している絵文字は mine=true', () => {
    const reactions = [
      { userId: 1, emoji: '👍' },
      { userId: 2, emoji: '👍' },
    ];
    expect(summarizeReactions(reactions, 2)).toEqual([{ emoji: '👍', count: 2, mine: true }]);
  });

  it('currentUserId が null なら全部 mine=false（誰が自分か分からないので安全側）', () => {
    const reactions = [{ userId: 1, emoji: '👍' }];
    expect(summarizeReactions(reactions, null)).toEqual([{ emoji: '👍', count: 1, mine: false }]);
  });
});
