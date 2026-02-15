import type { AxisScore } from '../types';
import Card from './Card';

interface SessionFeedbackSummaryProps {
  scores: AxisScore[];
  overallScore: number;
}

function getGrade(score: number): { label: string; color: string } {
  if (score >= 9.0) return { label: '秀', color: 'text-emerald-400' };
  if (score >= 8.0) return { label: '優', color: 'text-blue-400' };
  if (score >= 7.0) return { label: '良', color: 'text-primary-400' };
  if (score >= 5.0) return { label: '可', color: 'text-yellow-400' };
  return { label: '努力', color: 'text-red-400' };
}

function getBarColor(score: number): string {
  if (score >= 8.0) return 'bg-emerald-500';
  if (score >= 6.0) return 'bg-primary-500';
  if (score >= 4.0) return 'bg-yellow-500';
  return 'bg-red-500';
}

export default function SessionFeedbackSummary({ scores, overallScore }: SessionFeedbackSummaryProps) {
  if (scores.length === 0) return null;

  const grade = getGrade(overallScore);

  return (
    <Card>
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-medium text-[var(--color-text-secondary)]">スキル別スコア</p>
        <span className={`text-sm font-bold ${grade.color}`}>{grade.label}</span>
      </div>

      <div className="space-y-2">
        {scores.map((s) => (
          <div key={s.axis} className="flex items-center gap-2">
            <span className="text-[10px] text-[var(--color-text-muted)] w-20 flex-shrink-0 truncate">
              {s.axis}
            </span>
            <div
              role="progressbar"
              aria-valuenow={s.score}
              aria-valuemin={0}
              aria-valuemax={10}
              className="flex-1 bg-surface-3 rounded-full h-1.5"
            >
              <div
                className={`h-1.5 rounded-full ${getBarColor(s.score)}`}
                style={{ width: `${s.score * 10}%` }}
              />
            </div>
            <span className="text-xs font-semibold text-[var(--color-text-primary)] w-8 text-right">
              {s.score.toFixed(1)}
            </span>
          </div>
        ))}
      </div>
    </Card>
  );
}
