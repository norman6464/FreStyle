import { useState } from 'react';
import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import SNSSignInButton from '../components/SNSSignInButton';
import LinkText from '../components/LinkText';
import { getCognitoAuthUrl } from '../utils/auth';

export default function LoginPage() {
  const [form, setForm] = useState({ email: '', password: '' });

  const handleChange = (e) => {
    setForm({
      ...form,
      [e.target.name]: e.target.value,
    });
  };

  const handleLogin = (e) => {
    e.preventDefault();
    console.log('ログイン情報送信', form);
  };

  return (
    <AuthLayout>
      <h2 className="text-2xl font-bold mb-6 text-center">ログイン</h2>
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
      <div className="flex justify-between mt-4">
        <LinkText to="/forget-password">パスワードをお忘れですか？</LinkText>
        <LinkText to="/signup">アカウントを作成</LinkText>
      </div>
      <hr />
      <div className="my-6 text-center text-sm text-gray-500">
        またはSNSでログイン
      </div>
      <SNSSignInButton
        provider="google"
        onClick={() => {
          window.location.href = getCognitoAuthUrl('Google');
        }}
      />
      <SNSSignInButton
        provider="facebook"
        onClick={() => console.log('Facebook login')}
      />
      <SNSSignInButton provider="x" onClick={() => console.log('X login')} />
    </AuthLayout>
  );
}
