import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import authRepository from '../repositories/AuthRepository';
import { AxiosError } from 'axios';

interface FormMessage {
  type: 'success' | 'error';
  text: string;
}

export function useConfirmForgotPassword() {
  const location = useLocation();
  const navigate = useNavigate();

  const [form, setForm] = useState({
    email: (location.state as { email?: string })?.email || '',
    code: '',
    newPassword: '',
  });
  const [message, setMessage] = useState<FormMessage | null>(null);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setForm((prev) => ({
      ...prev,
      [name]: value,
    }));
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

  return { form, message, handleChange, handleConfirm };
}
