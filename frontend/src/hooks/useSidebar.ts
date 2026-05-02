import { useState, useCallback } from 'react';
import { useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { clearAuth } from '../store/authSlice';
import AuthRepository from '../repositories/AuthRepository';

export function useSidebar() {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const [loggingOut, setLoggingOut] = useState(false);

  const handleLogout = useCallback(async () => {
    setLoggingOut(true);
    try {
      await AuthRepository.logout();
      dispatch(clearAuth());
      navigate('/login', { state: { toast: 'ログアウトしました' } });
    } catch {
      setLoggingOut(false);
    }
  }, [dispatch, navigate]);

  return { handleLogout, loggingOut };
}
