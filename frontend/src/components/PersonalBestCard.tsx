import { TrophyIcon } from '@heroicons/react/24/solid';
import Card from './Card';
import { formatDate } from '../utils/formatters';

interface PersonalBestCardProps {
  history: { sessionId: number; overallScore: number; createdAt: string }[];
}

export default function PersonalBestCard({ history }: PersonalBestCardProps) {
  if (history.length === 0) return null;

  const best = history.reduce((prev, curr) =>
    curr.overallScore > prev.overallScore ? curr : prev
  );

  return (
    <Card>
      <div className="flex items-center gap-2 mb-3">
        <TrophyIcon className="w-5 h-5 text-amber-400" />
        <h3 className="text-xs font-semibold text-[var(--color-text-primary)]">自己ベスト</h3>
      </div>
      <div className="text-center">
        <p className="text-3xl font-bold text-amber-400">{best.overallScore}</p>
        <div className="mt-1">
          <p className="text-[10px] text-[var(--color-text-muted)]">達成日</p>
          <p className="text-xs text-[var(--color-text-secondary)]">
            {formatDate(best.createdAt)}
          </p>
        </div>
      </div>
    </Card>
  );
}
