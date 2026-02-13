import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import SecondaryPanel from '../components/layout/SecondaryPanel';
import EmptyState from '../components/EmptyState';
import SearchBox from '../components/SearchBox';
import { ChatBubbleLeftRightIcon } from '@heroicons/react/24/outline';
import SockJS from 'sockjs-client';
import { Client } from '@stomp/stompjs';
import { useChatList } from '../hooks/useChatList';

export default function ChatListPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const stompClientRef = useRef<Client | null>(null);
  const navigate = useNavigate();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const { chatUsers, loading, userId, fetchChatUsers, updateUnreadCount } = useChatList();

  // リアルタイム未読数更新のWebSocket購読
  useEffect(() => {
    if (!userId) return;

    const client = new Client({
      webSocketFactory: () =>
        new SockJS(`${API_BASE_URL}/ws/chat`, undefined, { withCredentials: true }),
      reconnectDelay: 5000,
      onConnect: () => {
        client.subscribe(`/topic/unread/${userId}`, (message) => {
          const data = JSON.parse(message.body);
          if (data.type === 'unread_update') {
            updateUnreadCount(data.roomId, data.increment);
          }
        });
      },
    });

    stompClientRef.current = client;
    client.activate();
    return () => client.deactivate();
  }, [userId, updateUnreadCount, API_BASE_URL]);

  // 検索クエリ変更時にデバウンス検索
  useEffect(() => {
    const timer = setTimeout(() => {
      fetchChatUsers(searchQuery || undefined);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery, fetchChatUsers]);

  const formatTime = (dateString: string): string => {
    if (!dateString) return '';
    const date = new Date(dateString);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const oneDay = 24 * 60 * 60 * 1000;
    const oneWeek = 7 * oneDay;

    if (diff < oneDay && date.getDate() === now.getDate()) {
      return date.toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' });
    } else if (diff < oneDay * 2 && date.getDate() === now.getDate() - 1) {
      return '昨日';
    } else if (diff < oneWeek) {
      const days = ['日', '月', '火', '水', '木', '金', '土'];
      return days[date.getDay()] + '曜日';
    } else {
      return date.toLocaleDateString('ja-JP', { month: 'short', day: 'numeric' });
    }
  };

  const truncateMessage = (message: string | undefined, maxLength = 30): string => {
    if (!message) return 'メッセージはありません';
    if (message.length <= maxLength) return message;
    return message.substring(0, maxLength) + '...';
  };

  return (
    <div className="flex h-full">
      <SecondaryPanel
        title="チャット"
        headerContent={
          <SearchBox
            value={searchQuery}
            onChange={setSearchQuery}
            placeholder="名前やメールで検索..."
          />
        }
      >
        {loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary-500" />
          </div>
        ) : chatUsers.length === 0 ? (
          <div className="p-4 text-center text-sm text-slate-500">
            チャット履歴がありません
          </div>
        ) : (
          chatUsers.map((user) => (
            <button
              key={user.roomId}
              onClick={() => navigate(`/chat/users/${user.roomId}`)}
              className="w-full px-4 py-3 flex items-center gap-3 hover:bg-primary-50 transition-colors border-b border-slate-100 text-left"
            >
              <div className="flex-shrink-0">
                {user.profileImage ? (
                  <img
                    src={user.profileImage}
                    alt={user.name}
                    className="w-10 h-10 rounded-full object-cover"
                  />
                ) : (
                  <div className="w-10 h-10 rounded-full bg-primary-500 flex items-center justify-center">
                    <span className="text-white text-sm font-bold">
                      {user.name?.charAt(0)?.toUpperCase() || '?'}
                    </span>
                  </div>
                )}
              </div>

              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-semibold text-slate-800 truncate">
                    {user.name || 'Unknown User'}
                  </span>
                  <span className="text-[11px] text-slate-400 flex-shrink-0 ml-2">
                    {formatTime(user.lastMessageAt!)}
                  </span>
                </div>
                <div className="flex items-center justify-between mt-0.5">
                  <p className="text-xs text-slate-500 truncate pr-2">
                    {user.lastMessageSenderId === user.userId
                      ? truncateMessage(user.lastMessage)
                      : `あなた: ${truncateMessage(user.lastMessage)}`}
                  </p>
                  {user.unreadCount > 0 && (
                    <span className="bg-rose-500 text-white text-[10px] font-bold px-1.5 py-0.5 rounded-full flex-shrink-0">
                      {user.unreadCount > 99 ? '99+' : user.unreadCount}
                    </span>
                  )}
                </div>
              </div>
            </button>
          ))
        )}
      </SecondaryPanel>

      <div className="flex-1">
        <EmptyState
          icon={ChatBubbleLeftRightIcon}
          title="チャットを選択してください"
          description="左のリストからチャット相手を選択するか、新しいチャットを始めましょう。"
          action={{ label: 'ユーザーを検索', onClick: () => navigate('/chat/users') }}
        />
      </div>
    </div>
  );
}
