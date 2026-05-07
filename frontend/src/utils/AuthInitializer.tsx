import { ReactNode, useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { setAuthData, clearAuth, finishLoading } from '../store/authSlice';
import type { RootState } from '../store';
import authRepository from '../repositories/AuthRepository';
import Loading from '../components/Loading';

interface AuthInitializerProps {
  children: ReactNode;
}

export default function AuthInitializer({ children }: AuthInitializerProps) {
  const dispatch = useDispatch();
  const loading = useSelector((state: RootState) => state.auth.loading);

  useEffect(() => {
    const checkAuth = async () => {
      try {
        const me = await authRepository.getCurrentUser();
        dispatch(
          setAuthData({
            isAdmin: !!me.isAdmin,
            // Welcome 画面に飛ばすかの判定を Protected で使う。
            // backend から undefined が返ってきた場合は完了扱い（既存ユーザー保護）。
            onboarded: me.onboarded ?? true,
            role: me.role ?? null,
          })
        );
      } catch {
        dispatch(clearAuth());
      } finally {
        dispatch(finishLoading());
      }
    };

    checkAuth();
  }, [dispatch]);

  if (loading) {
    return <Loading fullscreen />;
  }

  return children;
}
