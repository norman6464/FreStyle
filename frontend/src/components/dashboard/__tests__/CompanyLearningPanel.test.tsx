import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import CompanyLearningPanel from '../CompanyLearningPanel';
import type { CompanyLearningSummary } from '@/entities/member/api/adminMemberRepository';

function renderPanel(summary: CompanyLearningSummary) {
  return render(
    <MemoryRouter>
      <CompanyLearningPanel summary={summary} />
    </MemoryRouter>,
  );
}

describe('CompanyLearningPanel', () => {
  it('KPI と直近アクティブメンバーを表示する', () => {
    renderPanel({
      traineeCount: 4,
      activeToday: 1,
      activeThisWeek: 2,
      recentMembers: [
        { userId: 11, name: '山田 太郎', lastActiveDate: '2026-07-09', recentActivityCount: 3 },
        { userId: 12, name: '佐藤 花子', lastActiveDate: '2026-07-05', recentActivityCount: 1 },
      ],
    });
    expect(screen.getByText('メンバーの学習状況')).toBeInTheDocument();
    expect(screen.getByText('4 名')).toBeInTheDocument();
    expect(screen.getByText('山田 太郎')).toBeInTheDocument();
    expect(screen.getByText('7/9 ・ 3 回')).toBeInTheDocument();
    expect(screen.getByText('佐藤 花子')).toBeInTheDocument();
  });

  it('学習活動が無いときは空メッセージを出す', () => {
    renderPanel({ traineeCount: 2, activeToday: 0, activeThisWeek: 0, recentMembers: [] });
    expect(screen.getByText('まだ学習活動がありません。')).toBeInTheDocument();
  });

  it('従業員一覧へのリンクを持つ', () => {
    renderPanel({ traineeCount: 0, activeToday: 0, activeThisWeek: 0, recentMembers: [] });
    expect(screen.getByRole('link', { name: /従業員一覧へ/ })).toHaveAttribute(
      'href',
      '/admin/members',
    );
  });
});
