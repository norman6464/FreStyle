import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import WeeklyReportCard from '../WeeklyReportCard';
import type { ScoreHistory } from '../../types';

describe('WeeklyReportCard', () => {
  it('今週の練習回数を表示する', () => {
    const now = new Date();
    const today = now.toISOString();
    const scores: ScoreHistory[] = [
      { sessionId: 1, sessionTitle: 'テスト1', overallScore: 8.0, createdAt: today },
      { sessionId: 2, sessionTitle: 'テスト2', overallScore: 7.0, createdAt: today },
    ];

    render(<WeeklyReportCard allScores={scores} />);

    expect(screen.getByText('2')).toBeInTheDocument();
    expect(screen.getByText('回')).toBeInTheDocument();
  });

  it('今週の平均スコアを表示する', () => {
    const now = new Date();
    const today = now.toISOString();
    const scores: ScoreHistory[] = [
      { sessionId: 1, sessionTitle: 'テスト1', overallScore: 8.5, createdAt: today },
      { sessionId: 2, sessionTitle: 'テスト2', overallScore: 6.5, createdAt: today },
    ];

    render(<WeeklyReportCard allScores={scores} />);

    expect(screen.getByText('7.5')).toBeInTheDocument();
  });

  it('スコアがない場合は練習回数0を表示する', () => {
    render(<WeeklyReportCard allScores={[]} />);

    expect(screen.getByText('練習回数')).toBeInTheDocument();
    expect(screen.getByText('平均スコア')).toBeInTheDocument();
  });

  it('今週の練習タイトルを表示する', () => {
    render(<WeeklyReportCard allScores={[]} />);

    expect(screen.getByText('今週のレポート')).toBeInTheDocument();
  });

  it('曜日ラベルが表示される', () => {
    render(<WeeklyReportCard allScores={[]} />);
    expect(screen.getByText('月')).toBeInTheDocument();
    expect(screen.getByText('日')).toBeInTheDocument();
  });

  it('練習した曜日セクションが表示される', () => {
    render(<WeeklyReportCard allScores={[]} />);
    expect(screen.getByText('練習した曜日')).toBeInTheDocument();
  });
});
