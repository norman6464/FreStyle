import Loading from '../components/Loading';
import { useLoginCallback } from '../hooks/useLoginCallback';

export default function LoginCallback() {
  useLoginCallback();

  return <Loading fullscreen message="ログイン中..." />;
}
