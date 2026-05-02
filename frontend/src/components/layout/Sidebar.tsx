import { useLocation } from 'react-router-dom';
import { useSelector } from 'react-redux';
import SidebarItem from './SidebarItem';
import Loading from '../Loading';
import { useSidebar } from '../../hooks/useSidebar';
import { useTheme } from '../../hooks/useTheme';
import type { RootState } from '../../store';
import {
  HomeIcon,
  SparklesIcon,
  AcademicCapIcon,
  ChartBarIcon,
  StarIcon,
  DocumentTextIcon,
  BellIcon,
  DocumentChartBarIcon,
  UserCircleIcon,
  ArrowLeftOnRectangleIcon,
  SunIcon,
  MoonIcon,
  Cog6ToothIcon,
  UserPlusIcon,
} from '@heroicons/react/24/outline';

const navItems = [
  { icon: HomeIcon, label: 'ホーム', to: '/', matchExact: true },
  { icon: SparklesIcon, label: 'AI', to: '/chat/ask-ai', matchPrefix: '/chat/ask-ai' },
  { icon: AcademicCapIcon, label: '練習', to: '/practice', matchPrefix: '/practice' },
  { icon: ChartBarIcon, label: 'スコア履歴', to: '/scores', matchExact: true },
  { icon: StarIcon, label: 'お気に入り', to: '/favorites', matchExact: true },
  { icon: DocumentTextIcon, label: 'ノート', to: '/notes', matchPrefix: '/notes' },
  { icon: BellIcon, label: '通知', to: '/notifications', matchExact: true },
  { icon: DocumentChartBarIcon, label: 'レポート', to: '/reports', matchExact: true },
];

const bottomNavItems = [
  { icon: UserCircleIcon, label: 'プロフィール', to: '/profile/me', matchExact: true },
];

interface SidebarProps {
  onNavigate?: () => void;
}

export default function Sidebar({ onNavigate }: SidebarProps) {
  const location = useLocation();
  const { handleLogout, loggingOut } = useSidebar();
  const { theme, toggleTheme } = useTheme();
  const isAdmin = useSelector((state: RootState) => state.auth.isAdmin);

  const isActive = (item: typeof navItems[0]) => {
    if (item.matchExact) {
      return location.pathname === item.to;
    }
    if (item.matchPrefix) {
      return location.pathname.startsWith(item.matchPrefix);
    }
    return false;
  };

  return (
    <>
    {loggingOut && <Loading fullscreen message="ログアウト中..." />}
    <aside className="flex flex-col w-56 h-full bg-surface-1 border-r border-surface-3 flex-shrink-0">
      {/* メインナビ */}
      <nav className="flex-1 px-2 py-3 space-y-0.5 overflow-y-auto">
        {navItems.map(item => (
          <SidebarItem
            key={item.to}
            icon={item.icon}
            label={item.label}
            to={item.to}
            active={isActive(item)}
            onClick={onNavigate}
            badge={undefined}
          />
        ))}

        <div className="my-3 border-t border-surface-3" />

        {bottomNavItems.map(item => (
          <SidebarItem
            key={item.to}
            icon={item.icon}
            label={item.label}
            to={item.to}
            active={isActive(item)}
            onClick={onNavigate}
          />
        ))}

        {isAdmin && (
          <>
            <div className="my-3 border-t border-surface-3" />
            <SidebarItem
              icon={Cog6ToothIcon}
              label="管理: シナリオ"
              to="/admin/scenarios"
              active={location.pathname.startsWith('/admin/scenarios')}
              onClick={onNavigate}
            />
            <SidebarItem
              icon={UserPlusIcon}
              label="管理: 招待"
              to="/admin/invitations"
              active={location.pathname.startsWith('/admin/invitations')}
              onClick={onNavigate}
            />
          </>
        )}
      </nav>

      {/* テーマ切り替え・ログアウト */}
      <div className="px-2 py-3 border-t border-surface-3 space-y-0.5">
        <button
          onClick={toggleTheme}
          className="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium text-[var(--color-text-muted)] hover:bg-surface-2 hover:text-[var(--color-text-secondary)] transition-colors duration-150 w-full"
        >
          {theme === 'dark' ? (
            <SunIcon className="w-5 h-5 flex-shrink-0" />
          ) : (
            <MoonIcon className="w-5 h-5 flex-shrink-0" />
          )}
          <span>{theme === 'dark' ? 'ライトモード' : 'ダークモード'}</span>
        </button>
        <button
          onClick={() => { onNavigate?.(); handleLogout(); }}
          className="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium text-[var(--color-text-muted)] hover:bg-red-900/30 hover:text-red-400 transition-colors duration-150 w-full"
        >
          <ArrowLeftOnRectangleIcon className="w-5 h-5 flex-shrink-0" />
          <span>ログアウト</span>
        </button>
      </div>
    </aside>
    </>
  );
}
