import { useAuth } from '@/features/auth';

export function useSidebar() {
  // ログアウトの実体は useAuth（発行者からのサインアウト + Redux/authHint のクリア +
  // /login への遷移）に一本化する。以前はここに別実装を持っていた。
  const { logout, loading } = useAuth();

  return { handleLogout: logout, loggingOut: loading };
}
