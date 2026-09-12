import { useState, useEffect, useCallback } from 'react';
import { useDocumentMeta } from '@/shared/lib/hooks/useDocumentMeta';
import { Outlet } from 'react-router-dom';

import Header from './Header';
import SkipLink from './SkipLink';
import ScrollToTop from './ScrollToTop';
import CommandPalette from './CommandPalette';

export default function AppShell() {
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);

  // 認証必須ページ（AppShell 配下）はログイン前提なので検索インデックス対象外にする。
  useDocumentMeta({ robots: 'noindex, nofollow' });

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

      {/* ヘッダーは常時表示。本文とは縦に並べる（重ねない）ので、
          本文側に先頭の余白を入れる必要はない。 */}
      <Header onOpenSearch={() => setCommandPaletteOpen(true)} />

      {/*
        tabIndex は 0。ここは縦に流れるスクロール領域なので、キーボードだけの人が
        矢印キーで動かせるよう Tab で到達できる必要がある（-1 だと「本文へスキップ」から
        飛んだときしか触れず、そのまま Tab を続けると本文を飛び越してしまう）。
      */}
      <main
        id="main-content"
        tabIndex={0}
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
