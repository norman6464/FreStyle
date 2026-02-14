import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import SNSSignInButton from '../components/SNSSignInButton';
import LinkText from '../components/LinkText';
import { getCognitoAuthUrl } from '../utils/auth';
import { useLoginPage } from '../hooks/useLoginPage';
import { XCircleIcon } from '@heroicons/react/24/outline';

export default function LoginPage() {
  const { form, loginMessage, flashMessage, handleLogin, handleChange } = useLoginPage();

  return (
    <AuthLayout>
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
      <h2 className="text-3xl font-bold mb-2 text-center text-[var(--color-text-primary)]">
        ログイン
      </h2>
      <p className="text-center text-[var(--color-text-muted)] text-sm mb-6">
        アカウントにアクセスしてください
      </p>
      <form onSubmit={handleLogin}>
        <InputField
          label="メールアドレス"
          name="email"
          type="email"
          value={form.email}
          onChange={handleChange}
        />
        <InputField
          label="パスワード"
          name="password"
          type="password"
          value={form.password}
          onChange={handleChange}
        />
        <PrimaryButton type="submit">ログイン</PrimaryButton>
      </form>
      <div className="flex justify-between items-center mt-6 text-sm">
        <LinkText to="/forgot-password">パスワードをお忘れですか？</LinkText>
        <LinkText to="/signup">アカウント作成</LinkText>
      </div>
      <div className="relative my-6">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-surface-3"></div>
        </div>
        <div className="relative flex justify-center text-sm">
          <span className="px-2 bg-surface-1 text-[var(--color-text-muted)]">
            またはSNSでログイン
          </span>
        </div>
      </div>
      <SNSSignInButton
        provider="google"
        onClick={() => {
          window.location.href = getCognitoAuthUrl('Google');
        }}
      />
    </AuthLayout>
  );
}
