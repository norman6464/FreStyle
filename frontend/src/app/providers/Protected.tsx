import { ReactNode } from 'react';
import { useAppSelector } from '@/shared/lib/store';

import { Navigate } from 'react-router-dom';

interface ProtectedProps {
  children: ReactNode;
}

/**
 * 認証必須ルートのガード。
 *
 * 1. 未認証 → /login
 * 2. それ以外は子コンポーネントを描画
 */
export default function Protected({ children }: ProtectedProps) {
  const { isAuthenticated } = useAppSelector((state) => state.auth);

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return children;
}
