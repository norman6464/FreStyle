import { useCallback, useState } from 'react';

export type ResizablePanelSide = 'left' | 'right';

export interface UseResizablePanelOptions {
  /** 既定幅（px）。 */
  defaultWidth?: number;
  /** 最小幅（px）。既定 288（w-72 と同じ）。 */
  minWidth?: number;
  /** 最大幅を画面幅の何割までにするか。既定 0.5（画面の半分）。 */
  maxWidthRatio?: number;
  /**
   * パネルが本文の左右どちらにあるか。ドラッグの符号を決める
   * （左のパネルは右へドラッグで広がり、右のパネルは左へドラッグで広がる）。既定 'left'。
   */
  side?: ResizablePanelSide;
  /** 幅を覚える localStorage キー。省略時は永続化せず、再訪では defaultWidth に戻る。 */
  storageKey?: string;
}

export interface UseResizablePanelResult {
  /** 現在の幅（px）。 */
  width: number;
  /** ドラッグ中かどうか（ハンドルの見た目に使う）。 */
  isResizing: boolean;
  /** ハンドルの onMouseDown に渡す。 */
  onHandleMouseDown: (event: React.MouseEvent) => void;
  /** ハンドルの onKeyDown に渡す（← / → で操作。マウス操作の代替）。 */
  onHandleKeyDown: (event: React.KeyboardEvent) => void;
}

const DEFAULT_WIDTH = 288;
const DEFAULT_MIN_WIDTH = 288;
const DEFAULT_MAX_WIDTH_RATIO = 0.5;
// 矢印キー 1 回あたりの増減量。左右どちらのパネルでも「→ で広がる／← で狭まる」に統一する
// （マウスドラッグの向きは side で反転するが、キー操作は side を意識させない一貫した挙動にする）。
const KEY_STEP = 16;

function readInitialWidth(storageKey: string | undefined, defaultWidth: number): number {
  if (!storageKey) return defaultWidth;
  try {
    const raw = localStorage.getItem(storageKey);
    if (raw === null) return defaultWidth;
    const parsed = JSON.parse(raw);
    return typeof parsed === 'number' && Number.isFinite(parsed) ? parsed : defaultWidth;
  } catch {
    return defaultWidth;
  }
}

/**
 * useResizablePanel はサイドパネルの縦線（境界線）をドラッグして横幅を変える機構。
 *
 * 幅は storageKey ごとに localStorage へ保存し、再訪でも同じ幅で開く（省略時は保存しない）。
 * 最大幅は画面幅に対する割合で決める（ウィンドウを狭めた後の初回ドラッグにも追随する）。
 */
export function useResizablePanel(options: UseResizablePanelOptions = {}): UseResizablePanelResult {
  const {
    defaultWidth = DEFAULT_WIDTH,
    minWidth = DEFAULT_MIN_WIDTH,
    maxWidthRatio = DEFAULT_MAX_WIDTH_RATIO,
    side = 'left',
    storageKey,
  } = options;

  const [width, setWidth] = useState<number>(() => readInitialWidth(storageKey, defaultWidth));
  const [isResizing, setIsResizing] = useState(false);

  const persist = useCallback((value: number) => {
    if (!storageKey) return;
    try {
      localStorage.setItem(storageKey, JSON.stringify(value));
    } catch {
      // QuotaExceededError等 - 保存できなくても操作自体は継続してよい
    }
  }, [storageKey]);

  const onHandleMouseDown = useCallback((event: React.MouseEvent) => {
    // テキスト選択やドラッグ&ドロップの誤爆を防ぐ（境界線をつまむ操作であって選択操作ではない）。
    event.preventDefault();
    const startX = event.clientX;
    const startWidth = width;
    let latestWidth = startWidth;
    setIsResizing(true);

    const onMouseMove = (moveEvent: MouseEvent) => {
      const delta = moveEvent.clientX - startX;
      // 左のパネル: 右へドラッグ(delta>0)で広がる。右のパネル: 左へドラッグ(delta<0)で広がる。
      const signedDelta = side === 'right' ? -delta : delta;
      const maxWidth = window.innerWidth * maxWidthRatio;
      latestWidth = Math.min(maxWidth, Math.max(minWidth, startWidth + signedDelta));
      setWidth(latestWidth);
    };

    const onMouseUp = () => {
      setIsResizing(false);
      document.removeEventListener('mousemove', onMouseMove);
      document.removeEventListener('mouseup', onMouseUp);
      persist(latestWidth);
    };

    document.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseup', onMouseUp);
  }, [width, side, minWidth, maxWidthRatio, persist]);

  const onHandleKeyDown = useCallback((event: React.KeyboardEvent) => {
    let delta = 0;
    if (event.key === 'ArrowLeft') delta = -KEY_STEP;
    else if (event.key === 'ArrowRight') delta = KEY_STEP;
    else return;
    event.preventDefault();
    setWidth((prev) => {
      const maxWidth = window.innerWidth * maxWidthRatio;
      const next = Math.min(maxWidth, Math.max(minWidth, prev + delta));
      persist(next);
      return next;
    });
  }, [minWidth, maxWidthRatio, persist]);

  return { width, isResizing, onHandleMouseDown, onHandleKeyDown };
}
