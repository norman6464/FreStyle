import { useLocation } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { clearAuth } from '../../store/authSlice';
import SidebarItem from './SidebarItem';
import {
  HomeIcon,
  ChatBubbleLeftRightIcon,
  SparklesIcon,
  AcademicCapIcon,
  MagnifyingGlassIcon,
  ChartBarIcon,
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
  const dispatch = useDispatch();
  const navigate = useNavigate();

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

  const handleLogout = async () => {
    try {
      const res = await fetch(
        `${import.meta.env.VITE_API_BASE_URL}/api/auth/cognito/logout`,
        {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
        }
      );
      if (!res.ok) {
        alert('ログアウトに失敗しました');
        return;
      }
      dispatch(clearAuth());
      navigate('/login');
    } catch (err) {
      console.error('ログアウトエラー:', err);
      alert('ログアウト中にエラーが発生しました');
    }
  };

  return (
    <aside className="flex flex-col w-56 h-full bg-white border-r border-slate-200 flex-shrink-0">
      {/* ロゴ */}
      <div className="h-12 flex items-center px-4 border-b border-slate-200">
        <div className="bg-primary-600 w-7 h-7 rounded flex items-center justify-center mr-2.5">
          <span className="text-white font-bold text-sm">F</span>
        </div>
        <span className="text-sm font-bold text-slate-800">FreStyle</span>
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
