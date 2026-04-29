import apiClient from '../lib/axios';
import { NOTIFICATIONS } from '../constants/apiRoutes';
import type { Notification } from '../types';

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
