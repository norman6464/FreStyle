import Card from './Card';

interface AxisScore {
  axis: string;
  score: number;
  comment: string;
}

interface SkillMilestoneCardProps {
  scores: AxisScore[];
}

interface Milestone {
  label: string;
  threshold: number;
  color: string;
  bgColor: string;
}

const MILESTONES: Milestone[] = [
  { label: '入門', threshold: 0, color: 'text-[var(--color-text-muted)]', bgColor: 'bg-surface-3' },
  { label: '初級', threshold: 4, color: 'text-emerald-400', bgColor: 'bg-emerald-900/30' },
  { label: '中級', threshold: 6, color: 'text-blue-400', bgColor: 'bg-blue-900/30' },
  { label: '上級', threshold: 7, color: 'text-purple-600', bgColor: 'bg-purple-50' },
  { label: 'エキスパート', threshold: 9, color: 'text-amber-400', bgColor: 'bg-amber-900/30' },
];

function getCurrentMilestone(score: number): Milestone {
  for (let i = MILESTONES.length - 1; i >= 0; i--) {
    if (score >= MILESTONES[i].threshold) return MILESTONES[i];
  }
  return MILESTONES[0];
}

function getNextMilestone(score: number): Milestone | null {
  for (const m of MILESTONES) {
    if (score < m.threshold) return m;
  }
  return null;
}

export default function SkillMilestoneCard({ scores }: SkillMilestoneCardProps) {
  if (scores.length === 0) return null;

  return (
    <Card>
      <h3 className="text-xs font-semibold text-[var(--color-text-primary)] mb-3">スキル到達レベル</h3>
      <div className="space-y-3">
        {scores.map((s) => {
          const current = getCurrentMilestone(s.score);
          const next = getNextMilestone(s.score);
          const remaining = next ? Math.round((next.threshold - s.score) * 10) / 10 : null;
          const progress = next
            ? ((s.score - current.threshold) / (next.threshold - current.threshold)) * 100
            : 100;

          return (
            <div key={s.axis}>
              <div className="flex items-center justify-between mb-1">
                <span className="text-xs text-[var(--color-text-secondary)]">{s.axis}</span>
                <span className={`text-[10px] font-medium px-1.5 py-0.5 rounded ${current.bgColor} ${current.color}`}>
                  {current.label}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <div className="flex-1 bg-surface-3 rounded-full h-1.5">
                  <div
                    className="h-1.5 rounded-full bg-primary-500 transition-all"
                    style={{ width: `${Math.min(progress, 100)}%` }}
                  />
                </div>
                {remaining !== null && (
                  <span className="text-[10px] text-[var(--color-text-faint)] whitespace-nowrap">あと {remaining}</span>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </Card>
  );
}
