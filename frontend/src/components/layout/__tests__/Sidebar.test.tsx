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
      totalUnread: 0,
      handleLogout: vi.fn(),
    });
    mockUseTheme.mockReturnValue({
      theme: 'dark',
      toggleTheme: mockToggleTheme,
    });
  });

  it('全ナビゲーション項目を表示する', () => {
    renderSidebar();
    expect(screen.getByText('ホーム')).toBeDefined();
    expect(screen.getByText('チャット')).toBeDefined();
    expect(screen.getByText('AI')).toBeDefined();
    expect(screen.getByText('練習')).toBeDefined();
    expect(screen.getByText('ユーザー検索')).toBeDefined();
    expect(screen.getByText('プロフィール')).toBeDefined();
    expect(screen.getByText('パーソナリティ')).toBeDefined();
  });

  it('ログアウトボタンを表示する', () => {
    renderSidebar();
    expect(screen.getByText('ログアウト')).toBeDefined();
  });

  it('FreStyleロゴ画像を角丸で表示する', () => {
    renderSidebar();
    const logo = screen.getByAltText('FreStyle');
    expect(logo).toBeDefined();
    expect(logo.getAttribute('src')).toBe('/image.png');
    expect(logo.className).toContain('rounded-xl');
  });

  it('FreStyleテキストがロゴ横に表示される', () => {
    renderSidebar();
    expect(screen.getByText('FreStyle')).toBeInTheDocument();
  });

  it('ホームルートでホームがアクティブになる', () => {
    renderSidebar('/');
    const homeLink = screen.getByText('ホーム').closest('a');
    expect(homeLink?.className).toContain('bg-surface-3');
  });

  it('チャットルートでチャットがアクティブになる', () => {
    renderSidebar('/chat');
    const chatLink = screen.getByText('チャット').closest('a');
    expect(chatLink?.className).toContain('bg-surface-3');
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

  it('未読メッセージがある場合チャットにバッジを表示する', () => {
    mockUseSidebar.mockReturnValue({
      totalUnread: 5,
      handleLogout: vi.fn(),
    });

    renderSidebar();

    expect(screen.getByTestId('sidebar-badge')).toBeInTheDocument();
    expect(screen.getByTestId('sidebar-badge')).toHaveTextContent('5');
  });
});
