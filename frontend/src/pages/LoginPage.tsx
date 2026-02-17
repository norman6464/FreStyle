import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import SNSSignInButton from '../components/SNSSignInButton';
import LinkText from '../components/LinkText';
import { getCognitoAuthUrl } from '../utils/auth';
import { useLoginPage } from '../hooks/useLoginPage';
import { XCircleIcon } from '@heroicons/react/24/outline';

export default function LoginPage() {
  const { form, loginMessage, flashMessage, loading, handleLogin, handleChange } = useLoginPage();

  return (
    <AuthLayout
      title="ログイン"
      footer={
        <p className="text-sm text-[var(--color-text-muted)]">
          アカウントをお持ちでない方{' '}
          <LinkText to="/signup">新規登録</LinkText>
        </p>
      }
    >
      {/* flash Message */}
      <div>
        {flashMessage && (
          <p className="text-emerald-400 text-center mb-4 p-3 bg-emerald-900/30 rounded-lg font-medium">
            {flashMessage}
          </p>
        )}
        {loginMessage?.type === 'error' && (
          <p className="text-rose-400 text-center mb-4 p-3 bg-rose-900/30 rounded-lg font-medium flex items-center justify-center gap-1">
            <XCircleIcon className="w-4 h-4" />
            {loginMessage.text}
          </p>
        )}
      </div>

      {/* Googleログイン */}
      <SNSSignInButton
        provider="google"
        onClick={() => {
          window.location.href = getCognitoAuthUrl('Google');
        }}
      />

      {/* 区切り線 */}
      <div className="relative my-5">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-surface-3"></div>
        </div>
        <div className="relative flex justify-center text-sm">
          <span className="px-2 bg-surface-1 text-[var(--color-text-muted)]">
            または
          </span>
        </div>
      </div>

      {/* メール・パスワードフォーム */}
      <form onSubmit={handleLogin}>
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

      {/* パスワードリセットリンク */}
      <div className="mt-4 text-center">
        <LinkText to="/forgot-password">パスワードを忘れた方</LinkText>
      </div>
    </AuthLayout>
  );
}
