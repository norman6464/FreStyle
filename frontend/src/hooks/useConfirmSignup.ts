import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import authRepository from '../repositories/AuthRepository';
import { AxiosError } from 'axios';
import type { FormMessage } from '../types';
import { useFormField } from './useFormField';

export function useConfirmSignup() {
  const navigate = useNavigate();
  const { form, handleChange } = useFormField({ email: '', code: '' });
  const [message, setMessage] = useState<FormMessage | null>(null);
  const [loading, setLoading] = useState(false);

  const handleConfirm = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);

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
    } finally {
      setLoading(false);
    }
  };

  return { form, message, loading, handleChange, handleConfirm };
}
