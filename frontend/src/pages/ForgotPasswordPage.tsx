import { useState } from 'react';
import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import { useNavigate } from 'react-router-dom';

interface FormMessage {
  type: 'success' | 'error';
  text: string;
}

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState<FormMessage | null>(null);
  const navigate = useNavigate();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    try {
      const response = await fetch(
        `${API_BASE_URL}/api/auth/cognito/forgot-password`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email }),
        }
      );

      const data = await response.json();

      if (response.ok) {
        setMessage({ type: 'success', text: data.message });
        navigate('/confirm-forgot-password', { state: { email } });
      } else {
        setMessage({
          type: 'error',
          text: data.error || '送信に失敗しました。',
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
