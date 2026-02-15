import { useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from './useAuth';
import { useFormField } from './useFormField';

interface LoginMessage {
  type: 'success' | 'error';
  text: string;
}

/**
 * LoginPageフック
 *
 * <p>役割:</p>
 * <ul>
 *   <li>LoginPageのフォーム管理・バリデーション</li>
 *   <li>ログイン処理のロジック</li>
 * </ul>
 */
export function useLoginPage() {
  const { form, handleChange } = useFormField({ email: '', password: '' });
  const [loginMessage, setLoginMessage] = useState<LoginMessage | null>(null);

  const location = useLocation();
  const navigate = useNavigate();
  const { login } = useAuth();

  const flashMessage = (location.state as { message?: string })?.message || '';

  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const success = await login({ email: form.email, password: form.password });

    if (success) {
      navigate('/');
    } else {
      setLoginMessage({
        type: 'error',
        text: 'ログインに失敗しました。',
      });
    }
  };

  return {
    form,
    loginMessage,
    flashMessage,
    handleLogin,
    handleChange,
  };
}
