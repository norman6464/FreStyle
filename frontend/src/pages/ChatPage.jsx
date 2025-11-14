import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import MessageBubble from '../components/MessageBubble';
import MessageInput from '../components/MessageInput';
import HamburgerMenu from '../components/HamburgerMenu';
import { useSelector, useDispatch } from 'react-redux';
import { setAuthData, clearAuthData } from '../store/authSlice';

export default function ChatPage() {
  const [messages, setMessages] = useState([]);
  const wsRef = useRef(null);
  const { roomId } = useParams();
  const senderId = useSelector((state) => state.auth.sub);
  const accessToken = useSelector((state) => state.auth.accessToken);
  const email = useSelector((state) => state.auth.email);
  const navigate = useNavigate();
  const dispatch = useDispatch();

  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  // --- チャット履歴取得（JWT認証＋リフレッシュ対応） ---
  const fetchHistory = async () => {
    try {
      console.log('📡 履歴リクエスト開始');
      const res = await fetch(
        `${API_BASE_URL}/api/chat/users/${roomId}/history`,
        {
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${accessToken}`,
          },
          credentials: 'include', // Cookie（Refresh Token）送信
        }
      );

      // アクセストークン期限切れ
      if (res.status === 401) {
        console.warn('アクセストークン期限切れ。リフレッシュを試行します。');

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
          console.warn('新しいアクセストークンが取得できませんでした。');
          dispatch(clearAuthData());
          navigate('/login');
          return;
        }

        // Redux更新
        dispatch(setAuthData({ accessToken: newAccessToken }));
        console.log('✅ アクセストークン更新成功。再リクエストを実行します。');

        // 再試行
        const retryRes = await fetch(
          `${API_BASE_URL}/api/chat/users/${roomId}/history`,
          {
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${newAccessToken}`,
            },
            credentials: 'include',
          }
        );

        if (!retryRes.ok) throw new Error('再リクエスト失敗');

        const retryData = await retryRes.json();
        const formattedMessages = retryData.map((msg) => ({
          id: msg.timestamp,
          content: msg.content,
          isSender: msg.isUser,
        }));
        setMessages(formattedMessages);
        console.log('✅ 履歴再取得成功');
        return;
      }

      // 通常成功時
      if (!res.ok) throw new Error(`履歴取得失敗: ${res.status}`);

      const data = await res.json();

      data.map((msg) => {
        console.log(msg.isUser);
      });

      const formattedMessages = data.map((msg) => ({
        id: msg.timestamp,
        content: msg.content,
        isSender: msg.isUser,
      }));
      setMessages(formattedMessages);
      console.log('✅ 履歴取得成功');
    } catch (err) {
      console.error('❌ 履歴取得中エラー:', err);
    }
  };

  // --- WebSocket接続 ---
  useEffect(() => {
    if (!senderId) return;

    const wsUrl = `${
      import.meta.env.VITE_WEB_SOCKET_URL_CHAT
    }?user_id=${senderId}&room_id=${roomId}`;

    wsRef.current = new WebSocket(wsUrl);

    wsRef.current.onopen = () => {
      console.log('✅ WebSocket connected');
      fetchHistory();
    };

    wsRef.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      console.log('📩 WebSocket受信:', data);

      setMessages((prev) => [
        ...prev,
        {
          id: data.timestamp ?? Date.now(),
          content: data.content || data.message,
          isSender: data.sender_id === senderId,
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
  }, [roomId, senderId, accessToken]);

  // --- メッセージ送信 ---
  const handleSend = (text) => {
    const timestampNow = Date.now();

    // 即時反映
    setMessages((prev) => [
      ...prev,
      { id: timestampNow, content: text, isSender: true },
    ]);

    // WebSocket送信
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({
          room_id: roomId,
          sender_id: senderId,
          content: text,
        })
      );
    } else {
      console.warn('⚠️ WebSocket未接続: メッセージ送信できません');
    }
  };

  return (
    <>
      <HamburgerMenu title="個人チャット" />
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
