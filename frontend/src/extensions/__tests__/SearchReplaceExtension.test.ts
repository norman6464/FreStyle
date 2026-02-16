import { describe, it, expect } from 'vitest';
import { SearchReplaceExtension } from '../SearchReplaceExtension';

describe('SearchReplaceExtension', () => {
  it('拡張名がsearchReplaceである', () => {
    expect(SearchReplaceExtension.name).toBe('searchReplace');
  });

  it('configureで拡張が作成できる', () => {
    const ext = SearchReplaceExtension.configure({});
    expect(ext).toBeDefined();
  });
});
