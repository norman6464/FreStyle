import { memo } from 'react';

interface RoomListItemProps {
  name: string;
  lastMessage: string;
  unreadCount: number;
}

export default memo(function RoomListItem({ name, lastMessage, unreadCount }: RoomListItemProps) {
  return (
    <div className="p-4 rounded-xl hover:bg-surface-2 cursor-pointer flex justify-between items-center transition-colors duration-150 border border-surface-3">
      <div className="flex-1 min-w-0">
        <p className="font-semibold text-[var(--color-text-primary)]">{name}</p>
        <p className="text-sm text-[var(--color-text-muted)] truncate">{lastMessage}</p>
      </div>
      {unreadCount > 0 && (
        <span className="bg-primary-500 text-white text-xs px-3 py-1.5 rounded-full font-medium ml-3 flex-shrink-0">
          {unreadCount}
        </span>
      )}
    </div>
  );
});
