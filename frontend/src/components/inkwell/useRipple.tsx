import { useCallback, useState } from 'react';

interface RippleItem {
  key: number;
  x: number;
  y: number;
  size: number;
}

/**
 * 押した位置から波紋を広げる触感エフェクト。
 * pointer down の座標を中心に、要素を覆う円をスケール＋フェードで消す。
 */
export function useRipple(color = 'currentColor') {
  const [ripples, setRipples] = useState<RippleItem[]>([]);

  const addRipple = useCallback((e: React.PointerEvent<HTMLElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const px = e.clientX - rect.left;
    const py = e.clientY - rect.top;
    // クリック点から最遠角までの距離を半径にして要素を覆いきる。
    const radius = Math.hypot(Math.max(px, rect.width - px), Math.max(py, rect.height - py));
    const size = radius * 2;
    const x = px - radius;
    const y = py - radius;
    setRipples((prev) => [...prev, { key: prev.length ? prev[prev.length - 1].key + 1 : 0, x, y, size }]);
  }, []);

  const clear = useCallback((key: number) => {
    setRipples((prev) => prev.filter((r) => r.key !== key));
  }, []);

  const rippleOverlay = (
    <span aria-hidden="true" className="pointer-events-none absolute inset-0 overflow-hidden rounded-[inherit]">
      {ripples.map((r) => (
        <span
          key={r.key}
          onAnimationEnd={() => clear(r.key)}
          className="absolute rounded-full animate-inkwell-ripple"
          style={{ left: r.x, top: r.y, width: r.size, height: r.size, backgroundColor: color }}
        />
      ))}
    </span>
  );

  return { addRipple, rippleOverlay };
}
