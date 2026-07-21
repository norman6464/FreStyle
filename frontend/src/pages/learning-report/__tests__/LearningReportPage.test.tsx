import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import LearningReportPage from '../ui/LearningReportPage';
import * as useLearningReportModule from '@/hooks/useLearningReport';
import type { LearningReport } from '@/entities/learning-report';

vi.mock('@/hooks/useLearningReport');

const mockReport: LearningReport = {
  id: 1,
  userId: 7,
  periodFrom: '2026-02-01T00:00:00Z',
  periodTo: '2026-03-01T00:00:00Z',
  status: 'pending',
  createdAt: '2026-02-28T00:00:00Z',
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
