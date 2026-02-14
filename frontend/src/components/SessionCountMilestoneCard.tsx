import { StarIcon } from '@heroicons/react/24/solid';

interface SessionCountMilestoneCardProps {
  sessionCount: number;
}

const MILESTONES = [10, 25, 50, 100];

function getNextMilestone(count: number): number | null {
  return MILESTONES.find((m) => m > count) ?? null;
}

export default function SessionCountMilestoneCard({ sessionCount }: SessionCountMilestoneCardProps) {
  const nextMilestone = getNextMilestone(sessionCount);
  const allAchieved = nextMilestone === null;

  const prevMilestone = allAchieved
    ? MILESTONES[MILESTONES.length - 2]
    : MILESTONES[MILESTONES.indexOf(nextMilestone) - 1] ?? 0;

  const progress = allAchieved
    ? 100
    : Math.round(((sessionCount - prevMilestone) / (nextMilestone - prevMilestone)) * 100);

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <div className="flex items-center gap-2 mb-3">
        <StarIcon className="w-5 h-5 text-amber-400" />
        <h3 className="text-xs font-semibold text-[var(--color-text-primary)]">セッション達成</h3>
      </div>

      <div className="flex items-baseline gap-2 mb-2">
        <span className="text-2xl font-bold text-[var(--color-text-primary)]">{sessionCount}</span>
        <span className="text-xs text-[var(--color-text-muted)]">
          {allAchieved ? '全マイルストーン達成!' : `/ ${nextMilestone} 回`}
        </span>
      </div>

      <div
        role="progressbar"
        aria-valuenow={progress}
        className="w-full h-2 bg-surface-3 rounded-full overflow-hidden"
      >
        <div
          className="h-full bg-amber-400 rounded-full transition-all"
          style={{ width: `${Math.min(progress, 100)}%` }}
        />
      </div>

      {allAchieved && (
        <p className="text-xs text-amber-400 mt-2 font-medium">
          素晴らしい! 100回以上の練習を達成しました!
        </p>
      )}
    </div>
  );
}
