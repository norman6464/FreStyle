import { useState, useCallback, useEffect } from 'react';
import { NotificationRepository } from '../repositories/NotificationRepository';
import type { Notification } from '../types';

export function useNotification() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    try {
      const [notifs, count] = await Promise.all([
        NotificationRepository.getAll(),
        NotificationRepository.getUnreadCount(),
      ]);
      setNotifications(notifs);
      setUnreadCount(count);
    } catch {
      setNotifications([]);
      setUnreadCount(0);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const markAsRead = useCallback(async (notificationId: number) => {
    await NotificationRepository.markAsRead(notificationId);
    await fetchData();
  }, [fetchData]);

  const markAllAsRead = useCallback(async () => {
    await NotificationRepository.markAllAsRead();
    await fetchData();
  }, [fetchData]);

  return { notifications, unreadCount, loading, markAsRead, markAllAsRead, refresh: fetchData };
}
