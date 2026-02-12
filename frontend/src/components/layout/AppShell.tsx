import { useState } from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import Sidebar from './Sidebar';
import TopBar from './TopBar';

const pageTitles: Record<string, string> = {
  '/': 'ホーム',
  '/chat': 'チャット',
  '/chat/ask-ai': 'AI アシスタント',
  '/practice': '練習モード',
  '/chat/users': 'ユーザー検索',
  '/scores': 'スコア履歴',
  '/chat/members': 'メンバー',
  '/profile/me': 'プロフィール',
  '/profile/personality': 'パーソナリティ設定',
};

function getPageTitle(pathname: string): string {
  // 完全一致
  if (pageTitles[pathname]) return pageTitles[pathname];
  // プレフィックス一致
  if (pathname.startsWith('/chat/ask-ai/')) return 'AI アシスタント';
  if (pathname.startsWith('/chat/users/')) return 'チャット';
  return 'FreStyle';
}

export default function AppShell() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const location = useLocation();
  const title = getPageTitle(location.pathname);

  return (
    <div className="h-screen flex bg-slate-50">
      {/* デスクトップサイドバー */}
      <div className="hidden md:block">
        <Sidebar />
      </div>

      {/* モバイルオーバーレイ */}
      {mobileMenuOpen && (
        <div
          className="fixed inset-0 bg-black/40 z-40 md:hidden"
          onClick={() => setMobileMenuOpen(false)}
        />
      )}

      {/* モバイルサイドバー */}
      <div
        className={`fixed inset-y-0 left-0 z-50 w-56 transform transition-transform duration-200 md:hidden ${
          mobileMenuOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <Sidebar onNavigate={() => setMobileMenuOpen(false)} />
      </div>

      {/* メインコンテンツ */}
      <div className="flex-1 flex flex-col min-w-0">
        <TopBar
          title={title}
          onMenuToggle={() => setMobileMenuOpen(!mobileMenuOpen)}
        />
        <main className="flex-1 overflow-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
