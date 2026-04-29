import { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import ChatRepository from '../repositories/ChatRepository';
import { WS } from '../constants/apiRoutes';
import { ChatMessage } from '../types';
import { useMessageSelection } from './useMessageSelection';
import { useToast } from './useToast';
import { useWebSocketNative } from './useWebSocketNative';

/**
 * チャットページのコアロジックフック
 *
 * ChatPage からビジネスロジックを分離し、
 * メッセージ管理・WebSocket 接続・選択モード・言い換え提案を担う。
 *
 * WS は SockJS / STOMP を廃止し、native WebSocket + JSON プロトコルで通信する。
 * - 送信: { type: "send", content }
 * - 削除: { type: "delete", createdAtRef }
 * - 受信: { type: "message" | "delete", id, roomId, senderId, senderName, content, createdAt }
 */
export function useChat() {
  const { showToast } = useToast();
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [senderId, setSenderId] = useState<number | null>(null);
  const [deleteModal, setDeleteModal] = useState<{ isOpen: boolean; messageId: string | null }>({ isOpen: false, messageId: null });
  const [showSceneSelector, setShowSceneSelector] = useState(false);
  const [showRephraseModal, setShowRephraseModal] = useState(false);
  const [loading, setLoading] = useState(true);
  const [rephraseResult, setRephraseResult] = useState<{ formal: string; soft: string; concise: string } | null>(null);
  const [rephraseOriginalText, setRephraseOriginalText] = useState('');
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  const { roomId } = useParams<{ roomId: string }>();
  const navigate = useNavigate();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const selection = useMessageSelection(messages);

  // ユーザー情報取得
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
  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages, scrollToBottom]);

  // 未読カウントリセット
  useEffect(() => {
    if (!roomId || !senderId) return;
    ChatRepository.markAsRead(roomId).catch(() => {});
  }, [roomId, senderId]);

  // チャット履歴取得
  const fetchHistory = useCallback(async () => {
    try {
      const data = await ChatRepository.fetchHistory(roomId!);
      if (!Array.isArray(data)) return;
      const formatted = data.map((msg: ChatMessage) => ({
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
      // エラーは axios インターセプターが処理
    } finally {
      setLoading(false);
    }
  }, [roomId, senderId]);

  // WebSocket URL は http(s) → ws(s) に変換する。VITE_API_BASE_URL は http(s) で定義されている。
  const wsUrl = roomId && senderId && API_BASE_URL
    ? toWsUrl(`${API_BASE_URL}${WS.chatRoom(roomId)}`)
    : null;

  type ChatWsInbound =
    | { type: 'message'; id: string; roomId: string | number; senderId: number; senderName: string; content: string; createdAt: string }
    | { type: 'delete'; createdAt: string };

  const { send } = useWebSocketNative({
    url: wsUrl,
    onOpen: () => {
      fetchHistory();
    },
    onMessage: (raw) => {
      const data = raw as ChatWsInbound;
      if (data.type === 'delete') {
        setMessages((prev) =>
          prev.map((m) => (m.createdAt === data.createdAt ? { ...m, isDeleted: true } : m))
        );
        return;
      }
      if (data.type === 'message') {
        setMessages((prev) => [
          ...prev,
          {
            id: data.id,
            roomId: typeof data.roomId === 'string' ? parseInt(data.roomId, 10) : data.roomId,
            senderId: data.senderId,
            senderName: data.senderName,
            content: data.content,
            createdAt: data.createdAt,
            isSender: data.senderId === senderId,
          },
        ]);
      }
    },
    onError: () => setLoading(false),
    onClose: () => setLoading(false),
  });

  // メッセージ送信
  const handleSend = useCallback((text: string) => {
    send({ type: 'send', content: text });
  }, [send]);

  // メッセージ削除
  const handleDeleteMessage = useCallback((messageId: string) => {
    setDeleteModal({ isOpen: true, messageId });
  }, []);

  const confirmDelete = useCallback(() => {
    const messageId = deleteModal.messageId;
    setDeleteModal({ isOpen: false, messageId: null });
    if (!messageId) return;
    const msg = messages.find((m) => m.id === messageId);
    if (!msg?.createdAt) return;
    send({ type: 'delete', createdAtRef: msg.createdAt });
  }, [deleteModal.messageId, messages, send]);

  const cancelDelete = useCallback(() => {
    setDeleteModal({ isOpen: false, messageId: null });
  }, []);

  // AI選択モード
  const handleAiFeedback = useCallback(() => {
    selection.enterSelectionMode();
  }, [selection]);

  const handleCancelSelection = useCallback(() => {
    selection.cancelSelection();
  }, [selection]);

  const handleSendToAi = useCallback(() => {
    if (selection.selectedMessages.size === 0) {
      showToast('info', 'メッセージを選択してください');
      return;
    }
    setShowSceneSelector(true);
    // selection.selectedMessages は selection オブジェクト経由で参照しているため
    // ネスト依存は冗長。selection 全体を deps にすれば済む。
  }, [selection, showToast]);

  const handleSceneSelect = useCallback((scene: string | null) => {
    setShowSceneSelector(false);
    const selectedMsgs = messages.filter((msg) => selection.selectedMessages.has(msg.id));
    const chatHistory = selectedMsgs
      .map((msg) => `${msg.isSender ? '自分' : '相手'}: ${msg.content}`)
      .join('\n');
    selection.cancelSelection();
    navigate('/chat/ask-ai', {
      state: {
        initialPrompt: `【選択したチャット履歴】\n${chatHistory}`,
        fromChatFeedback: true,
        scene: scene,
      },
    });
  }, [messages, selection, navigate]);

  // 言い換え提案
  const handleRephrase = useCallback(async (content: string) => {
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
  }, []);

  return {
    messages,
    senderId,
    loading,
    deleteModal,
    selectionMode: selection.selectionMode,
    selectedMessages: selection.selectedMessages,
    showSceneSelector,
    showRephraseModal,
    rephraseResult,
    rephraseOriginalText,
    messagesEndRef,
    handleSend,
    handleDeleteMessage,
    confirmDelete,
    cancelDelete,
    handleAiFeedback,
    handleRangeClick: selection.handleRangeClick,
    handleQuickSelect: selection.handleQuickSelect,
    handleSelectAll: selection.handleSelectAll,
    handleDeselectAll: selection.handleDeselectAll,
    handleCancelSelection,
    handleSendToAi,
    handleSceneSelect,
    handleRephrase,
    setShowRephraseModal,
    isInRange: selection.isInRange,
    getRangeLabel: selection.getRangeLabel,
  };
}

/**
 * http(s) ベースの BASE_URL を WebSocket URL (ws/wss) に変換する。
 */
function toWsUrl(httpUrl: string): string {
  return httpUrl.replace(/^http:/, 'ws:').replace(/^https:/, 'wss:');
}
