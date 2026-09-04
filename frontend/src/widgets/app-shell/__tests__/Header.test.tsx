import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import authReducer from '@/entities/user/model/authSlice';
import { ToastProvider } from '@/app/providers/ToastProvider';
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

function renderHeader() {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: { auth: { isAuthenticated: true, loading: false } },
  });
  return render(
    <Provider store={store}>
      <ToastProvider>
        <MemoryRouter initialEntries={['/']}>
          <Header />
        </MemoryRouter>
      </ToastProvider>
    </Provider>,
  );
}

describe('Header', () => {
  beforeEach(() => vi.clearAllMocks());

  it('テキストのナビ項目を表示する', () => {
    renderHeader();
    expect(screen.getAllByText('コース').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('演習').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('ノート').length).toBeGreaterThanOrEqual(1);
  });

  it('通知ベルとハンバーガー(メニュー)を表示する', () => {
    renderHeader();
    expect(screen.getByRole('link', { name: /通知/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /メニュー/ })).toBeInTheDocument();
  });

  it('未読件数のバッジを表示する', async () => {
    renderHeader();
    await waitFor(() => expect(screen.getByText('3')).toBeInTheDocument());
  });

  it('AI のナビ項目は出ない（機能廃止の回帰）', () => {
    renderHeader();
    expect(screen.queryByText('AI')).not.toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'AI' })).not.toBeInTheDocument();
  });

  it('ハンバーガーでモバイルメニューが開き、設定/ログアウトが出る', () => {
    renderHeader();
    // 開く前はデスクトップ分のみ。
    expect(screen.getAllByText('コース').length).toBe(1);
    fireEvent.click(screen.getByRole('button', { name: /メニュー/ }));
    // モバイルメニュー分が増え、設定 / ログアウトも出る。
    expect(screen.getAllByText('コース').length).toBeGreaterThanOrEqual(2);
    expect(screen.getByText('設定')).toBeInTheDocument();
    expect(screen.getByText('ログアウト')).toBeInTheDocument();
  });

  it('ユーザーメニューを開くと設定/ログアウトが出る', async () => {
    renderHeader();
    const userButton = await screen.findByText('テスト太郎');
    fireEvent.click(userButton);
    expect(screen.getByRole('button', { name: '設定' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ログアウト' })).toBeInTheDocument();
  });
});
