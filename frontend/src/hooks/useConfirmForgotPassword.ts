import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import authRepository from '../repositories/AuthRepository';
import type { FormMessage } from '../types';
import { extractServerErrorMessage } from '../utils/classifyApiError';
import { useFormField } from './useFormField';

export function useConfirmForgotPassword() {
  const location = useLocation();
  const navigate = useNavigate();

  const { form, handleChange } = useFormField({
    email: (location.state as { email?: string })?.email || '',
    code: '',
    newPassword: '',
  });
  const [message, setMessage] = useState<FormMessage | null>(null);
  const [loading, setLoading] = useState(false);

  const handleConfirm = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);

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
      setMessage({ type: 'error', text: extractServerErrorMessage(error, '通信エラーが発生しました。') });
    } finally {
      setLoading(false);
    }
  };

  return { form, message, loading, handleChange, handleConfirm };
}
