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

function createTestStore() {
  return configureStore({
    reducer: { auth: authReducer },
    preloadedState: { auth: { isAuthenticated: true, loading: false } },
  });
}

function renderSidebar(currentPath = '/') {
  return render(
    <Provider store={createTestStore()}>
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
    // Teams スタイルサイドバーはボタンで実装。アクティブ時に bg-surface-3 クラスを持つ
    const homeBtn = screen.getByText('ホーム').closest('button');
    expect(homeBtn?.className).toContain('bg-surface-3');
  });

  it('ダークモード時にライトボタンを表示する', () => {
    renderSidebar();
    // 省略ラベル「ライト」が表示される
    expect(screen.getByText('ライト')).toBeInTheDocument();
  });

  it('ライトモード時にダークボタンを表示する', () => {
    mockUseTheme.mockReturnValue({
      theme: 'light',
      toggleTheme: mockToggleTheme,
    });
    renderSidebar();
    expect(screen.getByText('ダーク')).toBeInTheDocument();
  });

  it('テーマ切り替えボタンクリックでtoggleThemeが呼ばれる', () => {
    renderSidebar();
    fireEvent.click(screen.getByTitle('ライトモード'));
    expect(mockToggleTheme).toHaveBeenCalledTimes(1);
  });
});
