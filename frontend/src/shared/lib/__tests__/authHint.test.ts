import { describe, it, expect, beforeEach } from 'vitest';
import { AxiosError, AxiosHeaders } from 'axios';
import {
  AUTH_HINT_COOKIE,
  setAuthHint,
  clearAuthHint,
  clearAuthHintIfUnauthenticated,
  hasAuthHint,
} from '../authHint';

// 指定ステータスの AxiosError を作る。status 未指定は通信断（レスポンス無し）。
function axiosErrorWith(status?: number): AxiosError {
  const err = new AxiosError('failed');
  if (status !== undefined) {
    err.response = {
      status,
      statusText: '',
      data: {},
      headers: {},
      config: { headers: new AxiosHeaders() },
    };
  }
  return err;
}

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
      .split(';')
      .map((c) => c.trim())
      .find((c) => c.startsWith(`${AUTH_HINT_COOKIE}=`));
    expect(entry).toBe(`${AUTH_HINT_COOKIE}=1`);
  });

  it('似た名前の値（fs_signed_in=10）は目印として扱わない', () => {
    document.cookie = `${AUTH_HINT_COOKIE}=10; path=/`;
    expect(hasAuthHint()).toBe(false);
    document.cookie = `${AUTH_HINT_COOKIE}=; path=/; max-age=0`;
  });

  // 通信断や 5xx で目印を消すと、セッションは生きているのに次回トップの振り分けが
  // 効かなくなる（＝ちらつきが再発する）。認証切れが確定したときだけ消す。
  describe('clearAuthHintIfUnauthenticated', () => {
    it.each([401, 403])('%d は認証切れとして目印を消す', (status) => {
      setAuthHint();
      clearAuthHintIfUnauthenticated(axiosErrorWith(status));
      expect(hasAuthHint()).toBe(false);
    });

    it.each([500, 502, 503])('%d では目印を消さない', (status) => {
      setAuthHint();
      clearAuthHintIfUnauthenticated(axiosErrorWith(status));
      expect(hasAuthHint()).toBe(true);
    });

    it('通信断（レスポンス無し）では目印を消さない', () => {
      setAuthHint();
      clearAuthHintIfUnauthenticated(axiosErrorWith());
      expect(hasAuthHint()).toBe(true);
    });
  });
});
