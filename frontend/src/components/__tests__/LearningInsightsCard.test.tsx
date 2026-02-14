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

  it('単位が表示される', () => {
    render(
      <LearningInsightsCard
        totalSessions={5}
        averageScore={6.5}
        streakDays={3}
      />
    );

    expect(screen.getByText('回')).toBeInTheDocument();
    expect(screen.getByText('/10')).toBeInTheDocument();
    expect(screen.getByText('日')).toBeInTheDocument();
  });

  it('平均スコアが小数点1桁で表示される', () => {
    render(
      <LearningInsightsCard
        totalSessions={1}
        averageScore={7}
        streakDays={1}
      />
    );

    expect(screen.getByText('7.0')).toBeInTheDocument();
  });

  it('3カラムグリッドレイアウトが適用される', () => {
    const { container } = render(
      <LearningInsightsCard
        totalSessions={1}
        averageScore={5.0}
        streakDays={1}
      />
    );

    const grid = container.querySelector('.grid-cols-3');
    expect(grid).toBeTruthy();
  });
});
