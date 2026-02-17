import { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import SockJS from 'sockjs-client';
import { Client } from '@stomp/stompjs';
import ChatRepository from '../repositories/ChatRepository';
import { ChatMessage } from '../types';
import { useMessageSelection } from './useMessageSelection';

/**
 * チャットページのコアロジックフック
 *
 * ChatPageからビジネスロジックを分離し、
 * メッセージ管理・WebSocket接続・選択モード・言い換え提案を担う。
 */
export function useChat() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [senderId, setSenderId] = useState<number | null>(null);
  const [deleteModal, setDeleteModal] = useState<{ isOpen: boolean; messageId: number | null }>({ isOpen: false, messageId: null });
  const [showSceneSelector, setShowSceneSelector] = useState(false);
  const [showRephraseModal, setShowRephraseModal] = useState(false);
  const [loading, setLoading] = useState(true);
  const [rephraseResult, setRephraseResult] = useState<{ formal: string; soft: string; concise: string } | null>(null);
  const [rephraseOriginalText, setRephraseOriginalText] = useState('');
  const stompClientRef = useRef<Client | null>(null);
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
      // エラーはaxiosインターセプターが処理
    } finally {
      setLoading(false);
    }
  }, [roomId, senderId]);

  // WebSocket接続
  useEffect(() => {
    if (!senderId) return;
    const client = new Client({
      webSocketFactory: () =>
        new SockJS(`${API_BASE_URL}/ws/chat`, undefined, { withCredentials: true }),
      reconnectDelay: 5000,
      onConnect: () => {
        client.publish({
          destination: '/app/auth',
          body: JSON.stringify({ userId: senderId }),
        });
        client.subscribe(`/topic/chat/${roomId}`, (message) => {
          const data = JSON.parse(message.body);
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
    return () => { client.deactivate(); };
  }, [roomId, senderId, fetchHistory, API_BASE_URL]);

  // メッセージ送信
  const handleSend = useCallback((text: string) => {
    if (!stompClientRef.current?.connected) return;
    stompClientRef.current.publish({
      destination: '/app/chat/send',
      body: JSON.stringify({ roomId, senderId, content: text }),
    });
  }, [roomId, senderId]);

  // メッセージ削除
  const handleDeleteMessage = useCallback((messageId: number) => {
    setDeleteModal({ isOpen: true, messageId });
  }, []);

  const confirmDelete = useCallback(() => {
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
  }, [deleteModal.messageId, roomId, senderId]);

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
      alert('メッセージを選択してください');
      return;
    }
    setShowSceneSelector(true);
  }, [selection.selectedMessages]);

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
