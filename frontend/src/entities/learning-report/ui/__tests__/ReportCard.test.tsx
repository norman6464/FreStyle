import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ReportCard from '../ReportCard';
import type { LearningReport } from '@/entities/learning-report';

function makeReport(overrides: Partial<LearningReport> = {}): LearningReport {
  return {
    id: 1,
    userId: 7,
    periodFrom: '2026-06-01T00:00:00Z',
    periodTo: '2026-07-01T00:00:00Z',
    status: 'pending',
    createdAt: '2026-06-30T09:00:00Z',
    ...overrides,
  };
}

describe('ReportCard', () => {
  it('対象月（periodFrom 由来）を表示する', () => {
    render(<ReportCard report={makeReport()} />);
    expect(screen.getByText('2026年6月 の学習レポート')).toBeInTheDocument();
  });

  it('作成日を表示する', () => {
    render(<ReportCard report={makeReport()} />);
    expect(screen.getByText(/作成日: 2026\/06\/30/)).toBeInTheDocument();
  });

  it('pending は「作成中」バッジ', () => {
    render(<ReportCard report={makeReport({ status: 'pending' })} />);
    expect(screen.getByText('作成中')).toBeInTheDocument();
  });

  it('ready は「完成」バッジ', () => {
    render(<ReportCard report={makeReport({ status: 'ready' })} />);
    expect(screen.getByText('完成')).toBeInTheDocument();
  });

  it('failed は「失敗」バッジ', () => {
    render(<ReportCard report={makeReport({ status: 'failed' })} />);
    expect(screen.getByText('失敗')).toBeInTheDocument();
  });

  it('欠損データでもクラッシュしない（旧スキーマ由来の不整合耐性）', () => {
    const broken = { id: 2, status: 'pending' } as unknown as LearningReport;
    expect(() => render(<ReportCard report={broken} />)).not.toThrow();
    expect(screen.getByText('作成中')).toBeInTheDocument();
  });

  it('作成日が未設定なら「—」を表示する', () => {
    render(<ReportCard report={makeReport({ createdAt: '' })} />);
    expect(screen.getByText(/作成日: —/)).toBeInTheDocument();
  });
});
