import { FirebaseError } from 'firebase/app';

/**
 * Firebase Auth のエラーコードを日本語の利用者向けメッセージへ変換する。
 *
 * `shared/lib/classifyApiError.ts`（axios のエラー用）と同じ考え方。サーバーの
 * 生のエラーコードをそのまま出さず、対応が分かる文言に一度変換する。
 *
 * `auth/wrong-password` / `auth/user-not-found` は現行の Firebase Auth SDK では
 * 列挙攻撃対策として `auth/invalid-credential` に統合されて返る。個別に出し分けると
 * 「そのメールアドレスは登録されていない」ことが分かってしまうため、あえて
 * 1 つの文言にまとめる。
 */
const MESSAGES: Record<string, string> = {
  'auth/invalid-credential': 'メールアドレスまたはパスワードが正しくありません。',
  'auth/wrong-password': 'メールアドレスまたはパスワードが正しくありません。',
  'auth/user-not-found': 'メールアドレスまたはパスワードが正しくありません。',
  'auth/invalid-email': 'メールアドレスの形式が正しくありません。',
  'auth/email-already-in-use': 'このメールアドレスは既に登録されています。',
  'auth/weak-password': 'パスワードは 6 文字以上にしてください。',
  'auth/too-many-requests': '試行回数が多すぎます。しばらく待ってからお試しください。',
  'auth/network-request-failed': 'インターネット接続を確認してください。',
  'auth/user-disabled': 'このアカウントは無効化されています。',
  'auth/popup-closed-by-user': 'ログインがキャンセルされました。',
  'auth/cancelled-popup-request': 'ログインがキャンセルされました。',
  'auth/popup-blocked': 'ポップアップがブロックされました。ブラウザの設定を確認してください。',
};

/**
 * Firebase のエラーを日本語メッセージへ変換する。
 *
 * `FirebaseError` でなければ（想定外のエラー）fallback を返す。
 */
export function classifyFirebaseError(error: unknown, fallback: string): string {
  if (error instanceof FirebaseError) {
    return MESSAGES[error.code] ?? fallback;
  }
  return fallback;
}
