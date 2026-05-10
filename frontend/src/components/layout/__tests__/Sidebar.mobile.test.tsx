import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import authReducer from '../../../store/authSlice';
import Sidebar from '../Sidebar';

const mockHandleLogout = vi.fn();
vi.mock('../../../hooks/useSidebar', () => ({
  useSidebar: () => ({
    handleLogout: mockHandleLogout,
    loggingOut: false,
  }),
}));

vi.mock('../../../repositories/ProfileRepository', () => ({
  default: {
    fetchProfile: vi.fn().mockResolvedValue({
      userId: 1, displayName: 'テストユーザー', bio: '', avatarUrl: '', status: '',
    }),
    updateProfile: vi.fn(),
  },
}));

function createTestStore() {
  return configureStore({
    reducer: { auth: authReducer },
    preloadedState: { auth: { isAuthenticated: true, loading: false } },
  });
}

function renderSidebar(onNavigate?: () => void) {
  return render(
    <Provider store={createTestStore()}>
      <MemoryRouter initialEntries={['/']}>
        <Sidebar onNavigate={onNavigate} />
      </MemoryRouter>
    </Provider>
  );
}

describe('Sidebar モバイル動作', () => {
  it('ナビ項目クリック時にonNavigateが呼ばれる', () => {
    const onNavigate = vi.fn();
    renderSidebar(onNavigate);
    fireEvent.click(screen.getByText('AI'));
    expect(onNavigate).toHaveBeenCalledTimes(1);
  });

  it('ユーザーメニューからログアウトクリック時にonNavigateが呼ばれる', () => {
    const onNavigate = vi.fn();
    renderSidebar(onNavigate);
    // ユーザーボタンを押して dropdown を開く
    const userButton = screen.getAllByRole('button').find((b) => b.getAttribute('aria-haspopup') === 'menu');
    expect(userButton).toBeDefined();
    fireEvent.click(userButton!);
    fireEvent.click(screen.getByRole('menuitem', { name: /ログアウト/ }));
    expect(onNavigate).toHaveBeenCalledTimes(1);
  });

  it('onNavigateが未指定でもエラーにならない', () => {
    renderSidebar();
    expect(() => fireEvent.click(screen.getByText('AI'))).not.toThrow();
  });
});
