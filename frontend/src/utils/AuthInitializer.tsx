import { ReactNode, useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { setAuthData, clearAuth, finishLoading } from '../store/authSlice';
import type { RootState } from '../store';
import authRepository from '../repositories/AuthRepository';

interface AuthInitializerProps {
  children: ReactNode;
}

export default function AuthInitializer({ children }: AuthInitializerProps) {
  const dispatch = useDispatch();
  const loading = useSelector((state: RootState) => state.auth.loading);

  useEffect(() => {
    const checkAuth = async () => {
      try {
        await authRepository.getCurrentUser();
        dispatch(setAuthData());
      } catch {
        dispatch(clearAuth());
      } finally {
        dispatch(finishLoading());
      }
    };

    checkAuth();
  }, [dispatch]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <p className="text-lg">認証確認中...</p>
      </div>
    );
  }

  return children;
}
