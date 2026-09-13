import { useAuth } from '@/features/auth';

export function useSidebar() {
  // ログアウトの実体は useAuth（発行者からのサインアウト + Redux/authHint のクリア +
  // /login への遷移）に一本化する。
  const { logout, loading } = useAuth();

  return { handleLogout: logout, loggingOut: loading };
}
