import { useState } from 'react';
import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import { useNavigate } from 'react-router-dom';
import authRepository from '../repositories/AuthRepository';
import { AxiosError } from 'axios';

interface FormMessage {
  type: 'success' | 'error';
  text: string;
}

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState<FormMessage | null>(null);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    try {
      const data = await authRepository.forgotPassword({ email });
      setMessage({ type: 'success', text: data.message });
      navigate('/confirm-forgot-password', { state: { email } });
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
        パスワードリセット
      </h2>
      <form onSubmit={handleSubmit}>
        <InputField
          label="メールアドレス"
          name="email"
          type="email"
          value={email}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEmail(e.target.value)}
        />
        <PrimaryButton type="submit">確認コードを送信</PrimaryButton>
      </form>
    </AuthLayout>
  );
}
