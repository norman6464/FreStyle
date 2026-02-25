import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ReportCard from '../ReportCard';
import type { LearningReport } from '../../types';

const baseReport: LearningReport = {
  id: 1,
  year: 2024,
  month: 6,
  totalSessions: 15,
  averageScore: 7.8,
  practiceDays: 10,
  scoreChange: 0.5,
  bestAxis: '傾聴力',
  worstAxis: '説得力',
};

describe('ReportCard', () => {
  it('年月が表示される', () => {
    render(<ReportCard report={baseReport} />);
    expect(screen.getByText('2024年6月')).toBeInTheDocument();
  });

  it('統計データが表示される', () => {
    render(<ReportCard report={baseReport} />);
    expect(screen.getByText('15')).toBeInTheDocument();
    expect(screen.getByText('7.8')).toBeInTheDocument();
    expect(screen.getByText('10')).toBeInTheDocument();
  });

  it('スコア変動が表示される', () => {
    render(<ReportCard report={baseReport} />);
    expect(screen.getByText('+0.5')).toBeInTheDocument();
  });

  it('強みと課題が表示される', () => {
    render(<ReportCard report={baseReport} />);
    expect(screen.getByText('強み: 傾聴力')).toBeInTheDocument();
    expect(screen.getByText('課題: 説得力')).toBeInTheDocument();
  });
});
