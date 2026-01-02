import { useState, useEffect, useRef } from 'react';
import MessageBubbleAi from '../components/MessageBubbleAi';
import MessageInput from '../components/MessageInput';
import HamburgerMenu from '../components/HamburgerMenu';
import { useDispatch } from 'react-redux';
import { useNavigate, useLocation } from 'react-router-dom';

export default function AskAiPage() {
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  const WS_URL = import.meta.env.VITE_WEB_SOCKET_URL_AI_CHAT;

  const [messages, setMessages] = useState([]);
  const [initialPromptSent, setInitialPromptSent] = useState(false);
  const [historyLoaded, setHistoryLoaded] = useState(false);
  const [senderId, setSenderId] = useState(null);
  const wsRef = useRef(null);
  const messagesEndRef = useRef(null);

  const navigate = useNavigate();
  const location = useLocation();
  const dispatch = useDispatch();

  const initialPrompt = location.state?.initialPrompt;

  // ユーザー情報取得（senderId を取得）
  useEffect(() => {
    const fetchUserInfo = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/auth/me`, {
          credentials: 'include',
        });
        if (!res.ok) {
          navigate('/login');
          return;
        }
        const data = await res.json();
        setSenderId(data.sub);
      } catch (error) {
        console.error('ユーザー情報取得エラー:', error);
        navigate('/login');
      }
    };

    fetchUserInfo();
  }, [API_BASE_URL, navigate]);

  // --- メッセージ最下部へスクロール ---
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // --- チャット履歴取得（AIフィードバック時のみ） ---
  useEffect(() => {
    // initialPromptがない場合は履歴取得をスキップ（通常のAIチャット）
    if (!initialPrompt) {
      console.log('📝 通常のAIチャットモード（ユーザーチャット履歴なし）');
      setHistoryLoaded(true);
      return;
    }

    console.log('🔄 フィードバックモード：AI履歴取得開始');

    const fetchHistory = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/chat/ai/history`, {
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
        });

        if (res.status === 401) {
          console.warn('アクセストークン期限切れ → リフレッシュします');

          const refreshRes = await fetch(
            `${API_BASE_URL}/api/auth/cognito/refresh-token`,
            { method: 'POST', credentials: 'include' }
          );

          if (!refreshRes.ok) {
            navigate('/login');
            return;
          }

          // アクセストークン更新だがhttpOnlyなのでReduxには保存しない
          const refreshData = await refreshRes.json();

          const retryRes = await fetch(`${API_BASE_URL}/api/chat/ai/history`, {
            headers: {
              'Content-Type': 'application/json',
            },
            credentials: 'include',
          });

          const retryData = await retryRes.json();
          const formattedMessages = retryData.map((item) => ({
            id: item.timestamp,
            timestamp: item.timestamp,
            content: item.content,
            isSender: item.user === true || item.isUser === true,
          }));

          setMessages(formattedMessages);
          setHistoryLoaded(true);
          console.log('✅ AI履歴取得完了（フィードバックモード）');
          return;
        }

        const data = await res.json();
        const formattedMessages = data.map((item) => ({
          id: item.timestamp,
          timestamp: item.timestamp,
          content: item.content,
          isSender: item.user === true || item.isUser === true,
        }));
        setMessages(formattedMessages);
        setHistoryLoaded(true);
        console.log('✅ AI履歴取得完了（フィードバックモード）');
      } catch (err) {
        console.error('履歴取得失敗:', err);
        setHistoryLoaded(true);
      }
    };

  }, [initialPrompt, API_BASE_URL, dispatch, navigate]);

  // --- WebSocket ---
  useEffect(() => {
    if (!senderId || !historyLoaded) return;

    const socketUrl = `${WS_URL}?user_id=${senderId}&room_id=default`;
    wsRef.current = new WebSocket(socketUrl);

    wsRef.current.onopen = () => {
      console.log('✅ WebSocket Connected - AI履歴取得完了後に接続');

      // 初期プロンプトがあれば自動送信
      if (initialPrompt && !initialPromptSent) {
        console.log('📤 初期プロンプトを送信します');
        wsRef.current.send(
          JSON.stringify({
            sender_id: senderId,
            content: initialPrompt,
          })
        );

        // ユーザー側のメッセージも表示
        setMessages((prev) => [
          ...prev,
          {
            id: Date.now(),
            timestamp: Date.now(),
            content: initialPrompt,
            isSender: true,
          },
        ]);

        setInitialPromptSent(true);
      }
    };

    wsRef.current.onmessage = (event) => {
      const data = JSON.parse(event.data);

      setMessages((prev) => [
        ...prev,
        {
          id: data.timestamp ?? Date.now(),
          timestamp: data.timestamp ?? Date.now(),
          content: data.reply || data.message,
          isSender: data.from === senderId,
        },
      ]);
    };

    wsRef.current.onerror = (e) => console.error('WS Error:', e);
    wsRef.current.onclose = () => console.log('WS Closed');

    return () => {
      wsRef.current?.close();
    };
  }, [WS_URL, senderId, initialPrompt, initialPromptSent, historyLoaded]);

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

      // WebSocketで削除を送信
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.send(
          JSON.stringify({
            action: 'delete',
            timestamp: messageToDelete.timestamp,
            sender_id: senderId,
          })
        );
      }
    }
  };

  // --- メッセージ送信 ---
  const handleSend = (text) => {
    const timestampNow = Date.now();

    setMessages((prev) => [
      ...prev,
      { id: timestampNow, content: text, isSender: true },
    ]);

    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({
          sender_id: senderId,
          content: text,
        })
      );
    }
  };

  return (
    <>
      <HamburgerMenu title="AIに聞く" />

      {/* 全体レイアウト */}
      <div className="flex flex-col h-screen bg-gradient-to-br from-gray-50 to-pink-50 text-black pt-16">
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
        <div className="flex-1 overflow-y-auto px-4 py-6 space-y-3 max-w-4xl mx-auto w-full pb-[120px]">
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
            <MessageBubble
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
        <div className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 shadow-2xl p-4 z-10">
          <div className="max-w-4xl mx-auto w-full">
            <MessageInput onSend={handleSend} />
          </div>
        </div>
      </div>
    </>
  );
}
