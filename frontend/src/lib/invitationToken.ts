/**
 * 招待マジックリンクで受領した UUID トークンを sessionStorage で持ち回るためのヘルパー。
 *
 * フロー:
 *   1. /invitations/accept?token=<UUID> を踏むと AcceptInvitationPage が saveInvitationToken
 *   2. ユーザーが「ログインへ進む」ボタンで /login → Cognito Hosted UI へリダイレクト
 *   3. /login/callback (LoginCallback) が consumeInvitationToken で取り出して
 *      POST /auth/cognito/callback の body に invitationToken として含めて送る
 *
 * sessionStorage を選んだ理由:
 *   - localStorage だと別タブの誤読込みリスクがある
 *   - cookie だと CSRF / SameSite まわりが複雑
 *   - URL hash で渡し続けると history に残って他ユーザーが復元できてしまう
 *
 * consume 後は必ず sessionStorage から削除して再利用を防ぐ。
 */

const STORAGE_KEY = 'frestyle.invitationToken';

export function saveInvitationToken(token: string): void {
  if (!token) return;
  try {
    sessionStorage.setItem(STORAGE_KEY, token);
  } catch {
    // private mode などで sessionStorage が使えないケース。callback での
    // 招待ゲート fallback (email 一致) に任せる。
  }
}

export function readInvitationToken(): string | null {
  try {
    return sessionStorage.getItem(STORAGE_KEY);
  } catch {
    return null;
  }
}

export function clearInvitationToken(): void {
  try {
    sessionStorage.removeItem(STORAGE_KEY);
  } catch {
    // ignore
  }
}

/**
 * consumeInvitationToken は read + clear を 1 操作で行う。
 * callback で 1 度だけ使うため、意図せず再送されないように使い切りで削除する。
 */
export function consumeInvitationToken(): string | null {
  const t = readInvitationToken();
  if (t) clearInvitationToken();
  return t;
}
