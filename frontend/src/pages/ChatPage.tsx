import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import MessageBubble from '../components/MessageBubble';
import MessageInput from '../components/MessageInput';
import { ChatMessage } from '../types';

import ConfirmModal from '../components/ConfirmModal';
import SceneSelector from '../components/SceneSelector';
import RephraseModal from '../components/RephraseModal';
import ChatRepository from '../repositories/ChatRepository';

import SockJS from 'sockjs-client';
import { Client } from '@stomp/stompjs';

export default function ChatPage() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [senderId, setSenderId] = useState<number | null>(null);
  const [deleteModal, setDeleteModal] = useState<{ isOpen: boolean; messageId: number | null }>({ isOpen: false, messageId: null });
  const [selectionMode, setSelectionMode] = useState(false);
  const [selectedMessages, setSelectedMessages] = useState<Set<number>>(new Set());
  const [rangeStart, setRangeStart] = useState<number | null>(null);
  const [rangeEnd, setRangeEnd] = useState<number | null>(null);
  const [showSceneSelector, setShowSceneSelector] = useState(false);
  const [showRephraseModal, setShowRephraseModal] = useState(false);
  const [rephraseResult, setRephraseResult] = useState<{ formal: string; soft: string; concise: string } | null>(null);
  const [rephraseOriginalText, setRephraseOriginalText] = useState('');
  const stompClientRef = useRef<Client | null>(null);
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  const { roomId } = useParams<{ roomId: string }>();

  const navigate = useNavigate();

  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  // ユーザー情報取得（senderId を取得）
  useEffect(() => {
    const fetchUserInfo = async () => {
      try {
        const data = await ChatRepository.fetchCurrentUser();
        setSenderId(data.id);
      } catch {
        navigate('/login');
      }
    };

    fetchUserInfo();
  }, [navigate]);

  // スクロール
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // チャットルーム開封時に未読カウントをリセット
  useEffect(() => {
    if (!roomId || !senderId) return;
    ChatRepository.markAsRead(roomId).catch(() => {});
  }, [roomId, senderId]);

  // --- チャット履歴取得 ---
  const fetchHistory = async () => {
    try {
      const data = await ChatRepository.fetchHistory(roomId!);
      if (!Array.isArray(data)) return;

      const formatted = data.map((msg: any) => ({
        id: msg.id,
        roomId: msg.roomId,
        senderId: msg.senderId,
        senderName: msg.senderName,
        content: msg.content,
        createdAt: msg.createdAt,
        isSender: msg.senderId === senderId,
      }));

      setMessages(formatted);
    } catch {
      // エラーはaxiosインターセプターが処理
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
        // 認証メッセージを送信
        client.publish({
          destination: '/app/auth',
          body: JSON.stringify({ userId: senderId }),
        });

        // ルーム購読
        client.subscribe(`/topic/chat/${roomId}`, (message) => {
          const data = JSON.parse(message.body);

          // 削除通知の処理
          if (data.type === 'delete') {
            setMessages((prev) =>
              prev.map((m) =>
                m.id === data.messageId ? { ...m, isDeleted: true } : m
              )
            );
            return;
          }

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

      onStompError: () => {},
    });

    stompClientRef.current = client;
    client.activate();

    return () => {
      client.deactivate();
    };
  }, [roomId, senderId]);

  // --- メッセージ送信 ---
  const handleSend = (text: string) => {
    if (!stompClientRef.current?.connected) return;

    stompClientRef.current.publish({
      destination: '/app/chat/send',
      body: JSON.stringify({ roomId, senderId, content: text }),
    });
  };

  // --- メッセージ削除 ---
  const handleDeleteMessage = (messageId: number) => {
    setDeleteModal({ isOpen: true, messageId });
  };

  const confirmDelete = () => {
    const messageId = deleteModal.messageId;
    setDeleteModal({ isOpen: false, messageId: null });

    if (!stompClientRef.current?.connected) return;

    stompClientRef.current.publish({
      destination: '/app/chat/delete',
      body: JSON.stringify({
        messageId,
        roomId: parseInt(roomId!, 10),
        senderId,
      }),
    });
  };

  const cancelDelete = () => {
    setDeleteModal({ isOpen: false, messageId: null });
  };

  // --- AIフィードバック ---
  const handleAiFeedback = () => {
    setSelectionMode(true);
    setSelectedMessages(new Set());
    setRangeStart(null);
    setRangeEnd(null);
  };

  // 範囲選択: メッセージをクリックしたとき
  const handleRangeClick = (messageId: number) => {
    const messageIndex = messages.findIndex((msg) => msg.id === messageId);

    if (rangeStart === null) {
      setRangeStart(messageIndex);
      setRangeEnd(null);
      setSelectedMessages(new Set([messageId]));
    } else if (rangeEnd === null) {
      setRangeEnd(messageIndex);
      const start = Math.min(rangeStart, messageIndex);
      const end = Math.max(rangeStart, messageIndex);
      const rangeIds = new Set(
        messages.slice(start, end + 1).map((msg) => msg.id)
      );
      setSelectedMessages(rangeIds);
    } else {
      setRangeStart(messageIndex);
      setRangeEnd(null);
      setSelectedMessages(new Set([messageId]));
    }
  };

  // クイック選択: 直近N件
  const handleQuickSelect = (count: number) => {
    const recentMessages = messages.slice(-count);
    const recentIds = new Set(recentMessages.map((msg) => msg.id));
    setSelectedMessages(recentIds);
    if (recentMessages.length > 0) {
      setRangeStart(messages.length - count);
      setRangeEnd(messages.length - 1);
    }
  };

  const handleSelectAll = () => {
    const allIds = new Set(messages.map((msg) => msg.id));
    setSelectedMessages(allIds);
    setRangeStart(0);
    setRangeEnd(messages.length - 1);
  };

  const handleDeselectAll = () => {
    setSelectedMessages(new Set());
    setRangeStart(null);
    setRangeEnd(null);
  };

  const handleCancelSelection = () => {
    setSelectionMode(false);
    setSelectedMessages(new Set());
    setRangeStart(null);
    setRangeEnd(null);
  };

  const handleSendToAi = () => {
    if (selectedMessages.size === 0) {
      alert('メッセージを選択してください');
      return;
    }
    setShowSceneSelector(true);
  };

  const handleSceneSelect = (scene: string | null) => {
    setShowSceneSelector(false);

    const selectedMsgs = messages.filter((msg) => selectedMessages.has(msg.id));
    const chatHistory = selectedMsgs
      .map((msg) => `${msg.isSender ? '自分' : '相手'}: ${msg.content}`)
      .join('\n');

    setSelectionMode(false);
    setSelectedMessages(new Set());
    setRangeStart(null);
    setRangeEnd(null);

    navigate('/chat/ask-ai', {
      state: {
        initialPrompt: `【選択したチャット履歴】\n${chatHistory}`,
        fromChatFeedback: true,
        scene: scene,
      },
    });
  };

  // --- 言い換え提案 ---
  const handleRephrase = async (content: string) => {
    setRephraseResult(null);
    setRephraseOriginalText(content);
    setShowRephraseModal(true);

    try {
      const data = await ChatRepository.rephrase(content, null);
      try {
        const parsed = JSON.parse(data.result);
        setRephraseResult(parsed);
      } catch {
        setRephraseResult({ formal: data.result, soft: '', concise: '' });
      }
    } catch {
      setShowRephraseModal(false);
    }
  };

  // メッセージが範囲内かどうかを判定
  const isInRange = (index: number): boolean => {
    if (rangeStart === null) return false;
    if (rangeEnd === null) return index === rangeStart;
    const start = Math.min(rangeStart, rangeEnd);
    const end = Math.max(rangeStart, rangeEnd);
    return index >= start && index <= end;
  };

  // 範囲選択のラベルを取得
  const getRangeLabel = (index: number): string | null => {
    if (rangeStart === index && rangeEnd === null) return '開始';
    if (rangeEnd === null) return null;
    const start = Math.min(rangeStart!, rangeEnd);
    const end = Math.max(rangeStart!, rangeEnd);
    if (index === start) return '開始';
    if (index === end) return '終了';
    return null;
  };

  return (
    <div className="flex flex-col h-full">
      {/* メッセージエリア */}
      <div className="flex-1 overflow-y-auto px-4 py-4 space-y-3">
        {messages.length === 0 && (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <div className="bg-slate-100 rounded-full p-4 mb-4">
              <svg className="w-8 h-8 text-slate-400" fill="currentColor" viewBox="0 0 20 20">
                <path d="M2 5a2 2 0 012-2h12a2 2 0 012 2v10a2 2 0 01-2 2H4a2 2 0 01-2-2V5z" />
              </svg>
            </div>
            <h3 className="text-base font-semibold text-slate-700 mb-1">
              チャットへようこそ
            </h3>
            <p className="text-sm text-slate-500">
              相手とのチャットをここで行えます
            </p>
          </div>
        )}
        {messages.map((msg, index) => (
          <div key={msg.id} className="flex items-start gap-2 max-w-3xl mx-auto w-full">
            {selectionMode && (
              <div className="flex-shrink-0 flex flex-col items-center">
                {getRangeLabel(index) && (
                  <span className={`text-xs font-bold mb-1 px-2 py-0.5 rounded ${
                    getRangeLabel(index) === '開始'
                      ? 'bg-green-100 text-green-700'
                      : 'bg-red-100 text-red-700'
                  }`}>
                    {getRangeLabel(index)}
                  </span>
                )}
                <button
                  onClick={() => handleRangeClick(msg.id)}
                  className={`mt-1 w-7 h-7 rounded-full border-2 flex items-center justify-center transition-all ${
                    isInRange(index)
                      ? 'bg-primary-500 border-primary-500 text-white'
                      : 'border-slate-300 hover:border-primary-400 hover:bg-primary-50'
                  }`}
                >
                  {isInRange(index) ? (
                    <svg className="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                  ) : (
                    <span className="text-[10px] text-slate-400">{index + 1}</span>
                  )}
                </button>
              </div>
            )}
            <div className={`flex-1 transition-all ${
              selectionMode && isInRange(index) ? 'bg-primary-50 -mx-2 px-2 py-1 rounded-lg' : ''
            }`}>
              <MessageBubble
                {...msg}
                onDelete={selectionMode ? null : handleDeleteMessage}
                onRephrase={selectionMode ? null : handleRephrase}
              />
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* 入力欄 */}
      <div className="bg-white border-t border-slate-200 p-4">
        <div className="max-w-3xl mx-auto w-full space-y-3">
          {selectionMode ? (
            <div className="space-y-3">
              {/* ガイドメッセージ */}
              <div className="bg-primary-50 border border-primary-200 rounded-lg p-3">
                <p className="text-sm text-primary-700">
                  {rangeStart === null
                    ? '開始位置のメッセージをタップしてください'
                    : rangeEnd === null
                    ? '終了位置のメッセージをタップしてください'
                    : `${selectedMessages.size}件のメッセージを選択しました`
                  }
                </p>
              </div>

              {/* クイック選択ボタン */}
              <div className="flex gap-2 flex-wrap">
                <span className="text-xs text-slate-500 self-center">クイック選択:</span>
                {[5, 10, 20].map((n) => (
                  <button
                    key={n}
                    onClick={() => handleQuickSelect(n)}
                    className="px-3 py-1 text-xs bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-full transition-colors"
                  >
                    直近{n}件
                  </button>
                ))}
                <button
                  onClick={handleSelectAll}
                  className="px-3 py-1 text-xs bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-full transition-colors"
                >
                  すべて
                </button>
                <button
                  onClick={handleDeselectAll}
                  className="px-3 py-1 text-xs text-rose-500 hover:bg-rose-50 rounded-full transition-colors"
                >
                  リセット
                </button>
              </div>

              {/* アクションボタン */}
              <div className="flex gap-2">
                <button
                  onClick={handleCancelSelection}
                  className="flex-1 bg-slate-200 hover:bg-slate-300 text-slate-700 font-medium py-2.5 px-4 rounded-lg text-sm transition-colors"
                >
                  キャンセル
                </button>
                <button
                  onClick={handleSendToAi}
                  disabled={selectedMessages.size === 0}
                  className={`flex-1 font-medium py-2.5 px-4 rounded-lg text-sm transition-colors ${
                    selectedMessages.size > 0
                      ? 'bg-primary-500 hover:bg-primary-600 text-white'
                      : 'bg-slate-300 text-slate-500 cursor-not-allowed'
                  }`}
                >
                  {selectedMessages.size > 0
                    ? `${selectedMessages.size}件をAIに送信`
                    : '範囲を選択してください'
                  }
                </button>
              </div>
            </div>
          ) : (
            <>
              {messages.length > 0 && (
                <button
                  onClick={handleAiFeedback}
                  className="w-full bg-primary-500 hover:bg-primary-600 text-white font-medium py-2.5 px-4 rounded-lg text-sm transition-colors"
                >
                  AIにフィードバックしてもらう
                </button>
              )}
              <MessageInput onSend={handleSend} />
            </>
          )}
        </div>
      </div>

      {/* シーン選択モーダル */}
      {showSceneSelector && (
        <SceneSelector
          onSelect={(sceneId) => handleSceneSelect(sceneId)}
          onCancel={() => handleSceneSelect(null)}
        />
      )}

      {/* 言い換え提案モーダル */}
      {showRephraseModal && (
        <RephraseModal
          result={rephraseResult}
          onClose={() => setShowRephraseModal(false)}
          originalText={rephraseOriginalText}
        />
      )}

      {/* 削除確認モーダル */}
      <ConfirmModal
        isOpen={deleteModal.isOpen}
        title="メッセージを削除"
        message="このメッセージを削除しますか？この操作は取り消せません。"
        confirmText="削除する"
        cancelText="キャンセル"
        onConfirm={confirmDelete}
        onCancel={cancelDelete}
        isDanger={true}
      />
    </div>
  );
}
