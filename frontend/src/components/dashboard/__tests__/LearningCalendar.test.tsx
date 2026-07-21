import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import LearningCalendar from '../LearningCalendar';
import type { UserDailyActivity } from '@/entities/user';

function dayOffset(daysAgo: number): string {
  const t = new Date();
  const utc = new Date(Date.UTC(t.getUTCFullYear(), t.getUTCMonth(), t.getUTCDate()));
  utc.setUTCDate(utc.getUTCDate() - daysAgo);
  return utc.toISOString().slice(0, 10);
}

function activity(daysAgo: number, exerciseCount: number): UserDailyActivity {
  return {
    userId: 1,
    activityDate: `${dayOffset(daysAgo)}T00:00:00Z`,
    exerciseCount,
    correctCount: 0,
    lessonCount: 0,
    aiChatCount: 0,
    noteCount: 0,
  };
}

describe('LearningCalendar', () => {
  it('活動量の各段階（低/中/高）をツールチップ付きセルで描画する', () => {
    render(
      <LearningCalendar
        activities={[
          activity(1, 1), // <= 2
          activity(2, 4), // <= 5
          activity(3, 9), // > 5
        ]}
      />,
    );

    expect(screen.getByTitle(`${dayOffset(1)}: 1 アクティビティ`)).toBeInTheDocument();
    expect(screen.getByTitle(`${dayOffset(2)}: 4 アクティビティ`)).toBeInTheDocument();
    expect(screen.getByTitle(`${dayOffset(3)}: 9 アクティビティ`)).toBeInTheDocument();
  });

  it('活動がない日は 0 アクティビティのセルになる', () => {
    render(<LearningCalendar activities={[]} />);

    expect(screen.getByText('学習カレンダー（過去 90 日）')).toBeInTheDocument();
    // 今日のセルは活動 0
    expect(screen.getByTitle(`${dayOffset(0)}: 0 アクティビティ`)).toBeInTheDocument();
  });
});
