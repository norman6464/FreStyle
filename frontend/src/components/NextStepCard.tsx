interface NextStepCardProps {
  totalSessions: number;
  averageScore: number;
}

interface Step {
  emoji: string;
  title: string;
  description: string;
}

function getNextStep(totalSessions: number, averageScore: number): Step {
  if (totalSessions === 0) {
    return {
      emoji: '🚀',
      title: '最初の練習を始めよう',
      description: '練習モードでビジネスシナリオに挑戦し、コミュニケーションスキルを磨きましょう。',
    };
  }

  if (totalSessions <= 2) {
    return {
      emoji: '💪',
      title: '練習を続けましょう',
      description: 'まだ始めたばかりです。繰り返し練習することでスキルが定着します。',
    };
  }

  if (averageScore < 7) {
    return {
      emoji: '📈',
      title: 'スコアアップを目指そう',
      description: '弱点軸を意識して練習すると、効率よくスコアを伸ばせます。',
    };
  }

  return {
    emoji: '⭐',
    title: '新しいシナリオに挑戦',
    description: '素晴らしい成績です！まだ試していないシナリオで経験を広げましょう。',
  };
}

export default function NextStepCard({ totalSessions, averageScore }: NextStepCardProps) {
  const step = getNextStep(totalSessions, averageScore);

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <p className="text-xs font-medium text-[#D0D0D0] mb-2">次のステップ</p>
      <div className="flex items-start gap-3">
        <span className="text-2xl">{step.emoji}</span>
        <div>
          <p className="text-sm font-medium text-[#F0F0F0]">{step.title}</p>
          <p className="text-xs text-[#888888] mt-0.5">{step.description}</p>
        </div>
      </div>
    </div>
  );
}
