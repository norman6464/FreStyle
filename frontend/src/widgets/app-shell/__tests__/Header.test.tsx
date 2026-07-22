import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import authReducer from '@/entities/user/model/authSlice';
import Header from '../ui/Header';

vi.mock('@/entities/user/api/profileRepository', () => ({
  default: {
    fetchProfile: vi.fn().mockResolvedValue({
      displayName: 'テスト太郎',
      avatarUrl: null,
      email: 't@example.com',
    }),
  },
}));

vi.mock('@/entities/notification/api/notificationRepository', () => ({
  NotificationRepository: {
    getUnreadCount: vi.fn().mockResolvedValue(3),
  },
}));

function renderHeader(authState: Record<string, unknown>) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: { auth: { isAuthenticated: true, loading: false, ...authState } },
  });
  return render(
    <Provider store={store}>
      <MemoryRouter initialEntries={['/']}>
        <Header />
      </MemoryRouter>
    </Provider>,
  );
}

describe('Header', () => {
  beforeEach(() => vi.clearAllMocks());

  it('テキストのナビ項目を表示する', () => {
    renderHeader({ role: 'trainee', aiChatEnabledForTrainees: true });
    expect(screen.getAllByText('コース').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('演習').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('ノート').length).toBeGreaterThanOrEqual(1);
  });

  it('通知ベルとハンバーガー(メニュー)を表示する', () => {
    renderHeader({ role: 'trainee', aiChatEnabledForTrainees: true });
    expect(screen.getByRole('link', { name: /通知/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /メニュー/ })).toBeInTheDocument();
  });

  it('未読件数のバッジを表示する', async () => {
    renderHeader({ role: 'trainee', aiChatEnabledForTrainees: true });
    await waitFor(() => expect(screen.getByText('3')).toBeInTheDocument());
  });

  it('AI が無効な trainee には AI を出さない', () => {
    renderHeader({ role: 'trainee', aiChatEnabledForTrainees: false });
    expect(screen.queryByText('AI')).not.toBeInTheDocument();
  });

  it('非 admin には管理ドロップダウンを出さない', () => {
    renderHeader({ role: 'trainee', isAdmin: false, aiChatEnabledForTrainees: true });
    expect(screen.queryByRole('button', { name: '管理' })).not.toBeInTheDocument();
  });

  it('admin には管理ドロップダウンを出し、クリックでサブ項目が開く', () => {
    renderHeader({ role: 'super_admin', isAdmin: true });
    const adminBtn = screen.getByRole('button', { name: /管理/ });
    expect(adminBtn).toBeInTheDocument();
    fireEvent.click(adminBtn);
    expect(screen.getByText('会社一覧')).toBeInTheDocument();
    expect(screen.getByText('監査ログ')).toBeInTheDocument();
  });

  it('super_admin には学習系ナビ(コース)を出さない', () => {
    renderHeader({ role: 'super_admin', isAdmin: true });
    expect(screen.queryByText('コース')).not.toBeInTheDocument();
    expect(screen.getByText('ホーム')).toBeInTheDocument();
  });

  it('ハンバーガーでモバイルメニューが開き、設定/ログアウトが出る', () => {
    renderHeader({ role: 'trainee', aiChatEnabledForTrainees: true });
    // 開く前はデスクトップ分のみ。
    expect(screen.getAllByText('コース').length).toBe(1);
    fireEvent.click(screen.getByRole('button', { name: /メニュー/ }));
    // モバイルメニュー分が増え、設定 / ログアウトも出る。
    expect(screen.getAllByText('コース').length).toBeGreaterThanOrEqual(2);
    expect(screen.getByText('設定')).toBeInTheDocument();
    expect(screen.getByText('ログアウト')).toBeInTheDocument();
  });

  it('ユーザーメニューを開くと設定/ログアウトが出る', async () => {
    renderHeader({ role: 'trainee', aiChatEnabledForTrainees: true });
    const userButton = await screen.findByText('テスト太郎');
    fireEvent.click(userButton);
    expect(screen.getByRole('button', { name: '設定' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ログアウト' })).toBeInTheDocument();
  });
});
