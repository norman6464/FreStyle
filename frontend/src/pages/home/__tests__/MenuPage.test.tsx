import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter } from 'react-router-dom';
import MenuPage from '../ui/MenuPage';
import authReducer from '@/entities/user/model/authSlice';
import { useUserDashboard } from '../model/useUserDashboard';
import { useCompanyLearningSummary } from '../model/useCompanyLearningSummary';
import type { UserDashboard } from '@/entities/user';
import type { CompanyLearningSummary } from '@/entities/member/api/adminMemberRepository';
import { createMockStorage } from '@/test/mockStorage';

vi.mock('../model/useUserDashboard');
vi.mock('../model/useCompanyLearningSummary');

const mockUseUserDashboard = vi.mocked(useUserDashboard);
const mockUseCompanyLearningSummary = vi.mocked(useCompanyLearningSummary);

const sampleDashboard: UserDashboard = {
  streak: 2,
  totalExercises: 5,
  totalCorrect: 4,
  totalLessons: 6,
  recentActivity: [],
  recentChapterViews: [],
};

const sampleSummary: CompanyLearningSummary = {
  traineeCount: 4,
  activeToday: 1,
  activeThisWeek: 2,
  recentMembers: [
    { userId: 11, name: '山田 太郎', lastActiveDate: '2026-07-09', recentActivityCount: 3 },
  ],
};

function renderMenu(role: string, aiChatEnabledForTrainees = true) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: { role, aiChatEnabledForTrainees } as never,
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
    // 既定はどちらのサイドバー hook も「無効(取得なし)」相当。各テストで上書きする。
    mockUseUserDashboard.mockReturnValue({ dashboard: null, loading: false, error: null });
    mockUseCompanyLearningSummary.mockReturnValue({ summary: null, loading: false, error: null });
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

  it('company_admin は学習・ツール・管理セクションと AI カードを表示する', () => {
    mockUseCompanyLearningSummary.mockReturnValue({ summary: sampleSummary, loading: false, error: null });
    renderMenu('company_admin');

    expect(screen.getByRole('heading', { name: 'FreStyle へようこそ', level: 1 })).toBeInTheDocument();
    expect(screen.getByText('コース')).toBeInTheDocument();
    expect(screen.getByText('AI チャット')).toBeInTheDocument();
    expect(screen.getByText('従業員一覧')).toBeInTheDocument();
  });

  it('company_admin のサイドバーは自分の統計ではなくメンバーの学習状況を表示する (FRESTYLE-103)', () => {
    mockUseCompanyLearningSummary.mockReturnValue({ summary: sampleSummary, loading: false, error: null });
    renderMenu('company_admin');

    expect(screen.getByText('メンバーの学習状況')).toBeInTheDocument();
    expect(screen.getByText('在籍メンバー')).toBeInTheDocument();
    expect(screen.getByText('山田 太郎')).toBeInTheDocument();
    // 自分の学習統計(連続学習)は出さない。
    expect(screen.queryByText('連続学習')).not.toBeInTheDocument();
  });

  it('company_admin はサマリーロード中スケルトン待ちになる', () => {
    mockUseCompanyLearningSummary.mockReturnValue({ summary: null, loading: true, error: null });
    renderMenu('company_admin');

    expect(screen.queryByText('コース')).not.toBeInTheDocument();
    expect(screen.queryByText('メンバーの学習状況')).not.toBeInTheDocument();
  });

  it('trainee で AI 無効のとき AI チャットカードを出さない', () => {
    mockUseUserDashboard.mockReturnValue({ dashboard: sampleDashboard, loading: false, error: null });
    renderMenu('trainee', false);

    expect(screen.getByText('ノート')).toBeInTheDocument();
    expect(screen.queryByText('AI チャット')).not.toBeInTheDocument();
  });

  it('統計取得に失敗してもメニューは表示し、サイドバー統計は出さない', () => {
    mockUseUserDashboard.mockReturnValue({
      dashboard: null,
      loading: false,
      error: 'ダッシュボードの取得に失敗しました',
    });
    renderMenu('trainee');

    expect(screen.getByText('コース')).toBeInTheDocument();
    expect(screen.queryByText('連続学習')).not.toBeInTheDocument();
  });
});
