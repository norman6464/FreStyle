import { useNotification } from '../hooks/useNotification';
import EmptyState from '../components/EmptyState';
import Loading from '../components/Loading';
import NotificationItem from '../components/NotificationItem';
import { BellIcon } from '@heroicons/react/24/outline';

export default function NotificationPage() {
  const { notifications, unreadCount, loading, markAsRead, markAllAsRead } = useNotification();

  if (loading) {
    return (
      <div className="max-w-3xl mx-auto p-4">
        <Loading size="medium" message="通知を読み込み中..." className="py-12" />
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-3">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">
          通知
          {unreadCount > 0 && (
            <span className="ml-2 text-xs font-normal text-primary-500">{unreadCount}件の未読</span>
          )}
        </h2>
        {unreadCount > 0 && (
          <button
            onClick={markAllAsRead}
            className="text-xs text-primary-500 hover:text-primary-600 transition-colors"
          >
            すべて既読にする
          </button>
        )}
      </div>

      {notifications.length === 0 ? (
        <EmptyState
          icon={BellIcon}
          title="通知はありません"
          description="新しいメッセージや目標達成時に通知が届きます"
        />
      ) : (
        <div className="space-y-2">
          {notifications.map((notification) => (
            <NotificationItem
              key={notification.id}
              notification={notification}
              onMarkAsRead={markAsRead}
            />
          ))}
        </div>
      )}
    </div>
  );
}
