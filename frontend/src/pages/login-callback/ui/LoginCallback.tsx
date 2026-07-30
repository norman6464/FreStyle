import Loading from '@/shared/ui/Loading';
import { useDocumentMeta } from '@/shared/lib/hooks/useDocumentMeta';
import { useLoginCallback } from '../model/useLoginCallback';

export default function LoginCallback() {
  // OAuth コールバックの一時画面。検索エンジンに index させない。
  useDocumentMeta({ robots: 'noindex, nofollow' });
  useLoginCallback();

  return <Loading fullscreen message="ログイン中..." />;
}
