import { useState } from 'react';
import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import { useNavigate, useLocation } from 'react-router-dom';
import authRepository from '../repositories/AuthRepository';
import { AxiosError } from 'axios';

interface FormMessage {
  type: 'success' | 'error';
  text: string;
}

export default function ConfirmForgotPasswordPage() {
  const location = useLocation();
  const navigate = useNavigate();

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
      await authRepository.confirmForgotPassword({
        email: form.email,
        confirmationCode: form.code,
        newPassword: form.newPassword,
      });
      navigate('/login', {
        state: {
          message: 'パスワードリセットに成功しました。ログインしてください。',
        },
      });
    } catch (error) {
      if (error instanceof AxiosError && error.response?.data?.error) {
        setMessage({ type: 'error', text: error.response.data.error });
      } else {
        setMessage({ type: 'error', text: '通信エラーが発生しました。' });
      }
    }
  };

  return (
    <AuthLayout>
      {message?.type === 'error' && (
        <p className="text-rose-600 text-center">{message.text}</p>
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
