import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../lib/axios', () => ({
  default: {
    get: vi.fn(),
    put: vi.fn(),
  },
}));

import apiClient from '../../lib/axios';
import { NotificationRepository } from '../NotificationRepository';

const mockedGet = vi.mocked(apiClient.get);
const mockedPut = vi.mocked(apiClient.put);

describe('NotificationRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('getAll: 通知一覧を取得する', async () => {
    const notifications = [
      { id: 1, type: 'GOAL_ACHIEVED', title: 'テスト', message: '本文', isRead: false, createdAt: '2024-01-01' },
    ];
    mockedGet.mockResolvedValue({ data: notifications });

    const result = await NotificationRepository.getAll();
    expect(result).toEqual(notifications);
    expect(mockedGet).toHaveBeenCalledWith('/api/notifications');
  });

  it('markAsRead: 指定IDの通知を既読にする', async () => {
    mockedPut.mockResolvedValue({});

    await NotificationRepository.markAsRead(5);
    expect(mockedPut).toHaveBeenCalledWith('/api/notifications/5/read');
  });

  it('markAllAsRead: 全通知を既読にする', async () => {
    mockedPut.mockResolvedValue({});

    await NotificationRepository.markAllAsRead();
    expect(mockedPut).toHaveBeenCalledWith('/api/notifications/read-all');
  });

  it('getUnreadCount: 未読数を取得する', async () => {
    mockedGet.mockResolvedValue({ data: 3 });

    const result = await NotificationRepository.getUnreadCount();
    expect(result).toBe(3);
    expect(mockedGet).toHaveBeenCalledWith('/api/notifications/unread-count');
  });
});
