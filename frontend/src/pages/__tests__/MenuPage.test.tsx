import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter } from 'react-router-dom';
import MenuPage from '../MenuPage';
import authReducer from '../../store/authSlice';
import { useUserDashboard } from '../../hooks/useUserDashboard';
import type { UserDashboard } from '../../types';
import { createMockStorage } from '../../test/mockStorage';

vi.mock('../../hooks/useUserDashboard');

const mockUseUserDashboard = vi.mocked(useUserDashboard);

const sampleDashboard: UserDashboard = {
  streak: 2,
  totalExercises: 5,
  totalCorrect: 4,
  totalLessons: 6,
  recentActivity: [],
  recentChapterViews: [],
};

function renderMenu(role: string) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: { role, aiChatEnabledForTrainees: true } as never,
    },
  });
  return render(
    <Provider store={store}>
      <MemoryRouter>
        <MenuPage />
      </MemoryRouter>
    </Provider>,
  );
}

describe('MenuPage', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it('統計ロード中はメニューカードを出さず（スケルトン待ち）', () => {
    mockUseUserDashboard.mockReturnValue({ dashboard: null, loading: true, error: null });
    renderMenu('trainee');

    // ウェルカム見出しは即時表示される
    expect(screen.getByRole('heading', { name: 'FreStyle へようこそ', level: 1 })).toBeInTheDocument();
    // メニューカード（コース）はロード完了まで出さない
    expect(screen.queryByText('コース')).not.toBeInTheDocument();
  });

  it('統計ロード完了後にメニューカードと統計を同時表示する', () => {
    mockUseUserDashboard.mockReturnValue({ dashboard: sampleDashboard, loading: false, error: null });
    renderMenu('trainee');

    // メニューカードと統計（連続学習）が両方出ている
    expect(screen.getByText('コース')).toBeInTheDocument();
    expect(screen.getByText('コード演習')).toBeInTheDocument();
    expect(screen.getByText('連続学習')).toBeInTheDocument();
  });

  it('super_admin は統計を取得せず即時に管理メニューを表示する', () => {
    mockUseUserDashboard.mockReturnValue({ dashboard: null, loading: false, error: null });
    renderMenu('super_admin');

    expect(screen.getByRole('heading', { name: '管理メニュー', level: 1 })).toBeInTheDocument();
    expect(screen.getByText('会社一覧')).toBeInTheDocument();
    // 学習統計は出さない
    expect(screen.queryByText('連続学習')).not.toBeInTheDocument();
  });
});
