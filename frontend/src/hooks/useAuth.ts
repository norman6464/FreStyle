import { useState, useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { classifyApiError } from '../utils/classifyApiError';
import AuthRepository, {
  LoginRequest,
  SignupRequest,
  ConfirmSignupRequest,
  ForgotPasswordRequest,
  ConfirmForgotPasswordRequest,
  UserInfo,
} from '../repositories/AuthRepository';
import { setAuthData, clearAuth, finishLoading } from '../store/authSlice';
import { RootState } from '../store';

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

  const dispatch = useDispatch();
  const navigate = useNavigate();
  const authState = useSelector((state: RootState) => state.auth);

  /**
   * ログイン
   */
  const login = useCallback(
    async (request: LoginRequest): Promise<boolean> => {
      setLoading(true);
      setError(null);

      try {
        const userInfo = await AuthRepository.login(request);
        setUser(userInfo);
        dispatch(setAuthData());
        return true;
      } catch (err) {
        setError(classifyApiError(err, 'ログインに失敗しました。'));
        return false;
      } finally {
        setLoading(false);
      }
    },
    [dispatch]
  );

  /**
   * サインアップ
   */
  const signup = useCallback(async (request: SignupRequest): Promise<boolean> => {
    setLoading(true);
    setError(null);

    try {
      await AuthRepository.signup(request);
      return true;
    } catch (err) {
      setError(classifyApiError(err, 'サインアップに失敗しました。'));
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * サインアップ確認
   */
  const confirmSignup = useCallback(async (request: ConfirmSignupRequest): Promise<boolean> => {
    setLoading(true);
    setError(null);

    try {
      await AuthRepository.confirmSignup(request);
      return true;
    } catch (err) {
      setError(classifyApiError(err, '確認に失敗しました。'));
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * パスワード再設定リクエスト
   */
  const forgotPassword = useCallback(async (request: ForgotPasswordRequest): Promise<boolean> => {
    setLoading(true);
    setError(null);

    try {
      await AuthRepository.forgotPassword(request);
      return true;
    } catch (err) {
      setError(classifyApiError(err, 'パスワード再設定リクエストに失敗しました。'));
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * パスワード再設定確認
   */
  const confirmForgotPassword = useCallback(
    async (request: ConfirmForgotPasswordRequest): Promise<boolean> => {
      setLoading(true);
      setError(null);

      try {
        await AuthRepository.confirmForgotPassword(request);
        return true;
      } catch (err) {
        setError(classifyApiError(err, 'パスワード再設定確認に失敗しました。'));
        return false;
      } finally {
        setLoading(false);
      }
    },
    []
  );

  /**
   * ログアウト
   */
  const logout = useCallback(async (): Promise<void> => {
    setLoading(true);
    setError(null);

    try {
      await AuthRepository.logout();
      setUser(null);
      dispatch(clearAuth());
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
      dispatch(setAuthData());
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
    login,
    signup,
    confirmSignup,
    forgotPassword,
    confirmForgotPassword,
    logout,
    getCurrentUser,
    refreshToken,
  };
};
