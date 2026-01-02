import { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { setAuthData, clearAuth, finishLoading } from '../store/authSlice';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

export default function AuthInitializer({ children }) {
  const dispatch = useDispatch();
  const loading = useSelector((state) => state.auth.loading);

  useEffect(() => {
    const checkAuth = async () => {
      try {
        console.log('[AuthInitializer] Starting auth check...');
        const res = await fetch(`${API_BASE_URL}/api/auth/cognito/me`, {
          credentials: 'include',
        });

        console.log('[AuthInitializer] Auth check response status:', res.status);

        if (!res.ok) {
          console.log('[AuthInitializer] Auth check failed with status:', res.status);
          throw new Error(`Auth check failed: ${res.status}`);
        }

        console.log('[AuthInitializer] Auth check successful, setting auth state');
        dispatch(setAuthData());
      } catch (error) {
        console.error('[AuthInitializer] Auth check error:', error.message);
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
