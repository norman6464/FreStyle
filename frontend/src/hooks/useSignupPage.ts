import { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from './useAuth';
import type { FormMessage } from '../types';
import { useFormField } from './useFormField';

export function useSignupPage() {
  const { form, handleChange } = useFormField({ email: '', password: '', name: '' });
  const [message, setMessage] = useState<FormMessage | null>(null);
  const navigate = useNavigate();
  const { signup, loading } = useAuth();
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  const handleSignup = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!form.name.trim() || !form.email.trim() || !form.password.trim()) {
      setMessage({ type: 'error', text: 'すべてのフィールドを入力してください。' });
      return;
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(form.email)) {
      setMessage({ type: 'error', text: '有効なメールアドレスを入力してください。' });
      return;
    }

    const success = await signup({
      email: form.email,
      password: form.password,
      name: form.name,
    });

    if (success) {
      setMessage({ type: 'success', text: 'サインアップに成功しました！' });
      timerRef.current = setTimeout(() => {
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

  const clearMessage = () => setMessage(null);

  return {
    form,
    message,
    loading,
    handleChange,
    handleSignup,
    clearMessage,
  };
}
