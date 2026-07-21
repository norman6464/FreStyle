import Loading from '@/shared/ui/Loading';
import { useLoginCallback } from '../model/useLoginCallback';

export default function LoginCallback() {
  useLoginCallback();

  return <Loading fullscreen message="ログイン中..." />;
}
