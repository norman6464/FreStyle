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
  const navigate = useNavigate();
  const dispatch = useDispatch();

  const accessToken = useSelector((state) => state.auth.accessToken);
  const senderId = useSelector((state) => state.auth.sub);

  // --- チャット履歴取得 ---
  useEffect(() => {
    const fetchHistory = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/chat/ai/history`, {
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${accessToken}`,
          },
          credentials: 'include', // Cookie送信
        });

        if (res.status === 401) {
          console.warn('アクセストークン期限切れ。リフレッシュ試行開始。');

          // アクセストークン再発行
          const refreshRes = await fetch(
            `${API_BASE_URL}/api/auth/cognito/refresh-token`,
            {
              method: 'POST',
              credentials: 'include',
            }
          );

          if (!refreshRes.ok) {
            console.error('リフレッシュ失敗。再ログインへ遷移。');
            dispatch(clearAuthData());
            navigate('/login');
            return;
          }

          const refreshData = await refreshRes.json();
          const newAccessToken = refreshData.accessToken;

          if (!newAccessToken) {
            console.warn('新しいアクセストークン取得失敗。再ログインへ。');
            dispatch(clearAuthData());
            navigate('/login');
            return;
          }

          // Redux に新トークン反映
          dispatch(setAuthData({ accessToken: newAccessToken }));

          console.log('アクセストークン更新完了。再リクエストします。');

          // --- 再試行 ---
          const retryRes = await fetch(`${API_BASE_URL}/api/chat/ai/history`, {
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${newAccessToken}`,
            },
            credentials: 'include',
          });

          if (!retryRes.ok) {
            throw new Error('再リクエスト失敗');
          }

          const retryData = await retryRes.json();
          const formattedMessages = retryData.map((item) => ({
            id: item.timestamp,
            content: item.content,
            isSender: item.user === true || item.isUser === true,
          }));

          setMessages(formattedMessages);
          return;
        }

        if (!res.ok) {
          throw new Error(`履歴取得失敗: ${res.status}`);
        }

        const data = await res.json();
        const formattedMessages = data.map((item) => ({
          id: item.timestamp,
          content: item.content,
          isSender: item.user === true || item.isUser === true,
        }));
        setMessages(formattedMessages);
      } catch (err) {
        console.error('履歴取得中にエラー:', err);
      }
    };

    // アクセストークンがある時だけ実行
    if (accessToken) {
      fetchHistory();
    } else {
      console.warn('アクセストークンがありません。ログインへ遷移');
      navigate('/login');
    }
  }, [API_BASE_URL, accessToken, dispatch, navigate]);

  // --- WebSocket接続 ---
  useEffect(() => {
    if (!senderId) return;

    const socketUrl = `${WS_URL}?user_id=${senderId}&room_id=default`;
    wsRef.current = new WebSocket(socketUrl);

    wsRef.current.onopen = () => {
      console.log('✅ WebSocket connected');
    };

    wsRef.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      console.log('📩 WebSocket受信:', data);

      setMessages((prev) => [
        ...prev,
        {
          id: data.timestamp ?? Date.now(),
          content: data.reply || data.message,
          isSender: data.from === senderId,
        },
      ]);
    };

    wsRef.current.onerror = (err) => {
      console.error('❌ WebSocket error:', err);
    };

    wsRef.current.onclose = () => {
      console.log('❎ WebSocket closed');
    };

    return () => {
      if (wsRef.current) wsRef.current.close();
    };
  }, [WS_URL, senderId]);

  // --- メッセージ送信 ---
  const handleSend = (text) => {
    const timestampNow = Date.now();

    // 即時UI反映
    setMessages((prev) => [
      ...prev,
      { id: timestampNow, content: text, isSender: true },
    ]);

    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      const payload = {
        sender_id: senderId,
        content: text,
      };
      wsRef.current.send(JSON.stringify(payload));
    } else {
      console.warn('WebSocket未接続: メッセージ送信できません');
    }
  };

  return (
    <>
      <HamburgerMenu title="AIチャット" />
      <div className="flex flex-col h-screen bg-gray-100 text-black mt-16">
        <div className="flex-1 overflow-y-auto px-4 py-6 space-y-2 max-w-3xl mx-auto w-full">
          {messages.map((msg) => (
            <MessageBubble key={msg.id} {...msg} />
          ))}
        </div>
        <MessageInput onSend={handleSend} />
      </div>
    </>
  );
}
