import apiClient from '../lib/axios';
import type { Notification } from '../types';

export const NotificationRepository = {
  async getAll(): Promise<Notification[]> {
    const response = await apiClient.get<Notification[]>('/api/v2/notifications');
    return response.data;
  },

  async markAsRead(notificationId: number): Promise<void> {
    await apiClient.put(`/api/v2/notifications/${notificationId}/read`);
  },

  async markAllAsRead(): Promise<void> {
    await apiClient.put('/api/v2/notifications/read-all');
  },

  async getUnreadCount(): Promise<number> {
    const response = await apiClient.get<number>('/api/v2/notifications/unread-count');
    return response.data;
  },
};
