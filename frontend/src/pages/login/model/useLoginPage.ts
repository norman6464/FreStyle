import { useCallback, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { buildAuthorizeUrl } from '@/features/auth';
import { classifyApiError } from '@/shared/lib/classifyApiError';

/**
 * LoginPage 用フック。
 *
 * やることは 1 つだけ — 発行者のログイン画面へ送る。認可 URL を作るときに
 * state / nonce / PKCE の検証値を作って手元に置く（features/auth/lib/oidcAuthUrl）。
 *
 * メールとパスワードのフォームは持たない。アプリが受け取ると、二要素・ロックアウト・
 * パスワードの強さといった発行者側の守りを素通りする経路を自分で開くことになる。
 */
export function useLoginPage() {
  const location = useLocation();
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // 遷移元から渡される通知。鍵が 2 つあるのは呼び出し元が揃っていないため
  // （コールバックは toast、ログアウト等は message で渡してくる）。片方だけ読むと、
  // もう片方の経路の案内が黙って消える。
  const navState = location.state as { toast?: string; message?: string } | null;
  const flashMessage = navState?.toast ?? navState?.message ?? null;

  // 二重押しの見張り。state と検証値は呼ぶたびに作り直して sessionStorage を
  // 上書きするので、同時に 2 回走らせると、先に飛んだ方の戻りで state が合わなくなる
  // （利用者から見ると「押しただけでログインに失敗する」）。
  // setLoading は次の描画までしか効かないので、描画に依らない印で止める。
  const starting = useRef(false);

  const startLogin = useCallback(async (provider?: string) => {
    if (starting.current) return;
    starting.current = true;
    setLoading(true);
    setErrorMessage(null);
    try {
      // 認可 URL の組み立てには要約（code_challenge）の計算が入るので待つ。
      // 待たずに遷移すると、検証値を置く前にページが消えて次に進めなくなる。
      window.location.href = await buildAuthorizeUrl(provider);
    } catch (err) {
      starting.current = false;
      setLoading(false);
      setErrorMessage(classifyApiError(err, 'ログイン画面へ移動できませんでした。'));
    }
  }, []);

  return { flashMessage, errorMessage, loading, startLogin };
}
