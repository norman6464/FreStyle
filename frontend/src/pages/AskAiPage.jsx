import { useState, useEffect, useRef } from 'react';
import MessageBubbleAi from '../components/MessageBubbleAi';
import MessageInput from '../components/MessageInput';
import HamburgerMenu from '../components/HamburgerMenu';
import { useDispatch } from 'react-redux';
import { useNavigate, useLocation, useParams } from 'react-router-dom';

import SockJS from 'sockjs-client';
import { Client } from '@stomp/stompjs';

export default function AskAiPage() {
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  const WS_URL = import.meta.env.VITE_WEB_SOCKET_URL_AI_CHAT;

  const [messages, setMessages] = useState([]);
  const [sessions, setSessions] = useState([]);
  const [currentSessionId, setCurrentSessionId] = useState(null);
  const [initialPromptSent, setInitialPromptSent] = useState(false);
  const [historyLoaded, setHistoryLoaded] = useState(false);
  const [userId, setUserId] = useState(null);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  
  const stompClientRef = useRef(null);
  const messagesEndRef = useRef(null);

  const navigate = useNavigate();
  const location = useLocation();
  const dispatch = useDispatch();
  const { sessionId: urlSessionId } = useParams();

  const initialPrompt = location.state?.initialPrompt;

  // ユーザー情報取得（userId を取得）
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
        setUserId(data.id);
        console.log('[AskAiPage] Fetched user info, userId:', data.id);
      } catch (error) {
        console.error('ユーザー情報取得エラー:', error);
        navigate('/login');
      }
    };

    fetchUserInfo();
  }, [API_BASE_URL, navigate]);

  // --- セッション一覧取得 ---
  const fetchSessions = async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/chat/ai/sessions`, {
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
      });

      if (res.status === 401) {
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          { method: 'POST', credentials: 'include' }
        );
        if (!refreshRes.ok) {
          navigate('/login');
          return;
        }
        return fetchSessions();
      }

      if (!res.ok) {
        console.error('セッション一覧取得エラー:', res.status);
        return;
      }

      const data = await res.json();
      setSessions(data || []);
      console.log('[AskAiPage] セッション一覧取得成功:', data.length, '件');
    } catch (e) {
      console.error('セッション一覧取得失敗', e);
    }
  };

  // 初回ロード時にセッション一覧を取得
  useEffect(() => {
    if (userId) {
      fetchSessions();
    }
  }, [userId]);

  // --- メッセージ最下部へスクロール ---
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // --- URLパラメータのセッションID変更時にメッセージ履歴を取得 ---
  useEffect(() => {
    if (urlSessionId) {
      setCurrentSessionId(parseInt(urlSessionId));
    }
  }, [urlSessionId]);

  // --- セッション内のメッセージ履歴取得 ---
  const fetchSessionMessages = async (sessionId) => {
    if (!sessionId) {
      setMessages([]);
      setHistoryLoaded(true);
      return;
    }

    try {
      console.log('📥 セッションメッセージ取得開始 - sessionId:', sessionId);
      const res = await fetch(`${API_BASE_URL}/api/chat/ai/sessions/${sessionId}/messages`, {
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
      });

      if (res.status === 401) {
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          { method: 'POST', credentials: 'include' }
        );
        if (!refreshRes.ok) {
          navigate('/login');
          return;
        }
        return fetchSessionMessages(sessionId);
      }

      if (!res.ok) {
        console.error('メッセージ履歴取得エラー:', res.status);
        setHistoryLoaded(true);
        return;
      }

      const data = await res.json();
      const formattedMessages = data.map((item) => ({
        id: item.id,
        sessionId: item.sessionId,
        content: item.content,
        isSender: item.role === 'user',
        createdAt: item.createdAt,
      }));

      setMessages(formattedMessages);
      setHistoryLoaded(true);
      console.log('✅ メッセージ履歴取得完了 - 件数:', formattedMessages.length);
    } catch (err) {
      console.error('❌ メッセージ履歴取得失敗:', err);
      setHistoryLoaded(true);
    }
  };

  // currentSessionIdが変わったらメッセージを取得
  useEffect(() => {
    if (currentSessionId) {
      fetchSessionMessages(currentSessionId);
    } else {
      setMessages([]);
      setHistoryLoaded(true);
    }
  }, [currentSessionId]);

  // --- WebSocket (STOMP) 接続 ---
  useEffect(() => {
    if (!userId) {
      console.log('⏳ [WebSocket useEffect] userId未設定');
      return;
    }

    console.log('🔗 [WebSocket useEffect] STOMP接続開始 - userId:', userId);

    const client = new Client({
      webSocketFactory: () =>
        new SockJS(`${API_BASE_URL}/ws/ai-chat`, undefined, { withCredentials: true }),
      reconnectDelay: 5000,

      onConnect: () => {
        console.log('✅ STOMP connected for AI Chat');
        stompClientRef.current = client;

        // ユーザー単位のセッション更新通知を購読
        client.subscribe(`/topic/ai-chat/user/${userId}/session`, (message) => {
          const newSession = JSON.parse(message.body);
          console.log('📩 新規セッション作成通知:', newSession);
          setSessions((prev) => [newSession, ...prev]);
          setCurrentSessionId(newSession.id);
        });

        // セッション削除通知を購読
        client.subscribe(`/topic/ai-chat/user/${userId}/session-deleted`, (message) => {
          const data = JSON.parse(message.body);
          console.log('🗑️ セッション削除通知:', data);
          setSessions((prev) => prev.filter((s) => s.id !== data.sessionId));
          if (currentSessionId === data.sessionId) {
            setCurrentSessionId(null);
            setMessages([]);
          }
        });

        // 現在のセッションがあれば購読
        if (currentSessionId) {
          subscribeToSession(currentSessionId);
        }

        // 初期プロンプトがあれば自動送信
        if (initialPrompt && !initialPromptSent) {
          console.log('📤 初期プロンプトを送信します:', initialPrompt);
          handleSend(initialPrompt);
          setInitialPromptSent(true);
        }

        setHistoryLoaded(true);
      },

      onStompError: (frame) => {
        console.error('STOMP Error', frame);
      },
    });

    client.activate();

    return () => {
      console.log('🧹 [WebSocket cleanup] クリーンアップ実行中');
      client.deactivate();
    };
  }, [userId]);

  // セッション購読関数
  const subscribeToSession = (sessionId) => {
    if (!stompClientRef.current?.connected || !sessionId) return;

    console.log('📡 セッション購読開始:', sessionId);
    stompClientRef.current.subscribe(`/topic/ai-chat/session/${sessionId}`, (message) => {
      const data = JSON.parse(message.body);
      console.log('📩 AIチャットメッセージ受信:', data);

      setMessages((prev) => {
        // 重複チェック
        if (prev.some((m) => m.id === data.id)) {
          return prev;
        }
        return [
          ...prev,
          {
            id: data.id,
            sessionId: data.sessionId,
            content: data.content,
            isSender: data.role === 'user',
            createdAt: data.createdAt,
          },
        ];
      });
    });
  };

  // currentSessionIdが変わったら再購読
  useEffect(() => {
    if (currentSessionId && stompClientRef.current?.connected) {
      subscribeToSession(currentSessionId);
    }
  }, [currentSessionId]);

  // --- メッセージ削除処理 ---
  const handleDeleteMessage = (messageId) => {
    const messageToDelete = messages.find((msg) => msg.id === messageId);
    if (!messageToDelete) return;

    if (confirm('このメッセージを削除しますか？')) {
      // ローカルstateで削除済みマークをつける
      setMessages((prev) =>
        prev.map((msg) =>
          msg.id === messageId ? { ...msg, isDeleted: true } : msg
        )
      );
    }
  };

  // --- 新規セッション作成 ---
  const handleNewSession = () => {
    setCurrentSessionId(null);
    setMessages([]);
    console.log('🆕 新規セッション開始');
  };

  // --- セッション選択 ---
  const handleSelectSession = (sessionId) => {
    setCurrentSessionId(sessionId);
    navigate(`/ask-ai/${sessionId}`);
    console.log('📂 セッション選択:', sessionId);
  };

  // --- セッション削除 ---
  const handleDeleteSession = async (sessionId) => {
    if (!confirm('このセッションを削除しますか？')) return;

    try {
      const res = await fetch(`${API_BASE_URL}/api/chat/ai/sessions/${sessionId}`, {
        method: 'DELETE',
        credentials: 'include',
      });

      if (res.ok) {
        setSessions((prev) => prev.filter((s) => s.id !== sessionId));
        if (currentSessionId === sessionId) {
          setCurrentSessionId(null);
          setMessages([]);
        }
        console.log('✅ セッション削除成功');
      }
    } catch (e) {
      console.error('❌ セッション削除失敗:', e);
    }
  };

  // --- メッセージ送信 ---
  const handleSend = async (text) => {
    console.log('📤 [handleSend] メッセージ送信開始:', { text, userId, currentSessionId });
    
    if (!stompClientRef.current?.connected) {
      console.warn('⚠️ STOMP not connected');
      return;
    }

    // ローカルに即時表示（楽観的更新）
    const tempId = Date.now();
    setMessages((prev) => [
      ...prev,
      { id: tempId, content: text, isSender: true, isTemp: true },
    ]);

    // STOMPで送信
    const payload = {
      userId: userId,
      sessionId: currentSessionId,
      content: text,
      role: 'user',
    };

    console.log('📤 STOMP送信:', payload);
    stompClientRef.current.publish({
      destination: '/app/ai-chat/send',
      body: JSON.stringify(payload),
    });
  };

  return (
    <>
      <HamburgerMenu title="AIに聞く" />

      {/* 全体レイアウト */}
      <div className="flex h-screen bg-gradient-to-br from-gray-50 to-pink-50 text-black pt-16">
        
        {/* サイドバー（セッション一覧） */}
        <div className={`${sidebarOpen ? 'w-64' : 'w-0'} transition-all duration-300 bg-white border-r border-gray-200 flex flex-col overflow-hidden`}>
          <div className="p-4 border-b border-gray-200">
            <button
              onClick={handleNewSession}
              className="w-full bg-gradient-to-r from-pink-500 to-orange-400 text-white py-2 px-4 rounded-lg font-medium hover:opacity-90 transition-opacity flex items-center justify-center gap-2"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              新しいチャット
            </button>
          </div>
          
          <div className="flex-1 overflow-y-auto">
            <div className="p-2 space-y-1">
              {sessions.map((session) => (
                <div
                  key={session.id}
                  className={`group flex items-center justify-between p-3 rounded-lg cursor-pointer transition-colors ${
                    currentSessionId === session.id
                      ? 'bg-pink-100 text-pink-700'
                      : 'hover:bg-gray-100'
                  }`}
                  onClick={() => handleSelectSession(session.id)}
                >
                  <div className="flex-1 truncate">
                    <p className="text-sm font-medium truncate">{session.title || '新しいチャット'}</p>
                    <p className="text-xs text-gray-500">
                      {session.createdAt ? new Date(session.createdAt).toLocaleDateString('ja-JP') : ''}
                    </p>
                  </div>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      handleDeleteSession(session.id);
                    }}
                    className="opacity-0 group-hover:opacity-100 p-1 hover:bg-red-100 rounded transition-opacity"
                  >
                    <svg className="w-4 h-4 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* サイドバートグルボタン */}
        <button
          onClick={() => setSidebarOpen(!sidebarOpen)}
          className="absolute left-0 top-1/2 transform -translate-y-1/2 bg-white border border-gray-200 rounded-r-lg p-2 shadow-md z-20 ml-64 transition-all duration-300"
          style={{ marginLeft: sidebarOpen ? '256px' : '0' }}
        >
          <svg
            className={`w-4 h-4 text-gray-600 transition-transform ${sidebarOpen ? 'rotate-180' : ''}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
          </svg>
        </button>

        {/* メインコンテンツ */}
        <div className="flex-1 flex flex-col">
          {/* ヘッダー情報 */}
          <div className="bg-white border-b border-gray-200 px-4 py-4 shadow-sm">
            <div className="max-w-4xl mx-auto flex items-center space-x-3">
              <div className="w-10 h-10 bg-gradient-to-br from-pink-400 to-orange-400 rounded-full flex items-center justify-center">
                <svg
                  className="w-6 h-6 text-white"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                </svg>
              </div>
              <div>
                <h2 className="font-bold text-gray-800">AIアシスタント</h2>
                <p className="text-sm text-gray-600">何でも聞いてください</p>
              </div>
            </div>
          </div>

          {/* メッセージエリア */}
          <div className="flex-1 overflow-y-auto px-4 py-6 space-y-3 max-w-4xl mx-auto w-full mb-[100px]">
            {messages.length === 0 && (
              <div className="flex flex-col items-center justify-center h-full text-center">
                <div className="w-16 h-16 bg-gradient-to-br from-pink-200 to-orange-200 rounded-full flex items-center justify-center mb-4">
                  <svg
                    className="w-8 h-8 text-pink-600"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path d="M2 5a2 2 0 012-2h12a2 2 0 012 2v10a2 2 0 01-2 2H4a2 2 0 01-2-2V5z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-gray-800 mb-2">
                  AIアシスタントへようこそ
                </h3>
                <p className="text-gray-600 max-w-sm">
                  質問や相談を何でも聞いてください。AIがすぐに答えます
                </p>
              </div>
            )}
            {messages.map((msg) => (
              <MessageBubbleAi
                key={msg.id}
                {...msg}
                type={msg.isSender ? 'text' : 'bot'}
                onDelete={handleDeleteMessage}
              />
            ))}

            {/* スクロール最終地点 */}
            <div ref={messagesEndRef} />
          </div>

          {/* 入力欄固定 */}
          <div className="fixed bottom-0 right-0 bg-white border-t border-gray-200 shadow-2xl p-4 z-10"
               style={{ left: sidebarOpen ? '256px' : '0', transition: 'left 0.3s' }}>
            <div className="max-w-4xl mx-auto w-full">
              <MessageInput onSend={handleSend} />
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
