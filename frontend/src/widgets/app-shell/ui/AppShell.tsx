import { useState, useEffect, useCallback } from 'react';
import { useDocumentMeta } from '@/shared/lib/hooks/useDocumentMeta';
import { Outlet, useLocation } from 'react-router-dom';

import Header from './Header';
import { HeaderVisibilityContext, useHeaderVisibilityState } from '../model/headerVisibility';
import SkipLink from './SkipLink';
import ScrollToTop from './ScrollToTop';
import CommandPalette from './CommandPalette';

export default function AppShell() {
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);
  const { pathname } = useLocation();
  // ヘッダーの自動隠し（ノート本文の下スクロール等）。ページ側が setHeaderHidden で切り替える。
  const headerVisibility = useHeaderVisibilityState();

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

  // ページ遷移時はヘッダーを必ず表示へ戻す（前ページの隠し状態を持ち越さない）。
  const { setHeaderHidden } = headerVisibility;
  useEffect(() => {
    setHeaderHidden(false);
  }, [pathname, setHeaderHidden]);

  return (
    <HeaderVisibilityContext.Provider value={headerVisibility}>
    <div className="h-screen flex flex-col bg-surface overflow-hidden">
      <SkipLink targetId="main-content" />

      {/* ヘッダーを本文の**上に重ねる**。並べる（縦に積む）と背後に何も無く、
          半透明 + ぼかしの地が効かない。重ねたぶん本文の先頭に余白を入れて、
          最初の行がヘッダーの裏に隠れないようにする。 */}
      <div className="relative flex-1 min-h-0">
        {/* headerHidden のときは上へスライドして隠れる（本文が全高になる）。 */}
        <div
          className={`absolute inset-x-0 top-0 z-40 transition-transform duration-200 ease-out ${
            headerVisibility.headerHidden ? '-translate-y-full' : 'translate-y-0'
          }`}
        >
          <Header />
        </div>

        {/* メインコンテンツ。h-16 はヘッダーの高さ。隠れているときは余白も畳む。 */}
        <main
          id="main-content"
          tabIndex={-1}
          className={`h-full overflow-auto outline-none transition-[padding-top] duration-200 ease-out ${
            headerVisibility.headerHidden ? 'pt-0' : 'pt-16'
          }`}
        >
          <Outlet />
        </main>
      </div>
      <ScrollToTop targetId="main-content" />

      <CommandPalette
        isOpen={commandPaletteOpen}
        onClose={() => setCommandPaletteOpen(false)}
      />
    </div>
    </HeaderVisibilityContext.Provider>
  );
}
