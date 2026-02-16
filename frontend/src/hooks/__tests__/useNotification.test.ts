import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useNotification } from '../useNotification';

const mockGetAll = vi.fn();
const mockMarkAsRead = vi.fn();
const mockMarkAllAsRead = vi.fn();
const mockGetUnreadCount = vi.fn();

vi.mock('../../repositories/NotificationRepository', () => ({
  NotificationRepository: {
    getAll: (...args: unknown[]) => mockGetAll(...args),
    markAsRead: (...args: unknown[]) => mockMarkAsRead(...args),
    markAllAsRead: (...args: unknown[]) => mockMarkAllAsRead(...args),
    getUnreadCount: (...args: unknown[]) => mockGetUnreadCount(...args),
  },
}));

const mockNotifications = [
  { id: 1, type: 'NEW_MESSAGE', title: '新しいメッセージ', message: '田中さんから', isRead: false, relatedId: 5, createdAt: '2026-01-15T10:30:00' },
  { id: 2, type: 'GOAL_ACHIEVED', title: '目標達成', message: '日次目標を達成しました', isRead: true, createdAt: '2026-01-15T09:00:00' },
];

describe('useNotification', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetAll.mockResolvedValue(mockNotifications);
    mockGetUnreadCount.mockResolvedValue(1);
    mockMarkAsRead.mockResolvedValue(undefined);
    mockMarkAllAsRead.mockResolvedValue(undefined);
  });

  it('初期状態はloading=trueで空の通知リスト', () => {
    const { result } = renderHook(() => useNotification());
    expect(result.current.loading).toBe(true);
    expect(result.current.notifications).toEqual([]);
  });

  it('マウント時にAPIから通知一覧と未読数を取得する', async () => {
    const { result } = renderHook(() => useNotification());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.notifications).toHaveLength(2);
    expect(result.current.unreadCount).toBe(1);
    expect(mockGetAll).toHaveBeenCalled();
    expect(mockGetUnreadCount).toHaveBeenCalled();
  });

  it('通知を既読にできる', async () => {
    mockGetAll
      .mockResolvedValueOnce(mockNotifications)
      .mockResolvedValueOnce(mockNotifications.map(n => n.id === 1 ? { ...n, isRead: true } : n));
    mockGetUnreadCount
      .mockResolvedValueOnce(1)
      .mockResolvedValueOnce(0);

    const { result } = renderHook(() => useNotification());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.markAsRead(1);
    });

    expect(mockMarkAsRead).toHaveBeenCalledWith(1);
  });

  it('全通知を既読にできる', async () => {
    mockGetAll
      .mockResolvedValueOnce(mockNotifications)
      .mockResolvedValueOnce(mockNotifications.map(n => ({ ...n, isRead: true })));
    mockGetUnreadCount
      .mockResolvedValueOnce(1)
      .mockResolvedValueOnce(0);

    const { result } = renderHook(() => useNotification());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.markAllAsRead();
    });

    expect(mockMarkAllAsRead).toHaveBeenCalled();
  });

  it('API失敗時はエラーなく空リストを返す', async () => {
    mockGetAll.mockRejectedValue(new Error('API Error'));
    mockGetUnreadCount.mockRejectedValue(new Error('API Error'));

    const { result } = renderHook(() => useNotification());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.notifications).toEqual([]);
    expect(result.current.unreadCount).toBe(0);
  });
});
