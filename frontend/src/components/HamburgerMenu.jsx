import { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { clearAuthData } from '../store/authSlice';
import {
  Bars3Icon,
  XMarkIcon,
  HomeIcon,
  UserCircleIcon,
  MagnifyingGlassIcon,
  SparklesIcon,
  ArrowLeftOnRectangleIcon,
} from '@heroicons/react/24/solid';

export default function HamburgerMenu({ title }) {
  const [isOpen, setIsOpen] = useState(false);
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const accessToken = useSelector((state) => state.auth.accessToken);

  const handleLogout = async () => {
    try {
      const res = await fetch(
        `${import.meta.env.VITE_API_BASE_URL}/api/auth/cognito/logout`,
        {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${accessToken}`,
          },
        }
      );

      if (!res.ok) {
        return alert('ログアウトに失敗しました');
      }

      dispatch(clearAuthData());
      navigate('/login');
    } catch (err) {
      console.error('ログアウトエラー:', err);
      alert('ログアウト中にエラーが発生しました');
    }
  };

  const menuItems = [
    {
      label: 'ホーム',
      icon: HomeIcon,
      onClick: () => navigate('/'),
      color: 'text-primary-600',
    },
    {
      label: 'プロフィール',
      icon: UserCircleIcon,
      onClick: () => navigate('/profile/me'),
      color: 'text-secondary-600',
    },
    {
      label: 'ユーザー検索',
      icon: MagnifyingGlassIcon,
      onClick: () => navigate('/chat/users'),
      color: 'text-blue-600',
    },
    {
      label: 'AI',
      icon: SparklesIcon,
      onClick: () => navigate('/chat/ask-ai'),
      color: 'text-pink-600',
    },
    {
      label: 'ログアウト',
      icon: ArrowLeftOnRectangleIcon,
      onClick: handleLogout,
      color: 'text-red-600',
    },
  ];

  return (
    <header className="w-full bg-gradient-to-r from-white to-gray-50 shadow-md fixed top-0 left-0 z-50 border-b border-gray-100">
      <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
        <button
          onClick={() => navigate('/')}
          className="flex items-center space-x-2"
        >
          <div className="bg-gradient-primary w-10 h-10 rounded-lg flex items-center justify-center">
            <span className="text-white font-bold text-lg">F</span>
          </div>
          <h1 className="text-2xl font-bold bg-gradient-to-r from-primary-600 to-secondary-600 bg-clip-text text-transparent hidden sm:block">
            {title}
          </h1>
        </button>

        {/* デスクトップナビゲーション */}
        <nav className="hidden md:flex items-center space-x-1">
          {menuItems.map((item, index) => {
            const Icon = item.icon;
            return (
              <button
                key={index}
                onClick={item.onClick}
                className={`flex items-center space-x-2 px-4 py-2 rounded-lg transition-all duration-200 hover:bg-gray-100 ${item.color} font-medium group`}
              >
                <Icon className="w-5 h-5 group-hover:scale-110 transition-transform" />
                <span className="text-gray-700">{item.label}</span>
              </button>
            );
          })}
        </nav>

        {/* モバイルメニューボタン */}
        <button
          className="md:hidden p-2 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
          onClick={() => setIsOpen(!isOpen)}
        >
          {isOpen ? (
            <XMarkIcon className="w-6 h-6" />
          ) : (
            <Bars3Icon className="w-6 h-6" />
          )}
        </button>
      </div>

      {/* モバイルメニュー */}
      {isOpen && (
        <div className="md:hidden bg-white border-t border-gray-200 shadow-lg animate-fade-in">
          {menuItems.map((item, index) => {
            const Icon = item.icon;
            return (
              <button
                key={index}
                onClick={() => {
                  item.onClick();
                  setIsOpen(false);
                }}
                className={`flex items-center w-full px-4 py-4 border-b border-gray-100 hover:bg-gray-50 transition-colors ${item.color}`}
              >
                <Icon className="w-5 h-5 mr-3" />
                <span className="font-semibold text-gray-800">
                  {item.label}
                </span>
              </button>
            );
          })}
        </div>
      )}
    </header>
  );
}
