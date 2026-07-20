import { useState } from 'react';
import { useLocation } from 'react-router-dom';
import { useFormField } from './useFormField';
import authRepository from '@/entities/user/api/authRepository';
import { getApiError } from '@/shared/lib/classifyApiError';

interface LoginMessage {
  type: 'success' | 'error';
  text: string;
}

/**
 * LoginPage 用フック。
 *
 * メール / パスワードフォームの状態管理とログイン処理を担う。ログインは Cognito の
 * USER_PASSWORD_AUTH（backend `/auth/cognito/login`）で行う。Google は Hosted UI へ直行する。
 * 成功時は HttpOnly Cookie が発行されるので、フル再読み込みで AuthInitializer に
 * `/auth/me` を引かせて role / isAdmin を確定させる（SPA 内 navigate では再取得されないため）。
 */
export function useLoginPage() {
  const { form, handleChange } = useFormField({ email: '', password: '' });
  const [loginMessage, setLoginMessage] = useState<LoginMessage | null>(null);
  const [loading, setLoading] = useState(false);

  const location = useLocation();
  const flashMessage = (location.state as { message?: string })?.message || '';

  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);
    setLoginMessage(null);

    try {
      await authRepository.login({ email: form.email, password: form.password });
      window.location.assign('/');
    } catch (err) {
      // 招待なしの新規ユーザーは backend が 403 invitation_required を返す。専用文言を出す。
      const { status, serverMessage } = getApiError(err);
      if (status === 403) {
        setLoginMessage({
          type: 'error',
          text: serverMessage || 'FreStyle のご利用には管理者からの招待が必要です。',
        });
      } else {
        setLoginMessage({
          type: 'error',
          text: 'メールアドレスまたはパスワードが正しくありません。',
        });
      }
      setLoading(false);
    }
  };

  return { form, loginMessage, flashMessage, loading, handleLogin, handleChange };
}
