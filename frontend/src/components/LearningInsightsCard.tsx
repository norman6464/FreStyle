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
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <p className="text-xs font-medium text-[#D0D0D0] mb-3">学習インサイト</p>
      <div className="grid grid-cols-3 gap-3">
        {stats.map((stat) => (
          <div key={stat.label} className="text-center">
            <p className="text-lg font-semibold text-[#F0F0F0]">{stat.value}</p>
            <p className="text-[10px] text-[#666666]">
              {stat.unit}
            </p>
            <p className="text-[10px] text-[#888888] mt-0.5">{stat.label}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
