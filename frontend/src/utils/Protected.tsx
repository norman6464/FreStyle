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
 * 2. role === 'super_admin' + trainee 向けパス → /admin/companies
 * 3. それ以外は子コンポーネントを描画
 */
export default function Protected({ children }: ProtectedProps) {
  const { isAuthenticated, role } = useSelector((state: RootState) => state.auth);
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (
    role === 'super_admin' &&
    TRAINEE_ONLY_PATH_PREFIXES.some((p) => location.pathname.startsWith(p))
  ) {
    return <Navigate to="/admin/companies" replace />;
  }

  return children;
}
