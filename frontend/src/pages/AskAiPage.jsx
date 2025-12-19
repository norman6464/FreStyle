import { useState, useEffect, useRef } from 'react';
import MessageBubble from '../components/MessageBubble';
import MessageInput from '../components/MessageInput';
import HamburgerMenu from '../components/HamburgerMenu';
import { useSelector, useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { setAuthData, clearAuthData } from '../store/authSlice';

export default function AskAiPage() {
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  const WS_URL = import.meta.env.VITE_WEB_SOCKET_URL_AI_CHAT;

  const [messages, setMessages] = useState([]);
  const wsRef = useRef(null);
  const messagesEndRef = useRef(null);

  const navigate = useNavigate();
  const dispatch = useDispatch();

  const accessToken = useSelector((state) => state.auth.accessToken);
  const senderId = useSelector((state) => state.auth.sub);

  // --- メッセージ最下部へスクロール ---
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // --- チャット履歴取得 ---
  useEffect(() => {
    const fetchHistory = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/chat/ai/history`, {
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${accessToken}`,
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
            dispatch(clearAuthData());
            navigate('/login');
            return;
          }

          const refreshData = await refreshRes.json();
          const newAccessToken = refreshData.accessToken;

          if (!newAccessToken) {
            dispatch(clearAuthData());
            navigate('/login');
            return;
          }

          dispatch(setAuthData({ accessToken: newAccessToken }));

          const retryRes = await fetch(`${API_BASE_URL}/api/chat/ai/history`, {
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${newAccessToken}`,
            },
            credentials: 'include',
          });

          const retryData = await retryRes.json();
          const formattedMessages = retryData.map((item) => ({
            id: item.timestamp,
            content: item.content,
            isSender: item.user === true || item.isUser === true,
          }));

          setMessages(formattedMessages);
          return;
        }

        const data = await res.json();
        const formattedMessages = data.map((item) => ({
          id: item.timestamp,
          content: item.content,
          isSender: item.user === true || item.isUser === true,
        }));
        setMessages(formattedMessages);
      } catch (err) {
        console.error('履歴取得失敗:', err);
      }
    };

    if (accessToken) {
      fetchHistory();
    } else {
      navigate('/login');
    }
  }, [API_BASE_URL, accessToken, dispatch, navigate]);

  // --- WebSocket ---
  useEffect(() => {
    if (!senderId) return;

    const socketUrl = `${WS_URL}?user_id=${senderId}&room_id=default`;
    wsRef.current = new WebSocket(socketUrl);

    wsRef.current.onopen = () => console.log('WebSocket Connected');

    wsRef.current.onmessage = (event) => {
      const data = JSON.parse(event.data);

      setMessages((prev) => [
        ...prev,
        {
          id: data.timestamp ?? Date.now(),
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
  }, [WS_URL, senderId]);

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
