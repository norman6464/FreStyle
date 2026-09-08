import { initializeApp, type FirebaseApp } from 'firebase/app';
import { getAuth, type Auth } from 'firebase/auth';
import { readFirebaseAuthConfig, type ConfiguredFirebaseAuth } from './firebaseConfig';

/**
 * Firebase App / Auth のインスタンスを、設定が揃っているときだけ 1 度だけ作る。
 *
 * `initializeApp` を欠けた設定（空文字の apiKey 等）で呼ぶと、実際にサインインを
 * 試みるまでエラーにならず、押して初めて謎のエラーになる。呼び出し側
 * （`useFirebaseAuth`）は `readFirebaseAuthConfig` の `unconfigured` 枝で
 * このモジュール自体を使わないため、ここでは「呼ばれた時点で揃っている」前提でよい。
 */
let cached: { app: FirebaseApp; auth: Auth } | null = null;

function initialize(config: ConfiguredFirebaseAuth): { app: FirebaseApp; auth: Auth } {
  const app = initializeApp({
    apiKey: config.apiKey,
    authDomain: config.authDomain,
    projectId: config.projectId,
  });
  return { app, auth: getAuth(app) };
}

/**
 * Firebase Auth インスタンスを返す。設定が欠けていれば null
 * （呼び出し側は `useFirebaseAuth` の `available:false` 枝で弾いているはずで、
 * ここで null が返るのは想定外の呼び出し順のときだけ）。
 */
export function getFirebaseAuth(): Auth | null {
  if (cached) return cached.auth;
  const config = readFirebaseAuthConfig();
  if (config.status !== 'configured') return null;
  cached = initialize(config);
  return cached.auth;
}

/** テスト用: キャッシュされた Firebase App/Auth を破棄する。 */
export function resetFirebaseAppForTests(): void {
  cached = null;
}
