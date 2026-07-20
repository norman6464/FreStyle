import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import authRepository from '@/entities/user/api/authRepository';
import type { FormMessage } from '@/shared/ui/FormMessage';
import { extractServerErrorMessage } from '@/shared/lib/classifyApiError';
import { useFormField } from './useFormField';

export function useConfirmForgotPassword() {
  const location = useLocation();
  const navigate = useNavigate();

  const { form, handleChange } = useFormField({
    email: (location.state as { email?: string })?.email || '',
    code: '',
    newPassword: '',
    confirmPassword: '',
  });
  const [message, setMessage] = useState<FormMessage | null>(null);
  const [loading, setLoading] = useState(false);

  const handleConfirm = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!form.code.trim() || !form.newPassword.trim() || !form.confirmPassword.trim()) {
      setMessage({ type: 'error', text: 'すべてのフィールドを入力してください。' });
      return;
    }

    if (form.newPassword !== form.confirmPassword) {
      setMessage({ type: 'error', text: 'パスワードが一致しません。' });
      return;
    }

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

  const clearMessage = () => setMessage(null);

  return { form, message, loading, handleChange, handleConfirm, clearMessage };
}
