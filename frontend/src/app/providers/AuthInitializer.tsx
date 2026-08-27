import { ReactNode, useEffect } from 'react';
import { useAppSelector, useAppDispatch } from '@/shared/lib/store';

import { setAuthData, clearAuth, finishLoading } from '@/entities/user';

import { AuthRepository as authRepository } from '@/entities/user';
import Loading from '@/shared/ui/Loading';
import { setAuthHint, clearAuthHintIfUnauthenticated } from '@/shared/lib/authHint';

interface AuthInitializerProps {
  children: ReactNode;
}

export default function AuthInitializer({ children }: AuthInitializerProps) {
  const dispatch = useAppDispatch();
  const loading = useAppSelector((state) => state.auth.loading);

  useEffect(() => {
    const checkAuth = async () => {
      try {
        const me = await authRepository.getCurrentUser();
        dispatch(
          setAuthData({
            isAdmin: !!me.isAdmin,
            role: me.role ?? null,
            // 未定義(古い backend 等)は true にフォールバックし、誤って AI を隠さない。
          })
        );
        setAuthHint();
      } catch (err) {
        dispatch(clearAuth());
        // 認証切れが確定した(401/403)ときだけ目印を消す。通信断や 5xx で消すと、
        // セッションは生きているのに次回トップの振り分けが効かなくなる。
        clearAuthHintIfUnauthenticated(err);
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
