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
    <div className="h-screen flex flex-col bg-surface">
      <SkipLink targetId="main-content" />

      {/* 全幅ヘッダー */}
      <header className="h-14 w-full bg-surface-1 border-b border-surface-3 flex items-center px-4 flex-shrink-0">
        <button
          onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          className="md:hidden p-1.5 -ml-1.5 mr-3 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-surface-2 rounded-md transition-colors"
          aria-label="メニュー"
        >
          <Bars3Icon className="w-5 h-5" />
        </button>
        <img src="/logo.svg" alt="FreStyle" className="h-9 w-auto" />
      </header>

      {/* ヘッダー下: サイドバー + メインコンテンツ */}
      <div className="flex flex-1 min-h-0 overflow-hidden">
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
      </div>

      <CommandPalette
        isOpen={commandPaletteOpen}
        onClose={() => setCommandPaletteOpen(false)}
      />
    </div>
  );
}
