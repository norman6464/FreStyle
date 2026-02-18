import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import LearningReportPage from '../LearningReportPage';
import * as useLearningReportModule from '../../hooks/useLearningReport';
import type { LearningReport } from '../../types';

vi.mock('../../hooks/useLearningReport');

const mockReport: LearningReport = {
  id: 1,
  year: 2026,
  month: 2,
  totalSessions: 5,
  averageScore: 75.0,
  previousAverageScore: 70.0,
  scoreChange: 5.0,
  bestAxis: '論理的構成力',
  worstAxis: '配慮表現',
  practiceDays: 3,
  createdAt: '2026-02-01T00:00:00',
};

describe('LearningReportPage', () => {
  it('生成中はボタンが「生成中...」になり無効化される', () => {
    vi.mocked(useLearningReportModule.useLearningReport).mockReturnValue({
      reports: [mockReport],
      loading: false,
      generating: true,
      generateReport: vi.fn(),
      refresh: vi.fn(),
    });

    render(<LearningReportPage />);

    const button = screen.getByRole('button', { name: '生成中...' });
    expect(button).toBeDisabled();
  });

  it('生成中でない場合はボタンが有効', () => {
    vi.mocked(useLearningReportModule.useLearningReport).mockReturnValue({
      reports: [mockReport],
      loading: false,
      generating: false,
      generateReport: vi.fn(),
      refresh: vi.fn(),
    });

    render(<LearningReportPage />);

    const button = screen.getByRole('button', { name: '今月のレポートを生成' });
    expect(button).not.toBeDisabled();
  });

  it('ボタンクリックでgenerateReportが呼ばれる', () => {
    const mockGenerateReport = vi.fn();
    vi.mocked(useLearningReportModule.useLearningReport).mockReturnValue({
      reports: [mockReport],
      loading: false,
      generating: false,
      generateReport: mockGenerateReport,
      refresh: vi.fn(),
    });

    render(<LearningReportPage />);

    fireEvent.click(screen.getByRole('button', { name: '今月のレポートを生成' }));
    expect(mockGenerateReport).toHaveBeenCalled();
  });
});
