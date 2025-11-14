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
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
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
          console.warn("アクセストークン期限切れ → リフレッシュします");

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
        console.error("履歴取得失敗:", err);
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

    wsRef.current.onopen = () => console.log("WebSocket Connected");

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

    wsRef.current.onerror = (e) => console.error("WS Error:", e);
    wsRef.current.onclose = () => console.log("WS Closed");

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
      <HamburgerMenu title="AIチャット" />

      {/* 全体レイアウト */}
      <div className="flex flex-col h-screen bg-gray-100 text-black pt-16">

        {/* メッセージエリア */}
        <div className="flex-1 overflow-y-auto px-4 py-6 space-y-2 max-w-3xl mx-auto w-full pb-[100px]">
          {messages.map((msg) => (
            <MessageBubble key={msg.id} {...msg} />
          ))}

          {/* スクロール最終地点 */}
          <div ref={messagesEndRef} />
        </div>

        {/* 入力欄固定 */}
        <div className="fixed bottom-0 left-0 right-0 bg-gray-100 p-3 z-10">
          <div className="max-w-3xl mx-auto w-full">
            <MessageInput onSend={handleSend} />
          </div>
        </div>
      </div>
    </>
  );
}
