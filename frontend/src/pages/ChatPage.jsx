import { useState, useEffect, useRef } from 'react';
import { useParams } from 'react-router-dom';
import MessageBubble from '../components/MessageBubble';
import MessageInput from '../components/MessageInput';
import HamburgerMenu from '../components/HamburgerMenu';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

export default function ChatPage() {
  const [messages, setMessages] = useState([]);
  const wsRef = useRef(null);
  const { roomId } = useParams();
  const senderId = useSelector((state) => state.auth.sub);
  const token = useSelector((state) => state.auth.accessToken);
  const navigate = useNavigate();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  // ログインチェック
  useEffect(() => {
    if (!token || !senderId) {
      navigate('/login');
    }
  }, [token, senderId]);

  // --- チャット履歴取得 ---
  const fetchHistory = async () => {
    try {
      console.log('リクエスト開始');
      const response = await fetch(
        `${API_BASE_URL}/api/chat/users/${roomId}/history`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }
      );

      if (response.status === 401) {
        navigate('/login');
        return;
      }

      const data = await response.json();

      const formattedMessages = data.map((msg) => ({
        id: msg.timestamp,
        content: msg.content,
        isSender: msg.isUser, // Spring 側の isUser をそのまま使用
      }));

      setMessages(formattedMessages);
      console.log('✅ 履歴取得成功');
      console.log(data);
    } catch (err) {
      console.error('❌ 履歴取得失敗:', err);
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
  }, [roomId, senderId]);

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
