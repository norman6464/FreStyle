import { ReactNode, useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { setAuthData, clearAuth, finishLoading } from '@/entities/user';
import type { RootState } from '@/store';
import authRepository from '@/entities/user/api/authRepository';
import Loading from '@/shared/ui/Loading';

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
            role: me.role ?? null,
            // 未定義(古い backend 等)は true にフォールバックし、誤って AI を隠さない。
            aiChatEnabledForTrainees: me.aiChatEnabledForTrainees ?? true,
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
