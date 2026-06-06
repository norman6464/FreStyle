import { useLocation } from 'react-router-dom';

/**
 * LoginPage 用フック。
 *
 * ログインは Cognito Hosted UI に一本化したため、フォーム状態は持たない。
 * ログアウト後や招待受諾後などに遷移元から渡されるフラッシュメッセージだけを取り出す。
 */
export function useLoginPage() {
  const location = useLocation();
  const flashMessage = (location.state as { message?: string })?.message || '';

  return { flashMessage };
}
