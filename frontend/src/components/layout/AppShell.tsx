import { useState, useEffect, useCallback } from 'react';
import { Outlet } from 'react-router-dom';
import { Bars3Icon } from '@heroicons/react/24/outline';
import Sidebar from './Sidebar';
import SkipLink from '../SkipLink';
import ScrollToTop from '../ScrollToTop';
import CommandPalette from '../CommandPalette';

export default function AppShell() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);

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
    <div className="h-screen flex bg-surface overflow-hidden">
      <SkipLink targetId="main-content" />

      {/* モバイル用ハンバーガー（サイドバーが閉じているときのみ表示） */}
      {!mobileMenuOpen && (
        <button
          onClick={() => setMobileMenuOpen(true)}
          className="md:hidden fixed top-3 left-3 z-30 p-2 bg-[var(--color-nav)] border border-surface-3 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-nav-hover)] rounded-md transition-colors shadow-sm"
          aria-label="メニュー"
        >
          <Bars3Icon className="w-5 h-5" />
        </button>
      )}

      {/* モバイルオーバーレイ */}
      {mobileMenuOpen && (
        <div
          className="fixed inset-0 bg-black/40 z-40 md:hidden"
          onClick={() => setMobileMenuOpen(false)}
        />
      )}

      {/* モバイルサイドバー */}
      <div
        className={`fixed inset-y-0 left-0 z-50 transform transition-transform duration-200 md:hidden ${
          mobileMenuOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <Sidebar onNavigate={() => setMobileMenuOpen(false)} />
      </div>

      {/* デスクトップサイドバー */}
      <div className="hidden md:block h-full flex-shrink-0">
        <Sidebar />
      </div>

      {/* メインコンテンツ */}
      <main
        id="main-content"
        tabIndex={-1}
        className="flex-1 min-h-0 overflow-auto outline-none"
      >
        <Outlet />
      </main>
      <ScrollToTop targetId="main-content" />

      <CommandPalette
        isOpen={commandPaletteOpen}
        onClose={() => setCommandPaletteOpen(false)}
      />
    </div>
  );
}
