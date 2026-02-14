import { CheckCircleIcon } from '@heroicons/react/24/solid';
import Card from './Card';

interface SessionCompleteCardProps {
  duration: number;
  messageCount: number;
  onViewScores: () => void;
  onPracticeAgain: () => void;
}

function formatDuration(seconds: number): string {
  const minutes = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${minutes}:${String(secs).padStart(2, '0')}`;
}

export default function SessionCompleteCard({
  duration,
  messageCount,
  onViewScores,
  onPracticeAgain,
}: SessionCompleteCardProps) {
  return (
    <Card className="p-6 text-center">
      <CheckCircleIcon className="w-10 h-10 text-emerald-400 mx-auto mb-3" />
      <h3 className="text-sm font-bold text-[var(--color-text-primary)] mb-4">セッション完了</h3>

      <div className="grid grid-cols-2 gap-4 mb-5">
        <div>
          <p className="text-[10px] text-[var(--color-text-muted)]">経過時間</p>
          <p className="text-lg font-semibold text-[var(--color-text-primary)]">{formatDuration(duration)}</p>
        </div>
        <div>
          <p className="text-[10px] text-[var(--color-text-muted)]">メッセージ数</p>
          <p className="text-lg font-semibold text-[var(--color-text-primary)]">{messageCount}</p>
        </div>
      </div>

      <div className="flex gap-2">
        <button
          onClick={onViewScores}
          className="flex-1 bg-primary-500 hover:bg-primary-600 text-white font-medium py-2.5 px-4 rounded-lg text-sm transition-colors"
        >
          スコアを確認
        </button>
        <button
          onClick={onPracticeAgain}
          className="flex-1 bg-surface-3 hover:bg-surface-3 text-[var(--color-text-secondary)] font-medium py-2.5 px-4 rounded-lg text-sm transition-colors"
        >
          もう一度練習
        </button>
      </div>
    </Card>
  );
}
