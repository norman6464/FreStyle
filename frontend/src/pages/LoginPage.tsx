import AuthLayout from '../components/AuthLayout';
import PublicHeader from '../components/PublicHeader';
import SNSSignInButton from '../components/SNSSignInButton';
import LinkText from '../components/LinkText';
import { getCognitoAuthUrl } from '../utils/auth';
import { useLoginPage } from '../hooks/useLoginPage';
import { CheckCircleIcon } from '@heroicons/react/24/outline';

export default function LoginPage() {
  const { flashMessage } = useLoginPage();

  return (
    <AuthLayout title="ログイン" header={<PublicHeader />}>
      {flashMessage && (
        <p
          role="status"
          className="mb-4 flex items-center justify-center gap-1 rounded-lg bg-emerald-900/30 p-3 text-center font-medium text-emerald-400"
        >
          <CheckCircleIcon className="h-4 w-4" aria-hidden="true" />
          {flashMessage}
        </p>
      )}

      {/* ログインは Cognito Hosted UI に一本化(メール/パスワード + ソーシャルを Hosted UI 側で選択)。 */}
      <button
        type="button"
        onClick={() => {
          window.location.href = getCognitoAuthUrl();
        }}
        className="w-full rounded-lg bg-primary-500 py-2.5 font-medium text-white transition-colors duration-150 hover:bg-primary-600 active:bg-primary-700"
      >
        ログイン / 新規登録
      </button>

      <div className="relative my-5">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-surface-3"></div>
        </div>
        <div className="relative flex justify-center text-sm">
          <span className="bg-surface-1 px-2 text-[var(--color-text-muted)]">または</span>
        </div>
      </div>

      {/* Google で直接ログイン */}
      <SNSSignInButton
        provider="google"
        onClick={() => {
          window.location.href = getCognitoAuthUrl('Google');
        }}
      />

      <p className="mt-5 text-center text-sm text-[var(--color-text-muted)]">
        招待された方は招待メールのリンクからログインしてください。
        <br />
        企業の方は <LinkText to="/company-application">利用申請</LinkText> から。
      </p>
    </AuthLayout>
  );
}
