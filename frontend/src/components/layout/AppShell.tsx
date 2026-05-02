import { useState, useEffect, useCallback } from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import Sidebar from './Sidebar';
import TopBar from './TopBar';
import SkipLink from '../SkipLink';
import ScrollToTop from '../ScrollToTop';
import CommandPalette from '../CommandPalette';

const pageTitles: Record<string, string> = {
  '/': 'ホーム',
  '/chat/ask-ai': 'AI アシスタント',
  '/practice': '練習モード',
  '/scores': 'スコア履歴',
  '/profile/me': 'プロフィール',
  '/notes': 'ノート',
};

function getPageTitle(pathname: string): string {
  if (pageTitles[pathname]) return pageTitles[pathname];
  if (pathname.startsWith('/chat/ask-ai/')) return 'AI アシスタント';
  return 'FreStyle';
}

export default function AppShell() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);
  const location = useLocation();
  const title = getPageTitle(location.pathname);

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      setCommandPaletteOpen(prev => !prev);
    }
  }, []);

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  return (
    <div className="h-screen flex bg-surface">
      <SkipLink targetId="main-content" />
      {/* デスクトップサイドバー */}
      <div className="hidden md:block h-full">
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
      {/*
        flex-1 flex-col / min-h-0 / min-w-0 / overflow-hidden を組み合わせるのが
        Tailwind + Flexbox 内で「内側だけ overflow-auto してくれ」を成立させる
        必須レシピ。これらが無いと flex item の default min-height: auto により
        子コンテンツが column 全体を伸ばしてしまい、AppShell の h-screen を超え、
        body 側にも 2 本目のスクロールバーが発生する (実際 MenuPage でだけ顕在化していた)。
      */}
      <div className="flex-1 flex flex-col min-w-0 min-h-0 overflow-hidden">
        <TopBar
          title={title}
          onMenuToggle={() => setMobileMenuOpen(!mobileMenuOpen)}
        />
        <main
          id="main-content"
          tabIndex={-1}
          className="flex-1 min-h-0 overflow-auto outline-none"
        >
          <Outlet />
        </main>
        <ScrollToTop targetId="main-content" />
      </div>
      <CommandPalette
        isOpen={commandPaletteOpen}
        onClose={() => setCommandPaletteOpen(false)}
      />
    </div>
  );
}
