import { describe, it, expect, beforeEach } from 'vitest';
import { AUTH_HINT_COOKIE, setAuthHint, clearAuthHint, hasAuthHint } from '../authHint';

/**
 * ログイン済みの目印 Cookie（FRESTYLE-231）。
 * これが正しく置かれ・消えることが、配信側（CloudFront）のトップ振り分けの前提になる。
 */
describe('authHint', () => {
  beforeEach(() => {
    clearAuthHint();
  });

  it('目印が無い状態では false', () => {
    expect(hasAuthHint()).toBe(false);
  });

  it('setAuthHint で目印が置かれる', () => {
    setAuthHint();
    expect(hasAuthHint()).toBe(true);
    expect(document.cookie).toContain(`${AUTH_HINT_COOKIE}=1`);
  });

  it('clearAuthHint で目印が消える', () => {
    setAuthHint();
    clearAuthHint();
    expect(hasAuthHint()).toBe(false);
  });

  it('秘密情報を含まない（値は 1 のみ）', () => {
    setAuthHint();
    const entry = document.cookie
      .split('; ')
      .find((c) => c.startsWith(`${AUTH_HINT_COOKIE}=`));
    expect(entry).toBe(`${AUTH_HINT_COOKIE}=1`);
  });
});
