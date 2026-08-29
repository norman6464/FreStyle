import { AuthLayout } from '@/widgets/auth-layout';
import PublicHeader from '@/shared/ui/PublicHeader';
import Button from '@/shared/ui/Button';
import SNSSignInButton from '@/shared/ui/SNSSignInButton';
import LinkText from '@/shared/ui/LinkText';
import { getCognitoAuthUrl } from '@/features/auth';

export default function SignupPage() {
  return (
    <AuthLayout title="アカウントを作成" header={<PublicHeader />}>
      <p className="mb-6 text-center text-sm text-[var(--color-text-muted)]">
        メールアドレスだけで、すぐに使い始められます。
        <br />
        あなた専用のワークスペースが自動で用意されます。
      </p>

      <Button
        variant="primary"
        fullWidth
        type="button"
        onClick={() => {
          window.location.href = getCognitoAuthUrl(undefined, 'signup');
        }}
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
        onClick={() => {
          window.location.href = getCognitoAuthUrl('Google', 'signup');
        }}
      />

      <p className="mt-5 text-center text-sm text-[var(--color-text-muted)]">
        すでにアカウントをお持ちの方は <LinkText to="/login">ログイン</LinkText> へ。
      </p>
    </AuthLayout>
  );
}
