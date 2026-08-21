import { useState } from 'react';
import { useLocation } from 'react-router-dom';
import { useFormField } from '@/shared/lib/hooks/useFormField';
import { AuthRepository as authRepository } from '@/entities/user';
import { getApiError } from '@/shared/lib/classifyApiError';
import { setAuthHint } from '@/shared/lib/authHint';

interface LoginMessage {
  type: 'success' | 'error';
  text: string;
}

// 一時パスワード初回ログインの新パスワード設定フェーズの状態。
// null のとき通常のメール/パスワードフォームを表示する。
interface NewPasswordPhase {
  email: string;
  session: string;
}

/**
 * LoginPage 用フック。
 *
 * メール / パスワードフォームの状態管理とログイン処理を担う。ログインは Cognito の
 * USER_PASSWORD_AUTH（backend `/auth/cognito/login`）で行う。Google は Hosted UI へ直行する。
 * 成功時は HttpOnly Cookie が発行されるので、フル再読み込みで AuthInitializer に
 * `/auth/me` を引かせて role / isAdmin を確定させる（SPA 内 navigate では再取得されないため）。
 *
 * 一時パスワードでの初回ログイン（FRESTYLE-313）は backend が NEW_PASSWORD_REQUIRED を返す。
 * その場合は新パスワード設定フェーズに切り替え、本人に新パスワードを決めてもらう。
 */
export function useLoginPage() {
  const { form, handleChange } = useFormField({ email: '', password: '' });
  const [loginMessage, setLoginMessage] = useState<LoginMessage | null>(null);
  const [loading, setLoading] = useState(false);

  // 新パスワード設定フェーズ。null なら通常ログインフォーム。
  const [newPasswordPhase, setNewPasswordPhase] = useState<NewPasswordPhase | null>(null);
  const [newPassword, setNewPassword] = useState('');
  const [newPasswordConfirm, setNewPasswordConfirm] = useState('');

  const location = useLocation();
  const flashMessage = (location.state as { message?: string })?.message || '';

  const finishLogin = () => {
    // 配信側でトップを振り分けるための目印（FRESTYLE-231）
    setAuthHint();
    window.location.assign('/dashboard');
  };

  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);
    setLoginMessage(null);

    try {
      const outcome = await authRepository.loginWithChallenge({
        email: form.email,
        password: form.password,
      });
      if (outcome.kind === 'new_password_required') {
        // 一時パスワードでの初回ログイン。新パスワード設定フェーズへ。
        setNewPasswordPhase({ email: form.email, session: outcome.session });
        setLoginMessage({
          type: 'success',
          text: '初回ログインです。新しいパスワードを設定してください。',
        });
        setLoading(false);
        return;
      }
      finishLogin();
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

  const handleNewPassword = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!newPasswordPhase) return;
    if (newPassword !== newPasswordConfirm) {
      setLoginMessage({ type: 'error', text: 'パスワードが一致しません。' });
      return;
    }
    setLoading(true);
    setLoginMessage(null);

    try {
      await authRepository.submitNewPassword({
        email: newPasswordPhase.email,
        session: newPasswordPhase.session,
        newPassword,
      });
      finishLogin();
    } catch (err) {
      const { status } = getApiError(err);
      if (status === 401) {
        // session 失効。最初からやり直してもらう。
        setNewPasswordPhase(null);
        setLoginMessage({
          type: 'error',
          text: 'セッションの有効期限が切れました。もう一度ログインからやり直してください。',
        });
      } else {
        setLoginMessage({
          type: 'error',
          text: 'パスワードの条件を満たしていません。より長く複雑なパスワードを設定してください。',
        });
      }
      setLoading(false);
    }
  };

  return {
    form,
    loginMessage,
    flashMessage,
    loading,
    handleLogin,
    handleChange,
    // 新パスワード設定フェーズ
    newPasswordPhase,
    newPassword,
    newPasswordConfirm,
    setNewPassword,
    setNewPasswordConfirm,
    handleNewPassword,
  };
}
