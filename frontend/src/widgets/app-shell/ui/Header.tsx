import { useEffect, useRef, useState } from 'react';
import { useAppSelector } from '@/shared/lib/store';
import { useLocation, Link } from 'react-router-dom';

import {
  BellIcon,
  Bars3Icon,
  XMarkIcon,
  ChevronDownIcon,
} from '@heroicons/react/24/outline';
import Loading from '@/shared/ui/Loading';
import HeaderUserMenu from './HeaderUserMenu';
import HeaderWorkspaceSwitcher from './HeaderWorkspaceSwitcher';
import { useSidebar } from '../model/useSidebar';
import { NotificationRepository } from '@/entities/notification';
import { ProfileRepository } from '@/entities/user';

// ナビ項目・ロール出し分け・アクティブ判定は model/navigation に一元化してある
// （サイドバー・モバイルメニューと共用の正典）。ここでは描画だけを行う。
import { navActive, roleLabel, visibleAdminSubs, visibleMainNav } from '../model/navigation';

/**
 * Header — 上部固定のテキスト横並びナビ。
 *
 * 左: ロゴ ／ 中央左: テキストナビ（アイコンなし） ／ 右: 通知ベル + 管理ドロップダウン(admin) + ユーザーメニュー。
 * モバイルではハンバーガーで縦メニューを開く。 ロール出し分けはサイドバー時代の仕様を踏襲する。
 */
export default function Header() {
  const location = useLocation();
  const { handleLogout, loggingOut } = useSidebar();
  const isAdmin = useAppSelector((s) => s.auth.isAdmin);
  const role = useAppSelector((s) => s.auth.role);

  const [profile, setProfile] = useState<{ displayName: string; avatarUrl: string | null; email: string } | null>(null);
  const [unread, setUnread] = useState(0);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [adminOpen, setAdminOpen] = useState(false);
  const adminRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let cancelled = false;
    ProfileRepository.fetchProfile()
      .then((p) => {
        if (cancelled) return;
        setProfile({ displayName: p.displayName ?? '', avatarUrl: p.avatarUrl ?? null, email: p.email ?? '' });
      })
      .catch(() => { /* 表示が壊れない最低限のフォールバックは下で行う */ });
    // バッジ用に未読件数だけ取得する（全件取得は重いのでヘッダーでは行わない）。
    NotificationRepository.getUnreadCount()
      .then((c) => { if (!cancelled) setUnread(c); })
      .catch(() => { /* 取得失敗時はバッジ非表示 */ });
    return () => { cancelled = true; };
  }, []);

  // ルート遷移でモバイルメニュー / 管理ドロップダウンを閉じる。
  useEffect(() => {
    setMobileOpen(false);
    setAdminOpen(false);
  }, [location.pathname]);

  // 管理ドロップダウンの外側クリックで閉じる。
  useEffect(() => {
    if (!adminOpen) return;
    const onDoc = (e: MouseEvent) => {
      if (adminRef.current && e.target instanceof Node && adminRef.current.contains(e.target)) return;
      setAdminOpen(false);
    };
    document.addEventListener('mousedown', onDoc);
    return () => document.removeEventListener('mousedown', onDoc);
  }, [adminOpen]);

  const visibleNav = visibleMainNav(role);
  const adminSubs = visibleAdminSubs(role);

  const navLinkClass = (active: boolean) =>
    `px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
      active
        ? 'bg-[var(--color-nav-active)] text-[var(--color-text-primary)]'
        : 'text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)]'
    }`;

  return (
    <>
      {loggingOut && <Loading fullscreen message="ログアウト中..." />}
      {/*
        下の境界は線ではなく**ぼかし**で表す（見本合わせ）。線は面をはっきり分けるが、
        本文が続いていく画面では区切りが強すぎて視線が止まる。半透明の地色 +
        backdrop-blur なら、下の内容が透けたまま淡く沈み、境目だけが分かる。
        supports で backdrop-filter を持つ環境だけ半透明にし、無い環境では
        不透明の地色に落とす（透けたまま読めなくならないように）。
      */}
      <header className="app-header-surface flex-shrink-0 h-16 flex items-center gap-2 px-3">
        {/* ロゴは favicon と同じ画像（favicon.svg = 三角の飛翔マーク）に揃える。 */}
        <Link to="/" className="flex items-center gap-2 flex-shrink-0 mr-2" aria-label="FreStyle ホーム">
          <img src="/favicon.svg" alt="" aria-hidden="true" className="w-7 h-7 flex-shrink-0" />
          <span className="hidden sm:block text-sm font-semibold text-[var(--color-text-primary)]">FreStyle</span>
        </Link>

        <HeaderWorkspaceSwitcher />

        {/* デスクトップ: テキスト横並びナビ */}
        <nav className="hidden md:flex items-center gap-1" aria-label="メインナビゲーション">
          {visibleNav.map((item) => (
            <Link key={item.id} to={item.to} className={navLinkClass(navActive(item, location.pathname))}>
              {item.label}
            </Link>
          ))}
        </nav>

        {/* 右側 utilities */}
        <div className="ml-auto flex items-center gap-1">
          {/* 通知ベル（未読バッジ付き） */}
          <Link
            to="/notifications"
            aria-label={unread > 0 ? `通知 (未読 ${unread} 件)` : '通知'}
            className="relative p-2 rounded-md text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)] transition-colors"
          >
            <BellIcon className="w-5 h-5" />
            {unread > 0 && (
              <span className="absolute top-1 right-1 min-w-[16px] h-4 px-1 rounded-full bg-red-500 text-white text-[10px] leading-4 text-center">
                {unread > 99 ? '99+' : unread}
              </span>
            )}
          </Link>

          {/* 管理ドロップダウン（admin のみ・デスクトップ） */}
          {isAdmin && (
            <div ref={adminRef} className="relative hidden md:block">
              <button
                type="button"
                onClick={() => setAdminOpen((p) => !p)}
                aria-haspopup="true"
                aria-expanded={adminOpen}
                // 管理はドロップダウンのトリガーなので、 /admin 配下でもアクティブ背景を付けず
                // 他のナビ項目と同じ通常表示に揃える（実際のページは sub 項目側でハイライトする）。
                className={`${navLinkClass(false)} inline-flex items-center gap-1`}
              >
                管理
                <ChevronDownIcon className={`w-3.5 h-3.5 transition-transform ${adminOpen ? 'rotate-180' : ''}`} />
              </button>
              {adminOpen && (
                <div className="absolute top-full right-0 mt-2 w-44 bg-surface-1 border border-surface-3 rounded-lg shadow-lg overflow-hidden z-50 animate-fade-in">
                  {adminSubs.map((sub) => (
                    <Link
                      key={sub.to}
                      to={sub.to}
                      className={`block px-3 py-2 text-sm transition-colors ${
                        location.pathname.startsWith(sub.matchPrefix)
                          ? 'bg-[var(--color-nav-active)] text-[var(--color-text-primary)]'
                          : 'text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)]'
                      }`}
                    >
                      {sub.label}
                    </Link>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* ユーザーメニュー（デスクトップ） */}
          <div className="hidden md:block">
            <HeaderUserMenu
              displayName={profile?.displayName ?? ''}
              avatarUrl={profile?.avatarUrl}
              email={profile?.email ?? ''}
              subText={roleLabel(role)}
              onLogout={handleLogout}
            />
          </div>

          {/* モバイル: ハンバーガー */}
          <button
            type="button"
            onClick={() => setMobileOpen((p) => !p)}
            aria-label="メニュー"
            aria-expanded={mobileOpen}
            className="md:hidden p-2 rounded-md text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)] transition-colors"
          >
            {mobileOpen ? <XMarkIcon className="w-5 h-5" /> : <Bars3Icon className="w-5 h-5" />}
          </button>
        </div>
      </header>

      {/* モバイルメニュー */}
      {mobileOpen && (
        <div className="app-header-surface md:hidden">
          <nav className="px-3 py-2 space-y-0.5" aria-label="モバイルナビゲーション">
            {visibleNav.map((item) => (
              <Link key={item.id} to={item.to} className={`block ${navLinkClass(navActive(item, location.pathname))}`}>
                {item.label}
              </Link>
            ))}
            {isAdmin && adminSubs.length > 0 && (
              <>
                <div className="my-1 border-t border-surface-3" />
                <p className="px-3 py-1 text-xs text-[var(--color-text-muted)]">管理</p>
                {adminSubs.map((sub) => (
                  <Link key={sub.to} to={sub.to} className={`block ${navLinkClass(location.pathname.startsWith(sub.matchPrefix))}`}>
                    {sub.label}
                  </Link>
                ))}
              </>
            )}
            <div className="my-1 border-t border-surface-3" />
            <Link to="/settings" className={`block ${navLinkClass(location.pathname === '/settings')}`}>
              設定
            </Link>
            <button
              type="button"
              onClick={handleLogout}
              className="block w-full text-left px-3 py-1.5 rounded-md text-sm font-medium text-[var(--color-text-muted)] hover:bg-red-900/10 hover:text-red-500 transition-colors"
            >
              ログアウト
            </button>
          </nav>
        </div>
      )}
    </>
  );
}
