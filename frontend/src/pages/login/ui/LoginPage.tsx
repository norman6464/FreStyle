import { AuthLayout } from '@/widgets/auth-layout';
import PublicHeader from '@/shared/ui/PublicHeader';
import Button from '@/shared/ui/Button';
import SNSSignInButton from '@/shared/ui/SNSSignInButton';
import LinkText from '@/shared/ui/LinkText';
import { AuthUnavailableNotice } from '@/features/auth';
import { CheckCircleIcon } from '@heroicons/react/24/outline';
import { useLoginPage } from '../model/useLoginPage';

/**
 * ログイン画面。
 *
 * メールとパスワードのフォームは置かない。パスワードを受け取るのは発行者の
 * ログイン画面の役目で、アプリが受け取ると、二要素・ロックアウト・パスワードの
 * 強さといった発行者側の守りをすべて素通りする経路を自分で開くことになる。
 *
 * ここは「発行者へ送る」ことだけをする。
 *
 * 設定が揃っていないときは `login.available` が false になり、`start` が
 * 存在しない。押せるボタンを描く経路が型として無いので、「押しても何も
 * 起きない」状態は書こうとしても型検査で落ちる。
 */
export default function LoginPage() {
  const { flashMessage, login } = useLoginPage();
  const loading = login.available && login.loading;

  return (
    <AuthLayout title="ログイン" header={<PublicHeader />}>
      {flashMessage && (
        <p
          role="status"
          className="mb-4 flex items-center justify-center gap-1 rounded-lg border border-emerald-200 bg-emerald-50 p-3 text-center font-medium text-emerald-700"
        >
          <CheckCircleIcon className="h-4 w-4" aria-hidden="true" />
          {flashMessage}
        </p>
      )}

      {login.available && login.errorMessage && (
        <p
          role="alert"
          className="mb-4 rounded-lg border border-rose-200 bg-rose-50 p-3 text-center font-medium text-rose-700"
        >
          {login.errorMessage}
        </p>
      )}

      {!login.available && <AuthUnavailableNotice missing={login.missing} />}

      <Button
        variant="primary"
        fullWidth
        type="button"
        loading={loading}
        disabled={!login.available}
        onClick={() => login.available && login.start()}
      >
        {loading ? 'ログイン画面へ移動しています...' : 'ログインする'}
      </Button>

      <div className="relative my-5">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-surface-3"></div>
        </div>
        <div className="relative flex justify-center text-sm">
          <span className="bg-surface-1 px-2 text-[var(--color-text-muted)]">または</span>
        </div>
      </div>

      <SNSSignInButton
        provider="google"
        disabled={!login.available}
        onClick={() => login.available && login.start('Google')}
      />

      <p className="mt-5 text-center text-sm text-[var(--color-text-muted)]">
        招待された方は招待メールのリンクからログインできます。
        <br />
        アカウントをお持ちでない方は <LinkText to="/signup">アカウントを作成</LinkText>。
      </p>
    </AuthLayout>
  );
}
