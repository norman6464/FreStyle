/**
 * ログイン済みかどうかの「目印」Cookie。
 *
 * 元は配信側（CloudFront）でトップへのアクセスを振り分けるために置いていた。トップの
 * HTML に紹介ページが焼き込まれており、JS 起動前に描画されてしまうためアプリ内の判定
 * では間に合わなかった。公開ランディングと焼き込みを廃止したのでその用途は無くなり、
 * 今は Redux ストアが立ち上がる前に「ログインしていそうか」を知りたい画面が使う
 * （404 ページが戻り先の導線を出し分けるなど）。
 *
 * 秘密情報は入れない（値は "1" のみ）。認証の判定は従来どおりサーバー側で行うため、
 * この Cookie を偽装しても権限は得られない。
 */

import { getApiError } from './classifyApiError';

/** 目印 Cookie の名前。 */
export const AUTH_HINT_COOKIE = 'fs_signed_in';

/** 認証 Cookie（refresh token）と同じ 30 日。 */
const MAX_AGE_SECONDS = 30 * 24 * 60 * 60;

function isSecureContext(): boolean {
  return typeof window !== 'undefined' && window.location.protocol === 'https:';
}

/** ログインが成立したときに目印を置く。 */
export function setAuthHint(): void {
  if (typeof document === 'undefined') return;
  // SameSite=Lax: 通常の遷移では送られ、他サイトからの POST 等では送られない。
  // 目印の用途（自サイトを開いたときの表示判定）にはこれで足りる。
  const secure = isSecureContext() ? '; secure' : '';
  document.cookie = `${AUTH_HINT_COOKIE}=1; path=/; max-age=${MAX_AGE_SECONDS}; samesite=lax${secure}`;
}

/**
 * 認証切れが「確定した」ときだけ目印を消すためのヘルパ。
 *
 * 通信断や 5xx を認証切れ扱いにして消してしまうと、セッションは生きているのに
 * ログインしていない扱いの表示になる。サーバーが明確に 401 / 403 を返した場合のみ消す。
 */
export function clearAuthHintIfUnauthenticated(error: unknown): void {
  const { status } = getApiError(error);
  if (status === 401 || status === 403) {
    clearAuthHint();
  }
}

/** ログアウト時や認証切れを検知したときに目印を消す。 */
export function clearAuthHint(): void {
  if (typeof document === 'undefined') return;
  const secure = isSecureContext() ? '; secure' : '';
  document.cookie = `${AUTH_HINT_COOKIE}=; path=/; max-age=0; samesite=lax${secure}`;
}

/**
 * 目印の有無。ストアが立ち上がる前の表示判定に使う。
 *
 * 区切りは `;` で分割して trim する（`; ` 固定で分けるとスペースが無い環境で取りこぼす）。
 * 値は完全一致で見る（`startsWith` だと `fs_signed_in=10` も「ログイン済み」に見えてしまう）。
 */
export function hasAuthHint(): boolean {
  if (typeof document === 'undefined') return false;
  return document.cookie
    .split(';')
    .map((c) => c.trim())
    .includes(`${AUTH_HINT_COOKIE}=1`);
}
