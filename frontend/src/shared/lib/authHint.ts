/**
 * ログイン済みかどうかの「目印」Cookie（FRESTYLE-231）。
 *
 * 目的は配信側（CloudFront）でトップへのアクセスを振り分けること。トップの HTML には
 * 検索エンジン向けに紹介ページの内容が焼き込まれており、ブラウザは JS 起動前にそれを
 * 描画してしまう。そのためログイン済み判定をアプリ内で行っても紹介ページのちらつきは
 * 消せない。配信の段階で判断できるよう、ブラウザに読める形の目印を置く。
 *
 * 秘密情報は入れない（値は "1" のみ）。認証の判定は従来どおりサーバー側で行うため、
 * この Cookie を偽装しても権限は得られない（転送先で通常の認証チェックが走る）。
 */

/** 目印 Cookie の名前。CloudFront Function 側と一致させること。 */
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
  // 目印の用途（自サイトのトップを開いたときに判定する）にはこれで足りる。
  const secure = isSecureContext() ? '; secure' : '';
  document.cookie = `${AUTH_HINT_COOKIE}=1; path=/; max-age=${MAX_AGE_SECONDS}; samesite=lax${secure}`;
}

/** ログアウト時や認証切れを検知したときに目印を消す。 */
export function clearAuthHint(): void {
  if (typeof document === 'undefined') return;
  const secure = isSecureContext() ? '; secure' : '';
  document.cookie = `${AUTH_HINT_COOKIE}=; path=/; max-age=0; samesite=lax${secure}`;
}

/** 目印の有無。テストと、必要になった場合のアプリ内判定用。 */
export function hasAuthHint(): boolean {
  if (typeof document === 'undefined') return false;
  return document.cookie.split('; ').some((c) => c.startsWith(`${AUTH_HINT_COOKIE}=1`));
}
