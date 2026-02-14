import type { ComponentType, SVGProps } from 'react';
import {
  CheckCircleIcon,
  PencilSquareIcon,
  TrophyIcon,
  StarIcon,
  FireIcon,
  ShieldCheckIcon,
  SparklesIcon,
} from '@heroicons/react/24/outline';

interface AchievementBadgeCardProps {
  totalSessions: number;
}

interface Badge {
  name: string;
  threshold: number;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
}

const BADGES: Badge[] = [
  { name: 'ファーストステップ', threshold: 1, icon: CheckCircleIcon },
  { name: 'コツコツ練習', threshold: 3, icon: PencilSquareIcon },
  { name: 'ウィークリーチャレンジ', threshold: 5, icon: TrophyIcon },
  { name: 'テンセッション', threshold: 10, icon: StarIcon },
  { name: 'ハーフウェイ', threshold: 25, icon: FireIcon },
  { name: 'フィフティ達成', threshold: 50, icon: ShieldCheckIcon },
  { name: 'センチュリー', threshold: 100, icon: SparklesIcon },
];

export default function AchievementBadgeCard({ totalSessions }: AchievementBadgeCardProps) {
  const unlockedCount = BADGES.filter((b) => totalSessions >= b.threshold).length;
  const nextBadge = BADGES.find((b) => totalSessions < b.threshold);

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-medium text-[var(--color-text-secondary)]">達成バッジ</p>
        <span className="text-[10px] text-[var(--color-text-faint)]">
          {unlockedCount}/{BADGES.length}
        </span>
      </div>

      <div className="flex flex-wrap gap-2 mb-3">
        {BADGES.map((badge) => {
          const unlocked = totalSessions >= badge.threshold;
          const Icon = badge.icon;
          return (
            <div
              key={badge.name}
              data-testid={unlocked ? 'badge-unlocked' : 'badge-locked'}
              className={`flex flex-col items-center gap-0.5 w-16 ${
                unlocked ? 'opacity-100' : 'opacity-30'
              }`}
              title={`${badge.name}（${badge.threshold}回）`}
            >
              <Icon className="w-5 h-5 text-primary-400" />
              <span className="text-[9px] text-[var(--color-text-muted)] text-center leading-tight truncate w-full">
                {badge.name}
              </span>
            </div>
          );
        })}
      </div>

      {nextBadge ? (
        <p data-testid="next-badge-info" className="text-xs text-[var(--color-text-muted)]">
          次のバッジ「{nextBadge.name}」まであと{nextBadge.threshold - totalSessions}回
        </p>
      ) : (
        <p data-testid="next-badge-info" className="text-xs text-emerald-400 font-medium">
          全バッジを獲得しました！
        </p>
      )}
    </div>
  );
}
