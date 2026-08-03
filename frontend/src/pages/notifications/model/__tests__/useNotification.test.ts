import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useNotification } from '../useNotification';
import type { Notification } from '@/entities/notification';

const mockGetAll = vi.fn();
const mockMarkAsRead = vi.fn();
const mockMarkAllAsRead = vi.fn();
const mockGetUnreadCount = vi.fn();

vi.mock('@/entities/notification/api/notificationRepository', () => ({
  NotificationRepository: {
    getAll: (...args: unknown[]) => mockGetAll(...args),
    markAsRead: (...args: unknown[]) => mockMarkAsRead(...args),
    markAllAsRead: (...args: unknown[]) => mockMarkAllAsRead(...args),
    getUnreadCount: (...args: unknown[]) => mockGetUnreadCount(...args),
  },
}));

// backend が実際に作る形に合わせる（type=company_application / 本文は body）。
const mockNotifications: Notification[] = [
  {
    id: 1,
    type: 'company_application',
    title: '新しい利用申請が届きました',
    body: '株式会社サンプル（山田 太郎 / taro@example.com）から利用申請がありました。',
    isRead: false,
    createdAt: '2026-08-02T10:30:00Z',
  },
  {
    id: 2,
    type: 'company_application',
    title: '新しい利用申請が届きました',
    body: '株式会社テスト（鈴木 花子 / hanako@example.com）から利用申請がありました。',
    isRead: true,
    createdAt: '2026-08-02T09:00:00Z',
  },
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

  // 取得できなかったことを空配列で表すと「通知は 0 件」と区別がつかず、
  // 障害中に「通知はありません」という嘘を見せてしまう（FRESTYLE-94）。
  describe('取得に失敗したとき', () => {
    beforeEach(() => {
      mockGetAll.mockRejectedValue(new Error('API Error'));
      mockGetUnreadCount.mockRejectedValue(new Error('API Error'));
    });

    it('エラーを立てる（握りつぶさない）', async () => {
      const { result } = renderHook(() => useNotification());

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });
      expect(result.current.error).toBeTruthy();
    });

    it('直前に表示していた通知を空にしない', async () => {
      mockGetAll.mockReset().mockResolvedValueOnce(mockNotifications);
      mockGetUnreadCount.mockReset().mockResolvedValueOnce(1);

      const { result } = renderHook(() => useNotification());
      await waitFor(() => expect(result.current.notifications).toHaveLength(2));

      // 2 回目の取得だけ失敗させる
      mockGetAll.mockRejectedValue(new Error('API Error'));
      mockGetUnreadCount.mockRejectedValue(new Error('API Error'));
      await act(async () => {
        await result.current.refresh();
      });

      expect(result.current.error).toBeTruthy();
      expect(result.current.notifications).toHaveLength(2);
      expect(result.current.unreadCount).toBe(1);
    });

    it('再取得に成功したらエラーが消える', async () => {
      const { result } = renderHook(() => useNotification());
      await waitFor(() => expect(result.current.error).toBeTruthy());

      mockGetAll.mockResolvedValue(mockNotifications);
      mockGetUnreadCount.mockResolvedValue(1);
      await act(async () => {
        await result.current.refresh();
      });

      expect(result.current.error).toBeNull();
      expect(result.current.notifications).toHaveLength(2);
    });
  });
});
