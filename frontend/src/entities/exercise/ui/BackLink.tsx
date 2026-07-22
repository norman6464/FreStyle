import { Link } from 'react-router-dom';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';

/**
 * 演習詳細ページから演習一覧へ戻るリンク。
 * 主ビュー / QA ビュー 双方の ヘッダー で 共有 する。
 */
export default function BackLink() {
  return (
    <Link
      to="/code-editor"
      className="inline-flex items-center gap-1 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
    >
      <ArrowLeftIcon className="w-3.5 h-3.5" />
      問題一覧に戻る
    </Link>
  );
}
