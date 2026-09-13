import { useEffect } from 'react';

/**
 * useClickOutside は、ref の外側をクリック（mousedown）したときに onOutside を呼ぶ。
 * ポップオーバー・ドロップダウンを「他へ触ったら閉じる」ようにするための汎用 hook。
 *
 * click ではなく mousedown で見る — 同じ要素の click ハンドラより先に走るので、
 * 「開くボタンをもう一度押す」操作が「外側クリックで閉じてから即座に開き直す」
 * という余計な一往復にならない（押した瞬間に閉じるので、その後の click で
 * 自分の onClick がトグルを一度だけ実行する）。
 */
export function useClickOutside<T extends HTMLElement>(
  ref: React.RefObject<T | null>,
  enabled: boolean,
  onOutside: () => void,
): void {
  useEffect(() => {
    if (!enabled) return;
    const handleMouseDown = (event: MouseEvent) => {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        onOutside();
      }
    };
    document.addEventListener('mousedown', handleMouseDown);
    return () => document.removeEventListener('mousedown', handleMouseDown);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled]);
}
