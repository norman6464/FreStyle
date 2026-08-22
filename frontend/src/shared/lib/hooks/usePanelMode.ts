import { useCallback, useEffect, useRef, useState } from 'react';
import { useLocalStorage } from './useLocalStorage';

/**
 * PanelMode は折りたたみ式パネルの表示モード。
 * - collapsed: 普段は隠れており、左端等にポインタを寄せた間だけ一時表示（本文は動かない）
 * - pinned: 常に開いたままレイアウトに居座る
 */
export type PanelMode = 'collapsed' | 'pinned';

// ポインタが離れてから閉じるまでの猶予。トグルボタンへ移動する途中で消えないようにする。
const CLOSE_DELAY_MS = 220;

export interface UsePanelModeOptions {
  /** 初期モード（保存値が無いときに使う）。既定 'pinned'。 */
  defaultMode?: PanelMode;
  /** ⌘\ / Ctrl+\ でモードをトグルするか。既定 true。 */
  shortcut?: boolean;
}

export interface UsePanelModeResult {
  mode: PanelMode;
  /** 一時表示（collapsed のとき）で今パネルが見えているか。 */
  isPeeking: boolean;
  /** 実際にパネルの中身を見せるべきか（固定表示 or 一時表示中）。 */
  isVisible: boolean;
  pin: () => void;
  collapse: () => void;
  toggle: () => void;
  /** ホバーゾーン・パネル・トグルボタンへポインタが入ったとき。 */
  openPeek: () => void;
  /** 上記から離れたとき（猶予後に閉じる）。 */
  closePeek: () => void;
}

/**
 * usePanelMode は折りたたみ式パネルの「一時表示 / 固定表示」を管理する汎用 hook。
 *
 * モードは storageKey ごとに localStorage へ保存し、再訪時も同じ状態で開く。
 * 一時表示のホバー判定は閉じる側にだけ猶予を入れ、ホバーゾーン → パネル → ボタンへ
 * ポインタを移す途中でちらつかないようにする。
 */
export function usePanelMode(
  storageKey: string,
  options: UsePanelModeOptions = {},
): UsePanelModeResult {
  const { defaultMode = 'pinned', shortcut = true } = options;
  const [mode, setMode] = useLocalStorage<PanelMode>(storageKey, defaultMode);
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

  // ⌘\ / Ctrl+\ で固定表示を切り替える（パネルトグルとして一般的な割り当て）。
  useEffect(() => {
    if (!shortcut) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === '\\') {
        event.preventDefault();
        toggle();
      }
    };
    document.addEventListener('keydown', onKeyDown);
    return () => document.removeEventListener('keydown', onKeyDown);
  }, [shortcut, toggle]);

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
