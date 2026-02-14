import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import LinkText from '../components/LinkText';
import {
  CheckCircleIcon,
  ExclamationCircleIcon,
} from '@heroicons/react/24/solid';
import { useSignupPage } from '../hooks/useSignupPage';

export default function SignupPage() {
  const { form, message, loading, handleChange, handleSignup } = useSignupPage();

  return (
    <AuthLayout>
      <div>
        {message?.type === 'error' && (
          <div className="mb-6 p-4 bg-rose-900/30 border-l-4 border-rose-500 rounded-lg flex items-start">
            <ExclamationCircleIcon className="w-5 h-5 text-rose-400 mr-3 flex-shrink-0 mt-0.5" />
            <p className="text-rose-400 font-medium text-sm">{message.text}</p>
          </div>
        )}
        {message?.type === 'success' && (
          <div className="mb-6 p-4 bg-emerald-900/30 border-l-4 border-emerald-500 rounded-lg flex items-start">
            <CheckCircleIcon className="w-5 h-5 text-emerald-400 mr-3 flex-shrink-0" />
            <p className="text-emerald-400 font-medium text-sm">
              {message.text}
            </p>
          </div>
        )}
      </div>
      <h2 className="text-3xl font-bold mb-2 text-center text-[#F0F0F0]">
        アカウント作成
      </h2>
      <p className="text-center text-[#A0A0A0] text-sm mb-8">
        FreStyleに参加して、友達とチャットを始めましょう
      </p>
      <form onSubmit={handleSignup}>
        <InputField
          label="ニックネーム"
          name="name"
          value={form.name}
          onChange={handleChange}
        />
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
        <PrimaryButton type="submit" disabled={loading}>
          {loading ? '作成中...' : 'アカウント作成'}
        </PrimaryButton>
      </form>
      <div className="mt-6 text-center">
        <p className="text-[#A0A0A0] text-sm mb-2">
          すでにアカウントをお持ちですか？
        </p>
        <LinkText to="/login">ログインする</LinkText>
      </div>
    </AuthLayout>
  );
}
