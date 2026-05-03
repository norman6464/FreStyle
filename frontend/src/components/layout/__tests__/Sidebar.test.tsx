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

const mockToggleTheme = vi.fn();
const mockUseTheme = vi.fn();
vi.mock('../../../hooks/useTheme', () => ({
  useTheme: () => mockUseTheme(),
}));

function createTestStore(isAdmin = false) {
  return configureStore({
    reducer: { auth: authReducer },
    preloadedState: { auth: { isAuthenticated: true, loading: false, isAdmin } },
  });
}

function renderSidebar(currentPath = '/', isAdmin = false) {
  return render(
    <Provider store={createTestStore(isAdmin)}>
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
    mockUseTheme.mockReturnValue({
      theme: 'dark',
      toggleTheme: mockToggleTheme,
    });
  });

  it('全ナビゲーション項目を表示する', () => {
    renderSidebar();
    expect(screen.getByText('ホーム')).toBeDefined();
    expect(screen.getByText('AI')).toBeDefined();
    expect(screen.getByText('練習')).toBeDefined();
    expect(screen.getByText('プロフィール')).toBeDefined();
  });

  it('ログアウトボタンを表示する', () => {
    renderSidebar();
    expect(screen.getByText('ログアウト')).toBeDefined();
  });

  it('ホームルートでホームがアクティブになる', () => {
    renderSidebar('/');
    const homeLink = screen.getByText('ホーム').closest('a');
    expect(homeLink?.className).toContain('bg-surface-3');
  });

  it('ダークモード時にライトモードボタンを表示する', () => {
    renderSidebar();
    expect(screen.getByText('ライトモード')).toBeInTheDocument();
  });

  it('ライトモード時にダークモードボタンを表示する', () => {
    mockUseTheme.mockReturnValue({
      theme: 'light',
      toggleTheme: mockToggleTheme,
    });
    renderSidebar();
    expect(screen.getByText('ダークモード')).toBeInTheDocument();
  });

  it('テーマ切り替えボタンクリックでtoggleThemeが呼ばれる', () => {
    renderSidebar();
    fireEvent.click(screen.getByText('ライトモード'));
    expect(mockToggleTheme).toHaveBeenCalledTimes(1);
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

  it('admin メニュークリックでサブメニューが展開される', () => {
    renderSidebar('/', true);
    fireEvent.click(screen.getByText('管理'));
    expect(screen.getByText('会社一覧')).toBeInTheDocument();
    expect(screen.getByText('シナリオ管理')).toBeInTheDocument();
    expect(screen.getByText('招待管理')).toBeInTheDocument();
  });

  it('非 admin ユーザーには管理メニューが表示されない', () => {
    renderSidebar('/');
    expect(screen.queryByText('管理')).toBeNull();
  });
});
