import { FireIcon, CheckCircleIcon, BookOpenIcon, CodeBracketIcon } from '@heroicons/react/24/outline';
import type { UserDashboard } from '../../types';
import LearningCalendar from './LearningCalendar';

interface Props {
  dashboard: UserDashboard;
}

/**
 * DashboardStats はパーソナライズダッシュボードの統計セクション。
 * - streak（連続学習日数）
 * - 過去 90 日の合計演習数・正答数・完了章数
 * - 学習カレンダーヒートマップ
 */
export default function DashboardStats({ dashboard }: Props) {
  const correctRate = dashboard.totalExercises > 0
    ? Math.round((dashboard.totalCorrect / dashboard.totalExercises) * 100)
    : null;

  return (
    <div className="space-y-4 p-4 rounded-xl border border-[var(--color-surface-3)] bg-[var(--color-surface-1)]">
      {/* KPI バー */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <StatChip
          icon={<FireIcon className="w-4 h-4" />}
          label="連続学習"
          value={`${dashboard.streak} 日`}
          color="text-orange-500"
        />
        <StatChip
          icon={<CodeBracketIcon className="w-4 h-4" />}
          label="演習 (90 日)"
          value={`${dashboard.totalExercises} 問`}
          color="text-taupe-500"
        />
        {correctRate !== null && (
          <StatChip
            icon={<CheckCircleIcon className="w-4 h-4" />}
            label="正答率"
            value={`${correctRate}%`}
            color="text-emerald-500"
          />
        )}
        <StatChip
          icon={<BookOpenIcon className="w-4 h-4" />}
          label="章完了 (90 日)"
          value={`${dashboard.totalLessons} 章`}
          color="text-brand-500"
        />
      </div>

      {/* 学習カレンダー */}
      <LearningCalendar activities={dashboard.recentActivity} />
    </div>
  );
}

function StatChip({
  icon,
  label,
  value,
  color,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  color: string;
}) {
  return (
    <div className="flex flex-col gap-1 p-3 rounded-lg bg-[var(--color-surface-2)]">
      <span className={`${color}`}>{icon}</span>
      <span className="text-xs text-[var(--color-text-muted)]">{label}</span>
      <span className="text-base font-bold text-[var(--color-text-primary)]">{value}</span>
    </div>
  );
}
