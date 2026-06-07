import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import authReducer from '../../../store/authSlice';
import Sidebar from '../Sidebar';

const mockUseSidebar = vi.fn();
vi.mock('../../../hooks/useSidebar', () => ({
  useSidebar: () => mockUseSidebar(),
}));

// Sidebar はマウント時に ProfileRepository.fetchProfile を呼んで user メニュー (avatar / 名前) を出す。
// テストでは固定値を返すモックに差し替える。
vi.mock('../../../repositories/ProfileRepository', () => ({
  default: {
    fetchProfile: vi.fn().mockResolvedValue({
      userId: 1, displayName: 'テストユーザー', bio: '', avatarUrl: '', status: '',
    }),
    updateProfile: vi.fn(),
  },
}));

function createTestStore(isAdmin = false, role: string | null = null, aiEnabled = true) {
  return configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: {
        isAuthenticated: true,
        loading: false,
        isAdmin,
        role,
        aiChatEnabledForTrainees: aiEnabled,
      },
    },
  });
}

function renderSidebar(
  currentPath = '/',
  isAdmin = false,
  role: string | null = null,
  aiEnabled = true
) {
  return render(
    <Provider store={createTestStore(isAdmin, role, aiEnabled)}>
      <MemoryRouter initialEntries={[currentPath]}>
        <Sidebar />
      </MemoryRouter>
    </Provider>
  );
}

describe('Sidebar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseSidebar.mockReturnValue({
      handleLogout: vi.fn(),
      loggingOut: false,
    });
  });

  it('全ナビゲーション項目を表示する', () => {
    renderSidebar();
    expect(screen.getByText('ホーム')).toBeDefined();
    expect(screen.getByText('AI')).toBeDefined();
    expect(screen.getByText('演習')).toBeDefined();
    expect(screen.getByText('コース')).toBeDefined();
    expect(screen.getByText('ノート')).toBeDefined();
    expect(screen.getByText('レポート')).toBeDefined();
    // プロフィールは sidebar bottom のユーザーメニュー（クリックで開くドロップダウン）に移動済。
  });

  it('削除済み機能のリンクは表示されない', () => {
    renderSidebar();
    // 練習 / お気に入り / スコア履歴 / ランキング などはコア機能整理で削除済
    expect(screen.queryByText('練習')).toBeNull();
    expect(screen.queryByText('お気に入り')).toBeNull();
    expect(screen.queryByText('スコア履歴')).toBeNull();
  });

  it('ユーザーメニューを開くと「設定」「ログアウト」が表示される', async () => {
    renderSidebar();
    // 初期表示ではログアウトボタンは非表示（ドロップダウンの中）。
    expect(screen.queryByText('ログアウト')).toBeNull();
    // ユーザーボタン（aria-haspopup="menu"）をクリックすると展開する。
    const userButton = screen.getByRole('button', { expanded: false, name: /./ });
    fireEvent.click(userButton);
    expect(screen.getByRole('menuitem', { name: /設定/ })).toBeDefined();
    expect(screen.getByRole('menuitem', { name: /ログアウト/ })).toBeDefined();
  });

  it('ホームルートでホームがアクティブになる', () => {
    renderSidebar('/');
    const homeLink = screen.getByText('ホーム').closest('a');
    expect(homeLink?.className).toContain('bg-[var(--color-nav-active)]');
  });

  it('折りたたみボタンでサイドバーが折りたたまれる', () => {
    renderSidebar();
    // 展開状態ではラベルが見える
    expect(screen.getByText('ホーム')).toBeVisible();
    // 折りたたみボタンをクリック
    fireEvent.click(screen.getByTitle('サイドバーを閉じる'));
    // 折りたたみ後はラベルが消える（expanded=false）
    expect(screen.queryByText('ホーム')).toBeNull();
  });

  it('折りたたみ後に展開ボタンでラベルが戻る', () => {
    renderSidebar();
    fireEvent.click(screen.getByTitle('サイドバーを閉じる'));
    fireEvent.click(screen.getByTitle('サイドバーを開く'));
    expect(screen.getByText('ホーム')).toBeVisible();
  });

  it('admin ユーザーには管理メニューが表示される', () => {
    renderSidebar('/', true);
    expect(screen.getByText('管理')).toBeInTheDocument();
  });

  it('super_admin の管理メニューには会社一覧と招待管理が表示される', () => {
    renderSidebar('/', true, 'super_admin');
    fireEvent.click(screen.getByText('管理'));
    expect(screen.getByText('会社一覧')).toBeInTheDocument();
    expect(screen.getByText('招待管理')).toBeInTheDocument();
    // シナリオ管理は廃止済（コア機能整理）
    expect(screen.queryByText('シナリオ管理')).toBeNull();
  });

  it('company_admin の管理メニューには招待管理のみが表示され、会社一覧は出ない', () => {
    renderSidebar('/', true, 'company_admin');
    fireEvent.click(screen.getByText('管理'));
    expect(screen.queryByText('会社一覧')).toBeNull();
    expect(screen.getByText('招待管理')).toBeInTheDocument();
  });

  it('非 admin ユーザーには管理メニューが表示されない', () => {
    renderSidebar('/');
    expect(screen.queryByText('管理')).toBeNull();
  });

  it('super_admin には trainee 向けメニュー (AI / 演習 / ノート / レポート / コース) が表示されない', () => {
    renderSidebar('/', true, 'super_admin');
    expect(screen.queryByText('AI')).toBeNull();
    expect(screen.queryByText('演習')).toBeNull();
    expect(screen.queryByText('ノート')).toBeNull();
    expect(screen.queryByText('レポート')).toBeNull();
    expect(screen.queryByText('コース')).toBeNull();
    // ホーム / 通知 / 管理は表示される
    expect(screen.getByText('ホーム')).toBeInTheDocument();
    expect(screen.getByText('通知')).toBeInTheDocument();
    expect(screen.getByText('管理')).toBeInTheDocument();
  });

  it('company_admin には trainee 向けメニューも表示される', () => {
    renderSidebar('/', true, 'company_admin');
    expect(screen.getByText('AI')).toBeInTheDocument();
    expect(screen.getByText('演習')).toBeInTheDocument();
    expect(screen.getByText('ノート')).toBeInTheDocument();
    expect(screen.getByText('レポート')).toBeInTheDocument();
    expect(screen.getByText('管理')).toBeInTheDocument();
  });

  it('trainee は会社が AI 無効なら「AI」が非表示になる', () => {
    renderSidebar('/', false, 'trainee', false);
    expect(screen.queryByText('AI')).toBeNull();
    // 他の trainee 向けメニューは残る
    expect(screen.getByText('演習')).toBeInTheDocument();
    expect(screen.getByText('ノート')).toBeInTheDocument();
  });

  it('trainee は会社が AI 有効なら「AI」が表示される', () => {
    renderSidebar('/', false, 'trainee', true);
    expect(screen.getByText('AI')).toBeInTheDocument();
  });

  it('company_admin は会社が AI 無効でも自分の「AI」は表示される', () => {
    renderSidebar('/', true, 'company_admin', false);
    expect(screen.getByText('AI')).toBeInTheDocument();
  });
});
