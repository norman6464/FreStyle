import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import authRepository from '../repositories/AuthRepository';
import type { FormMessage } from '../types';
import { extractServerErrorMessage } from '../utils/classifyApiError';

export function useForgotPassword() {
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState<FormMessage | null>(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!email.trim()) {
      setMessage({ type: 'error', text: 'メールアドレスを入力してください。' });
      return;
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      setMessage({ type: 'error', text: '有効なメールアドレスを入力してください。' });
      return;
    }

    setLoading(true);

    try {
      await authRepository.forgotPassword({ email });
      setMessage({ type: 'success', text: 'コード送信済み' });
      navigate('/confirm-forgot-password', { state: { email } });
    } catch (error) {
      setMessage({ type: 'error', text: extractServerErrorMessage(error, '通信エラーが発生しました。') });
    } finally {
      setLoading(false);
    }
  };

  const clearMessage = () => setMessage(null);

  return { email, setEmail, message, loading, handleSubmit, clearMessage };
}
