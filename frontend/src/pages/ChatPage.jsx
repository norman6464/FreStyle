import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import MessageBubble from '../components/MessageBubble';
import MessageInput from '../components/MessageInput';
import HamburgerMenu from '../components/HamburgerMenu';
import { useDispatch } from 'react-redux';
import { clearAuth } from '../store/authSlice';

import SockJS from 'sockjs-client';
import { Client } from '@stomp/stompjs';

export default function ChatPage() {
  const [messages, setMessages] = useState([]);
  const [senderId, setSenderId] = useState(null);
  const stompClientRef = useRef(null);
  const messagesEndRef = useRef(null);

  const { roomId } = useParams();

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
      } catch (error) {
        console.error('ユーザー情報取得エラー:', error);
        navigate('/login');
      }
    };

    fetchUserInfo();
  }, [API_BASE_URL, navigate]);

  // スクロール
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

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

      const formatted = data.map((msg) => ({
        id: msg.id,
        roomId: msg.roomId,
        senderId: msg.senderId,
        senderName: msg.senderName,
        content: msg.content,
        createdAt: msg.createdAt,
        isSender: msg.isSender, // バックエンドから直接取得
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
  const handleSend = (text) => {
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

    // ローカルでも先に追加（楽観的 UI 更新）
    setMessages((prev) => [
      ...prev,
      {
        id: Date.now(), // 仮のID
        roomId,
        senderId,
        senderName: '自分',
        content: text,
        createdAt: new Date().toISOString(),
        isSender: true,
      },
    ]);
  };

  // --- メッセージ削除（拡張用） ---
  const handleDeleteMessage = (messageId) => {
    if (!confirm('このメッセージを削除しますか？')) return;

    setMessages((prev) =>
      prev.map((m) => (m.id === messageId ? { ...m, isDeleted: true } : m))
    );

    stompClientRef.current.publish({
      destination: '/app/chat/delete',
      body: JSON.stringify({
        messageId,
        roomId,
        senderId,
      }),
    });
  };

  // --- AIフィードバック ---
  const handleAiFeedback = () => {
    const chatHistory = messages
      .map((msg) => `${msg.isSender ? '自分' : '相手'}: ${msg.content}`)
      .join('\n');

    navigate('/chat/ask-ai', {
      state: {
        initialPrompt: `【チャット履歴】\n${chatHistory}`,
      },
    });
  };

  return (
    <>
      <HamburgerMenu title="個人チャット" />

      <div className="flex flex-col h-screen bg-gradient-to-br from-gray-50 to-primary-50 text-black pt-16">
        <div className="flex-1 overflow-y-auto px-4 py-6 space-y-2 max-w-4xl mx-auto w-full pb-[120px]">
          {messages.length === 0 && (
            <div className="flex items-center justify-center h-full">
              <p className="text-gray-400 text-lg">メッセージがありません</p>
            </div>
          )}

          {messages.map((msg) => (
            <MessageBubble
              key={msg.id}
              {...msg}
              onDelete={handleDeleteMessage}
            />
          ))}

          <div ref={messagesEndRef} />
        </div>

        <div className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-100 shadow-2xl p-4 z-10">
          <div className="max-w-4xl mx-auto w-full space-y-3">
            {messages.length > 0 && (
              <button
                onClick={handleAiFeedback}
                className="w-full bg-gradient-to-r from-blue-500 to-purple-500 text-white font-semibold py-3 px-4 rounded-lg"
              >
                AIにフィードバックしてもらう
              </button>
            )}
            <MessageInput onSend={handleSend} />
          </div>
        </div>
      </div>
    </>
  );
}
