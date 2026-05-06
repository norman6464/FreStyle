import { ReactNode } from 'react';
import { useSelector } from 'react-redux';
import { Navigate, useLocation } from 'react-router-dom';
import type { RootState } from '../store';

interface ProtectedProps {
  children: ReactNode;
}

/**
 * 認証必須ルートのガード。
 *
 * 1. 未認証 → /login
 * 2. 認証済 + onboarded === false → /welcome（初回オンボーディング）
 * 3. 認証済 + onboarded === true → 子コンポーネントを描画
 *
 * /welcome 自体への navigate ループを避けるため、現在 path が /welcome の場合は
 * onboarded 判定を skip する（WelcomePage 側でガード）。
 */
export default function Protected({ children }: ProtectedProps) {
  const { isAuthenticated, onboarded } = useSelector((state: RootState) => state.auth);
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (!onboarded && location.pathname !== '/welcome') {
    return <Navigate to="/welcome" replace />;
  }

  return children;
}
