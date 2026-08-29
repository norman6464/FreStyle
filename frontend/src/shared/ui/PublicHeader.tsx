import { Link } from 'react-router-dom';
import { UserPlusIcon } from '@heroicons/react/24/outline';

/** 公開ページ(ログイン / サインアップ)共通のヘッダー。 */
export default function PublicHeader() {
  return (
    <header className="w-full border-b border-surface-3 bg-surface-1">
      <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-3">
        <Link to="/login" className="flex items-center gap-2" aria-label="FreStyle ホーム">
          <img src="/favicon.svg" alt="" aria-hidden="true" className="h-7 w-7" />
          <span className="text-lg font-bold tracking-tight text-[var(--color-text-primary)]">
            FreStyle
          </span>
        </Link>

        <nav className="flex items-center gap-2">
          <Link
            to="/signup"
            className="flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm font-medium text-[var(--color-text-secondary)] transition-colors hover:bg-surface-2"
          >
            <UserPlusIcon className="h-4 w-4" aria-hidden="true" />
            アカウントを作成
          </Link>
        </nav>
      </div>
    </header>
  );
}
