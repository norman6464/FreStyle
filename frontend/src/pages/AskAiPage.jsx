import { useState, useEffect, useRef } from 'react';
import MessageBubble from '../components/MessageBubble';
import MessageInput from '../components/MessageInput';
import { useSelector } from 'react-redux';

export default function AskAiPage() {
  const [messages, setMessages] = useState([]);
  const senderId = useSelector((state) => state.auth.sub); // subをsenderIdにする
  const wsRef = useRef(null);
  const token = useSelector((state) => state.auth.accessToken); // アクセストークン

  // --- WebSocket & 履歴取得 ---
  useEffect(() => {
    // ① 履歴取得（Spring Boot 経由）
    const fetchHistory = async () => {
      try {
        const response = await fetch(
          'http://localhost:8080/api/chat/ai/history',
          {
            headers: {
              Authorization: `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
          }
        );

        if (!response.ok) throw new Error(`HTTP Error: ${response.status}`);

        const data = await response.json();

        const formattedMessages = data.map((item) => ({
          id: item.timestamp,
          content: item.content,
          isSender: item.user === true || item.isUser === true, // DTO次第で両対応
        }));

        console.log('リクエスト成功', response.status);
        setMessages(formattedMessages);
      } catch (err) {
        console.error('履歴取得失敗:', err);
      }
    };

    fetchHistory(); // 最初に呼び出し

    // ② WebSocket 接続

    wsRef.current = new WebSocket(
      `${
        import.meta.env.VITE_WEB_SOCKET_URL
      }?user_id=${senderId}&room_id=default`
    );

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

    // ③ クリーンアップ
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [senderId]);

  // --- メッセージ送信 ---
  const handleSend = (text) => {
    const timestampNow = Date.now();

    // ① UI即時反映
    setMessages((prev) => [
      ...prev,
      {
        id: timestampNow,
        content: text,
        isSender: true,
      },
    ]);

    // ② WebSocket 送信
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
    <div className="flex flex-col h-screen bg-gray-100 text-black">
      <div className="flex-1 overflow-y-auto px-4 py-6 space-y-2 max-w-3xl mx-auto w-full">
        {messages.map((msg) => (
          <MessageBubble key={msg.id} {...msg} />
        ))}
      </div>
      <MessageInput onSend={handleSend} />
    </div>
  );
}
