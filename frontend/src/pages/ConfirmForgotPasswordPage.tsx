import { useState } from 'react';
import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import { useNavigate, useLocation } from 'react-router-dom';

interface FormMessage {
  type: 'success' | 'error';
  text: string;
}

export default function ConfirmForgotPasswordPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const [form, setForm] = useState({
    email: (location.state as { email?: string })?.email || '',
    code: '',
    newPassword: '',
  });
  const [message, setMessage] = useState<FormMessage | null>(null);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm({
      ...form,
      [e.target.name]: e.target.value,
    });
  };

  const handleConfirm = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    try {
      const response = await fetch(
        `${API_BASE_URL}/api/auth/cognito/confirm-forgot-password`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(form),
        }
      );

      const data = await response.json();

      if (response.ok) {
        setMessage({ type: 'success', text: data.message });
        navigate('/login', {
          state: {
            message: 'パスワードリセットに成功しました。ログインしてください。',
          },
        });
      } else {
        setMessage({
          type: 'error',
          text: data.error || 'リセットに失敗しました。',
        });
      }
    } catch (error) {
      console.log(error);
      setMessage({ type: 'error', text: '通信エラーが発生しました。' });
    }
  };

  return (
    <AuthLayout>
      {message?.type === 'error' && (
        <p className="text-red-600 text-center">{message.text}</p>
      )}
      <h2 className="text-2xl font-bold mb-6 text-center">
        パスワードリセット確認
      </h2>
      <form onSubmit={handleConfirm}>
        <InputField
          label="メールアドレス"
          name="email"
          type="email"
          value={form.email}
          onChange={handleChange}
        />
        <InputField
          label="確認コード"
          name="code"
          value={form.code}
          onChange={handleChange}
        />
        <InputField
          label="新しいパスワード"
          name="newPassword"
          type="password"
          value={form.newPassword}
          onChange={handleChange}
        />
        <PrimaryButton type="submit">パスワードをリセット</PrimaryButton>
      </form>
    </AuthLayout>
  );
}
