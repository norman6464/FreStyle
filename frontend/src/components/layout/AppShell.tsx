import { useState, useEffect, useCallback } from 'react';
import { Outlet } from 'react-router-dom';
import Header from './Header';
import SkipLink from '../SkipLink';
import ScrollToTop from '../ScrollToTop';
import CommandPalette from '../CommandPalette';

export default function AppShell() {
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      setCommandPaletteOpen((prev) => !prev);
    }
  }, []);

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  return (
    <div className="h-screen flex flex-col bg-surface overflow-hidden">
      <SkipLink targetId="main-content" />

      {/* 上部ヘッダー（テキスト横並びナビ + 右側に通知/管理/ユーザー）。モバイルメニューも Header が持つ。 */}
      <Header />

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
