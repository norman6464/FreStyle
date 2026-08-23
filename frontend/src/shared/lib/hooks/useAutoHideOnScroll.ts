import { useCallback, useEffect, useRef, useState } from 'react';

export interface UseAutoHideOnScrollOptions {
  /** この位置（px）までは常に表示する（先頭付近でヘッダーが消えないように）。既定 80。 */
  topThreshold?: number;
  /** 方向判定に使う最小移動量（px）。小刻みな揺れでちらつかせない。既定 8。 */
  delta?: number;
  /**
   * hidden を切り替えた直後に方向判定を止める時間（ms）。既定 350。
   * ヘッダーの表示/非表示で本文の表示域が伸縮すると、最下部ではブラウザが scrollTop を
   * 自動クランプして「逆方向スクロール」の偽イベントが発火し、表示⇔非表示が無限に
   * 振動する（ヘッダーがガタつく）。切替直後を無視して このフィードバックループを断つ。
   */
  cooldownMs?: number;
}

export interface UseAutoHideOnScrollResult {
  /** 対象要素を下へスクロール中なら true（＝ヘッダー等を隠してよい）。 */
  hidden: boolean;
  /** 監視するスクロールコンテナに張る callback ref（要素の差し替え・再マウントに追随する）。 */
  scrollRef: (node: HTMLElement | null) => void;
}

/**
 * useAutoHideOnScroll はスクロール方向で「隠すべきか」を判定する汎用 hook。
 *
 * - 下へスクロール: hidden = true（本文に集中させる）
 * - 上へスクロール: 即座に hidden = false
 * - 先頭付近（topThreshold 以内）: 常に false
 * - 監視対象が消えた（ノート切替のローディング等）: false に戻す
 */
export function useAutoHideOnScroll(
  options: UseAutoHideOnScrollOptions = {},
): UseAutoHideOnScrollResult {
  const { topThreshold = 80, delta = 8, cooldownMs = 350 } = options;
  const [hidden, setHidden] = useState(false);
  const [element, setElement] = useState<HTMLElement | null>(null);
  const lastTopRef = useRef(0);
  // hidden 切替直後の判定停止期限。クランプ由来の偽イベントを飲み込む（基準位置だけ追随させる）。
  const suppressUntilRef = useRef(0);

  const scrollRef = useCallback((node: HTMLElement | null) => {
    setElement(node);
  }, []);

  useEffect(() => {
    if (!element) {
      setHidden(false);
      return;
    }
    lastTopRef.current = element.scrollTop;

    const onScroll = () => {
      const top = element.scrollTop;
      if (top <= topThreshold) {
        setHidden(false);
        lastTopRef.current = top;
        return;
      }
      // 切替直後はレイアウト変化によるクランプイベントを無視し、基準位置だけ更新する。
      if (Date.now() < suppressUntilRef.current) {
        lastTopRef.current = top;
        return;
      }
      const diff = top - lastTopRef.current;
      if (Math.abs(diff) < delta) return;
      setHidden((prev) => {
        const next = diff > 0;
        if (next !== prev) suppressUntilRef.current = Date.now() + cooldownMs;
        return next;
      });
      lastTopRef.current = top;
    };

    element.addEventListener('scroll', onScroll, { passive: true });
    return () => element.removeEventListener('scroll', onScroll);
  }, [element, topThreshold, delta, cooldownMs]);

  return { hidden, scrollRef };
}
