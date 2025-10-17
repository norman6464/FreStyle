import { useState } from 'react';
import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import LinkText from '../components/LinkText';
import { useNavigate } from 'react-router-dom';

export default function SignupPage() {
  const [form, setForm] = useState({ email: '', password: '', name: '' });
  const [message, setMessage] = useState(null);
  const navigate = useNavigate();

  const handleChange = (e) => {
    setForm({
      ...form,
      [e.target.name]: e.target.value,
    });
  };

  const handleSignup = async (e) => {
    e.preventDefault(); // フォームのデフォルト送信を防止

    try {
      const response = await fetch(
        'http://localhost:8080/api/auth/cognito/signup',
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            email: form.email,
            password: form.password,
            name: form.name, // Spring boot側ではnameで受け取る
          }),
        }
      );

      const data = await response.json();

      if (response.ok) {
        setMessage({ type: 'success', text: data.message });
        navigate('/confirm', {
          state: {
            message:
              'サインアップに成功しました！メッセージを送信したので確認して下さい。',
          },
        });
      } else {
        setMessage({
          type: 'error',
          text: data.error || '登録に失敗しました。',
        });
      }
    } catch (error) {
      console.log(error);
      setMessage({ type: 'error', text: '通信エラーが発生しました。' });
    }
  };

  return (
    <AuthLayout>
      <div>
        {message?.type === 'error' && (
          <p className="text-red-600 text-center">{message.text}</p>
        )}
      </div>
      <h2 className="text-2xl font-bold mb-6 text-center">アカウント作成</h2>
      <form onSubmit={handleSignup}>
        <InputField
          label="ニックネーム"
          // typeがないのはInputFieldコンポーネントでデフォルトでの引数を使うため
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
        <PrimaryButton type="submit">サインアップ</PrimaryButton>
      </form>
      <div className="mt-4 text-center">
        <LinkText to="/login">すでにアカウントをお持ちですか？</LinkText>
      </div>
    </AuthLayout>
  );
}
