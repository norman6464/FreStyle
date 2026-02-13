import { useEffect } from 'react';
import { usePracticeTimer } from '../hooks/usePracticeTimer';

export default function PracticeTimer() {
  const { formattedTime, start, stop } = usePracticeTimer();

  useEffect(() => {
    start();
    return () => stop();
  }, [start, stop]);

  return (
    <div className="flex items-center gap-1.5 text-xs text-slate-500">
      <span>経過時間</span>
      <span className="font-mono font-medium text-slate-700">{formattedTime}</span>
    </div>
  );
}
