import { useEffect } from 'react';
import { usePracticeTimer } from '../hooks/usePracticeTimer';

export default function PracticeTimer() {
  const { formattedTime, start, stop } = usePracticeTimer();

  useEffect(() => {
    start();
    return () => stop();
  }, [start, stop]);

  return (
    <div className="flex items-center gap-1.5 text-xs text-[var(--color-text-muted)]">
      <span>経過時間</span>
      <span className="font-mono font-medium text-[var(--color-text-secondary)]">{formattedTime}</span>
    </div>
  );
}
