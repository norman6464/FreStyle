import { describe, it, expect } from 'vitest';
import { Callout } from '../CalloutExtension';

describe('CalloutExtension', () => {
  it('ノード名がcalloutである', () => {
    expect(Callout.name).toBe('callout');
  });

  it('グループがblockである', () => {
    expect(Callout.config.group).toBe('block');
  });

  it('contentがblock+である', () => {
    expect(Callout.config.content).toBe('block+');
  });

  it('デフォルト属性にtypeとemojiがある', () => {
    const attrs = Callout.config.addAttributes?.();
    expect(attrs).toHaveProperty('type');
    expect(attrs).toHaveProperty('emoji');
    expect(attrs.type.default).toBe('info');
    expect(attrs.emoji.default).toBe('💡');
  });
});
