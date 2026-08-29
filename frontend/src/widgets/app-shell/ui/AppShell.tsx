import { useState, useEffect, useCallback } from 'react';
import { useAppSelector } from '@/shared/lib/store';
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
  const role = useAppSelector((s) => s.auth.role);
  // ヘッダーの自動隠し（ノート本文の下スクロール等）。ページ側が setHeaderHidden で切り替える。
  const headerVisibility = useHeaderVisibilityState();

  // 認証必須ページ（AppShell 配下）はログイン前提なので検索インデックス対象外にする。
  useDocumentMeta({ robots: 'noindex, nofollow' });

  // 受講者の教材閲覧(/courses/:id)はヘッダーごとスクロールで画面外に流す(FRESTYLE-122)。
  // チャット / ノート / コース編集などのパネル型ページは main の固定高さに依存しているため、
  // このルート + 受講者のときだけスクロールコンテナをヘッダーの外側に広げる。
  const canManage = role === 'company_admin' || role === 'super_admin';
  const documentScroll = /^\/courses\/\d+\/?$/.test(pathname) && !canManage;

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

      {documentScroll ? (
        // ドキュメントスクロール: ヘッダーと main を 1 つのスクロールコンテナに入れ、
        // スクロールするとヘッダーが本文と一緒に流れる。章切替時の先頭スクロールも
        // このコンテナ([data-app-scroll])に対して行う。
        <div
          id="app-scroll"
          data-app-scroll
          className="flex-1 min-h-0 overflow-y-auto bg-[var(--color-reading-surface)]"
        >
          {/* 上部ヘッダー。**本文の上に留める**（sticky）。
              ヘッダーの地は半透明 + ぼかしなので、下を本文が通って初めてぼけて見える。
              並べるだけ（通常フロー）だと背後に何も無く、ぼかしは効かない。 */}
          <div className="sticky top-0 z-40">
            <Header />
          </div>
          <main id="main-content" tabIndex={-1} className="outline-none">
            <Outlet />
          </main>
        </div>
      ) : (
        // ヘッダーを本文の**上に重ねる**。並べる（縦に積む）と背後に何も無く、
        // 半透明 + ぼかしの地が効かない。重ねたぶん本文の先頭に余白を入れて、
        // 最初の行がヘッダーの裏に隠れないようにする。
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
      )}
      <ScrollToTop targetId={documentScroll ? 'app-scroll' : 'main-content'} />

      <CommandPalette
        isOpen={commandPaletteOpen}
        onClose={() => setCommandPaletteOpen(false)}
      />
    </div>
    </HeaderVisibilityContext.Provider>
  );
}
