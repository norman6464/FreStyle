import { ReactNode } from 'react';
import { useSelector } from 'react-redux';
import { Navigate, useLocation } from 'react-router-dom';
import type { RootState } from '../store';

interface ProtectedProps {
  children: ReactNode;
}

// super_admin は trainee 向け学習機能を利用しないため、これらのパスにアクセスしたら
// /admin/companies へリダイレクトする。Sidebar の filter とセットで運用する。
const TRAINEE_ONLY_PATH_PREFIXES = ['/chat/ask-ai', '/code-editor', '/notes', '/reports'];

/**
 * 認証必須ルートのガード。
 *
 * 1. 未認証 → /login
 * 2. 認証済 + onboarded === false → /welcome（初回オンボーディング）
 * 3. role === 'super_admin' + trainee 向けパス → /admin/companies
 * 4. それ以外は子コンポーネントを描画
 *
 * /welcome 自体への navigate ループを避けるため、現在 path が /welcome の場合は
 * onboarded 判定を skip する（WelcomePage 側でガード）。
 */
export default function Protected({ children }: ProtectedProps) {
  const { isAuthenticated, onboarded, role } = useSelector((state: RootState) => state.auth);
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (!onboarded && location.pathname !== '/welcome') {
    return <Navigate to="/welcome" replace />;
  }

  if (
    role === 'super_admin' &&
    TRAINEE_ONLY_PATH_PREFIXES.some((p) => location.pathname.startsWith(p))
  ) {
    return <Navigate to="/admin/companies" replace />;
  }

  return children;
}
