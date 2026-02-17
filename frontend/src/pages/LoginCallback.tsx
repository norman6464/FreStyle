import Loading from '../components/Loading';
import { useLoginCallback } from '../hooks/useLoginCallback';

export default function LoginCallback() {
  useLoginCallback();

  return <Loading message="ログイン中..." className="min-h-screen" />;
}
