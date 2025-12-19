export default function RoomListItem({ name, lastMessage, unreadCount }) {
  return (
    <div className="p-4 rounded-xl hover:bg-gradient-to-r hover:from-primary-50 hover:to-secondary-50 cursor-pointer flex justify-between items-center transition-all duration-200 border border-gray-100 hover:border-primary-200 shadow-sm hover:shadow-md">
      <div className="flex-1 min-w-0">
        <p className="font-semibold text-gray-800">{name}</p>
        <p className="text-sm text-gray-500 truncate">{lastMessage}</p>
      </div>
      {unreadCount > 0 && (
        <span className="bg-gradient-primary text-white text-xs px-3 py-1.5 rounded-full font-bold ml-3 flex-shrink-0">
          {unreadCount}
        </span>
      )}
    </div>
  );
}
