import { useState } from 'react';
import { useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { clearAuthData } from '../store/authSlice';

export default function HamburgerMenu({ title }) {
  const [isOpen, setIsOpen] = useState(false);
  const dispatch = useDispatch();
  const navigate = useNavigate();

  const handleLogout = async () => {
    fetch(`${import.meta.env.VITE_API_BASE_URL}/api/auth/cognito/logout`, {
      method: 'POST',
      credentials: 'include',
    }).then((res) => {
      if (!res.ok) {
        return alert('失敗しました');
      }
      dispatch(clearAuthData());
      navigate('/login');
    });
  };

  const menuItems = [
    { label: 'プロフィールを編集', onClick: () => navigate('/profile/me') },
    { label: 'ユーザー検索', onClick: () => navigate('/chat/users') },
    { label: 'AIに聞いてみる', onClick: () => navigate('/chat/ask-ai') },
    { label: 'ログアウト', onClick: handleLogout },
  ];

  return (
    <header className="w-full bg-gray-50 shadow-sm fixed top-0 left-0 z-50">
      <div className="max-w-4xl mx-auto px-3 py-5 flex justify-between items-center">
        {/* 左側ロゴ（タイトル） */}
        <button onClick={() => navigate('/')}>
          <h1 className="text-lg font-semibold text-gray-800">{title}</h1>
        </button>
        {/* ハンバーガーボタン（モバイル用） */}
        <button
          className="md:hidden p-2 text-gray-700 rounded-md hover:bg-gray-200 focus:outline-none"
          onClick={() => setIsOpen(!isOpen)}
        >
          <svg
            className="w-6 h-6"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            {isOpen ? (
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M6 18L18 6M6 6l12 12"
              />
            ) : (
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M4 6h16M4 12h16M4 18h16"
              />
            )}
          </svg>
        </button>

        {/* デスクトップ用ナビゲーション */}
        <nav className="hidden md:flex flex-1 justify-center space-x-10">
          {menuItems.map((item, index) => (
            <button
              key={index}
              onClick={item.onClick}
              className="text-gray-700 hover:text-blue-600 font-medium transition"
            >
              {item.label}
            </button>
          ))}
        </nav>

        {/* 右側の空スペースでバランスを取る */}
        <div className="hidden md:block w-10" />
      </div>

      {/* モバイル用ドロップダウン（オーバーレイ） */}
      {isOpen && (
        <div className="absolute top-full left-0 w-full bg-white shadow-lg z-50">
          {menuItems.map((item, index) => (
            <button
              key={index}
              onClick={() => {
                item.onClick();
                setIsOpen(false);
              }}
              className="block w-full text-left px-4 py-3 border-b border-gray-200 hover:bg-gray-100 text-gray-700 font-medium"
            >
              {item.label}
            </button>
          ))}
        </div>
      )}
    </header>
  );
}
