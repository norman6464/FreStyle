import Loading from '../components/Loading';
import { useLoginCallback } from '../hooks/useLoginCallback';

export default function LoginCallback() {
  useLoginCallback();

  return <Loading message="読み込み中..." className="py-12" />;
}
