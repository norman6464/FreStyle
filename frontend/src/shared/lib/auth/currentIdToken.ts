/**
 * 「いまサインインしている人」まわりの操作（ID トークン取得・状態購読・サインアウト）を、
 * 発行者の違いを問わず 1 か所に閉じる。呼び出し側（axios の interceptor・
 * AuthInitializer・useAuth のログアウト）は GCIP（Firebase）とローカルの Dex の
 * どちらが有効かを意識しなくてよい。
 *
 * 優先順位は Firebase → Dex → 未設定。両方の設定が揃うことは通常無い
 * （本番は Firebase 設定だけを、ローカルは Dex 設定だけをビルドに焼き込む）が、
 * 両方揃っていた場合は本番向けの Firebase を優先する。
 */

import { onIdTokenChanged, signOut as firebaseSignOut, type User } from 'firebase/auth';
import { readFirebaseAuthConfig } from './firebaseConfig';
import { readAuthConfig } from './authConfig';
import { getFirebaseAuth } from './firebaseApp';
import { getValidDexIdToken, hasDexSession, clearDexSession } from './dexSession';

export type AuthMode = 'firebase' | 'dex' | 'unconfigured';

export function resolveAuthMode(): AuthMode {
  if (readFirebaseAuthConfig().status === 'configured') return 'firebase';
  if (readAuthConfig().status === 'configured') return 'dex';
  return 'unconfigured';
}

/**
 * いま有効な ID トークンを返す。サインインしていなければ null。
 *
 * `forceRefresh` は 401 を受けたときの再試行に使う。期限的には有効に見えても
 * backend 側の理由（時計のずれ・失効）で無効になっているケースを拾うため、
 * 期限チェックを飛ばして必ず更新を試みる。
 */
export async function getCurrentIdToken(forceRefresh = false): Promise<string | null> {
  const mode = resolveAuthMode();

  if (mode === 'firebase') {
    const auth = getFirebaseAuth();
    const user = auth?.currentUser;
    if (!user) return null;
    return user.getIdToken(forceRefresh);
  }

  if (mode === 'dex') {
    const cfg = readAuthConfig();
    if (cfg.status !== 'configured') return null;
    return getValidDexIdToken(cfg, forceRefresh);
  }

  return null;
}

/**
 * サインイン状態が変わるたびに呼ばれるコールバックを登録する。
 *
 * Firebase は `onIdTokenChanged` がサインイン・サインアウト・トークン自動更新の
 * たびに非同期で発火する（初回はマウント時、SDK が永続化済みセッションを
 * 読み終えたタイミングで一度呼ばれる）。
 *
 * Dex にはこの種の購読の仕組みが無い（サーバー側 Cookie に頼らずブラウザだけで
 * 完結する SDK が無いため）。localStorage に保存したセッションの有無を
 * マウント時に一度だけ確認し、その結果を 1 回だけコールバックへ渡す
 * （ローカル開発は元々この一発確認方式で動いていた。差分はトークンの出し先が
 * Cookie から Bearer に変わっただけ）。
 *
 * 戻り値は購読解除関数（Firebase 版のみ意味を持つ。Dex 版は一度呼んで終わりなので
 * 呼んでも何も起きない no-op を返す）。
 */
export function subscribeAuthState(callback: (signedIn: boolean) => void): () => void {
  const mode = resolveAuthMode();

  if (mode === 'firebase') {
    const auth = getFirebaseAuth();
    if (!auth) {
      callback(false);
      return () => {};
    }
    return onIdTokenChanged(auth, (user: User | null) => {
      callback(user !== null);
    });
  }

  if (mode === 'dex') {
    callback(hasDexSession());
    return () => {};
  }

  callback(false);
  return () => {};
}

/**
 * いま有効な発行者からサインアウトする。
 *
 * Firebase はクライアント側の永続化を消すだけ（サーバー側セッションという概念を
 * 持たない）。Dex はこのブラウザが保持しているセッション（`dexSession.ts`）を
 * 消すだけで、発行者側のセッションは残る（ローカル開発の割り切り。本番の
 * GCIP には同種の懸念が無いため、発行者側セッション終了 URL のような仕組みは
 * どちらの発行者にも実装しない）。
 */
export async function signOutCurrentProvider(): Promise<void> {
  const mode = resolveAuthMode();

  if (mode === 'firebase') {
    const auth = getFirebaseAuth();
    if (auth) await firebaseSignOut(auth);
    return;
  }

  if (mode === 'dex') {
    clearDexSession();
    return;
  }
}
