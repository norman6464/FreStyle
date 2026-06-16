import { useEffect, useRef, useState } from 'react';
import { useLocation, Link } from 'react-router-dom';
import { useSelector } from 'react-redux';
import {
  BellIcon,
  Bars3Icon,
  XMarkIcon,
  ChevronDownIcon,
} from '@heroicons/react/24/outline';
import Loading from '../Loading';
import HeaderUserMenu from './HeaderUserMenu';
import { useSidebar } from '../../hooks/useSidebar';
import { NotificationRepository } from '../../repositories/NotificationRepository';
import ProfileRepository from '../../repositories/ProfileRepository';
import type { RootState } from '../../store';

interface NavItem {
  id: string;
  label: string;
  to: string;
  matchExact?: boolean;
  matchPrefix?: string;
}

// ヘッダーのメインナビ（テキストのみ。 アイコンは使わない）。
// 通知はベル、 管理はドロップダウンに分けるため、 ここには含めない。
const mainNavItems: NavItem[] = [
  { id: 'home', label: 'ホーム', to: '/', matchExact: true },
  { id: 'ai', label: 'AI', to: '/chat/ask-ai', matchPrefix: '/chat/ask-ai' },
  { id: 'code', label: '演習', to: '/code-editor', matchPrefix: '/code-editor' },
  { id: 'courses', label: 'コース', to: '/courses', matchPrefix: '/courses' },
  { id: 'notes', label: 'ノート', to: '/notes', matchPrefix: '/notes' },
  { id: 'reports', label: 'レポート', to: '/reports', matchExact: true },
];

// super_admin は企業管理に専念するロールなので学習系メニューは出さない（ホームのみ）。
const SUPER_ADMIN_MAIN_NAV_IDS = new Set(['home']);

interface AdminSub {
  label: string;
  to: string;
  matchPrefix: string;
  allowedRoles?: ReadonlyArray<'super_admin' | 'company_admin'>;
}

const adminSubItems: AdminSub[] = [
  { label: '概況', to: '/admin/dashboard', matchPrefix: '/admin/dashboard', allowedRoles: ['super_admin'] },
  { label: '会社一覧', to: '/admin/companies', matchPrefix: '/admin/companies', allowedRoles: ['super_admin'] },
  { label: '利用申請', to: '/admin/applications', matchPrefix: '/admin/applications', allowedRoles: ['super_admin'] },
  { label: '従業員一覧', to: '/admin/members', matchPrefix: '/admin/members' },
  { label: '招待管理', to: '/admin/invitations', matchPrefix: '/admin/invitations' },
  { label: '監査ログ', to: '/admin/audit', matchPrefix: '/admin/audit', allowedRoles: ['super_admin'] },
];

function navActive(item: NavItem, pathname: string): boolean {
  if (item.matchExact) return pathname === item.to;
  if (item.matchPrefix) return pathname.startsWith(item.matchPrefix);
  return pathname === item.to;
}

function roleLabel(role: string | null): string {
  switch (role) {
    case 'super_admin': return '運営管理者';
    case 'company_admin': return '会社管理者';
    case 'trainee': return '受講者';
    default: return '';
  }
}

/**
 * Header — 上部固定のテキスト横並びナビ。
 *
 * 左: ロゴ ／ 中央左: テキストナビ（アイコンなし） ／ 右: 通知ベル + 管理ドロップダウン(admin) + ユーザーメニュー。
 * モバイルではハンバーガーで縦メニューを開く。 ロール出し分けはサイドバー時代の仕様を踏襲する。
 */
export default function Header() {
  const location = useLocation();
  const { handleLogout, loggingOut } = useSidebar();
  const isAdmin = useSelector((s: RootState) => s.auth.isAdmin);
  const role = useSelector((s: RootState) => s.auth.role);
  const aiChatEnabledForTrainees = useSelector((s: RootState) => s.auth.aiChatEnabledForTrainees);
  const isSuperAdmin = role === 'super_admin';

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

  const visibleNav = (
    isSuperAdmin
      ? mainNavItems.filter((i) => SUPER_ADMIN_MAIN_NAV_IDS.has(i.id))
      : mainNavItems
  ).filter((i) => !(i.id === 'ai' && role === 'trainee' && !aiChatEnabledForTrainees));

  const visibleAdminSubs = adminSubItems.filter(
    (s) => !s.allowedRoles || (role !== null && s.allowedRoles.includes(role as 'super_admin' | 'company_admin')),
  );

  const navLinkClass = (active: boolean) =>
    `px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
      active
        ? 'bg-[var(--color-nav-active)] text-[var(--color-text-primary)]'
        : 'text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)]'
    }`;

  return (
    <>
      {loggingOut && <Loading fullscreen message="ログアウト中..." />}
      <header className="flex-shrink-0 h-14 bg-[var(--color-nav)] border-b border-surface-3 flex items-center gap-2 px-3">
        {/* ロゴ（色付き箱 + サービス名） */}
        <Link to="/" className="flex items-center gap-2 flex-shrink-0 mr-2" aria-label="FreStyle ホーム">
          <span className="flex items-center justify-center w-8 h-8 rounded-md bg-brand-500">
            <img src="/favicon.svg" alt="" className="w-5 h-5" />
          </span>
          <span className="hidden sm:block text-sm font-semibold text-[var(--color-text-primary)]">FreStyle</span>
        </Link>

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
                aria-haspopup="menu"
                aria-expanded={adminOpen}
                className={`${navLinkClass(location.pathname.startsWith('/admin'))} inline-flex items-center gap-1`}
              >
                管理
                <ChevronDownIcon className={`w-3.5 h-3.5 transition-transform ${adminOpen ? 'rotate-180' : ''}`} />
              </button>
              {adminOpen && (
                <div role="menu" className="absolute top-full right-0 mt-2 w-44 bg-surface-1 border border-surface-3 rounded-lg shadow-lg overflow-hidden z-50 animate-fade-in">
                  {visibleAdminSubs.map((sub) => (
                    <Link
                      key={sub.to}
                      to={sub.to}
                      role="menuitem"
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
        <div className="md:hidden border-b border-surface-3 bg-[var(--color-nav)]">
          <nav className="px-3 py-2 space-y-0.5" aria-label="モバイルナビゲーション">
            {visibleNav.map((item) => (
              <Link key={item.id} to={item.to} className={`block ${navLinkClass(navActive(item, location.pathname))}`}>
                {item.label}
              </Link>
            ))}
            {isAdmin && visibleAdminSubs.length > 0 && (
              <>
                <div className="my-1 border-t border-surface-3" />
                <p className="px-3 py-1 text-xs text-[var(--color-text-muted)]">管理</p>
                {visibleAdminSubs.map((sub) => (
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
