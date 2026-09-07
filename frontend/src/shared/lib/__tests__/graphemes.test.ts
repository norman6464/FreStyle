import { describe, it, expect } from 'vitest';
import { countGraphemes } from '../graphemes';

describe('countGraphemes', () => {
  it('空文字は 0', () => {
    expect(countGraphemes('')).toBe(0);
  });

  it('素の ASCII 1 文字は 1', () => {
    expect(countGraphemes('A')).toBe(1);
  });

  it('サロゲートペアの絵文字（📘）は 1', () => {
    expect(countGraphemes('📘')).toBe(1);
  });

  it('ZWJ で連結した家族の絵文字は見た目どおり 1', () => {
    expect(countGraphemes('👨‍👩‍👧')).toBe(1);
  });

  it('国旗（地域指示記号のペア）は 1', () => {
    expect(countGraphemes('🇯🇵')).toBe(1);
  });

  it('ASCII 2 文字は 2（1 文字だけ入力してくださいの判定対象）', () => {
    expect(countGraphemes('ab')).toBe(2);
  });
});
