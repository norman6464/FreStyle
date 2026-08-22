import { useCallback, useEffect, useRef, useState } from 'react';
import { useLocalStorage } from '@/shared/lib/hooks/useLocalStorage';

/**
 * SidebarMode はサイドバーの表示モード。
 * - collapsed: 普段は隠れており、左端にポインタを寄せた間だけ一時表示（本文は動かない）
 * - pinned: 常に開いたままレイアウトに居座る（本文が右に寄る）
 */
export type SidebarMode = 'collapsed' | 'pinned';

const STORAGE_KEY = 'frestyle.sidebar.mode';

// ポインタが離れてから閉じるまでの猶予。ボタンへ移動する途中で消えないようにする。
const CLOSE_DELAY_MS = 220;

export interface UseSidebarModeResult {
  mode: SidebarMode;
  /** 一時表示（collapsed のとき）で今サイドバーが見えているか。 */
  isPeeking: boolean;
  /** 実際にサイドバーの中身を描画すべきか（固定表示 or 一時表示中）。 */
  isVisible: boolean;
  pin: () => void;
  collapse: () => void;
  toggle: () => void;
  /** 左端・サイドバー・トグルボタンへポインタが入ったとき。 */
  openPeek: () => void;
  /** 上記から離れたとき（猶予後に閉じる）。 */
  closePeek: () => void;
}

/**
 * useSidebarMode はサイドバーの「一時表示 / 固定表示」を管理する。
 *
 * モードは localStorage に保存し、再訪時も同じ状態で開く。⌘\ / Ctrl+\ で切り替えられる。
 * 一時表示のホバー判定は閉じる側にだけ猶予を入れ、左端 → サイドバー → ボタンへ
 * ポインタを移す途中でちらつかないようにする。
 */
export function useSidebarMode(): UseSidebarModeResult {
  const [mode, setMode] = useLocalStorage<SidebarMode>(STORAGE_KEY, 'collapsed');
  const [isPeeking, setIsPeeking] = useState(false);
  const closeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearCloseTimer = useCallback(() => {
    if (closeTimerRef.current !== null) {
      clearTimeout(closeTimerRef.current);
      closeTimerRef.current = null;
    }
  }, []);

  useEffect(() => clearCloseTimer, [clearCloseTimer]);

  const openPeek = useCallback(() => {
    clearCloseTimer();
    setIsPeeking(true);
  }, [clearCloseTimer]);

  const closePeek = useCallback(() => {
    clearCloseTimer();
    closeTimerRef.current = setTimeout(() => setIsPeeking(false), CLOSE_DELAY_MS);
  }, [clearCloseTimer]);

  const pin = useCallback(() => {
    clearCloseTimer();
    setIsPeeking(false);
    setMode('pinned');
  }, [clearCloseTimer, setMode]);

  const collapse = useCallback(() => {
    clearCloseTimer();
    setIsPeeking(false);
    setMode('collapsed');
  }, [clearCloseTimer, setMode]);

  const toggle = useCallback(() => {
    clearCloseTimer();
    setIsPeeking(false);
    setMode((prev) => (prev === 'pinned' ? 'collapsed' : 'pinned'));
  }, [clearCloseTimer, setMode]);

  // ⌘\ / Ctrl+\ で固定表示を切り替える（サイドバーのトグルとして一般的な割り当て）。
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === '\\') {
        event.preventDefault();
        toggle();
      }
    };
    document.addEventListener('keydown', onKeyDown);
    return () => document.removeEventListener('keydown', onKeyDown);
  }, [toggle]);

  return {
    mode,
    isPeeking,
    isVisible: mode === 'pinned' || isPeeking,
    pin,
    collapse,
    toggle,
    openPeek,
    closePeek,
  };
}
