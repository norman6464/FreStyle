import { useLocation } from 'react-router-dom';
import SidebarItem from './SidebarItem';
import { useSidebar } from '../../hooks/useSidebar';
import {
  HomeIcon,
  ChatBubbleLeftRightIcon,
  SparklesIcon,
  AcademicCapIcon,
  MagnifyingGlassIcon,
  ChartBarIcon,
  StarIcon,
  UserCircleIcon,
  LightBulbIcon,
  ArrowLeftOnRectangleIcon,
} from '@heroicons/react/24/outline';

const navItems = [
  { icon: HomeIcon, label: 'ホーム', to: '/', matchExact: true },
  { icon: ChatBubbleLeftRightIcon, label: 'チャット', to: '/chat', matchPrefix: '/chat', excludePrefix: ['/chat/ask-ai', '/chat/users'] },
  { icon: SparklesIcon, label: 'AI', to: '/chat/ask-ai', matchPrefix: '/chat/ask-ai' },
  { icon: AcademicCapIcon, label: '練習', to: '/practice', matchPrefix: '/practice' },
  { icon: MagnifyingGlassIcon, label: 'ユーザー検索', to: '/chat/users', matchExact: true },
  { icon: ChartBarIcon, label: 'スコア履歴', to: '/scores', matchExact: true },
  { icon: StarIcon, label: 'お気に入り', to: '/favorites', matchExact: true },
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
    <aside className="flex flex-col w-56 h-full bg-white border-r border-slate-200 flex-shrink-0">
      {/* ロゴ */}
      <div className="h-14 flex items-center px-4 border-b border-slate-200 gap-2.5">
        <img src="/image.png" alt="FreStyle" className="w-11 h-11 rounded-xl object-contain" />
        <span className="text-base font-bold text-slate-800 tracking-tight">FreStyle</span>
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

        <div className="my-3 border-t border-slate-200" />

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

      {/* ログアウト */}
      <div className="px-2 py-3 border-t border-slate-200">
        <button
          onClick={() => { onNavigate?.(); handleLogout(); }}
          className="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium text-slate-500 hover:bg-red-50 hover:text-red-600 transition-colors duration-150 w-full"
        >
          <ArrowLeftOnRectangleIcon className="w-5 h-5 flex-shrink-0" />
          <span>ログアウト</span>
        </button>
      </div>
    </aside>
  );
}
