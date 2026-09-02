import { AuthLayout } from '@/widgets/auth-layout';
import PublicHeader from '@/shared/ui/PublicHeader';
import Button from '@/shared/ui/Button';
import SNSSignInButton from '@/shared/ui/SNSSignInButton';
import LinkText from '@/shared/ui/LinkText';
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
 */
export default function LoginPage() {
  const { flashMessage, errorMessage, loading, startLogin } = useLoginPage();

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

      {errorMessage && (
        <p
          role="alert"
          className="mb-4 rounded-lg border border-rose-200 bg-rose-50 p-3 text-center font-medium text-rose-700"
        >
          {errorMessage}
        </p>
      )}

      <Button
        variant="primary"
        fullWidth
        type="button"
        loading={loading}
        onClick={() => startLogin()}
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

      <SNSSignInButton provider="google" onClick={() => startLogin('Google')} />

      <p className="mt-5 text-center text-sm text-[var(--color-text-muted)]">
        招待された方は招待メールのリンクからログインできます。
        <br />
        アカウントをお持ちでない方は <LinkText to="/signup">アカウントを作成</LinkText>。
      </p>
    </AuthLayout>
  );
}
