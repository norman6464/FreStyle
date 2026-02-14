import Card from './Card';

interface LearningInsightsCardProps {
  totalSessions: number;
  averageScore: number;
  streakDays: number;
}

export default function LearningInsightsCard({
  totalSessions,
  averageScore,
  streakDays,
}: LearningInsightsCardProps) {
  const stats = [
    { label: '総練習回数', value: String(totalSessions), unit: '回' },
    { label: '平均スコア', value: averageScore.toFixed(1), unit: '/10' },
    { label: '連続練習日', value: String(streakDays), unit: '日' },
  ];

  return (
    <Card>
      <p className="text-xs font-medium text-[var(--color-text-secondary)] mb-3">学習インサイト</p>
      <div className="grid grid-cols-3 gap-3">
        {stats.map((stat) => (
          <div key={stat.label} className="text-center">
            <p className="text-lg font-semibold text-[var(--color-text-primary)]">{stat.value}</p>
            <p className="text-[10px] text-[var(--color-text-faint)]">
              {stat.unit}
            </p>
            <p className="text-[10px] text-[var(--color-text-muted)] mt-0.5">{stat.label}</p>
          </div>
        ))}
      </div>
    </Card>
  );
}
