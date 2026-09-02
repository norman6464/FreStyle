import { useState, useCallback } from 'react';
import { useAppSelector, useAppDispatch } from '@/shared/lib/store';

import { useNavigate } from 'react-router-dom';
import { classifyApiError } from '@/shared/lib/classifyApiError';
import { clearAuthHint } from '@/shared/lib/authHint';
import { AuthRepository, UserInfo } from '@/entities/user';
import { setAuthData, clearAuth, finishLoading } from '@/entities/user';

/**
 * 認証フック
 *
 * <p>役割:</p>
 * <ul>
 *   <li>認証状態管理（ログイン、ログアウト、ユーザー情報取得）</li>
 *   <li>AuthRepositoryとRedux storeの統合</li>
 * </ul>
 *
 * <p>Hooks層（Presentation Layer - Business Logic）:</p>
 * <ul>
 *   <li>コンポーネントからビジネスロジックを分離</li>
 *   <li>Repository層を使用してAPI呼び出し</li>
 * </ul>
 */
export const useAuth = () => {
  const [user, setUser] = useState<UserInfo | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const authState = useAppSelector((state) => state.auth);

  /**
   * ログアウト
   */
  const logout = useCallback(async (): Promise<void> => {
    setLoading(true);
    setError(null);

    try {
      const { endSessionUrl } = await AuthRepository.logout();
      setUser(null);
      dispatch(clearAuth());
      // 認証ヒント（次回の初期描画を早めるための印）も消す。
      // 残すと、ログアウト後の再訪でログイン済みとして描き始めてしまう。
      clearAuthHint();
      // 発行者側のセッションも終わらせる（手元の Cookie を消すだけでは残る）。
      if (endSessionUrl) {
        window.location.href = endSessionUrl;
        return;
      }
      navigate('/login');
    } catch (err) {
      setError(classifyApiError(err, 'ログアウトに失敗しました。'));
    } finally {
      setLoading(false);
    }
  }, [dispatch, navigate]);

  /**
   * 現在のユーザー情報を取得
   */
  const getCurrentUser = useCallback(async (): Promise<UserInfo | null> => {
    setLoading(true);
    setError(null);

    try {
      const userInfo = await AuthRepository.getCurrentUser();
      setUser(userInfo);
      dispatch(
        setAuthData({
          isAdmin: !!userInfo.isAdmin,
          role: userInfo.role,
        }),
      );
      return userInfo;
    } catch (err) {
      setError(classifyApiError(err, 'ユーザー情報の取得に失敗しました。'));
      dispatch(finishLoading());
      return null;
    } finally {
      setLoading(false);
    }
  }, [dispatch]);

  /**
   * トークンリフレッシュ
   */
  const refreshToken = useCallback(async (): Promise<boolean> => {
    try {
      await AuthRepository.refreshToken();
      return true;
    } catch (err) {
      setError('トークンのリフレッシュに失敗しました。');
      dispatch(clearAuth());
      navigate('/login');
      return false;
    }
  }, [dispatch, navigate]);

  return {
    user,
    loading,
    error,
    isAuthenticated: authState.isAuthenticated,
    logout,
    getCurrentUser,
    refreshToken,
  };
};
