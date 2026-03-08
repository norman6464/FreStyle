import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import WeeklyChallengePage from '../WeeklyChallengePage';
import { useWeeklyChallenge } from '../../hooks/useWeeklyChallenge';

vi.mock('../../hooks/useWeeklyChallenge');
const mockedUseWeeklyChallenge = vi.mocked(useWeeklyChallenge);

function renderPage() {
  return render(<MemoryRouter><WeeklyChallengePage /></MemoryRouter>);
}

describe('WeeklyChallengePage', () => {
  const mockChallenge = { id: 1, title: '会議スキル強化週間', description: '会議関連のシナリオを3回練習', category: 'meeting', targetSessions: 3, completedSessions: 1, isCompleted: false, weekStart: '2024-01-01', weekEnd: '2024-01-07' };

  beforeEach(() => {
    vi.clearAllMocks();
    mockedUseWeeklyChallenge.mockReturnValue({ challenge: mockChallenge, loading: false, incrementProgress: vi.fn() });
  });

  it('タイトルを表示する', () => {
    renderPage();
    expect(screen.getByText('ウィークリーチャレンジ')).toBeInTheDocument();
  });

  it('チャレンジ名を表示する', () => {
    renderPage();
    expect(screen.getByText('会議スキル強化週間')).toBeInTheDocument();
  });

  it('進捗を表示する', () => {
    renderPage();
    expect(screen.getByText('1 / 3 セッション')).toBeInTheDocument();
  });

  it('ローディング中はローディング表示する', () => {
    mockedUseWeeklyChallenge.mockReturnValue({ challenge: null, loading: true, incrementProgress: vi.fn() });
    renderPage();
    expect(screen.getByText('チャレンジを読み込み中...')).toBeInTheDocument();
  });

  it('チャレンジがない場合は空メッセージを表示する', () => {
    mockedUseWeeklyChallenge.mockReturnValue({ challenge: null, loading: false, incrementProgress: vi.fn() });
    renderPage();
    expect(screen.getByText('今週のチャレンジはまだ設定されていません')).toBeInTheDocument();
  });

  it('達成時にバッジを表示する', () => {
    mockedUseWeeklyChallenge.mockReturnValue({ challenge: { ...mockChallenge, isCompleted: true, completedSessions: 3 }, loading: false, incrementProgress: vi.fn() });
    renderPage();
    expect(screen.getByTestId('completed-badge')).toBeInTheDocument();
  });
});
