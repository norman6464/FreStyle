import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import SNSSignInButton from '../components/SNSSignInButton';
import LinkText from '../components/LinkText';
import { getCognitoAuthUrl } from '../utils/auth';
import { useLoginPage } from '../hooks/useLoginPage';

export default function LoginPage() {
  const { form, loginMessage, flashMessage, handleLogin, handleChange } = useLoginPage();

  return (
    <AuthLayout>
      {/* flash Message */}
      <div>
        {flashMessage && (
          <p className="text-emerald-600 text-center mb-4 p-3 bg-emerald-50 rounded-lg font-medium">
            {flashMessage}
          </p>
        )}
        {loginMessage?.type === 'error' && (
          <p className="text-rose-600 text-center mb-4 p-3 bg-rose-50 rounded-lg font-medium">
            ✕ {loginMessage.text}
          </p>
        )}
      </div>
      <h2 className="text-3xl font-bold mb-2 text-center text-slate-800">
        ログイン
      </h2>
      <p className="text-center text-slate-500 text-sm mb-6">
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
          <div className="w-full border-t border-slate-200"></div>
        </div>
        <div className="relative flex justify-center text-sm">
          <span className="px-2 bg-white text-slate-500">
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
