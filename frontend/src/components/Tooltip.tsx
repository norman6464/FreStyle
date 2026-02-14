import { ReactNode, useCallback, useRef, useState } from 'react';

interface TooltipProps {
  content: string;
  position?: 'top' | 'bottom';
  children: ReactNode;
}

export default function Tooltip({ content, position = 'top', children }: TooltipProps) {
  const [visible, setVisible] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleMouseEnter = useCallback(() => {
    timerRef.current = setTimeout(() => {
      setVisible(true);
    }, 300);
  }, []);

  const handleMouseLeave = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
    setVisible(false);
  }, []);

  const positionClass = position === 'top'
    ? 'bottom-full left-1/2 -translate-x-1/2 mb-1.5'
    : 'top-full left-1/2 -translate-x-1/2 mt-1.5';

  return (
    <div
      className="relative inline-flex"
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      {children}
      {visible && (
        <div
          role="tooltip"
          className={`absolute ${positionClass} px-2 py-1 text-xs text-white bg-gray-900 rounded whitespace-nowrap pointer-events-none z-50 animate-fade-in`}
        >
          {content}
        </div>
      )}
    </div>
  );
}
