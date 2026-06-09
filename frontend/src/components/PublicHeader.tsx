import { Link, useLocation } from 'react-router-dom';
import { BuildingOffice2Icon } from '@heroicons/react/24/outline';

/**
 * 公開ページ(ログイン / 新規登録 / 企業利用申請)共通のヘッダー。
 *
 * 見込み企業の「利用申請」と、ログイン↔新規登録を行き来する CTA を提供する。
 * CTA は現在のページに応じて切り替える（ログインページでは「新規登録」、
 * 新規登録ページでは「ログイン」）。一般的な認証ページの導線パターンに合わせる。
 */
export default function PublicHeader() {
  const { pathname } = useLocation();
  const onSignup = pathname.startsWith('/signup');
  const cta = onSignup ? { to: '/login', label: 'ログイン' } : { to: '/signup', label: '新規登録' };

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
            to={cta.to}
            className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-brand-600"
          >
            {cta.label}
          </Link>
        </nav>
      </div>
    </header>
  );
}
