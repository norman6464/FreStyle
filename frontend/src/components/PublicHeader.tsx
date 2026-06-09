import { Link } from 'react-router-dom';
import { BuildingOffice2Icon } from '@heroicons/react/24/outline';

/**
 * 公開ページ(ログイン / 企業利用申請)共通のヘッダー。
 *
 * 既存ユーザーの「ログイン」と、見込み企業の「利用申請」の 2 導線を提供する。
 * ログインは role を問わず 1 つに統一し、ログイン後にダッシュボードを出し分ける方針。
 */
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
            to="/company-application"
            className="flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm font-medium text-[var(--color-text-secondary)] transition-colors hover:bg-surface-2"
          >
            <BuildingOffice2Icon className="h-4 w-4" aria-hidden="true" />
            企業の利用申請
          </Link>
          <Link
            to="/login"
            className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-brand-600"
          >
            ログイン
          </Link>
        </nav>
      </div>
    </header>
  );
}
