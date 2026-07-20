import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import DashboardStats from '../DashboardStats';
import type { UserDashboard } from '@/types';

function makeDashboard(overrides: Partial<UserDashboard> = {}): UserDashboard {
  return {
    streak: 3,
    totalExercises: 10,
    totalCorrect: 8,
    totalLessons: 6,
    recentActivity: [],
    recentChapterViews: [],
    ...overrides,
  };
}

describe('DashboardStats', () => {
  it('KPI（連続学習 / 演習 / 正答率 / 章完了）を表示する', () => {
    render(<DashboardStats dashboard={makeDashboard()} />);

    expect(screen.getByText('連続学習')).toBeInTheDocument();
    expect(screen.getByText('3 日')).toBeInTheDocument();
    expect(screen.getByText('10 問')).toBeInTheDocument();
    // 8/10 = 80%
    expect(screen.getByText('80%')).toBeInTheDocument();
    expect(screen.getByText('6 章')).toBeInTheDocument();
  });

  it('演習数が 0 のとき正答率カードを出さない', () => {
    render(<DashboardStats dashboard={makeDashboard({ totalExercises: 0, totalCorrect: 0 })} />);

    expect(screen.queryByText('正答率')).not.toBeInTheDocument();
  });

  it('学習カレンダーのヒートマップを表示し、活動日にツールチップを付ける', () => {
    const today = new Date().toISOString().slice(0, 10);
    render(
      <DashboardStats
        dashboard={makeDashboard({
          recentActivity: [
            {
              userId: 1,
              activityDate: `${today}T00:00:00Z`,
              exerciseCount: 2,
              correctCount: 2,
              lessonCount: 1,
              aiChatCount: 0,
              noteCount: 0,
            },
          ],
        })}
      />,
    );

    expect(screen.getByText('学習カレンダー（過去 90 日）')).toBeInTheDocument();
    // 活動 3 件（2+1）の日のセルが title 付きで存在する
    expect(screen.getByTitle(`${today}: 3 アクティビティ`)).toBeInTheDocument();
  });
});
