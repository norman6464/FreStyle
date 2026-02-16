import apiClient from '../lib/axios';
import type { Notification } from '../types';

export const NotificationRepository = {
  async getAll(): Promise<Notification[]> {
    const response = await apiClient.get<Notification[]>('/api/notifications');
    return response.data;
  },

  async markAsRead(notificationId: number): Promise<void> {
    await apiClient.put(`/api/notifications/${notificationId}/read`);
  },

  async markAllAsRead(): Promise<void> {
    await apiClient.put('/api/notifications/read-all');
  },

  async getUnreadCount(): Promise<number> {
    const response = await apiClient.get<number>('/api/notifications/unread-count');
    return response.data;
  },
};
