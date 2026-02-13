import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import LearningInsightsCard from '../LearningInsightsCard';

describe('LearningInsightsCard', () => {
  it('統計情報が表示される', () => {
    render(
      <LearningInsightsCard
        totalSessions={12}
        averageScore={7.8}
        streakDays={5}
      />
    );

    expect(screen.getByText('学習インサイト')).toBeInTheDocument();
    expect(screen.getByText('12')).toBeInTheDocument();
    expect(screen.getByText('7.8')).toBeInTheDocument();
    expect(screen.getByText('5')).toBeInTheDocument();
  });

  it('ラベルが表示される', () => {
    render(
      <LearningInsightsCard
        totalSessions={0}
        averageScore={0}
        streakDays={0}
      />
    );

    expect(screen.getByText('総練習回数')).toBeInTheDocument();
    expect(screen.getByText('平均スコア')).toBeInTheDocument();
    expect(screen.getByText('連続練習日')).toBeInTheDocument();
  });

  it('データなしでもクラッシュしない', () => {
    render(
      <LearningInsightsCard
        totalSessions={0}
        averageScore={0}
        streakDays={0}
      />
    );

    expect(screen.getByText('学習インサイト')).toBeInTheDocument();
  });
});
