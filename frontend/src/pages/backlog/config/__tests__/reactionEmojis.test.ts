import { describe, it, expect } from 'vitest';
import { REACTION_EMOJIS } from '../reactionEmojis';

describe('REACTION_EMOJIS', () => {
  it('12 種の固定', () => {
    expect(REACTION_EMOJIS).toHaveLength(12);
  });

  it('重複が無い', () => {
    expect(new Set(REACTION_EMOJIS).size).toBe(REACTION_EMOJIS.length);
  });

  it('backend の上限（UTF-8 で 32 バイト）に必ず収まる', () => {
    const encoder = new TextEncoder();
    for (const emoji of REACTION_EMOJIS) {
      expect(encoder.encode(emoji).length).toBeLessThanOrEqual(32);
    }
  });
});
