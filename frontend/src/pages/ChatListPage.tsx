import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { clearAuth } from '../store/authSlice';
import HamburgerMenu from '../components/HamburgerMenu';
import {
  MagnifyingGlassIcon,
  ChatBubbleLeftRightIcon,
  SparklesIcon,
} from '@heroicons/react/24/solid';
import { ChevronRightIcon } from '@heroicons/react/24/outline';
import SockJS from 'sockjs-client';
import { Client } from '@stomp/stompjs';
import { ChatUser } from '../types';

export default function ChatListPage() {
  const [chatUsers, setChatUsers] = useState<ChatUser[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [userId, setUserId] = useState<number | null>(null);
  const stompClientRef = useRef<Client | null>(null);
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  // チャット履歴のあるユーザー一覧取得
  const fetchChatUsers = async (query = ''): Promise<void> => {
    try {
      setLoading(true);
      const url = query
        ? `${API_BASE_URL}/api/chat/rooms?query=${encodeURIComponent(query)}`
        : `${API_BASE_URL}/api/chat/rooms`;

      const res = await fetch(url, {
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
      });

      if (res.status === 401) {
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          { method: 'POST', credentials: 'include' }
        );
        if (!refreshRes.ok) {
          dispatch(clearAuth());
          navigate('/login');
          return;
        }
        return fetchChatUsers(query);
      }

      if (!res.ok) {
        console.error('チャットユーザー取得エラー:', res.status);
        return;
      }

      const data = await res.json();
      setChatUsers(data.chatUsers || []);
    } catch (e) {
      console.error('チャットユーザー取得失敗', e);
    } finally {
      setLoading(false);
    }
  };

  // ユーザーID取得
  useEffect(() => {
    const fetchUserId = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/auth/cognito/me`, {
          credentials: 'include',
        });
        if (res.ok) {
          const data = await res.json();
          setUserId(data.id);
        }
      } catch (e) {
        console.error('ユーザー情報取得エラー:', e);
      }
    };
    fetchUserId();
  }, []);

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
            setChatUsers((prev) =>
              prev.map((u) =>
                u.roomId === data.roomId
                  ? { ...u, unreadCount: (u.unreadCount || 0) + data.increment }
                  : u
              )
            );
          }
        });
      },
    });

    stompClientRef.current = client;
    client.activate();
    return () => client.deactivate();
  }, [userId]);

  // 初回ロード
  useEffect(() => {
    fetchChatUsers();
  }, []);

  // 検索クエリ変更時にデバウンス検索
  useEffect(() => {
    const timer = setTimeout(() => {
      fetchChatUsers(searchQuery);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  // 最終メッセージの時間をフォーマット
  const formatTime = (dateString: string): string => {
    if (!dateString) return '';
    const date = new Date(dateString);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const oneDay = 24 * 60 * 60 * 1000;
    const oneWeek = 7 * oneDay;

    if (diff < oneDay && date.getDate() === now.getDate()) {
      // 今日
      return date.toLocaleTimeString('ja-JP', {
        hour: '2-digit',
        minute: '2-digit',
      });
    } else if (diff < oneDay * 2 && date.getDate() === now.getDate() - 1) {
      // 昨日
      return '昨日';
    } else if (diff < oneWeek) {
      // 1週間以内
      const days = ['日', '月', '火', '水', '木', '金', '土'];
      return days[date.getDay()] + '曜日';
    } else {
      // それ以外
      return date.toLocaleDateString('ja-JP', {
        month: 'short',
        day: 'numeric',
      });
    }
  };

  // メッセージを短縮表示
  const truncateMessage = (message: string | undefined, maxLength = 30): string => {
    if (!message) return 'メッセージはありません';
    if (message.length <= maxLength) return message;
    return message.substring(0, maxLength) + '...';
  };

  return (
    <>
      <HamburgerMenu title="チャット" />
      <div className="min-h-screen bg-gray-50 pt-16 pb-24">
        {/* ヘッダーセクション */}
        <div className="sticky top-16 z-10 bg-white border-b border-gray-200 px-4 py-4">
          <div className="max-w-2xl mx-auto">
            {/* 検索ボックス */}
            <div className="relative">
              <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearchQuery(e.target.value)}
                placeholder="名前やメールで検索..."
                className="w-full pl-10 pr-4 py-3 bg-gray-100 rounded-xl border-none focus:outline-none focus:ring-1 focus:ring-primary-500 transition-colors duration-150"
              />
            </div>
          </div>
        </div>

        {/* メインコンテンツ */}
        <div className="max-w-2xl mx-auto px-4 pt-4">
          {/* ユーザーがいない場合 */}
          {loading ? (
            <div className="flex flex-col items-center justify-center py-16">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500 mb-4"></div>
              <p className="text-gray-500">読み込み中...</p>
            </div>
          ) : chatUsers.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <div className="bg-primary-100 rounded-full p-6 mb-6">
                <ChatBubbleLeftRightIcon className="w-16 h-16 text-primary-400" />
              </div>
              <h2 className="text-xl font-bold text-gray-700 mb-2">
                まだチャットがありません
              </h2>
              <p className="text-gray-500 mb-6 max-w-xs">
                新しい友達を追加して、チャットを始めましょう！
              </p>
              <button
                onClick={() => navigate('/chat/users')}
                className="bg-primary-500 text-white font-medium px-6 py-3 rounded-xl hover:bg-primary-600 transition-colors duration-150"
              >
                友達を追加する
              </button>
            </div>
          ) : (
            <>
              {/* チャット一覧 */}
              <div className="space-y-2">
                {chatUsers.map((user) => (
                  <div
                    key={user.roomId}
                    onClick={() => navigate(`/chat/users/${user.roomId}`)}
                    className="bg-white rounded-2xl cursor-pointer overflow-hidden group border border-gray-200 hover:bg-gray-50 transition-colors duration-150"
                  >
                    <div className="flex items-center p-4">
                      {/* アバター */}
                      <div className="relative flex-shrink-0">
                        {user.profileImage ? (
                          <img
                            src={user.profileImage}
                            alt={user.name}
                            className="w-14 h-14 rounded-full object-cover border border-gray-200"
                          />
                        ) : (
                          <div className="w-14 h-14 rounded-full bg-primary-500 flex items-center justify-center">
                            <span className="text-white text-xl font-bold">
                              {user.name?.charAt(0)?.toUpperCase() || '?'}
                            </span>
                          </div>
                        )}
                        {/* オンラインインジケーター（将来的に） */}
                        {/* <div className="absolute bottom-0 right-0 w-4 h-4 bg-green-500 border-2 border-white rounded-full"></div> */}
                      </div>

                      {/* ユーザー情報 */}
                      <div className="flex-1 ml-4 min-w-0">
                        <div className="flex items-center justify-between mb-1">
                          <h3 className="text-base font-bold text-gray-800 truncate">
                            {user.name || 'Unknown User'}
                          </h3>
                          <span className="text-xs text-gray-400 flex-shrink-0 ml-2">
                            {formatTime(user.lastMessageAt)}
                          </span>
                        </div>
                        <div className="flex items-center justify-between">
                          <p className="text-sm text-gray-500 truncate pr-2">
                            {user.lastMessageSenderId === user.userId
                              ? truncateMessage(user.lastMessage)
                              : `あなた: ${truncateMessage(user.lastMessage)}`}
                          </p>
                          {user.unreadCount > 0 && (
                            <span className="bg-pink-500 text-white text-xs font-bold px-2 py-0.5 rounded-full flex-shrink-0">
                              {user.unreadCount > 99 ? '99+' : user.unreadCount}
                            </span>
                          )}
                        </div>
                      </div>

                      {/* 矢印 */}
                      <ChevronRightIcon className="w-5 h-5 text-gray-300 ml-2 flex-shrink-0" />
                    </div>
                  </div>
                ))}
              </div>

              {/* AI分析への導線 */}
              <div className="mt-8 mb-4">
                <div
                  onClick={() => navigate('/chat/ask-ai')}
                  className="bg-white rounded-2xl p-5 cursor-pointer border border-gray-200 hover:bg-gray-50 transition-colors duration-150"
                >
                  <div className="flex items-center">
                    <div className="bg-primary-100 rounded-xl p-3 mr-4">
                      <SparklesIcon className="w-6 h-6 text-primary-500" />
                    </div>
                    <div className="flex-1">
                      <h3 className="text-gray-800 font-bold text-lg">
                        AIにチャットを分析してもらう
                      </h3>
                      <p className="text-gray-500 text-sm">
                        印象のギャップを発見しよう
                      </p>
                    </div>
                    <ChevronRightIcon className="w-5 h-5 text-gray-400" />
                  </div>
                </div>
              </div>
            </>
          )}
        </div>

        {/* フローティングアクションボタン */}
        <div className="fixed bottom-6 right-6 z-20">
          <button
            onClick={() => navigate('/chat/users')}
            className="bg-primary-500 text-white p-4 rounded-full shadow-md hover:bg-primary-600 transition-colors duration-150"
            title="新しいチャット"
          >
            <svg
              className="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 4v16m8-8H4"
              />
            </svg>
          </button>
        </div>
      </div>
    </>
  );
}
