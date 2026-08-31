import { ReactNode } from 'react';
import { useAppSelector } from '@/shared/lib/store';

import { Navigate, useLocation } from 'react-router-dom';

interface ProtectedProps {
  children: ReactNode;
}

// super_admin は trainee 向け学習機能を利用しないため、これらのパスにアクセスしたら
// ホーム（/）へリダイレクトする。Header のナビ filter とセットで運用する。
// ノート（/notes・/p）は旧ナレッジを統合した共有の面なので super_admin にも開く。
const TRAINEE_ONLY_PATH_PREFIXES = ['/code-editor', '/reports'];

/**
 * 認証必須ルートのガード。
 *
 * 1. 未認証 → /login
 * 2. role === 'super_admin' + trainee 向けパス → /
 * 3. それ以外は子コンポーネントを描画
 */
export default function Protected({ children }: ProtectedProps) {
  const { isAuthenticated, role } = useAppSelector((state) => state.auth);
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (
    role === 'super_admin' &&
    TRAINEE_ONLY_PATH_PREFIXES.some((p) => location.pathname.startsWith(p))
  ) {
    return <Navigate to="/" replace />;
  }

  return children;
}
