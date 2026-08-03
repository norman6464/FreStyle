import { useNotification } from '../model/useNotification';
import EmptyState from '@/shared/ui/EmptyState';
import Loading from '@/shared/ui/Loading';
import { NotificationItem } from '@/entities/notification';
import { BellIcon, ExclamationTriangleIcon } from '@heroicons/react/24/outline';

export default function NotificationPage() {
  const { notifications, unreadCount, loading, error, markAsRead, markAllAsRead, refresh } =
    useNotification();

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
            <span className="ml-2 text-xs font-normal text-taupe-500">{unreadCount}件の未読</span>
          )}
        </h2>
        {unreadCount > 0 && (
          <button
            onClick={markAllAsRead}
            className="text-xs text-taupe-500 hover:text-taupe-600 transition-colors"
          >
            すべて既読にする
          </button>
        )}
      </div>

      {/* 取得に失敗したことは独立した帯で伝える。取得済みの通知は隠さない（FRESTYLE-94）。 */}
      {error && (
        <div
          role="alert"
          className="rounded-lg border border-surface-3 bg-surface-1 px-4 py-4 flex items-start gap-3"
        >
          <ExclamationTriangleIcon className="w-5 h-5 flex-shrink-0 text-[var(--color-text-muted)] mt-0.5" />
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-[var(--color-text-primary)]">{error}</p>
            <p className="mt-1 text-xs text-[var(--color-text-muted)]">
              通知が無いのではなく、読み込めていない状態です。
            </p>
          </div>
          <button
            type="button"
            onClick={refresh}
            className="flex-shrink-0 inline-flex items-center justify-center px-3 py-1.5 rounded-md border border-surface-3 text-xs font-medium text-[var(--color-text-primary)] hover:bg-[var(--color-nav-hover)] transition-colors"
          >
            再試行
          </button>
        </div>
      )}

      {notifications.length > 0 ? (
        <div className="space-y-2">
          {notifications.map((notification) => (
            <NotificationItem
              key={notification.id}
              notification={notification}
              onMarkAsRead={markAsRead}
            />
          ))}
        </div>
      ) : (
        // 取得に失敗しているときは「0 件」と断定できないので空状態を出さない。
        !error && (
          <EmptyState
            icon={BellIcon}
            title="通知はありません"
            description="お知らせが届くとここに表示されます"
          />
        )
      )}
    </div>
  );
}
