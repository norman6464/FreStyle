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
    return <Loading fullscreen />;
  }

  return children;
}
