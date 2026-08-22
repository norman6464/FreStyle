import { useState, useEffect, useCallback } from 'react';
import { useAppSelector } from '@/shared/lib/store';
import { useDocumentMeta } from '@/shared/lib/hooks/useDocumentMeta';
import { Outlet, useLocation } from 'react-router-dom';

import Header from './Header';
import SkipLink from './SkipLink';
import ScrollToTop from './ScrollToTop';
import CommandPalette from './CommandPalette';
import NavSidebar from './NavSidebar';
import { useSidebarMode } from '../model/useSidebarMode';

export default function AppShell() {
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);
  const { pathname } = useLocation();
  const role = useAppSelector((s) => s.auth.role);
  const sidebar = useSidebarMode();

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

  return (
    <div className="h-screen flex bg-surface overflow-hidden">
      <SkipLink targetId="main-content" />

      {/* 固定表示: サイドバーがレイアウトに居座り、本文が右に寄る（デスクトップのみ）。 */}
      {sidebar.mode === 'pinned' && (
        <div className="hidden md:block h-full flex-shrink-0 border-r border-surface-3">
          <NavSidebar mode="pinned" onPin={sidebar.pin} onCollapse={sidebar.collapse} />
        </div>
      )}

      {/* 一時表示: 左端ホバーで浮かぶオーバーレイ（本文は動かない・デスクトップのみ）。 */}
      {sidebar.mode === 'collapsed' && (
        <>
          {/* 左端のホバー検知ゾーン。 */}
          <div
            aria-hidden="true"
            className="hidden md:block fixed left-0 top-0 z-30 h-full w-2"
            onMouseEnter={sidebar.openPeek}
            onMouseLeave={sidebar.closePeek}
          />
          <div
            className={`hidden md:block fixed left-0 top-2 bottom-2 z-40 overflow-hidden rounded-r-xl border border-surface-3 shadow-xl transition-all duration-200 ease-out ${
              sidebar.isPeeking
                ? 'translate-x-0 opacity-100'
                : '-translate-x-full opacity-0 pointer-events-none'
            }`}
          >
            <NavSidebar
              mode="collapsed"
              onPin={sidebar.pin}
              onCollapse={sidebar.collapse}
              onPointerEnter={sidebar.openPeek}
              onPointerLeave={sidebar.closePeek}
            />
          </div>
        </>
      )}

      <div className="flex-1 min-w-0 h-full flex flex-col overflow-hidden">
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
            <Header
              sidebarMode={sidebar.mode}
              onSidebarPin={sidebar.pin}
              onSidebarHoverStart={sidebar.openPeek}
              onSidebarHoverEnd={sidebar.closePeek}
            />
            <main id="main-content" tabIndex={-1} className="outline-none">
              <Outlet />
            </main>
          </div>
        ) : (
          <>
            {/* 上部ヘッダー（テキスト横並びナビ + 右側に通知/管理/ユーザー）。モバイルメニューも Header が持つ。 */}
            <Header
              sidebarMode={sidebar.mode}
              onSidebarPin={sidebar.pin}
              onSidebarHoverStart={sidebar.openPeek}
              onSidebarHoverEnd={sidebar.closePeek}
            />

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
      </div>

      <CommandPalette
        isOpen={commandPaletteOpen}
        onClose={() => setCommandPaletteOpen(false)}
      />
    </div>
  );
}
