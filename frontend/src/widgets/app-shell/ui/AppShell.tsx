import { useState, useEffect, useCallback } from 'react';
import { useAppSelector } from '@/shared/lib/store';
import { Outlet, useLocation } from 'react-router-dom';

import Header from './Header';
import SkipLink from './SkipLink';
import ScrollToTop from './ScrollToTop';
import CommandPalette from './CommandPalette';

export default function AppShell() {
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);
  const { pathname } = useLocation();
  const role = useAppSelector((s) => s.auth.role);

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

  return (
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
          {/* 上部ヘッダー（テキスト横並びナビ + 右側に通知/管理/ユーザー）。モバイルメニューも Header が持つ。 */}
          <Header />
          <main id="main-content" tabIndex={-1} className="outline-none">
            <Outlet />
          </main>
        </div>
      ) : (
        <>
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
        </>
      )}
      <ScrollToTop targetId={documentScroll ? 'app-scroll' : 'main-content'} />

      <CommandPalette
        isOpen={commandPaletteOpen}
        onClose={() => setCommandPaletteOpen(false)}
      />
    </div>
  );
}
