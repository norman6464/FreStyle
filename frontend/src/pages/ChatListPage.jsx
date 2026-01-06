import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { clearAuth } from '../store/authSlice';
import HamburgerMenu from '../components/HamburgerMenu';
import {
  MagnifyingGlassIcon,
  ChatBubbleLeftRightIcon,
  SparklesIcon,
  UserCircleIcon,
} from '@heroicons/react/24/solid';
import { ChevronRightIcon } from '@heroicons/react/24/outline';

export default function ChatListPage() {
  const [chatUsers, setChatUsers] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  // チャット履歴のあるユーザー一覧取得
  const fetchChatUsers = async (query = '') => {
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
  const formatTime = (dateString) => {
    if (!dateString) return '';
    const date = new Date(dateString);
    const now = new Date();
    const diff = now - date;
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
  const truncateMessage = (message, maxLength = 30) => {
    if (!message) return 'メッセージはありません';
    if (message.length <= maxLength) return message;
    return message.substring(0, maxLength) + '...';
  };

  return (
    <>
      <HamburgerMenu title="チャット" />
      <div className="min-h-screen bg-gradient-to-br from-gray-50 to-primary-50 pt-16 pb-24">
        {/* ヘッダーセクション */}
        <div className="sticky top-16 z-10 bg-white/80 backdrop-blur-md border-b border-gray-100 px-4 py-4">
          <div className="max-w-2xl mx-auto">
            {/* 検索ボックス */}
            <div className="relative">
              <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="名前やメールで検索..."
                className="w-full pl-10 pr-4 py-3 bg-gray-100 rounded-xl border-none focus:outline-none focus:ring-2 focus:ring-primary-400 transition-all"
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
              <div className="bg-gradient-to-br from-primary-100 to-secondary-100 rounded-full p-6 mb-6">
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
                className="bg-gradient-to-r from-primary-500 to-secondary-500 text-white font-semibold px-6 py-3 rounded-xl shadow-lg hover:shadow-xl transition-all transform hover:scale-105"
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
                    className="bg-white rounded-2xl shadow-sm hover:shadow-lg transition-all duration-300 cursor-pointer overflow-hidden group border border-gray-100 hover:border-primary-200"
                  >
                    <div className="flex items-center p-4">
                      {/* アバター */}
                      <div className="relative flex-shrink-0">
                        {user.profileImage ? (
                          <img
                            src={user.profileImage}
                            alt={user.name}
                            className="w-14 h-14 rounded-full object-cover border-2 border-gray-100 group-hover:border-primary-300 transition-colors"
                          />
                        ) : (
                          <div className="w-14 h-14 rounded-full bg-gradient-to-br from-primary-400 to-secondary-400 flex items-center justify-center border-2 border-gray-100 group-hover:border-primary-300 transition-colors">
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
                          <h3 className="text-base font-bold text-gray-800 truncate group-hover:text-primary-600 transition-colors">
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
                          {/* 未読バッジ（将来的に） */}
                          {/* {user.unreadCount > 0 && (
                            <span className="bg-pink-500 text-white text-xs font-bold px-2 py-0.5 rounded-full flex-shrink-0">
                              {user.unreadCount}
                            </span>
                          )} */}
                        </div>
                      </div>

                      {/* 矢印 */}
                      <ChevronRightIcon className="w-5 h-5 text-gray-300 group-hover:text-primary-400 transition-colors ml-2 flex-shrink-0" />
                    </div>
                  </div>
                ))}
              </div>

              {/* AI分析への導線 */}
              <div className="mt-8 mb-4">
                <div
                  onClick={() => navigate('/chat/ask-ai')}
                  className="relative bg-gradient-to-r from-pink-500 via-orange-400 to-yellow-400 rounded-2xl shadow-lg p-5 cursor-pointer hover:shadow-xl transition-all transform hover:scale-[1.02] overflow-hidden"
                >
                  {/* 背景装飾 */}
                  <div className="absolute top-0 left-0 w-full h-full opacity-20">
                    <div className="absolute top-2 left-6 text-xl">✨</div>
                    <div className="absolute bottom-2 right-8 text-lg">💬</div>
                  </div>
                  <div className="relative flex items-center">
                    <div className="bg-white/30 backdrop-blur-sm rounded-xl p-3 mr-4">
                      <SparklesIcon className="w-6 h-6 text-white" />
                    </div>
                    <div className="flex-1">
                      <h3 className="text-white font-bold text-lg">
                        AIにチャットを分析してもらう
                      </h3>
                      <p className="text-white/80 text-sm">
                        印象のギャップを発見しよう
                      </p>
                    </div>
                    <ChevronRightIcon className="w-5 h-5 text-white/70" />
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
            className="bg-gradient-to-r from-primary-500 to-secondary-500 text-white p-4 rounded-full shadow-lg hover:shadow-xl transition-all transform hover:scale-110"
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
