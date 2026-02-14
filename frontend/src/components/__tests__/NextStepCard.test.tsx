import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import NextStepCard from '../NextStepCard';

describe('NextStepCard', () => {
  it('タイトルが表示される', () => {
    render(<NextStepCard totalSessions={0} averageScore={0} />);
    expect(screen.getByText('次のステップ')).toBeInTheDocument();
  });

  it('練習0回の場合は最初の練習を促す', () => {
    render(<NextStepCard totalSessions={0} averageScore={0} />);
    expect(screen.getByText(/最初の練習/)).toBeInTheDocument();
  });

  it('練習1-2回の場合は継続を促す', () => {
    render(<NextStepCard totalSessions={2} averageScore={5} />);
    expect(screen.getByText(/練習を続けましょう/)).toBeInTheDocument();
  });

  it('3回以上で低スコアの場合は弱点改善を提案', () => {
    render(<NextStepCard totalSessions={5} averageScore={5} />);
    expect(screen.getByText(/スコアアップ/)).toBeInTheDocument();
  });

  it('3回以上で高スコアの場合は新シナリオを提案', () => {
    render(<NextStepCard totalSessions={5} averageScore={8} />);
    expect(screen.getByText(/新しいシナリオ/)).toBeInTheDocument();
  });
});
