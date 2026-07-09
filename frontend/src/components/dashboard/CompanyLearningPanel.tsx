import { Link } from 'react-router-dom';
import { UsersIcon, FireIcon, CalendarDaysIcon, ArrowRightIcon } from '@heroicons/react/24/outline';
import type { CompanyLearningSummary } from '../../repositories/AdminMemberRepository';

interface Props {
  summary: CompanyLearningSummary;
}

/**
 * CompanyLearningPanel は company_admin のホームに出す「メンバーの学習状況」サイドバー(FRESTYLE-103)。
 * 自分の学習統計の代わりに、自社 trainee の学習アクティビティ集計を表示する。
 */
export default function CompanyLearningPanel({ summary }: Props) {
  return (
    <div className="space-y-4 p-4 rounded-xl border border-[var(--color-surface-3)] bg-[var(--color-surface-1)]">
      <h2 className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-widest">
        メンバーの学習状況
      </h2>

      {/* KPI */}
      <div className="grid grid-cols-1 gap-3">
        <StatChip
          icon={<UsersIcon className="w-4 h-4" />}
          label="在籍メンバー"
          value={`${summary.traineeCount} 名`}
          color="text-taupe-500"
        />
        <StatChip
          icon={<FireIcon className="w-4 h-4" />}
          label="今日学習した人"
          value={`${summary.activeToday} 名`}
          color="text-orange-500"
        />
        <StatChip
          icon={<CalendarDaysIcon className="w-4 h-4" />}
          label="直近 7 日で学習した人"
          value={`${summary.activeThisWeek} 名`}
          color="text-emerald-500"
        />
      </div>

      {/* 直近アクティブメンバー */}
      <div>
        <h3 className="text-xs text-[var(--color-text-muted)] mb-2">直近アクティブ</h3>
        {summary.recentMembers.length === 0 ? (
          <p className="text-xs text-[var(--color-text-muted)]">
            まだ学習活動がありません。
          </p>
        ) : (
          <ul className="space-y-2">
            {summary.recentMembers.map((m) => (
              <li key={m.userId} className="flex items-center justify-between gap-2 text-sm">
                <span className="truncate text-[var(--color-text-primary)]">{m.name}</span>
                <span className="shrink-0 text-xs text-[var(--color-text-muted)]">
                  {formatShortDate(m.lastActiveDate)} ・ {m.recentActivityCount} 回
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>

      <Link
        to="/admin/members"
        className="inline-flex items-center gap-1.5 text-sm text-brand-600 hover:text-brand-700 transition-colors"
      >
        従業員一覧へ
        <ArrowRightIcon className="w-4 h-4" />
      </Link>
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
    <div className="flex items-center gap-3 px-3 py-2.5 rounded-lg bg-[var(--color-surface-2)]">
      <span className={color}>{icon}</span>
      <span className="flex-1 text-xs text-[var(--color-text-muted)]">{label}</span>
      <span className="text-base font-bold text-[var(--color-text-primary)]">{value}</span>
    </div>
  );
}

/** YYYY-MM-DD を「M/D」に短縮する（不正な値はそのまま返す）。 */
function formatShortDate(iso: string): string {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(iso);
  if (!m) return iso;
  return `${Number(m[2])}/${Number(m[3])}`;
}
