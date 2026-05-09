import { useState, ComponentType, SVGProps } from 'react';
import { useLocation, Link } from 'react-router-dom';
import { useSelector } from 'react-redux';
import {
  HomeIcon,
  SparklesIcon,
  DocumentTextIcon,
  BellIcon,
  DocumentChartBarIcon,
  UserCircleIcon,
  ArrowLeftOnRectangleIcon,
  CodeBracketIcon,
  BuildingOffice2Icon,
  ChevronDoubleLeftIcon,
  ChevronDoubleRightIcon,
} from '@heroicons/react/24/outline';
import Loading from '../Loading';
import { useSidebar } from '../../hooks/useSidebar';
import type { RootState } from '../../store';

interface SubItem {
  label: string;
  to: string;
  matchPrefix?: string;
  /**
   * このサブメニューを表示できる役職集合。
   * undefined のときは全 admin（super_admin / company_admin 両方）に表示する。
   */
  allowedRoles?: ReadonlyArray<'super_admin' | 'company_admin'>;
}

interface NavItem {
  id: string;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  label: string;
  to?: string;
  matchExact?: boolean;
  matchPrefix?: string;
  subItems?: SubItem[];
}

const mainNavItems: NavItem[] = [
  { id: 'home', icon: HomeIcon, label: 'ホーム', to: '/', matchExact: true },
  { id: 'ai', icon: SparklesIcon, label: 'AI', to: '/chat/ask-ai', matchPrefix: '/chat/ask-ai' },
  { id: 'code', icon: CodeBracketIcon, label: 'コード学習', to: '/code-editor', matchPrefix: '/code-editor' },
  { id: 'notes', icon: DocumentTextIcon, label: 'ノート', to: '/notes', matchPrefix: '/notes' },
  { id: 'notifications', icon: BellIcon, label: '通知', to: '/notifications', matchExact: true },
  { id: 'reports', icon: DocumentChartBarIcon, label: 'レポート', to: '/reports', matchExact: true },
];

// super_admin に表示する main nav の id 集合。
// super_admin は企業管理に専念するロールで、trainee 向け学習機能は表示しない。
const SUPER_ADMIN_MAIN_NAV_IDS = new Set(['home', 'notifications']);

const bottomNavItems: NavItem[] = [
  { id: 'profile', icon: UserCircleIcon, label: 'プロフィール', to: '/profile/me', matchExact: true },
];

const adminNavItem: NavItem = {
  id: 'admin',
  icon: BuildingOffice2Icon,
  label: '管理',
  matchPrefix: '/admin',
  subItems: [
    // 会社一覧は全テナントの管理画面なので運営 (super_admin) 専用。
    // company_admin が見えると越権が起きる印象を与えるため非表示にする。
    {
      label: '会社一覧',
      to: '/admin/companies',
      matchPrefix: '/admin/companies',
      allowedRoles: ['super_admin'],
    },
    // 招待管理は自社内で trainee を招待する操作のため company_admin / super_admin 双方が利用する。
    { label: '招待管理', to: '/admin/invitations', matchPrefix: '/admin/invitations' },
  ],
};

interface SidebarProps {
  onNavigate?: () => void;
}

function isActive(item: NavItem, pathname: string): boolean {
  if (item.matchExact) return pathname === item.to;
  if (item.matchPrefix) return pathname.startsWith(item.matchPrefix);
  if (item.to) return pathname === item.to;
  return false;
}

export default function Sidebar({ onNavigate }: SidebarProps) {
  const location = useLocation();
  const { handleLogout, loggingOut } = useSidebar();
  const isAdmin = useSelector((state: RootState) => state.auth.isAdmin);
  const role = useSelector((state: RootState) => state.auth.role);
  const isSuperAdmin = role === 'super_admin';

  // expanded: パネル（ラベル）表示状態。false = アイコンのみの折りたたみ状態
  const [expanded, setExpanded] = useState(true);
  // 管理サブメニューの開閉
  const [adminOpen, setAdminOpen] = useState(false);

  // super_admin は trainee 向けメニュー（AI / コード / ノート / レポート）を非表示。
  const visibleMainNavItems = isSuperAdmin
    ? mainNavItems.filter((item) => SUPER_ADMIN_MAIN_NAV_IDS.has(item.id))
    : mainNavItems;
  const adminActive = isActive(adminNavItem, location.pathname);

  const handleAdminClick = () => {
    if (!expanded) {
      setExpanded(true);
      setAdminOpen(true);
    } else {
      setAdminOpen((prev) => !prev);
    }
  };

  return (
    <>
      {loggingOut && <Loading fullscreen message="ログアウト中..." />}
      <aside
        className={`flex flex-col h-full bg-[var(--color-nav)] border-r border-surface-3 flex-shrink-0 transition-all duration-200 ${
          expanded ? 'w-56' : 'w-14'
        }`}
      >
        {/* トグルボタン */}
        <div className={`flex ${expanded ? 'justify-end' : 'justify-center'} px-2 pt-3 pb-1`}>
          <button
            onClick={() => { setExpanded((v) => !v); if (expanded) setAdminOpen(false); }}
            title={expanded ? 'サイドバーを閉じる' : 'サイドバーを開く'}
            className="p-1.5 rounded-md text-[var(--color-text-muted)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)] transition-colors"
          >
            {expanded
              ? <ChevronDoubleLeftIcon className="w-4 h-4" />
              : <ChevronDoubleRightIcon className="w-4 h-4" />
            }
          </button>
        </div>

        {/* メインナビ */}
        <nav className="flex-1 px-2 space-y-0.5 overflow-y-auto">
          {visibleMainNavItems.map((item) => {
            const active = isActive(item, location.pathname);
            return (
              <Link
                key={item.id}
                to={item.to!}
                onClick={onNavigate}
                title={!expanded ? item.label : undefined}
                className={`flex items-center gap-3 px-2 py-2 rounded-md text-sm font-medium transition-colors duration-150 ${
                  active
                    ? 'bg-[var(--color-nav-active)] text-[var(--color-text-primary)]'
                    : 'text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)]'
                }`}
              >
                <item.icon className="w-5 h-5 flex-shrink-0" />
                {expanded && <span className="truncate">{item.label}</span>}
              </Link>
            );
          })}

          <div className="my-2 border-t border-surface-3" />

          {/* プロフィール */}
          {bottomNavItems.map((item) => {
            const active = isActive(item, location.pathname);
            return (
              <Link
                key={item.id}
                to={item.to!}
                onClick={onNavigate}
                title={!expanded ? item.label : undefined}
                className={`flex items-center gap-3 px-2 py-2 rounded-md text-sm font-medium transition-colors duration-150 ${
                  active
                    ? 'bg-[var(--color-nav-active)] text-[var(--color-text-primary)]'
                    : 'text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)]'
                }`}
              >
                <item.icon className="w-5 h-5 flex-shrink-0" />
                {expanded && <span className="truncate">{item.label}</span>}
              </Link>
            );
          })}

          {/* 管理メニュー (admin only) */}
          {isAdmin && (
            <>
              <div className="my-2 border-t border-surface-3" />
              <button
                onClick={handleAdminClick}
                title={!expanded ? adminNavItem.label : undefined}
                className={`flex items-center gap-3 px-2 py-2 rounded-md text-sm font-medium w-full transition-colors duration-150 ${
                  adminActive
                    ? 'bg-[var(--color-nav-active)] text-[var(--color-text-primary)]'
                    : 'text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)]'
                }`}
              >
                <adminNavItem.icon className="w-5 h-5 flex-shrink-0" />
                {expanded && (
                  <>
                    <span className="flex-1 text-left truncate">{adminNavItem.label}</span>
                    <ChevronDoubleRightIcon
                      className={`w-3.5 h-3.5 flex-shrink-0 transition-transform ${adminOpen ? 'rotate-90' : ''}`}
                    />
                  </>
                )}
              </button>

              {/* 管理サブメニュー */}
              {expanded && adminOpen && (
                <div className="ml-4 space-y-0.5">
                  {adminNavItem.subItems
                    ?.filter((sub) => {
                      if (!sub.allowedRoles) return true;
                      return role !== null && sub.allowedRoles.includes(role as 'super_admin' | 'company_admin');
                    })
                    .map((sub) => {
                    const subActive = sub.matchPrefix
                      ? location.pathname.startsWith(sub.matchPrefix)
                      : location.pathname === sub.to;
                    return (
                      <Link
                        key={sub.to}
                        to={sub.to}
                        onClick={onNavigate}
                        className={`flex items-center px-2 py-1.5 rounded-md text-xs transition-colors ${
                          subActive
                            ? 'bg-[var(--color-nav-active)] text-[var(--color-text-primary)]'
                            : 'text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)]'
                        }`}
                      >
                        {sub.label}
                      </Link>
                    );
                  })}
                </div>
              )}
            </>
          )}
        </nav>

        {/* ログアウト（ダークモード切替は廃止） */}
        <div className="px-2 py-3 border-t border-surface-3 space-y-0.5">
          <button
            onClick={() => { onNavigate?.(); handleLogout(); }}
            title="ログアウト"
            className="flex items-center gap-3 px-2 py-2 rounded-md text-sm font-medium text-[var(--color-text-muted)] hover:bg-red-900/30 hover:text-red-400 transition-colors duration-150 w-full"
          >
            <ArrowLeftOnRectangleIcon className="w-5 h-5 flex-shrink-0" />
            {expanded && <span>ログアウト</span>}
          </button>
        </div>
      </aside>
    </>
  );
}
