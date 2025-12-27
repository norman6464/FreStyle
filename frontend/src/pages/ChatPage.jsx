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
  const messagesEndRef = useRef(null); // メッセージの最下部へスクロールするためのRef
  const { roomId } = useParams();
  const senderId = useSelector((state) => state.auth.sub);
  const accessToken = useSelector((state) => state.auth.accessToken);
  const email = useSelector((state) => state.auth.email);
  const navigate = useNavigate();
  const dispatch = useDispatch();

  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  // メッセージリストの最下部にスクロールする関数
  const scrollToBottom = () => {
    // スクロール時に、アニメーションを伴ってスムーズに移動
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  // メッセージが更新されたらスクロール
  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // --- チャット履歴取得（JWT認証＋リフレッシュ対応） ---
  const fetchHistory = async () => {
    try {
      console.log('📡 履歴リクエスト開始');
      console.log('🔍 senderId:', senderId, 'タイプ:', typeof senderId);
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
        console.log('📋 再取得成功。user フィールドで判定します');

        const formattedMessages = retryData.map((msg) => ({
          id: msg.timestamp,
          content: msg.content,
          isSender: msg.user === true,
        }));
        setMessages(formattedMessages);
        console.log('✅ 履歴再取得成功');
        return;
      }

      // 通常成功時
      if (!res.ok) throw new Error(`履歴取得失敗: ${res.status}`);

      const data = await res.json();
      console.log('📋 取得成功。user フィールドで判定します');

      const formattedMessages = data.map((msg) => ({
        id: msg.timestamp,
        timestamp: msg.timestamp,
        content: msg.content,
        isSender: msg.user === true,
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
          timestamp: data.timestamp ?? Date.now(),
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
            room_id: roomId,
            timestamp: messageToDelete.timestamp,
            sender_id: senderId,
          })
        );
      }
    }
  };

  // --- AIにフィードバックをもらう処理 ---
  const handleAiFeedback = () => {
    // チャット履歴をプロンプト用に整形
    const chatHistory = messages
      .map((msg) => {
        const sender = msg.isSender ? '自分' : '相手';
        return `${sender}: ${msg.content}`;
      })
      .join('\n');

    const feedbackPrompt = `以下は私と友人との最近のチャット履歴です。この会話から、私がどのような性格・特性を持っているかを分析して教えてください。

【チャット履歴】
${chatHistory}

この会話に基づいて、以下について教えてください：
1. あなたの会話スタイル
2. 推測される性格特性
3. コミュニケーションの強み
4. より良いコミュニケーションのための改善案`;

    // stateを通じてプロンプトを渡してAIチャット画面へ遷移
    navigate('/chat/ask-ai', { state: { initialPrompt: feedbackPrompt } });
  };

  return (
    <>
      {/* 画面上部の固定ヘッダー */}
      <HamburgerMenu title="個人チャット" />

      {/* メインのコンテナ: ヘッダーの下、入力欄の上全体を使う */}
      <div className="flex flex-col h-screen bg-gradient-to-br from-gray-50 to-primary-50 text-black pt-16">
        {/* メッセージ表示エリア: flex-1 で残りのスペースを全て使い、スクロール可能にする */}
        {/* pb-[100px] は、可変長の入力欄が最大高さになった場合でもメッセージが隠れないようにするためのスペース */}
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
          {/* 最下部スクロール用エレメント */}
          <div ref={messagesEndRef} />
        </div>

        {/* メッセージ入力エリアのコンテナ: 画面下部に固定し、適切なパディングを設定 */}
        <div className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-100 shadow-2xl p-4 z-10">
          {/* MessageInput自体の幅を max-w-4xl に制限し、メッセージバブルの幅に合わせる */}
          <div className="max-w-4xl mx-auto w-full space-y-3">
            {messages.length > 0 && (
              <button
                onClick={handleAiFeedback}
                className="w-full bg-gradient-to-r from-blue-500 to-purple-500 hover:from-blue-600 hover:to-purple-600 text-white font-semibold py-3 px-4 rounded-lg transition-all duration-300 flex items-center justify-center space-x-2 shadow-md"
              >
                <svg
                  className="w-5 h-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M13 10V3L4 14h7v7l9-11h-7z"
                  />
                </svg>
                <span>AIにフィードバックしてもらう</span>
              </button>
            )}
            <MessageInput onSend={handleSend} />
          </div>
        </div>
      </div>
    </>
  );
}
