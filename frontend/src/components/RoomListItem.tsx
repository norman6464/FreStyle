interface RoomListItemProps {
  name: string;
  lastMessage: string;
  unreadCount: number;
}

export default function RoomListItem({ name, lastMessage, unreadCount }: RoomListItemProps) {
  return (
    <div className="p-4 rounded-xl hover:bg-slate-50 cursor-pointer flex justify-between items-center transition-colors duration-150 border border-slate-200">
      <div className="flex-1 min-w-0">
        <p className="font-semibold text-slate-800">{name}</p>
        <p className="text-sm text-slate-500 truncate">{lastMessage}</p>
      </div>
      {unreadCount > 0 && (
        <span className="bg-primary-500 text-white text-xs px-3 py-1.5 rounded-full font-medium ml-3 flex-shrink-0">
          {unreadCount}
        </span>
      )}
    </div>
  );
}
