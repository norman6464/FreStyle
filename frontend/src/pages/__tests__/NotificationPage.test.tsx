import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import NotificationPage from '../NotificationPage';

const mockMarkAsRead = vi.fn();
const mockMarkAllAsRead = vi.fn();

vi.mock('../../hooks/useNotification', () => ({
  useNotification: vi.fn(),
}));

import { useNotification } from '../../hooks/useNotification';
const mockedUseNotification = vi.mocked(useNotification);

describe('NotificationPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('ローディング中はスピナーが表示される', () => {
    mockedUseNotification.mockReturnValue({
      notifications: [],
      unreadCount: 0,
      loading: true,
      markAsRead: mockMarkAsRead,
      markAllAsRead: mockMarkAllAsRead,
      refresh: vi.fn(),
    });

    render(<NotificationPage />);
    expect(screen.getByText('通知を読み込み中...')).toBeInTheDocument();
  });

  it('通知がない場合はEmptyStateが表示される', () => {
    mockedUseNotification.mockReturnValue({
      notifications: [],
      unreadCount: 0,
      loading: false,
      markAsRead: mockMarkAsRead,
      markAllAsRead: mockMarkAllAsRead,
      refresh: vi.fn(),
    });

    render(<NotificationPage />);
    expect(screen.getByText('通知はありません')).toBeInTheDocument();
  });

  it('通知一覧が表示される', () => {
    mockedUseNotification.mockReturnValue({
      notifications: [
        { id: 1, type: 'GOAL_ACHIEVED', title: '月間目標を達成しました', message: 'おめでとう', isRead: false, createdAt: '2024-06-15T10:00:00Z' },
        { id: 2, type: 'SYSTEM', title: 'システム通知', message: 'メンテナンス', isRead: true, createdAt: '2024-06-14T10:00:00Z' },
      ],
      unreadCount: 1,
      loading: false,
      markAsRead: mockMarkAsRead,
      markAllAsRead: mockMarkAllAsRead,
      refresh: vi.fn(),
    });

    render(<NotificationPage />);
    expect(screen.getByText('月間目標を達成しました')).toBeInTheDocument();
    expect(screen.getByText('システム通知')).toBeInTheDocument();
    expect(screen.getByText('1件の未読')).toBeInTheDocument();
  });

  it('未読がある場合「すべて既読にする」ボタンが表示される', () => {
    mockedUseNotification.mockReturnValue({
      notifications: [
        { id: 1, type: 'GOAL_ACHIEVED', title: '目標達成', message: 'テスト', isRead: false, createdAt: '2024-06-15T10:00:00Z' },
      ],
      unreadCount: 1,
      loading: false,
      markAsRead: mockMarkAsRead,
      markAllAsRead: mockMarkAllAsRead,
      refresh: vi.fn(),
    });

    render(<NotificationPage />);
    const btn = screen.getByText('すべて既読にする');
    fireEvent.click(btn);
    expect(mockMarkAllAsRead).toHaveBeenCalled();
  });

  it('未読が0件の場合「すべて既読にする」ボタンが非表示', () => {
    mockedUseNotification.mockReturnValue({
      notifications: [
        { id: 1, type: 'SYSTEM', title: '通知', message: 'テスト', isRead: true, createdAt: '2024-06-15T10:00:00Z' },
      ],
      unreadCount: 0,
      loading: false,
      markAsRead: mockMarkAsRead,
      markAllAsRead: mockMarkAllAsRead,
      refresh: vi.fn(),
    });

    render(<NotificationPage />);
    expect(screen.queryByText('すべて既読にする')).not.toBeInTheDocument();
  });
});
