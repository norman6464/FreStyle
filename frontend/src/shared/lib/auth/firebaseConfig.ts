/**
 * GCIP（Firebase Authentication 互換）の設定を「欠けうるもの」として 1 か所で読む。
 *
 * `shared/lib/auth/authConfig.ts`（Dex 向け）と同じ設計。設定はビルド時にバンドルへ
 * 焼き込まれ、焼き込まれなければ**無い**。判別可能な合併にして、揃っている枝でしか
 * Firebase SDK を初期化できないようにする。
 */

/** Firebase Auth SDK の初期化に必要な設定がすべて揃っている状態。 */
export type ConfiguredFirebaseAuth = {
  readonly status: 'configured';
  readonly apiKey: string;
  readonly authDomain: string;
  readonly projectId: string;
};

/** 設定が欠けている状態。何が欠けているかを持つ（表示ではなく記録・検査のため）。 */
export type UnconfiguredFirebaseAuth = {
  readonly status: 'unconfigured';
  readonly missing: readonly string[];
};

export type FirebaseAuthConfig = ConfiguredFirebaseAuth | UnconfiguredFirebaseAuth;

/**
 * 焼き込まれた設定を読んで、揃っているかを判定する。
 *
 * apiKey / authDomain / projectId の 3 つを必須にする（Firebase Auth SDK の
 * `initializeApp` が実際に使う値。`storageBucket` 等の他プロダクト向け設定は
 * このアプリでは Auth しか使わないため要求しない）。
 */
export function readFirebaseAuthConfig(): FirebaseAuthConfig {
  const env = import.meta.env;
  const apiKey = env.VITE_FIREBASE_API_KEY;
  const authDomain = env.VITE_FIREBASE_AUTH_DOMAIN;
  const projectId = env.VITE_FIREBASE_PROJECT_ID;

  const missing: string[] = [];
  if (!apiKey) missing.push('VITE_FIREBASE_API_KEY');
  if (!authDomain) missing.push('VITE_FIREBASE_AUTH_DOMAIN');
  if (!projectId) missing.push('VITE_FIREBASE_PROJECT_ID');

  if (apiKey && authDomain && projectId) {
    return { status: 'configured', apiKey, authDomain, projectId };
  }
  return { status: 'unconfigured', missing };
}
