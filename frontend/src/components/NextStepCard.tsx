import type { ComponentType, SVGProps } from 'react';
import {
  RocketLaunchIcon,
  ArrowPathIcon,
  ArrowTrendingUpIcon,
  StarIcon,
} from '@heroicons/react/24/outline';
import Card from './Card';

interface NextStepCardProps {
  totalSessions: number;
  averageScore: number;
}

interface Step {
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  title: string;
  description: string;
}

function getNextStep(totalSessions: number, averageScore: number): Step {
  if (totalSessions === 0) {
    return {
      icon: RocketLaunchIcon,
      title: '最初の練習を始めよう',
      description: '練習モードでビジネスシナリオに挑戦し、コミュニケーションスキルを磨きましょう。',
    };
  }

  if (totalSessions <= 2) {
    return {
      icon: ArrowPathIcon,
      title: '練習を続けましょう',
      description: 'まだ始めたばかりです。繰り返し練習することでスキルが定着します。',
    };
  }

  if (averageScore < 7) {
    return {
      icon: ArrowTrendingUpIcon,
      title: 'スコアアップを目指そう',
      description: '弱点軸を意識して練習すると、効率よくスコアを伸ばせます。',
    };
  }

  return {
    icon: StarIcon,
    title: '新しいシナリオに挑戦',
    description: '素晴らしい成績です！まだ試していないシナリオで経験を広げましょう。',
  };
}

export default function NextStepCard({ totalSessions, averageScore }: NextStepCardProps) {
  const step = getNextStep(totalSessions, averageScore);
  const Icon = step.icon;

  return (
    <Card>
      <p className="text-xs font-medium text-[var(--color-text-secondary)] mb-2">次のステップ</p>
      <div className="flex items-start gap-3">
        <Icon className="w-6 h-6 text-primary-400 flex-shrink-0" />
        <div>
          <p className="text-sm font-medium text-[var(--color-text-primary)]">{step.title}</p>
          <p className="text-xs text-[var(--color-text-muted)] mt-0.5">{step.description}</p>
        </div>
      </div>
    </Card>
  );
}
