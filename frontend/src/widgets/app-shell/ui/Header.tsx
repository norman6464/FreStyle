import { useEffect, useState } from 'react';
import { useLocation, Link } from 'react-router-dom';

import {
  BellIcon,
  Bars3Icon,
  MagnifyingGlassIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';
import Loading from '@/shared/ui/Loading';
import HeaderUserMenu from './HeaderUserMenu';
import { useSidebar } from '../model/useSidebar';
import { NotificationRepository } from '@/entities/notification';
import { ProfileRepository } from '@/entities/user';

// ナビ項目・アクティブ判定は model/navigation に一元化してある
// （サイドバー・モバイルメニューと共用の正典）。ここでは描画だけを行う。
import { MAIN_NAV_ITEMS, navActive } from '../model/navigation';

interface HeaderProps {
  /** 中央の検索ボタン押下時に呼ぶ。AppShell が持つ既存の ⌘K パレットを開くだけで、
   *  ここでは検索の状態を持たない。 */
  onOpenSearch: () => void;
}

/**
 * Header — 上部固定のテキスト横並びナビ。常時表示（本文には重ねない・自動的には隠れない）。
 *
 * 左: ロゴ ／ 中央左: テキストナビ（アイコンなし） ／ 中央: 検索ボタン ／
 * 右: 通知ベル + ユーザーメニュー。モバイルではハンバーガーで縦メニューを開く。
 *
 * ワークスペース切替は置かない（`KbSidebar` 先頭に既にあり、二重にしない）。
 */
export default function Header({ onOpenSearch }: HeaderProps) {
  const location = useLocation();
  const { handleLogout, loggingOut } = useSidebar();

  const [profile, setProfile] = useState<{ displayName: string; avatarUrl: string | null; email: string } | null>(null);
  const [unread, setUnread] = useState(0);
  const [mobileOpen, setMobileOpen] = useState(false);

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

  // ルート遷移でモバイルメニューを閉じる。
  useEffect(() => {
    setMobileOpen(false);
  }, [location.pathname]);

  const navLinkClass = (active: boolean) =>
    `px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
      active
        ? 'bg-[var(--color-nav-active)] text-[var(--color-text-primary)]'
        : 'text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)]'
    }`;

  return (
    <>
      {loggingOut && <Loading fullscreen message="ログアウト中..." />}
      {/* 常時表示・不透明。本文とは縦に並ぶだけで重ねないので、半透明やぼかしは不要。 */}
      <header className="app-header-surface flex-shrink-0 h-14 flex items-center gap-2 px-3">
        {/* ロゴは favicon と同じ画像（favicon.svg = 三角の飛翔マーク）に揃える。 */}
        <Link to="/" className="flex items-center gap-2 flex-shrink-0 mr-2" aria-label="FreStyle ホーム">
          <img src="/favicon.svg" alt="" aria-hidden="true" className="w-7 h-7 flex-shrink-0" />
          <span className="hidden sm:block text-sm font-semibold text-[var(--color-text-primary)]">FreStyle</span>
        </Link>

        {/* デスクトップ: テキスト横並びナビ */}
        <nav className="hidden md:flex items-center gap-1" aria-label="メインナビゲーション">
          {MAIN_NAV_ITEMS.map((item) => (
            <Link key={item.id} to={item.to} className={navLinkClass(navActive(item, location.pathname))}>
              {item.label}
            </Link>
          ))}
        </nav>

        {/* 中央の検索ボタン。flex-1 の帯の中で justify-center することで、左（ロゴ＋ナビ）・
            右（utilities）の幅に関わらず帯の中央に来る（mx-auto だと右の ml-auto と
            auto マージンを取り合って中央からズレるため使わない）。 */}
        <div className="hidden md:flex flex-1 justify-center px-4">
          <button
            type="button"
            onClick={onOpenSearch}
            className="flex items-center gap-2 w-64 px-3 py-1.5 rounded-md border border-surface-3 bg-surface text-sm text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] transition-colors"
          >
            <MagnifyingGlassIcon className="w-4 h-4 flex-shrink-0" />
            <span className="truncate">検索</span>
            <span className="ml-auto text-xs text-[var(--color-text-muted)]" aria-hidden="true">⌘K</span>
          </button>
        </div>

        {/* 右側 utilities。モバイルでは中央帯が隠れて自動の余白が無くなるので ml-auto で右へ寄せる。 */}
        <div className="ml-auto flex items-center gap-1">
          {/* モバイル: 検索は虫眼鏡アイコンのボタンに畳む。 */}
          <button
            type="button"
            onClick={onOpenSearch}
            aria-label="検索"
            className="md:hidden p-2 rounded-md text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)] transition-colors"
          >
            <MagnifyingGlassIcon className="w-5 h-5" />
          </button>
          {/* 通知ベル（未読バッジ付き） */}
          <Link
            to="/notifications"
            aria-label={unread > 0 ? `通知 (未読 ${unread} 件)` : '通知'}
            className="relative p-2 rounded-md text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)] transition-colors"
          >
            <BellIcon className="w-5 h-5" />
            {unread > 0 && (
              <span className="absolute top-1 right-1 min-w-[16px] h-4 px-1 rounded-full bg-red-600 text-white text-[10px] leading-4 text-center">
                {unread > 99 ? '99+' : unread}
              </span>
            )}
          </Link>

          {/* ユーザーメニュー（デスクトップ） */}
          <div className="hidden md:block">
            <HeaderUserMenu
              displayName={profile?.displayName ?? ''}
              avatarUrl={profile?.avatarUrl}
              email={profile?.email ?? ''}
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
            {MAIN_NAV_ITEMS.map((item) => (
              <Link key={item.id} to={item.to} className={`block ${navLinkClass(navActive(item, location.pathname))}`}>
                {item.label}
              </Link>
            ))}
            <div className="my-1 border-t border-surface-3" />
            <Link to="/settings" className={`block ${navLinkClass(location.pathname === '/settings')}`}>
              設定
            </Link>
            <button
              type="button"
              onClick={handleLogout}
              className="block w-full text-left px-3 py-1.5 rounded-md text-sm font-medium text-[var(--color-text-muted)] hover:bg-red-900/10 hover:text-red-700 transition-colors"
            >
              ログアウト
            </button>
          </nav>
        </div>
      )}
    </>
  );
}
