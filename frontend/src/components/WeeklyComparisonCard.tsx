interface Session {
  overallScore: number;
  createdAt: string;
}

interface WeeklyComparisonCardProps {
  sessions: Session[];
}

function getMonday(date: Date): Date {
  const d = new Date(date);
  const day = d.getDay();
  d.setDate(d.getDate() - (day === 0 ? 6 : day - 1));
  d.setHours(0, 0, 0, 0);
  return d;
}

export default function WeeklyComparisonCard({ sessions }: WeeklyComparisonCardProps) {
  const now = new Date();
  const thisMonday = getMonday(now);
  const lastMonday = new Date(thisMonday);
  lastMonday.setDate(thisMonday.getDate() - 7);

  const thisWeek = sessions.filter((s) => new Date(s.createdAt) >= thisMonday);
  const lastWeek = sessions.filter((s) => {
    const d = new Date(s.createdAt);
    return d >= lastMonday && d < thisMonday;
  });

  const thisAvg =
    thisWeek.length > 0
      ? Math.round((thisWeek.reduce((sum, s) => sum + s.overallScore, 0) / thisWeek.length) * 10) / 10
      : 0;
  const lastAvg =
    lastWeek.length > 0
      ? Math.round((lastWeek.reduce((sum, s) => sum + s.overallScore, 0) / lastWeek.length) * 10) / 10
      : 0;

  const showDelta = thisWeek.length > 0 && lastWeek.length > 0;
  const delta = showDelta ? Math.round((thisAvg - lastAvg) * 10) / 10 : 0;

  return (
    <div className="bg-white rounded-lg border border-slate-200 p-4">
      <p className="text-xs font-medium text-slate-700 mb-3">週間比較</p>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <p className="text-xs text-slate-500 mb-1">今週</p>
          <p data-testid="avg-score" className="text-lg font-bold text-slate-800">
            {thisAvg.toFixed(1)}
          </p>
          <p className="text-xs text-slate-400">
            <span data-testid="session-count">{thisWeek.length}</span> セッション
          </p>
        </div>
        <div>
          <p className="text-xs text-slate-500 mb-1">先週</p>
          <p data-testid="avg-score" className="text-lg font-bold text-slate-800">
            {lastAvg.toFixed(1)}
          </p>
          <p className="text-xs text-slate-400">
            <span data-testid="session-count">{lastWeek.length}</span> セッション
          </p>
        </div>
      </div>

      {showDelta && (
        <div className="mt-3 pt-3 border-t border-slate-100">
          <span
            data-testid="score-delta"
            className={`text-sm font-medium ${
              delta > 0 ? 'text-emerald-600' : delta < 0 ? 'text-rose-600' : 'text-slate-500'
            }`}
          >
            {delta > 0 ? `+${delta.toFixed(1)}` : delta < 0 ? `\u2212${Math.abs(delta).toFixed(1)}` : '\u00B10.0'}
          </span>
          <span className="text-xs text-slate-400 ml-1">先週比</span>
        </div>
      )}
    </div>
  );
}
