import { AuthLayout } from '@/widgets/auth-layout';
import PublicHeader from '@/shared/ui/PublicHeader';
import Button from '@/shared/ui/Button';
import SNSSignInButton from '@/shared/ui/SNSSignInButton';
import LinkText from '@/shared/ui/LinkText';
import { AuthUnavailableNotice, useOidcLogin } from '@/features/auth';

/**
 * アカウント作成画面。
 *
 * ログイン画面と同じく、発行者へ送るだけ。登録画面へ直行させるために
 * screenHint に 'signup' を渡す。
 */
export default function SignupPage() {
  const login = useOidcLogin();

  return (
    <AuthLayout title="アカウントを作成" header={<PublicHeader />}>
      <p className="mb-6 text-center text-sm text-[var(--color-text-muted)]">
        メールアドレスだけで、すぐに使い始められます。
        <br />
        あなた専用のワークスペースが自動で用意されます。
      </p>

      {!login.available && <AuthUnavailableNotice missing={login.missing} />}

      <Button
        variant="primary"
        fullWidth
        type="button"
        loading={login.available && login.loading}
        disabled={!login.available}
        onClick={() => login.available && login.start(undefined, 'signup')}
      >
        メールで始める
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
        onClick={() => login.available && login.start('Google', 'signup')}
      />

      <p className="mt-5 text-center text-sm text-[var(--color-text-muted)]">
        すでにアカウントをお持ちの方は <LinkText to="/login">ログイン</LinkText> へ。
      </p>
    </AuthLayout>
  );
}
