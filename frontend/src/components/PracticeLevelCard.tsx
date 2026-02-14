interface PracticeLevelCardProps {
  totalSessions: number;
}

interface Level {
  level: number;
  title: string;
  threshold: number;
}

const LEVELS: Level[] = [
  { level: 1, title: 'ビギナー', threshold: 0 },
  { level: 2, title: 'ルーキー', threshold: 5 },
  { level: 3, title: 'レギュラー', threshold: 10 },
  { level: 4, title: 'ベテラン', threshold: 20 },
  { level: 5, title: 'エキスパート', threshold: 30 },
  { level: 6, title: 'マスター', threshold: 50 },
  { level: 7, title: 'グランドマスター', threshold: 75 },
  { level: 8, title: 'レジェンド', threshold: 100 },
];

function getCurrentLevel(sessions: number): { current: Level; next: Level | null } {
  let current = LEVELS[0];
  for (let i = LEVELS.length - 1; i >= 0; i--) {
    if (sessions >= LEVELS[i].threshold) {
      current = LEVELS[i];
      break;
    }
  }
  const nextIndex = LEVELS.indexOf(current) + 1;
  const next = nextIndex < LEVELS.length ? LEVELS[nextIndex] : null;
  return { current, next };
}

export default function PracticeLevelCard({ totalSessions }: PracticeLevelCardProps) {
  const { current, next } = getCurrentLevel(totalSessions);

  const progress = next
    ? ((totalSessions - current.threshold) / (next.threshold - current.threshold)) * 100
    : 100;

  const remaining = next ? next.threshold - totalSessions : 0;

  return (
    <div className="bg-white rounded-lg border border-slate-200 p-4">
      <p className="text-xs font-medium text-slate-700 mb-2">練習レベル</p>

      <div className="flex items-baseline gap-2 mb-2">
        <span className="text-xl font-bold text-primary-600">Lv.{current.level}</span>
        <span className="text-sm font-medium text-slate-600">{current.title}</span>
      </div>

      <div className="w-full bg-slate-100 rounded-full h-2 mb-2">
        <div
          data-testid="level-progress"
          className="h-2 rounded-full bg-primary-500 transition-all"
          style={{ width: `${Math.min(progress, 100)}%` }}
        />
      </div>

      <p className="text-xs text-slate-500">
        {next ? `次のレベルまであと${remaining}回` : '最高レベル到達！'}
      </p>
    </div>
  );
}
