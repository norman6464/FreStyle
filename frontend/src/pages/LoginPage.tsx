import AuthLayout from '../components/AuthLayout';
import PublicHeader from '../components/PublicHeader';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import SNSSignInButton from '../components/SNSSignInButton';
import LinkText from '../components/LinkText';
import { getCognitoAuthUrl } from '../utils/auth';
import { useLoginPage } from '../hooks/useLoginPage';
import { XCircleIcon, CheckCircleIcon } from '@heroicons/react/24/outline';

export default function LoginPage() {
  const { form, loginMessage, flashMessage, loading, handleLogin, handleChange } = useLoginPage();

  return (
    <AuthLayout title="ログイン" header={<PublicHeader />}>
      {/* フラッシュメッセージ（ログアウト後・招待受諾後などの成功通知） */}
      {flashMessage && (
        <p
          role="status"
          className="mb-4 flex items-center justify-center gap-1 rounded-lg bg-emerald-900/30 p-3 text-center font-medium text-emerald-400"
        >
          <CheckCircleIcon className="h-4 w-4" aria-hidden="true" />
          {flashMessage}
        </p>
      )}

      {/* エラーメッセージ */}
      {loginMessage?.type === 'error' && (
        <p
          role="alert"
          className="mb-4 flex items-center justify-center gap-1 rounded-lg bg-rose-900/30 p-3 text-center font-medium text-rose-400"
        >
          <XCircleIcon className="h-4 w-4" aria-hidden="true" />
          {loginMessage.text}
        </p>
      )}

      {/* メール・パスワードフォーム（Cognito USER_PASSWORD_AUTH） */}
      <form onSubmit={handleLogin} aria-label="ログインフォーム">
        <InputField
          label="メールアドレス"
          name="email"
          type="email"
          value={form.email}
          onChange={handleChange}
          disabled={loading}
        />
        <InputField
          label="パスワード"
          name="password"
          type="password"
          value={form.password}
          onChange={handleChange}
          disabled={loading}
        />
        <PrimaryButton type="submit" loading={loading}>
          {loading ? 'ログイン中...' : 'ログイン'}
        </PrimaryButton>
      </form>

      {/* 区切り線 */}
      <div className="relative my-5">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-surface-3"></div>
        </div>
        <div className="relative flex justify-center text-sm">
          <span className="bg-surface-1 px-2 text-[var(--color-text-muted)]">または</span>
        </div>
      </div>

      {/* Google で直接ログイン（Hosted UI） */}
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
