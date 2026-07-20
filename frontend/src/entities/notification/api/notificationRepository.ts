import apiClient from '@/shared/api/axios';
import { NOTIFICATIONS } from '@/shared/config/apiRoutes';
import type { Notification } from '../model/types';

// Go バックエンドは PATCH を REST 標準として提供する。フロントは PATCH に揃える。
export const NotificationRepository = {
  async getAll(): Promise<Notification[]> {
    const response = await apiClient.get<Notification[]>(NOTIFICATIONS.list);
    return response.data;
  },

  async markAsRead(notificationId: number): Promise<void> {
    await apiClient.patch(NOTIFICATIONS.read(notificationId));
  },

  async markAllAsRead(): Promise<void> {
    await apiClient.patch(NOTIFICATIONS.readAll);
  },

  async getUnreadCount(): Promise<number> {
    const response = await apiClient.get<number>(NOTIFICATIONS.unreadCount);
    return response.data;
  },
};
