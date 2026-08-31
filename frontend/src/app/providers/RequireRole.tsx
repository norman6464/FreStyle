import { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';

import { useAppSelector } from '@/shared/lib/store';
import Loading from '@/shared/ui/Loading';
import type { UserRole } from '@/entities/user';

type RequireRoleProps = { children: ReactNode } & (
  | {
      /** 通過を許す role の許可リスト。 */
      allow: UserRole[];
      /** true なら role に加えて auth.isAdmin フラグも要求する。 */
      requireAdminFlag?: boolean;
    }
  | {
      /**
       * role を問わない。isAdmin フラグだけで通す画面のためにあるので、
       * 単独では使えない（requireAdminFlag の指定を型で必須にしている）。
       */
      allow: 'any';
      requireAdminFlag: true;
    }
);

/**
 * ロール単位のルートガード。
 *
 * 1. 認証情報の確認中 → ローディング（判定を保留する）
 * 2. 許可リストに無い role、または isAdmin フラグ不足 → /
 * 3. それ以外は子コンポーネントを描画
 *
 * isAdmin は role から導けない独立した事実（backend は Cognito の admin グループにも
 * true を返し、DB の role が trainee のままでも true になりうる）ため、role の許可リストと
 * 別のフラグとして扱う。認証済みかどうかの判定は Protected が担う。
 */
export default function RequireRole({
  allow,
  requireAdminFlag = false,
  children,
}: RequireRoleProps) {
  const loading = useAppSelector((state) => state.auth.loading);
  const isAdmin = useAppSelector((state) => state.auth.isAdmin);
  const role = useAppSelector((state) => state.auth.role);

  if (loading) {
    return <Loading message="認証情報を確認中..." className="min-h-[50vh]" />;
  }

  const roleAllowed = allow === 'any' || allow.some((allowed) => allowed === role);
  if (!roleAllowed || (requireAdminFlag && !isAdmin)) {
    return <Navigate to="/" replace />;
  }

  return children;
}
