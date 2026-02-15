import { describe, it, expect } from 'vitest';
import { getFieldBorderClass } from '../fieldStyles';

describe('getFieldBorderClass', () => {
  it('エラーありの場合はrose系のクラスを返す', () => {
    const result = getFieldBorderClass(true);
    expect(result).toContain('border-rose-500');
    expect(result).toContain('focus:border-rose-500');
  });

  it('エラーなしの場合はprimary系のクラスを返す', () => {
    const result = getFieldBorderClass(false);
    expect(result).toContain('border-surface-3');
    expect(result).toContain('focus:border-primary-500');
  });
});
