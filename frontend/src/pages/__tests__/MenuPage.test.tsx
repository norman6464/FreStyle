import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import MenuPage from '../MenuPage';
import { createMockStorage } from '../../test/mockStorage';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock('../../components/DailyGoalCard', () => ({
  default: () => <div data-testid="daily-goal-card">DailyGoalCard</div>,
}));

vi.mock('../../components/LearningInsightsCard', () => ({
  default: ({ totalSessions, averageScore, streakDays }: { totalSessions: number; averageScore: number; streakDays: number }) => (
    <div data-testid="learning-insights">{totalSessions} sessions, {averageScore} avg, {streakDays} days</div>
  ),
}));

vi.mock('../../components/WeeklyReportCard', () => ({
  default: () => <div data-testid="weekly-report">WeeklyReportCard</div>,
}));

vi.mock('../../components/RecentNotesCard', () => ({
  default: () => <div data-testid="recent-notes">RecentNotesCard</div>,
}));

vi.mock('../../components/CommunicationTipCard', () => ({
  default: () => <div data-testid="communication-tip">CommunicationTipCard</div>,
}));

vi.mock('../../components/DailyChallengeCard', () => ({
  default: () => <div data-testid="daily-challenge">DailyChallengeCard</div>,
}));

const mockUseMenuData = vi.fn();
vi.mock('../../hooks/useMenuData', () => ({
  useMenuData: () => mockUseMenuData(),
}));

function defaultMenuData() {
  return {
    stats: { chatPartnerCount: 5 },
    totalUnread: 5,
    latestScore: { sessionId: 1, sessionTitle: 'テスト', overallScore: 7.5, createdAt: '2026-02-13' },
    allScores: [{ sessionId: 1, sessionTitle: 'テスト', overallScore: 7.5, createdAt: '2026-02-13' }],
    totalSessions: 1,
    averageScore: 7.5,
    uniqueDays: 1,
    practiceDates: ['2026-02-13'],
    sessionsThisWeek: 1,
    loading: false,
  };
}

describe('MenuPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal('localStorage', createMockStorage());
    mockUseMenuData.mockReturnValue(defaultMenuData());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('MenuNavigationCard のメニュー項目がすべて表示される', () => {
    render(<BrowserRouter><MenuPage /></BrowserRouter>);

    // MenuNavigationCard 内のナビゲーション項目は role=button + aria-label で描画されるので
    // ロール＆アクセシブルネームで検証し、PageIntroのGlossaryTerm "練習モード" と混同しない
    expect(screen.getByRole('button', { name: 'チャット' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'AI アシスタント' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '練習モード' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'スコア履歴' })).toBeInTheDocument();
  });

  it('会話した人数を表示する', () => {
    render(<BrowserRouter><MenuPage /></BrowserRouter>);

    expect(screen.getByText('5')).toBeInTheDocument();
  });

  it('未読メッセージ数バッジをチャットカードに表示する', () => {
    render(<BrowserRouter><MenuPage /></BrowserRouter>);

    expect(screen.getByText('5件の未読')).toBeInTheDocument();
  });

  it('最新スコアをスコア履歴カードに表示する', () => {
    render(<BrowserRouter><MenuPage /></BrowserRouter>);

    expect(screen.getByText(/最新: 7\.5/)).toBeInTheDocument();
  });

  it('初回ユーザー (totalSessions=0) には FirstTimeWelcome ウェルカムカードを表示する', () => {
    mockUseMenuData.mockReturnValue({
      stats: { chatPartnerCount: 0 },
      totalUnread: 0,
      latestScore: null,
      allScores: [],
      totalSessions: 0,
      averageScore: 0,
      uniqueDays: 0,
      practiceDates: [],
      sessionsThisWeek: 0,
      loading: false,
    });

    render(<BrowserRouter><MenuPage /></BrowserRouter>);

    expect(screen.getByRole('heading', { name: 'ようこそ FreStyle へ' })).toBeInTheDocument();
    expect(screen.getByText(/シナリオを選んで AI と練習/)).toBeInTheDocument();
  });

  it('未読がない場合は未読バッジを表示しない', () => {
    mockUseMenuData.mockReturnValue({
      ...defaultMenuData(),
      totalUnread: 0,
    });

    render(<BrowserRouter><MenuPage /></BrowserRouter>);

    expect(screen.getByText('チャット')).toBeInTheDocument();
    expect(screen.queryByText(/件の未読/)).not.toBeInTheDocument();
  });

  it('ローディング中はLoadingコンポーネントが表示される', () => {
    mockUseMenuData.mockReturnValue({
      ...defaultMenuData(),
      loading: true,
    });

    render(<BrowserRouter><MenuPage /></BrowserRouter>);

    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.queryByText('チャット')).not.toBeInTheDocument();
  });
});
