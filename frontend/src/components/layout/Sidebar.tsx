import { useLocation } from 'react-router-dom';
import SidebarItem from './SidebarItem';
import { useSidebar } from '../../hooks/useSidebar';
import { useTheme } from '../../hooks/useTheme';
import {
  HomeIcon,
  ChatBubbleLeftRightIcon,
  SparklesIcon,
  AcademicCapIcon,
  MagnifyingGlassIcon,
  ChartBarIcon,
  StarIcon,
  DocumentTextIcon,
  BellIcon,
  UserCircleIcon,
  LightBulbIcon,
  ArrowLeftOnRectangleIcon,
  SunIcon,
  MoonIcon,
} from '@heroicons/react/24/outline';

const navItems = [
  { icon: HomeIcon, label: 'ホーム', to: '/', matchExact: true },
  { icon: ChatBubbleLeftRightIcon, label: 'チャット', to: '/chat', matchPrefix: '/chat', excludePrefix: ['/chat/ask-ai', '/chat/users'] },
  { icon: SparklesIcon, label: 'AI', to: '/chat/ask-ai', matchPrefix: '/chat/ask-ai' },
  { icon: AcademicCapIcon, label: '練習', to: '/practice', matchPrefix: '/practice' },
  { icon: MagnifyingGlassIcon, label: 'ユーザー検索', to: '/chat/users', matchExact: true },
  { icon: ChartBarIcon, label: 'スコア履歴', to: '/scores', matchExact: true },
  { icon: StarIcon, label: 'お気に入り', to: '/favorites', matchExact: true },
  { icon: DocumentTextIcon, label: 'ノート', to: '/notes', matchPrefix: '/notes' },
  { icon: BellIcon, label: '通知', to: '/notifications', matchExact: true },
];

const bottomNavItems = [
  { icon: UserCircleIcon, label: 'プロフィール', to: '/profile/me', matchExact: true },
  { icon: LightBulbIcon, label: 'パーソナリティ', to: '/profile/personality', matchExact: true },
];

interface SidebarProps {
  onNavigate?: () => void;
}

export default function Sidebar({ onNavigate }: SidebarProps) {
  const location = useLocation();
  const { totalUnread, handleLogout } = useSidebar();
  const { theme, toggleTheme } = useTheme();

  const isActive = (item: typeof navItems[0]) => {
    if (item.matchExact) {
      return location.pathname === item.to;
    }
    if (item.matchPrefix) {
      const matches = location.pathname.startsWith(item.matchPrefix);
      if (matches && item.excludePrefix) {
        return !item.excludePrefix.some(p => location.pathname.startsWith(p));
      }
      return matches;
    }
    return false;
  };

  return (
    <aside className="flex flex-col w-56 h-full bg-surface-1 border-r border-surface-3 flex-shrink-0">
      {/* ロゴ */}
      <div className="h-14 flex items-center px-4 border-b border-surface-3 gap-2.5">
        <img src="/image.png" alt="FreStyle" className="w-11 h-11 rounded-xl object-contain" />
        <span className="text-base font-bold text-[var(--color-text-primary)] tracking-tight">FreStyle</span>
      </div>

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
            badge={item.label === 'チャット' ? totalUnread : undefined}
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
  );
}
