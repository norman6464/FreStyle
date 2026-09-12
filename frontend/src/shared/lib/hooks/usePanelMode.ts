import { useCallback, useEffect, useState } from 'react';
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

// isPeeking は永続化しない値だが、同じ storageKey を複数箇所（例: ヘッダーの
// 再表示ボタンとサイドバー本体）で使うときは、そこでも同じプレビュー状態を共有する
// 必要がある（ヘッダー側にポインタが乗った瞬間、本文側のオーバーレイも浮かせたい）。
// mode は useLocalStorage 側で同じ仕組みにより同期されるが、isPeeking はそこに乗らない
// ためここで同様の購読レジストリを持つ。閉じる猶予のタイマーも同じ理由で key 単位の
// 共有にする（片方のインスタンスで開いても、もう片方が予約した「閉じる」タイマーを
// 消せないと、ホバーし続けているのに閉じてしまう）。
const peekListeners = new Map<string, Set<(peeking: boolean) => void>>();
const closeTimers = new Map<string, ReturnType<typeof setTimeout>>();

function broadcastPeek(key: string, peeking: boolean) {
  peekListeners.get(key)?.forEach((fn) => fn(peeking));
}

function subscribePeek(key: string, fn: (peeking: boolean) => void) {
  let set = peekListeners.get(key);
  if (!set) {
    set = new Set();
    peekListeners.set(key, set);
  }
  set.add(fn);
  return () => {
    set!.delete(fn);
    if (set!.size === 0) peekListeners.delete(key);
  };
}

function clearSharedCloseTimer(key: string) {
  const timer = closeTimers.get(key);
  if (timer !== undefined) {
    clearTimeout(timer);
    closeTimers.delete(key);
  }
}

/**
 * usePanelMode は折りたたみ式パネルの「一時表示 / 固定表示」を管理する汎用 hook。
 *
 * モードは storageKey ごとに localStorage へ保存し、再訪時も同じ状態で開く。
 * 一時表示のホバー判定は閉じる側にだけ猶予を入れ、ホバーゾーン → パネル → ボタンへ
 * ポインタを移す途中でちらつかないようにする。同じ storageKey を使う複数箇所（ヘッダーの
 * 再表示ボタンとサイドバー本体）はこの猶予・プレビュー状態も共有する。
 */
export function usePanelMode(
  storageKey: string,
  options: UsePanelModeOptions = {},
): UsePanelModeResult {
  const { defaultMode = 'pinned', shortcut = true } = options;
  const [mode, setMode] = useLocalStorage<PanelMode>(storageKey, defaultMode);
  const [isPeeking, setIsPeeking] = useState(false);

  useEffect(() => subscribePeek(storageKey, setIsPeeking), [storageKey]);

  const openPeek = useCallback(() => {
    clearSharedCloseTimer(storageKey);
    broadcastPeek(storageKey, true);
  }, [storageKey]);

  const closePeek = useCallback(() => {
    clearSharedCloseTimer(storageKey);
    const timer = setTimeout(() => {
      closeTimers.delete(storageKey);
      broadcastPeek(storageKey, false);
    }, CLOSE_DELAY_MS);
    closeTimers.set(storageKey, timer);
  }, [storageKey]);

  const pin = useCallback(() => {
    clearSharedCloseTimer(storageKey);
    broadcastPeek(storageKey, false);
    setMode('pinned');
  }, [storageKey, setMode]);

  const collapse = useCallback(() => {
    clearSharedCloseTimer(storageKey);
    broadcastPeek(storageKey, false);
    setMode('collapsed');
  }, [storageKey, setMode]);

  const toggle = useCallback(() => {
    clearSharedCloseTimer(storageKey);
    broadcastPeek(storageKey, false);
    setMode((prev) => (prev === 'pinned' ? 'collapsed' : 'pinned'));
  }, [storageKey, setMode]);

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
