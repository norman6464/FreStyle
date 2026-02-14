import { useMemo } from 'react';

interface ScoreHistory {
  sessionId: number;
  sessionTitle: string;
  overallScore: number;
  createdAt: string;
}

interface WeeklyReportCardProps {
  allScores: ScoreHistory[];
}

function getWeekRange(weeksAgo: number): { start: Date; end: Date } {
  const now = new Date();
  const dayOfWeek = now.getDay();
  const mondayOffset = dayOfWeek === 0 ? -6 : 1 - dayOfWeek;

  const start = new Date(now);
  start.setDate(now.getDate() + mondayOffset - weeksAgo * 7);
  start.setHours(0, 0, 0, 0);

  const end = new Date(start);
  end.setDate(start.getDate() + 7);

  return { start, end };
}

export default function WeeklyReportCard({ allScores }: WeeklyReportCardProps) {
  const { thisWeek, lastWeek } = useMemo(() => {
    const thisWeekRange = getWeekRange(0);
    const lastWeekRange = getWeekRange(1);

    const thisWeekScores = allScores.filter(s => {
      const d = new Date(s.createdAt);
      return d >= thisWeekRange.start && d < thisWeekRange.end;
    });

    const lastWeekScores = allScores.filter(s => {
      const d = new Date(s.createdAt);
      return d >= lastWeekRange.start && d < lastWeekRange.end;
    });

    return { thisWeek: thisWeekScores, lastWeek: lastWeekScores };
  }, [allScores]);

  const thisWeekCount = thisWeek.length;
  const thisWeekAvg = thisWeekCount > 0
    ? Math.round((thisWeek.reduce((s, v) => s + v.overallScore, 0) / thisWeekCount) * 10) / 10
    : 0;

  const lastWeekCount = lastWeek.length;
  const lastWeekAvg = lastWeekCount > 0
    ? Math.round((lastWeek.reduce((s, v) => s + v.overallScore, 0) / lastWeekCount) * 10) / 10
    : 0;

  const countDiff = thisWeekCount - lastWeekCount;
  const scoreDiff = Math.round((thisWeekAvg - lastWeekAvg) * 10) / 10;

  const practiceDays = new Set(thisWeek.map(s => new Date(s.createdAt).getDay()));
  const dayLabels = ['日', '月', '火', '水', '木', '金', '土'];

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <p className="text-xs font-medium text-[var(--color-text-secondary)] mb-3">今週のレポート</p>

      <div className="grid grid-cols-2 gap-4 mb-3">
        <div>
          <p className="text-xs text-[var(--color-text-muted)]">練習回数</p>
          <div className="flex items-baseline gap-1">
            <span className="text-lg font-semibold text-[var(--color-text-primary)]">{thisWeekCount}</span>
            <span className="text-xs text-[var(--color-text-muted)]">回</span>
            {countDiff !== 0 && (
              <span className={`text-xs ${countDiff > 0 ? 'text-emerald-400' : 'text-rose-400'}`}>
                {countDiff > 0 ? '↑' : '↓'}{Math.abs(countDiff)}
              </span>
            )}
          </div>
        </div>
        <div>
          <p className="text-xs text-[var(--color-text-muted)]">平均スコア</p>
          <div className="flex items-baseline gap-1">
            <span className="text-lg font-semibold text-[var(--color-text-primary)]">{thisWeekAvg || 0}</span>
            {scoreDiff !== 0 && thisWeekCount > 0 && (
              <span className={`text-xs ${scoreDiff > 0 ? 'text-emerald-400' : 'text-rose-400'}`}>
                {scoreDiff > 0 ? '↑' : '↓'}{Math.abs(scoreDiff)}
              </span>
            )}
          </div>
        </div>
      </div>

      <div>
        <p className="text-xs text-[var(--color-text-muted)] mb-1">練習した曜日</p>
        <div className="flex gap-1">
          {dayLabels.map((label, i) => (
            <span
              key={label}
              className={`w-7 h-7 flex items-center justify-center rounded-full text-[11px] font-medium ${
                practiceDays.has(i)
                  ? 'bg-primary-500 text-white'
                  : 'bg-surface-3 text-[var(--color-text-faint)]'
              }`}
            >
              {label}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}
