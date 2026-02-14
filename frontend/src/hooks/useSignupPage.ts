import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from './useAuth';

interface SignupForm {
  email: string;
  password: string;
  name: string;
}

interface FormMessage {
  type: 'success' | 'error';
  text: string;
}

export function useSignupPage() {
  const [form, setForm] = useState<SignupForm>({ email: '', password: '', name: '' });
  const [message, setMessage] = useState<FormMessage | null>(null);
  const navigate = useNavigate();
  const { signup, loading } = useAuth();

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm({
      ...form,
      [e.target.name]: e.target.value,
    });
  };

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
            message: '✓ サインアップに成功しました！メール確認をお願いします。',
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
