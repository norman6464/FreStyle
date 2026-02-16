import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import LinkText from '../components/LinkText';
import FormMessage from '../components/FormMessage';
import { useSignupPage } from '../hooks/useSignupPage';

export default function SignupPage() {
  const { form, message, loading, handleChange, handleSignup } = useSignupPage();

  return (
    <AuthLayout>
      <FormMessage message={message} />
      <h2 className="text-3xl font-bold mb-2 text-center text-[var(--color-text-primary)]">
        アカウント作成
      </h2>
      <p className="text-center text-[var(--color-text-tertiary)] text-sm mb-8">
        FreStyleに参加して、友達とチャットを始めましょう
      </p>
      <form onSubmit={handleSignup}>
        <InputField
          label="ニックネーム"
          name="name"
          value={form.name}
          onChange={handleChange}
          disabled={loading}
        />
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
          {loading ? '作成中...' : 'アカウント作成'}
        </PrimaryButton>
      </form>
      <div className="mt-6 text-center">
        <p className="text-[var(--color-text-tertiary)] text-sm mb-2">
          すでにアカウントをお持ちですか？
        </p>
        <LinkText to="/login">ログインする</LinkText>
      </div>
    </AuthLayout>
  );
}
