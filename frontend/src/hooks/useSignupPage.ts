import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from './useAuth';
import type { FormMessage } from '../types';
import { useFormField } from './useFormField';

export function useSignupPage() {
  const { form, handleChange } = useFormField({ email: '', password: '', name: '' });
  const [message, setMessage] = useState<FormMessage | null>(null);
  const navigate = useNavigate();
  const { signup, loading } = useAuth();

  const handleSignup = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const success = await signup({
      email: form.email,
      password: form.password,
      name: form.name,
    });

    if (success) {
      setMessage({ type: 'success', text: 'サインアップに成功しました！' });
      setTimeout(() => {
        navigate('/confirm', {
          state: {
            message: 'サインアップに成功しました！メール確認をお願いします。',
          },
        });
      }, 1500);
    } else {
      setMessage({
        type: 'error',
        text: '登録に失敗しました。',
      });
    }
  };

  return {
    form,
    message,
    loading,
    handleChange,
    handleSignup,
  };
}
