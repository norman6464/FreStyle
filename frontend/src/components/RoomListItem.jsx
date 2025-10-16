export default function RoomListItem({ name, lastMessage, unreadCount }) {
  return (
    <div className="p-3 rounded-lg hover:bg-gray-100 cursor-pointer flex justify-between items-center">
      <div>
        <p className="font-semibold">{name}</p>
        <p className="text-sm text-gray-500 truncate">{lastMessage}</p>
      </div>
      {unreadCount > 0 && (
        <span className="bg-blue-500 text-white text-xs px-2 py-1 rounded-full">
          {unreadCount}
        </span>
      )}
    </div>
  );
}
