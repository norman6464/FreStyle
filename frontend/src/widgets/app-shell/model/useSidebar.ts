import { useState, useCallback } from 'react';
import { useAppDispatch } from '@/shared/lib/store';

import { useNavigate } from 'react-router-dom';
import { clearAuth } from '@/entities/user';
import { clearAuthHint } from '@/shared/lib/authHint';
import { AuthRepository } from '@/entities/user';

export function useSidebar() {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const [loggingOut, setLoggingOut] = useState(false);

  const handleLogout = useCallback(async () => {
    setLoggingOut(true);
    try {
      await AuthRepository.logout();
      dispatch(clearAuth());
      clearAuthHint();
      navigate('/login');
    } catch {
      setLoggingOut(false);
    }
  }, [dispatch, navigate]);

  return { handleLogout, loggingOut };
}
