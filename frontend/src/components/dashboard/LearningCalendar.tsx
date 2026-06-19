import type { UserDailyActivity } from '../../types';

interface Props {
  activities: UserDailyActivity[];
}

/**
 * LearningCalendar は過去 90 日間の学習活動を GitHub スタイルのヒートマップで表示する。
 * 活動量（exercise + lesson + ai + note の合計）に応じて 4 段階の色を割り当てる。
 */
export default function LearningCalendar({ activities }: Props) {
  // date string → 合計アクティビティ数 のマップを作る。
  const actMap = new Map<string, number>();
  for (const a of activities) {
    const dateKey = a.activityDate.slice(0, 10); // "2026-06-19"
    actMap.set(dateKey, (a.exerciseCount + a.lessonCount + a.aiChatCount + a.noteCount));
  }

  // 今日から 90 日分の日付リストを生成（UTC 基準 — activityDate の UTC 文字列と一致させる）。
  const todayUtc = new Date(Date.UTC(
    new Date().getUTCFullYear(),
    new Date().getUTCMonth(),
    new Date().getUTCDate(),
  ));
  const days: string[] = [];
  for (let i = 89; i >= 0; i--) {
    const d = new Date(todayUtc);
    d.setUTCDate(todayUtc.getUTCDate() - i);
    days.push(d.toISOString().slice(0, 10));
  }

  const colorClass = (count: number) => {
    if (count === 0) return 'bg-[var(--color-surface-3)]';
    if (count <= 2) return 'bg-emerald-200 dark:bg-emerald-900';
    if (count <= 5) return 'bg-emerald-400 dark:bg-emerald-600';
    return 'bg-emerald-600 dark:bg-emerald-400';
  };

  return (
    <div className="space-y-1.5">
      <h3 className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-widest">
        学習カレンダー（過去 90 日）
      </h3>
      <div className="flex flex-wrap gap-[3px]">
        {days.map((day) => {
          const count = actMap.get(day) ?? 0;
          return (
            <div
              key={day}
              title={`${day}: ${count} アクティビティ`}
              className={`w-3 h-3 rounded-[2px] ${colorClass(count)}`}
            />
          );
        })}
      </div>
      <div className="flex items-center gap-1.5 text-[10px] text-[var(--color-text-muted)]">
        <span>少ない</span>
        <div className="w-2.5 h-2.5 rounded-[2px] bg-[var(--color-surface-3)]" />
        <div className="w-2.5 h-2.5 rounded-[2px] bg-emerald-200 dark:bg-emerald-900" />
        <div className="w-2.5 h-2.5 rounded-[2px] bg-emerald-400 dark:bg-emerald-600" />
        <div className="w-2.5 h-2.5 rounded-[2px] bg-emerald-600 dark:bg-emerald-400" />
        <span>多い</span>
      </div>
    </div>
  );
}
