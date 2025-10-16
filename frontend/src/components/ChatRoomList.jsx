import RoomListItem from './RoomListItem';

const mockRooms = [
  {
    id: 1,
    name: 'サポートBot',
    lastMessage: 'どうされましたか？',
    unreadCount: 2,
  },
  {
    id: 2,
    name: '開発チーム',
    lastMessage: 'ミーティングは10時から',
    unreadCount: 0,
  },
];

export default function ChatRoomList() {
  return (
    <div className="p-4 space-y-2">
      {mockRooms.map((room) => (
        <RoomListItem key={room.id} {...room} />
      ))}
    </div>
  );
}
