import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter } from 'react-router-dom';
import MenuPage from '../ui/MenuPage';
import authReducer, { setAuthData } from '@/entities/user/model/authSlice';
import { createMockStorage } from '@/test/mockStorage';

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
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it('trainee は統計取得を行わず即時にメニューカードを表示する', () => {
    renderMenu('trainee');

    expect(screen.getByRole('heading', { name: 'FreStyle へようこそ', level: 1 })).toBeInTheDocument();
    expect(screen.getByText('コース')).toBeInTheDocument();
    expect(screen.getByText('コード演習')).toBeInTheDocument();
  });

  it('コース/演習カードに学べる技術ロゴ(Devicon)が出る (FRESTYLE-179)', () => {
    const { container } = renderMenu('trainee');
    // LanguageIcon は /lang/<key>.svg を img で描画する。コース(git 等)・演習(go 等)のロゴが出る。
    expect(container.querySelector('img[src="/lang/git.svg"]')).not.toBeNull();
    expect(container.querySelector('img[src="/lang/typescript.svg"]')).not.toBeNull();
    // go は両カードに出るため 2 つ以上。
    expect(container.querySelectorAll('img[src="/lang/go.svg"]').length).toBeGreaterThanOrEqual(2);
  });

  it('super_admin は統計を取得せず即時に管理メニューを表示する', () => {
    renderMenu('super_admin');

    expect(screen.getByRole('heading', { name: '管理メニュー', level: 1 })).toBeInTheDocument();
    expect(screen.getByText('招待管理')).toBeInTheDocument();
  });

  it('company_admin は学習・ツール・管理セクションを表示する', () => {
    renderMenu('company_admin');

    expect(screen.getByRole('heading', { name: 'FreStyle へようこそ', level: 1 })).toBeInTheDocument();
    expect(screen.getByText('コース')).toBeInTheDocument();
    expect(screen.getByText('従業員一覧')).toBeInTheDocument();
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
