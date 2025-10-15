import { useState } from 'react';
import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import LinkText from '../components/LinkText';

export default function SignupPage() {
  const [form, setForm] = useState({ email: '', password: '', nickname: '' });

  const handleChange = (e) => {
    setForm({
      ...form,
      [e.target.name]: e.target.value,
    });
  };

  const handleSignup = () => {
    console.log('サインアップ', form);
  };

  return (
    <AuthLayout>
      <h2 className="text-2xl font-bold mb-6 text-center">アカウント作成</h2>
      <form onSubmit={handleSignup}>
        <InputField
          label="ニックネーム"
          // typeがないのはInputFieldコンポーネントでデフォルトでの引数を使うため
          name="nickname"
          value={form.nickname}
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
        <PrimaryButton type="submit">サインアップ</PrimaryButton>
      </form>
      <div className="mt-4 text-center">
        <LinkText to="/login">すでにアカウントをお持ちですか？</LinkText>
      </div>
    </AuthLayout>
  );
}
