import { memo } from 'react';
import { CheckIcon } from '@heroicons/react/24/outline';
import type { Notification } from '../types';

const TYPE_LABELS: Record<string, string> = {
  NEW_MESSAGE: 'メッセージ',
  GOAL_ACHIEVED: '目標達成',
  SCORE_IMPROVED: 'スコア向上',
  PRACTICE_REMINDER: '練習リマインダー',
  SYSTEM: 'システム',
};

interface NotificationItemProps {
  notification: Notification;
  onMarkAsRead: (id: number) => void;
}

export default memo(function NotificationItem({ notification, onMarkAsRead }: NotificationItemProps) {
  return (
    <div
      className={`p-4 rounded-lg border transition-colors ${
        notification.isRead
          ? 'bg-surface-1 border-surface-3'
          : 'bg-surface-2 border-primary-200'
      }`}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <span className="text-[10px] font-medium text-primary-400 bg-surface-2 px-2 py-0.5 rounded">
              {TYPE_LABELS[notification.type] ?? notification.type}
            </span>
            {!notification.isRead && (
              <span className="w-2 h-2 rounded-full bg-primary-500 flex-shrink-0" />
            )}
          </div>
          <p className="text-sm font-medium text-[var(--color-text-primary)] mb-0.5">{notification.title}</p>
          <p className="text-xs text-[var(--color-text-muted)]">{notification.message}</p>
          <p className="text-[10px] text-[var(--color-text-faint)] mt-1">
            {new Date(notification.createdAt).toLocaleString('ja-JP')}
          </p>
        </div>
        {!notification.isRead && (
          <button
            onClick={() => onMarkAsRead(notification.id)}
            aria-label="既読にする"
            className="ml-2 p-1 text-[var(--color-text-faint)] hover:text-primary-500 transition-colors"
          >
            <CheckIcon className="w-4 h-4" />
          </button>
        )}
      </div>
    </div>
  );
});
