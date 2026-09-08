import { FormEvent, useMemo, useState } from 'react';
import { useOidcLogin, useFirebaseAuth, resolveAuthMode, type AuthMode } from '@/features/auth';

export interface PasswordResetPageState {
  readonly mode: AuthMode;
  readonly email: string;
  readonly setEmail: (value: string) => void;
  /** 送信操作を終えたか（成否は問わない。理由は handleSubmit の doc を参照）。 */
  readonly submitted: boolean;
  readonly loading: boolean;
  readonly handleSubmit: (e: FormEvent) => void;
  /** mode === 'unconfigured' のときだけ意味を持つ。 */
  readonly missing: readonly string[];
}

/**
 * PasswordResetPage 用フック。
 *
 * パスワード再設定は GCIP（Firebase）だけの機能。ローカル開発（Dex）は固定 1 アカウント
 * （`docker/idp/config.yaml` の `staticPasswords`）のみで自己登録も再設定も無いため、
 * 画面側でその旨を案内する（このフックは Dex 向けの操作を持たない）。
 */
export function usePasswordResetPage(): PasswordResetPageState {
  const mode = useMemo(() => resolveAuthMode(), []);
  const firebaseAuth = useFirebaseAuth();
  // 操作としては使わない。mode === 'unconfigured' のとき、どちらの設定が
  // 欠けているかを合わせて示すためだけに読む（useLoginPage / useSignupPage と同じ形）。
  const dexLogin = useOidcLogin();

  const [email, setEmail] = useState('');
  const [submitted, setSubmitted] = useState(false);

  /**
   * 送信結果（成功したか、そのメールアドレスのアカウントが存在しなかったか）を
   * **画面に反映しない**。存在確認の結果で表示を変えると、それ自体が
   * 「このメールアドレスは登録されているか」を外部から探る手がかりになる
   * （`shared/lib/auth/firebaseErrorMessage.ts` が wrong-password / user-not-found を
   * 同じ文言にまとめているのと同じ理由）。成功・失敗を問わず同じ案内を出す。
   *
   * **これだけでは列挙対策として完結しない。** GCIP の REST API 自体は、プロジェクトの
   * 「メール列挙保護」設定が無効だと存在確認の結果によって応答（ステータス・本文）を
   * 変える。ここでの一律表示は画面の見た目を揃えるだけで、ネットワークタブを見れば
   * 分かってしまう可能性がある。実際に塞ぐには GCIP コンソール側の設定が要る
   * （フロントのコードだけでは閉じられない）。
   */
  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (!firebaseAuth.available) return;
    firebaseAuth.sendPasswordReset(email).finally(() => setSubmitted(true));
  };

  const missing =
    mode === 'unconfigured'
      ? [...(!firebaseAuth.available ? firebaseAuth.missing : []), ...(!dexLogin.available ? dexLogin.missing : [])]
      : [];

  return {
    mode,
    email,
    setEmail: (value: string) => setEmail(value),
    submitted,
    loading: firebaseAuth.available ? firebaseAuth.loading : false,
    handleSubmit,
    missing,
  };
}
