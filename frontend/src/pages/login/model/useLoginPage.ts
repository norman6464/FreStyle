import { useLocation } from 'react-router-dom';
import { useOidcLogin, type OidcLogin } from '@/features/auth';

/**
 * LoginPage 用フック。
 *
 * やることは 1 つだけ — 発行者のログイン画面へ送る。認可 URL を作るときに
 * state / nonce / PKCE の検証値を作って手元に置く（features/auth/lib/oidcAuthUrl）。
 *
 * メールとパスワードのフォームは持たない。アプリが受け取ると、二要素・ロックアウト・
 * パスワードの強さといった発行者側の守りを素通りする経路を自分で開くことになる。
 */
export function useLoginPage(): { flashMessage: string | null; login: OidcLogin } {
  const location = useLocation();

  // 遷移元から渡される通知。鍵が 2 つあるのは呼び出し元が揃っていないため
  // （コールバックは toast、ログアウト等は message で渡してくる）。片方だけ読むと、
  // もう片方の経路の案内が黙って消える。
  const navState = location.state as { toast?: string; message?: string } | null;
  const flashMessage = navState?.toast ?? navState?.message ?? null;

  return { flashMessage, login: useOidcLogin() };
}
