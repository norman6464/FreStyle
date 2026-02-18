import type { ProfileStats } from '../../repositories/ProfileStatsRepository';
import {
  ChatBubbleLeftRightIcon,
  ChartBarIcon,
  FireIcon,
  TrophyIcon,
  UsersIcon,
} from '@heroicons/react/24/outline';

interface StatCardProps {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  value: string;
  color: string;
}

function StatCard({ icon: Icon, label, value, color }: StatCardProps) {
  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4 flex items-center gap-3">
      <div className={`rounded-full p-2 ${color}`}>
        <Icon className="w-5 h-5" />
      </div>
      <div>
        <p className="text-xs text-[var(--color-text-muted)]">{label}</p>
        <p className="text-lg font-bold text-[var(--color-text-primary)]">{value}</p>
      </div>
    </div>
  );
}

interface ProfileStatsSectionProps {
  stats: ProfileStats | null;
  loading: boolean;
}

export default function ProfileStatsSection({ stats, loading }: ProfileStatsSectionProps) {
  if (loading) {
    return (
      <div className="space-y-3">
        <h3 className="text-sm font-semibold text-[var(--color-text-secondary)]">学習統計</h3>
        <div className="grid grid-cols-2 gap-3">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="bg-surface-1 rounded-lg border border-surface-3 p-4 h-20 animate-pulse" />
          ))}
        </div>
      </div>
    );
  }

  if (!stats) return null;

  return (
    <div className="space-y-3">
      <h3 className="text-sm font-semibold text-[var(--color-text-secondary)]">学習統計</h3>
      <div className="grid grid-cols-2 gap-3">
        <StatCard
          icon={ChatBubbleLeftRightIcon}
          label="AIセッション"
          value={`${stats.totalSessions}回`}
          color="bg-blue-500/20 text-blue-400"
        />
        <StatCard
          icon={ChartBarIcon}
          label="平均スコア"
          value={stats.averageScore > 0 ? `${stats.averageScore.toFixed(1)}点` : '未計測'}
          color="bg-amber-500/20 text-amber-400"
        />
        <StatCard
          icon={FireIcon}
          label="連続学習"
          value={`${stats.currentStreak}日`}
          color="bg-red-500/20 text-red-400"
        />
        <StatCard
          icon={TrophyIcon}
          label="最長連続"
          value={`${stats.longestStreak}日`}
          color="bg-purple-500/20 text-purple-400"
        />
        <StatCard
          icon={UsersIcon}
          label="フォロワー"
          value={`${stats.followerCount}人`}
          color="bg-emerald-500/20 text-emerald-400"
        />
        <StatCard
          icon={UsersIcon}
          label="フォロー中"
          value={`${stats.followingCount}人`}
          color="bg-cyan-500/20 text-cyan-400"
        />
      </div>
    </div>
  );
}
