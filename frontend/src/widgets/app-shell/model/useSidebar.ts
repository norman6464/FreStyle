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
      const { endSessionUrl } = await AuthRepository.logout();
      dispatch(clearAuth());
      clearAuthHint();
      // 発行者側のセッションも終わらせる。
      //
      // 手元の Cookie を消すだけだと、発行者には「ログイン済み」が残る。同じ端末で
      // もう一度ログインを始めると、ログイン画面すら出ずにそのまま入り直せてしまう。
      // 共用端末では、前の人のアカウントに次の人が入れることになる。
      if (endSessionUrl) {
        window.location.href = endSessionUrl;
        return;
      }
      navigate('/login');
    } catch {
      setLoggingOut(false);
    }
  }, [dispatch, navigate]);

  return { handleLogout, loggingOut };
}
