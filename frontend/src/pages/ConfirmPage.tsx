import { useState } from 'react';
import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import LinkText from '../components/LinkText';
import { useNavigate } from 'react-router-dom';
import authRepository from '../repositories/AuthRepository';
import { AxiosError } from 'axios';

interface FormMessage {
  type: 'success' | 'error';
  text: string;
}

export default function ConfirmPage() {
  const [form, setForm] = useState({ email: '', code: '' });
  const [message, setMessage] = useState<FormMessage | null>(null);
  const navigate = useNavigate();

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm({
      ...form,
      [e.target.name]: e.target.value,
    });
  };

  const handleConfirm = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    try {
      await authRepository.confirmSignup(form);
      navigate('/login', {
        state: {
          message: '確認に成功しました。ログインしてください。',
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
      <h2 className="text-2xl font-bold mb-6 text-center">確認コードの入力</h2>
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
        <PrimaryButton type="submit">確認する</PrimaryButton>
      </form>
      <div className="mt-4 text-center">
        <LinkText to="/signup">アカウント作成に戻る</LinkText>
      </div>
    </AuthLayout>
  );
}
