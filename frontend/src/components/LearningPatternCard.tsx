import { useMemo } from 'react';

const DAY_LABELS = ['日', '月', '火', '水', '木', '金', '土'];

type PatternType = '平日集中型' | '週末集中型' | '毎日コツコツ型' | '不定期型';

const PATTERN_MESSAGES: Record<PatternType, string> = {
  '平日集中型': '平日の空き時間を活用して練習を続けています。業務と並行してスキルアップする姿勢が素晴らしいです！',
  '週末集中型': '週末にまとめて練習するスタイルです。集中して取り組めるのが強みです！',
  '毎日コツコツ型': '毎日少しずつ練習を続けるスタイルです。継続は力なり、理想的な学習パターンです！',
  '不定期型': 'まだ学習パターンが定まっていません。毎日少しずつ取り組むと効果的です！',
};

const PATTERN_COLORS: Record<PatternType, string> = {
  '平日集中型': 'text-blue-400 bg-blue-900/30',
  '週末集中型': 'text-purple-600 bg-purple-50',
  '毎日コツコツ型': 'text-emerald-400 bg-emerald-900/30',
  '不定期型': 'text-amber-400 bg-amber-900/30',
};

function analyzePattern(practiceDates: string[]): { pattern: PatternType; dayCounts: number[] } {
  const dayCounts = [0, 0, 0, 0, 0, 0, 0]; // 日〜土

  for (const dateStr of practiceDates) {
    const day = new Date(dateStr).getDay();
    dayCounts[day]++;
  }

  const weekdayCount = dayCounts[1] + dayCounts[2] + dayCounts[3] + dayCounts[4] + dayCounts[5];
  const weekendCount = dayCounts[0] + dayCounts[6];
  const total = weekdayCount + weekendCount;
  const activeDays = dayCounts.filter(c => c > 0).length;

  let pattern: PatternType;
  if (activeDays >= 6) {
    pattern = '毎日コツコツ型';
  } else if (total >= 3 && weekdayCount / total >= 0.7) {
    pattern = '平日集中型';
  } else if (total >= 3 && weekendCount / total >= 0.7) {
    pattern = '週末集中型';
  } else {
    pattern = '不定期型';
  }

  return { pattern, dayCounts };
}

interface Props {
  practiceDates: string[];
}

export default function LearningPatternCard({ practiceDates }: Props) {
  const { pattern, dayCounts } = useMemo(() => analyzePattern(practiceDates), [practiceDates]);

  if (practiceDates.length === 0) return null;

  const maxCount = Math.max(...dayCounts, 1);

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-medium text-[#D0D0D0]">学習パターン</p>
        <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${PATTERN_COLORS[pattern]}`}>
          {pattern}
        </span>
      </div>

      <div className="flex items-end gap-1 h-12 mb-2">
        {dayCounts.map((count, i) => (
          <div key={i} className="flex-1 flex flex-col items-center gap-0.5">
            <div
              className="w-full bg-primary-400 rounded-t transition-all"
              style={{ height: `${(count / maxCount) * 100}%`, minHeight: count > 0 ? '4px' : '0' }}
            />
            <span className="text-[10px] text-[#666666]">{DAY_LABELS[i]}</span>
          </div>
        ))}
      </div>

      <p className="text-xs text-[#888888]">{PATTERN_MESSAGES[pattern]}</p>
    </div>
  );
}
