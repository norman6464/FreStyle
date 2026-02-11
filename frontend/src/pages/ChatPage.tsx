import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import MessageBubble from '../components/MessageBubble';
import MessageInput from '../components/MessageInput';
import HamburgerMenu from '../components/HamburgerMenu';
import SearchBox from '../components/SearchBox';
import { ChatUser, ChatMessage } from '../types';

import ConfirmModal from '../components/ConfirmModal';
import SceneSelector from '../components/SceneSelector';
import RephraseModal from '../components/RephraseModal';
import { useDispatch } from 'react-redux';
import { clearAuth } from '../store/authSlice';

import SockJS from 'sockjs-client';
import { Client } from '@stomp/stompjs';

export default function ChatPage() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [senderId, setSenderId] = useState<number | null>(null);
  const [deleteModal, setDeleteModal] = useState<{ isOpen: boolean; messageId: number | null }>({ isOpen: false, messageId: null });
  const [chatUsers, setChatUsers] = useState<ChatUser[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [selectionMode, setSelectionMode] = useState(false);
  const [selectedMessages, setSelectedMessages] = useState<Set<number>>(new Set());
  const [rangeStart, setRangeStart] = useState<number | null>(null); // 範囲選択の開始点
  const [rangeEnd, setRangeEnd] = useState<number | null>(null);     // 範囲選択の終了点
  const [showSceneSelector, setShowSceneSelector] = useState(false); // シーンセレクター表示
  const [showRephraseModal, setShowRephraseModal] = useState(false);
  const [rephraseResult, setRephraseResult] = useState<{ formal: string; soft: string; concise: string } | null>(null);
  const stompClientRef = useRef<Client | null>(null);
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  const { roomId } = useParams<{ roomId: string }>();

  const navigate = useNavigate();
  const dispatch = useDispatch();

  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  // ユーザー情報取得（senderId を取得）
  useEffect(() => {
    const fetchUserInfo = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/auth/cognito/me`, {
          credentials: 'include',
        });
        if (!res.ok) {
          navigate('/login');
          return;
        }
        const data = await res.json();
        // このdata.idはリアルタイムで相手から送信してきた。それとも自分で送信をしたの判断をつけるためのフラグ
        setSenderId(data.id);
        console.log('[ChatPage] Fetched user info, senderId:', data.id);
      } catch (error) {
        console.error('ユーザー情報取得エラー:', error);
        navigate('/login');
      }
    };

    fetchUserInfo();
  }, [API_BASE_URL, navigate]);

  // --- チャット履歴のあるユーザー一覧取得 ---
  const fetchChatUsers = async (query = '') => {
    try {
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
    }
  };

  // 初回ロード時にチャットユーザー一覧を取得
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

  // スクロール
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // チャットルーム開封時に未読カウントをリセット
  useEffect(() => {
    if (!roomId || !senderId) return;

    const markAsRead = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/chat/rooms/${roomId}/read`, {
          method: 'POST',
          credentials: 'include',
        });
        if (res.status === 401) {
          const refreshRes = await fetch(
            `${API_BASE_URL}/api/auth/cognito/refresh-token`,
            { method: 'POST', credentials: 'include' }
          );
          if (!refreshRes.ok) {
            dispatch(clearAuth());
            return;
          }
          return markAsRead();
        }
      } catch (e) {
        console.error('既読リセットエラー:', e);
      }
    };
    markAsRead();

    // サイドバーの未読数もローカルでリセット
    setChatUsers((prev) =>
      prev.map((u) =>
        u.roomId === parseInt(roomId, 10) ? { ...u, unreadCount: 0 } : u
      )
    );
  }, [roomId, senderId]);

  // サイドバーの未読数リアルタイム更新（WebSocket購読）
  useEffect(() => {
    if (!senderId) return;

    const unreadClient = new Client({
      webSocketFactory: () =>
        new SockJS(`${API_BASE_URL}/ws/chat`, undefined, { withCredentials: true }),
      reconnectDelay: 5000,
      onConnect: () => {
        unreadClient.subscribe(`/topic/unread/${senderId}`, (message) => {
          const data = JSON.parse(message.body);
          if (data.type === 'unread_update') {
            // 現在開いているルームの場合は即座にリセット
            if (data.roomId === parseInt(roomId!, 10)) {
              fetch(`${API_BASE_URL}/api/chat/rooms/${roomId}/read`, {
                method: 'POST',
                credentials: 'include',
              }).catch(() => {});
              return;
            }
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

    unreadClient.activate();
    return () => unreadClient.deactivate();
  }, [senderId, roomId]);

  // --- チャット履歴取得 ---
  const fetchHistory = async () => {
    try {
      const res = await fetch(
        `${API_BASE_URL}/api/chat/users/${roomId}/history`,
        {
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
        }
      );

      if (res.status === 401) {
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          { method: 'POST', credentials: 'include' }
        );

        if (!refreshRes.ok) {
          dispatch(clearAuth());
          return;
        }

        return fetchHistory();
      }

      if (!res.ok) {
        console.error('チャット履歴取得エラー:', res.status, res.statusText);
        return;
      }

      const data = await res.json();

      // レスポンスが配列でない場合のチェック
      if (!Array.isArray(data)) {
        console.error('レスポンスが配列ではありません:', data);
        return;
      }

      const formatted = data.map((msg: any) => ({
        id: msg.id,
        roomId: msg.roomId,
        senderId: msg.senderId,
        senderName: msg.senderName,
        content: msg.content,
        createdAt: msg.createdAt,
        isSender: msg.senderId === senderId,
      }));

      setMessages(formatted);
    } catch (e) {
      console.error('履歴取得失敗', e);
    }
  };

  // --- WebSocket (STOMP) 接続 ---
  useEffect(() => {
    if (!senderId) return;

    const client = new Client({
      webSocketFactory: () =>
        new SockJS(`${API_BASE_URL}/ws/chat`, undefined, { withCredentials: true }),
      reconnectDelay: 5000,

      onConnect: () => {
        console.log('✅ STOMP connected');
        console.log('Connected status:', stompClientRef.current?.connected);

        // 認証メッセージを送信（接続時のみ）
        client.publish({
          destination: '/app/auth',
          body: JSON.stringify({
            userId: senderId,
          }),
        });
        console.log('📤 Auth message sent');

        // ルーム購読（相手ユーザーがリアルタイムでチャットをしてきたらそれを取得して表示をする）
        client.subscribe(`/topic/chat/${roomId}`, (message) => {
          const data = JSON.parse(message.body);
          console.log('📩 Received message from topic:', data);

          // 削除通知の処理
          if (data.type === 'delete') {
            console.log('🗑️ Delete notification received for messageId:', data.messageId);
            setMessages((prev) =>
              prev.map((m) =>
                m.id === data.messageId ? { ...m, isDeleted: true } : m
              )
            );
            return;
          }

          // バックエンドから返却された ChatMessageDto をそのまま使用
          // data.isSender は既にバックエンドで計算されている
          setMessages((prev) => [
            ...prev,
            {
              id: data.id,
              roomId: data.roomId,
              senderId: data.senderId,
              senderName: data.senderName,
              content: data.content,
              createdAt: data.createdAt,
              isSender: data.senderId === senderId,
            },
          ]);
        });

        fetchHistory();
      },

      onStompError: (frame) => {
        console.error('STOMP Error', frame);
      },
    });

    stompClientRef.current = client;
    client.activate();

    return () => {
      client.deactivate();
    };
  }, [roomId, senderId]);

  // --- メッセージ送信 ---
  const handleSend = (text: string) => {
    if (!stompClientRef.current?.connected) {
      console.warn('⚠️ STOMP not connected');
      return;
    }

    console.log('📤 Sending message:', { roomId, senderId, content: text });

    stompClientRef.current.publish({
      destination: '/app/chat/send',
      body: JSON.stringify({
        roomId,
        senderId,
        content: text,
      }),
    });

    // 💡 楽観的UI更新を削除：バックエンドからの返却を待つ
    // WebSocket経由でバックエンドが /topic/chat/{roomId} にメッセージをブロードキャストするので
    // 自動的にメッセージが画面に追加される
  };

  // --- メッセージ削除 ---
  const handleDeleteMessage = (messageId: number) => {
    setDeleteModal({ isOpen: true, messageId });
  };

  const confirmDelete = () => {
    const messageId = deleteModal.messageId;
    setDeleteModal({ isOpen: false, messageId: null });

    if (!stompClientRef.current?.connected) {
      console.warn('⚠️ STOMP not connected');
      return;
    }

    console.log('🗑️ Sending delete request for messageId:', messageId);

    // WebSocket経由でバックエンドに削除リクエストを送信
    // バックエンドが削除後、/topic/chat/{roomId} に削除通知をブロードキャストする
    stompClientRef.current.publish({
      destination: '/app/chat/delete',
      body: JSON.stringify({
        messageId,
        roomId: parseInt(roomId!, 10),
        senderId,
      }),
    });
  };

  const cancelDelete = () => {
    setDeleteModal({ isOpen: false, messageId: null });
  };

  // --- AIフィードバック ---
  const handleAiFeedback = () => {
    // 選択モードを開始
    setSelectionMode(true);
    setSelectedMessages(new Set());
    setRangeStart(null);
    setRangeEnd(null);
  };

  // 範囲選択: メッセージをクリックしたとき
  const handleRangeClick = (messageId: number) => {
    const messageIndex = messages.findIndex((msg) => msg.id === messageId);

    if (rangeStart === null) {
      // 開始点を設定
      setRangeStart(messageIndex);
      setRangeEnd(null);
      setSelectedMessages(new Set([messageId]));
    } else if (rangeEnd === null) {
      // 終了点を設定し、範囲内のメッセージを選択
      setRangeEnd(messageIndex);
      const start = Math.min(rangeStart, messageIndex);
      const end = Math.max(rangeStart, messageIndex);
      const rangeIds = new Set(
        messages.slice(start, end + 1).map((msg) => msg.id)
      );
      setSelectedMessages(rangeIds);
    } else {
      // 既に範囲が選択されている場合はリセットして新しい開始点を設定
      setRangeStart(messageIndex);
      setRangeEnd(null);
      setSelectedMessages(new Set([messageId]));
    }
  };

  // クイック選択: 直近N件
  const handleQuickSelect = (count: number) => {
    const recentMessages = messages.slice(-count);
    const recentIds = new Set(recentMessages.map((msg) => msg.id));
    setSelectedMessages(recentIds);
    // 範囲表示用に開始・終了を設定
    if (recentMessages.length > 0) {
      setRangeStart(messages.length - count);
      setRangeEnd(messages.length - 1);
    }
  };

  const handleSelectAll = () => {
    const allIds = new Set(messages.map((msg) => msg.id));
    setSelectedMessages(allIds);
    setRangeStart(0);
    setRangeEnd(messages.length - 1);
  };

  const handleDeselectAll = () => {
    setSelectedMessages(new Set());
    setRangeStart(null);
    setRangeEnd(null);
  };

  const handleCancelSelection = () => {
    setSelectionMode(false);
    setSelectedMessages(new Set());
    setRangeStart(null);
    setRangeEnd(null);
  };

  const handleSendToAi = () => {
    if (selectedMessages.size === 0) {
      alert('メッセージを選択してください');
      return;
    }

    // シーンセレクターを表示
    setShowSceneSelector(true);
  };

  const handleSceneSelect = (scene: string | null) => {
    setShowSceneSelector(false);

    // 選択されたメッセージを時系列順にフィルタリング
    const selectedMsgs = messages.filter((msg) => selectedMessages.has(msg.id));
    const chatHistory = selectedMsgs
      .map((msg) => `${msg.isSender ? '自分' : '相手'}: ${msg.content}`)
      .join('\n');

    setSelectionMode(false);
    setSelectedMessages(new Set());
    setRangeStart(null);
    setRangeEnd(null);

    navigate('/chat/ask-ai', {
      state: {
        initialPrompt: `【選択したチャット履歴】\n${chatHistory}`,
        fromChatFeedback: true,
        scene: scene,
      },
    });
  };

  // --- 言い換え提案 ---
  const handleRephrase = async (content: string) => {
    setRephraseResult(null);
    setShowRephraseModal(true);

    try {
      const res = await fetch(`${API_BASE_URL}/api/chat/ai/rephrase`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ originalMessage: content, scene: null }),
      });

      if (res.status === 401) {
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          { method: 'POST', credentials: 'include' }
        );
        if (!refreshRes.ok) {
          dispatch(clearAuth());
          return;
        }
        return handleRephrase(content);
      }

      if (!res.ok) {
        console.error('言い換えエラー:', res.status);
        setShowRephraseModal(false);
        return;
      }

      const data = await res.json();
      try {
        const parsed = JSON.parse(data.result);
        setRephraseResult(parsed);
      } catch {
        setRephraseResult({ formal: data.result, soft: '', concise: '' });
      }
    } catch (e) {
      console.error('言い換え失敗:', e);
      setShowRephraseModal(false);
    }
  };

  // メッセージが範囲内かどうかを判定
  const isInRange = (index: number): boolean => {
    if (rangeStart === null) return false;
    if (rangeEnd === null) return index === rangeStart;
    const start = Math.min(rangeStart, rangeEnd);
    const end = Math.max(rangeStart, rangeEnd);
    return index >= start && index <= end;
  };

  // 範囲選択のラベルを取得
  const getRangeLabel = (index: number): string | null => {
    if (rangeStart === index && rangeEnd === null) return '開始';
    if (rangeEnd === null) return null;
    const start = Math.min(rangeStart, rangeEnd);
    const end = Math.max(rangeStart, rangeEnd);
    if (index === start) return '開始';
    if (index === end) return '終了';
    return null;
  };

  // --- ユーザー選択してルームに移動 ---
  const handleSelectUser = async (userId: number) => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/chat/users/${userId}/create`, {
        method: 'POST',
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
          return;
        }
        return handleSelectUser(userId);
      }

      if (!res.ok) {
        console.error('ルーム作成エラー:', res.status);
        return;
      }

      const data = await res.json();
      if (data.roomId) {
        navigate(`/chat/users/${data.roomId}`);
      }
    } catch (e) {
      console.error('ルーム作成失敗', e);
    }
  };

  // --- 日付フォーマット ---
  const formatDate = (dateString: string): string => {
    if (!dateString) return '';
    const date = new Date(dateString);
    const now = new Date();
    const diffDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24));

    if (diffDays === 0) {
      return date.toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' });
    } else if (diffDays === 1) {
      return '昨日';
    } else if (diffDays < 7) {
      return ['日', '月', '火', '水', '木', '金', '土'][date.getDay()] + '曜日';
    } else {
      return date.toLocaleDateString('ja-JP', { month: 'short', day: 'numeric' });
    }
  };

  return (
    <>
      <HamburgerMenu title="個人チャット" />

      {/* 全体レイアウト - サイドバー付き */}
      <div className="flex h-screen bg-gray-50 text-black pt-16">

        {/* サイドバー */}
        <div className={`${sidebarOpen ? 'w-80' : 'w-0'} bg-white border-r border-gray-200 flex flex-col overflow-hidden`}>
          {/* サイドバーヘッダー */}
          <div className="p-4 border-b border-gray-100">
            <h3 className="font-bold text-gray-800 mb-3">チャット履歴</h3>
            <SearchBox
              value={searchQuery}
              onChange={setSearchQuery}
              placeholder="ユーザーを検索..."
            />
          </div>

          {/* ユーザーリスト */}
          <div className="flex-1 overflow-y-auto">
            {chatUsers.length === 0 ? (
              <div className="p-4 text-center text-gray-500">
                <p className="text-sm">チャット履歴がありません</p>
              </div>
            ) : (
              chatUsers.map((user) => (
                <button
                  key={user.roomId}
                  onClick={() => handleSelectUser(user.userId)}
                  className={`w-full p-4 flex items-start space-x-3 hover:bg-gray-50 transition-colors border-b border-gray-100 text-left ${
                    parseInt(roomId!) === user.roomId ? 'bg-primary-50 border-l-4 border-l-primary-500' : ''
                  }`}
                >
                  {/* アバター */}
                  <div className="w-12 h-12 bg-primary-500 rounded-full flex-shrink-0 flex items-center justify-center">
                    <span className="text-white font-bold text-lg">
                      {user.name?.charAt(0)?.toUpperCase() || 'U'}
                    </span>
                  </div>

                  {/* ユーザー情報 */}
                  <div className="flex-1 min-w-0">
                    <div className="flex justify-between items-start">
                      <h4 className="font-semibold text-gray-800 truncate">
                        {user.name || 'Unknown'}
                      </h4>
                      <span className="text-xs text-gray-400 flex-shrink-0 ml-2">
                        {formatDate(user.lastMessageAt)}
                      </span>
                    </div>
                    <div className="flex items-center justify-between mt-1">
                      <p className="text-sm text-gray-500 truncate">
                        {user.lastMessageSenderId === senderId && (
                          <span className="text-gray-400">あなた: </span>
                        )}
                        {user.lastMessage || 'メッセージがありません'}
                      </p>
                      {user.unreadCount > 0 && (
                        <span className="bg-pink-500 text-white text-xs font-bold px-2 py-0.5 rounded-full flex-shrink-0 ml-2">
                          {user.unreadCount > 99 ? '99+' : user.unreadCount}
                        </span>
                      )}
                    </div>
                  </div>
                </button>
              ))
            )}
          </div>
        </div>

        {/* サイドバートグルボタン */}
        <button
          onClick={() => setSidebarOpen(!sidebarOpen)}
          className="absolute left-0 top-1/2 -translate-y-1/2 z-20 bg-white border border-gray-200 rounded-r-lg p-2 shadow-sm hover:bg-gray-50 transition-colors"
          style={{ left: sidebarOpen ? '320px' : '0' }}
        >
          <svg
            className={`w-5 h-5 text-gray-600 transition-transform ${sidebarOpen ? 'rotate-180' : ''}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
          </svg>
        </button>

        {/* メインチャットエリア */}
        <div className="flex-1 flex flex-col">
          {/* ヘッダー情報 */}
          <div className="bg-white border-b border-gray-200 px-4 py-4 shadow-sm">
            <div className="max-w-4xl mx-auto flex items-center space-x-3">
              <div className="w-10 h-10 bg-primary-500 rounded-full flex items-center justify-center">
                <svg
                  className="w-6 h-6 text-white"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path d="M2 5a2 2 0 012-2h12a2 2 0 012 2v10a2 2 0 01-2 2H4a2 2 0 01-2-2V5z" />
                </svg>
              </div>
              <div>
                <h2 className="font-bold text-gray-800">チャット</h2>
                <p className="text-sm text-gray-600">メッセージをお送りください</p>
              </div>
            </div>
          </div>

          {/* メッセージエリア */}
          <div className="flex-1 overflow-y-auto px-4 py-6 space-y-3 max-w-4xl mx-auto w-full mb-[160px]">
            {messages.length === 0 && (
              <div className="flex flex-col items-center justify-center h-full text-center">
                <div className="w-16 h-16 bg-primary-100 rounded-full flex items-center justify-center mb-4">
                  <svg
                    className="w-8 h-8 text-blue-600"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path d="M2 5a2 2 0 012-2h12a2 2 0 012 2v10a2 2 0 01-2 2H4a2 2 0 01-2-2V5z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-gray-800 mb-2">
                  チャットへようこそ
                </h3>
                <p className="text-gray-600 max-w-sm">
                  相手とのチャットをここで行えます
                </p>
              </div>
            )}
            {messages.map((msg, index) => (
              <div key={msg.id} className="flex items-start gap-2">
                {selectionMode && (
                  <div className="flex-shrink-0 flex flex-col items-center">
                    {/* 範囲ラベル */}
                    {getRangeLabel(index) && (
                      <span className={`text-xs font-bold mb-1 px-2 py-0.5 rounded ${
                        getRangeLabel(index) === '開始'
                          ? 'bg-green-100 text-green-700'
                          : 'bg-red-100 text-red-700'
                      }`}>
                        {getRangeLabel(index)}
                      </span>
                    )}
                    <button
                      onClick={() => handleRangeClick(msg.id)}
                      className={`mt-1 w-8 h-8 rounded-full border-2 flex items-center justify-center transition-all ${
                        isInRange(index)
                          ? 'bg-blue-500 border-blue-500 text-white'
                          : 'border-gray-300 hover:border-blue-400 hover:bg-blue-50'
                      }`}
                    >
                      {isInRange(index) ? (
                        <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                        </svg>
                      ) : (
                        <span className="text-xs text-gray-400">{index + 1}</span>
                      )}
                    </button>
                  </div>
                )}
                <div className={`flex-1 transition-all ${
                  selectionMode && isInRange(index) ? 'bg-blue-50 -mx-2 px-2 py-1 rounded-lg' : ''
                }`}>
                  <MessageBubble
                    {...msg}
                    onDelete={selectionMode ? null : handleDeleteMessage}
                    onRephrase={selectionMode ? null : handleRephrase}
                  />
                </div>
              </div>
            ))}

            {/* スクロール最終地点 */}
            <div ref={messagesEndRef} />
          </div>

          {/* 入力欄固定 */}
          <div className="fixed bottom-0 right-0 bg-white border-t border-gray-200 shadow-sm p-4 z-10" style={{ left: sidebarOpen ? '320px' : '0' }}>
            <div className="max-w-4xl mx-auto w-full space-y-3">
              {selectionMode ? (
                /* 選択モードUI */
                <div className="space-y-3">
                  {/* ガイドメッセージ */}
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
                    <p className="text-sm text-blue-700">
                      {rangeStart === null
                        ? '💡 「開始」位置のメッセージをタップしてください'
                        : rangeEnd === null
                        ? '💡 「終了」位置のメッセージをタップしてください'
                        : `✅ ${selectedMessages.size}件のメッセージを選択しました`
                      }
                    </p>
                  </div>

                  {/* クイック選択ボタン */}
                  <div className="flex gap-2 flex-wrap">
                    <span className="text-sm text-gray-500 self-center">クイック選択:</span>
                    <button
                      onClick={() => handleQuickSelect(5)}
                      className="px-3 py-1.5 text-sm bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-full transition-colors"
                    >
                      直近5件
                    </button>
                    <button
                      onClick={() => handleQuickSelect(10)}
                      className="px-3 py-1.5 text-sm bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-full transition-colors"
                    >
                      直近10件
                    </button>
                    <button
                      onClick={() => handleQuickSelect(20)}
                      className="px-3 py-1.5 text-sm bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-full transition-colors"
                    >
                      直近20件
                    </button>
                    <button
                      onClick={handleSelectAll}
                      className="px-3 py-1.5 text-sm bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-full transition-colors"
                    >
                      すべて
                    </button>
                    <button
                      onClick={handleDeselectAll}
                      className="px-3 py-1.5 text-sm text-red-500 hover:bg-red-50 rounded-full transition-colors"
                    >
                      リセット
                    </button>
                  </div>

                  {/* アクションボタン */}
                  <div className="flex gap-2">
                    <button
                      onClick={handleCancelSelection}
                      className="flex-1 bg-gray-200 hover:bg-gray-300 text-gray-700 font-semibold py-3 px-4 rounded-lg transition-colors duration-150"
                    >
                      キャンセル
                    </button>
                    <button
                      onClick={handleSendToAi}
                      disabled={selectedMessages.size === 0}
                      className={`flex-1 font-semibold py-3 px-4 rounded-lg transition-colors duration-150 ${
                        selectedMessages.size > 0
                          ? 'bg-primary-500 hover:bg-primary-600 text-white'
                          : 'bg-gray-300 text-gray-500 cursor-not-allowed'
                      }`}
                    >
                      {selectedMessages.size > 0
                        ? `${selectedMessages.size}件をAIに送信`
                        : '範囲を選択してください'
                      }
                    </button>
                  </div>
                </div>
              ) : (
                /* 通常モードUI */
                <>
                  {messages.length > 0 && (
                    <button
                      onClick={handleAiFeedback}
                      className="w-full bg-primary-500 hover:bg-primary-600 text-white font-semibold py-3 px-4 rounded-lg transition-colors duration-150"
                    >
                      AIにフィードバックしてもらう
                    </button>
                  )}
                  <MessageInput onSend={handleSend} />
                </>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* シーン選択モーダル */}
      {showSceneSelector && (
        <SceneSelector
          onSelect={(sceneId) => handleSceneSelect(sceneId)}
          onCancel={() => handleSceneSelect(null)}
        />
      )}

      {/* 言い換え提案モーダル */}
      {showRephraseModal && (
        <RephraseModal
          result={rephraseResult}
          onClose={() => setShowRephraseModal(false)}
        />
      )}

      {/* 削除確認モーダル */}
      <ConfirmModal
        isOpen={deleteModal.isOpen}
        title="メッセージを削除"
        message="このメッセージを削除しますか？この操作は取り消せません。"
        confirmText="削除する"
        cancelText="キャンセル"
        onConfirm={confirmDelete}
        onCancel={cancelDelete}
        isDanger={true}
      />
    </>
  );
}
