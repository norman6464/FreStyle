import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter } from 'react-router-dom';
import MenuPage from '../ui/MenuPage';
import authReducer, { setAuthData } from '@/entities/user/model/authSlice';
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

function renderMenu(role: string | null) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: { role } as never,
    },
  });
  const view = render(
    <Provider store={store}>
      <MemoryRouter>
        <MenuPage />
      </MemoryRouter>
    </Provider>,
  );
  return { ...view, store };
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

  it('コース/演習カードに学べる技術ロゴ(Devicon)が出る (FRESTYLE-179)', () => {
    mockUseUserDashboard.mockReturnValue({ dashboard: sampleDashboard, loading: false, error: null });
    const { container } = renderMenu('trainee');
    // LanguageIcon は /lang/<key>.svg を img で描画する。コース(git 等)・演習(go 等)のロゴが出る。
    expect(container.querySelector('img[src="/lang/git.svg"]')).not.toBeNull();
    expect(container.querySelector('img[src="/lang/typescript.svg"]')).not.toBeNull();
    // go は両カードに出るため 2 つ以上。
    expect(container.querySelectorAll('img[src="/lang/go.svg"]').length).toBeGreaterThanOrEqual(2);
  });

  it('super_admin は統計を取得せず即時に管理メニューを表示する', () => {
    mockUseUserDashboard.mockReturnValue({ dashboard: null, loading: false, error: null });
    renderMenu('super_admin');

    expect(screen.getByRole('heading', { name: '管理メニュー', level: 1 })).toBeInTheDocument();
    expect(screen.getByText('会社一覧')).toBeInTheDocument();
    // 学習統計は出さない
    expect(screen.queryByText('連続学習')).not.toBeInTheDocument();
  });

  it('company_admin は学習・ツール・管理セクションを表示する', () => {
    mockUseCompanyLearningSummary.mockReturnValue({ summary: sampleSummary, loading: false, error: null });
    renderMenu('company_admin');

    expect(screen.getByRole('heading', { name: 'FreStyle へようこそ', level: 1 })).toBeInTheDocument();
    expect(screen.getByText('コース')).toBeInTheDocument();
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

  // role が null のときは「未認証」と「未確定」を区別できない。確定前に描画すると
  // すべてのロール判定が false になり、既定として学習者向けが出てしまう。
  // 管理者のログイン直後に学習者画面が一瞬映る原因だった（FRESTYLE-233）。
  describe('ロールが未確定のとき', () => {
    it('学習者向けのカードを一切描画しない', () => {
      renderMenu(null);

      expect(screen.queryByText('コース')).not.toBeInTheDocument();
      expect(screen.queryByText('コード演習')).not.toBeInTheDocument();
      expect(screen.queryByText('ノート')).not.toBeInTheDocument();
      expect(screen.queryByText('学習レポート')).not.toBeInTheDocument();
    });

    it('管理者向けのカードも見出しも描画しない', () => {
      renderMenu(null);

      expect(screen.queryByText('会社一覧')).not.toBeInTheDocument();
      expect(screen.queryByText('招待管理')).not.toBeInTheDocument();
      expect(screen.queryByText('管理メニュー')).not.toBeInTheDocument();
      expect(screen.queryByText('FreStyle へようこそ')).not.toBeInTheDocument();
    });

    it('読み込み中であることを支援技術に伝える', () => {
      renderMenu(null);

      expect(screen.getByRole('status')).toHaveTextContent('読み込み中');
    });

    // 同じ store を保ったままロールを流し込み、実際の遷移として検証する
    // （別マウントで比較すると「起動時の状態」を 2 通り見ているだけになる）。
    it('ロールが確定したら本来の画面に切り替わる', () => {
      const { store } = renderMenu(null);
      expect(screen.getByRole('status')).toBeInTheDocument();
      expect(screen.queryByText('管理メニュー')).not.toBeInTheDocument();

      act(() => {
        store.dispatch(setAuthData({ role: 'super_admin', isAdmin: true }));
      });

      expect(screen.queryByRole('status')).not.toBeInTheDocument();
      expect(screen.getByText('管理メニュー')).toBeInTheDocument();
      expect(screen.queryByText('コース')).not.toBeInTheDocument();
    });
  });
});
