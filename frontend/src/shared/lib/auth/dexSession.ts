/**
 * Dex（ローカル開発用の発行者）のセッションをブラウザに保持する。
 *
 * GCIP（本番）は Firebase SDK が自分でセッションを永続化する（IndexedDB）。
 * Dex にはクライアント SDK が無く、`exchangeCodeForToken` で得た ID トークン・
 * refresh_token をこちらで保存しないと、ページを再読み込みするたびに
 * サインインし直しになる（ローカル開発の使い勝手が大きく落ちる）。
 *
 * localStorage を使う（sessionStorage だとタブを閉じるたびに切れる）。本番の
 * トークンではなくローカル発行者のトークンなので、Firebase の既定の永続化先
 * （IndexedDB）と保存場所が違っても実害は無い。
 */

import type { ConfiguredAuth } from './authConfig';
import { refreshDexToken, DexTokenExchangeError } from './oidcAuthUrl';

const STORAGE_KEY = 'dex.session';

/** 期限までの残りがこれを切ったら、使う前に更新しておく（リクエストの途中で切れるのを防ぐ）。 */
const REFRESH_LEEWAY_MS = 60_000;

type StoredSession = {
  idToken: string;
  refreshToken: string | null;
  /** UNIX ミリ秒。 */
  expiresAt: number;
};

export function saveDexSession(idToken: string, refreshToken: string | null, expiresInSeconds: number): void {
  const session: StoredSession = {
    idToken,
    refreshToken,
    expiresAt: Date.now() + expiresInSeconds * 1000,
  };
  localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
}

export function clearDexSession(): void {
  localStorage.removeItem(STORAGE_KEY);
}

function loadDexSession(): StoredSession | null {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as Partial<StoredSession>;
    if (!parsed.idToken || typeof parsed.expiresAt !== 'number') return null;
    return { idToken: parsed.idToken, refreshToken: parsed.refreshToken ?? null, expiresAt: parsed.expiresAt };
  } catch {
    return null;
  }
}

/**
 * 有効な ID トークンを返す。保存が無い／期限切れで更新もできない場合は null。
 *
 * 期限が近い（`REFRESH_LEEWAY_MS` を切っている）ときは、使う前に
 * `refreshDexToken` で更新してから返す。更新に失敗したら保存を消して null にする
 * （中途半端に古い値を残すと、次回もまた失敗する更新を繰り返すだけになる）。
 *
 * `forceRefresh` は、期限内に見えても backend が 401 を返した場合の再試行に使う
 * （backend とこのブラウザの時計がずれている・トークンが失効させられた等、
 * 期限だけでは分からない理由で無効になっているケースを拾う）。
 */
export async function getValidDexIdToken(cfg: ConfiguredAuth, forceRefresh = false): Promise<string | null> {
  const session = loadDexSession();
  if (!session) return null;

  if (!forceRefresh && Date.now() < session.expiresAt - REFRESH_LEEWAY_MS) {
    return session.idToken;
  }

  if (!session.refreshToken) {
    clearDexSession();
    return null;
  }

  try {
    const refreshed = await refreshDexToken(cfg, session.refreshToken);
    saveDexSession(refreshed.idToken, refreshed.refreshToken, refreshed.expiresInSeconds);
    return refreshed.idToken;
  } catch (err) {
    if (err instanceof DexTokenExchangeError) {
      clearDexSession();
      return null;
    }
    throw err;
  }
}

/** 保存されたセッションがあるか（有効期限は問わない）。UIの初期描画判定に使う。 */
export function hasDexSession(): boolean {
  return loadDexSession() !== null;
}
