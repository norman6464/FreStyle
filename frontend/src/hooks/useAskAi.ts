import { useEffect, useRef, useMemo, useState, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { useAuth } from './useAuth';
import { useAiChat } from './useAiChat';
import { useWebSocketNative } from './useWebSocketNative';
import { useAiSession } from './useAiSession';
import { WS } from '../constants/apiRoutes';

/**
 * AskAiPage フック（汎用 AI チャット版）。
 *
 * 旧版は「練習モード / scoreCard / シナリオ受け渡し / chat-feedback」など複合機能を
 * 引き受けていたが、PR-A の機能整理でアプリは「自由対話の AI チャット」に集約した。
 * このフックも以下のみを扱う:
 *   - セッション一覧 / 検索 / 切替
 *   - メッセージ送受信（WebSocket; SSE への置換は PR-C）
 *   - 編集中タイトルやモーダル状態
 *
 * 削除した機能: 練習モード / scoreCard / scenario / chat-feedback / scene
 */
export function useAskAi() {
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const [sessionSearchQuery, setSessionSearchQuery] = useState('');
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  const { sessionId: urlSessionId } = useParams<{ sessionId: string }>();

  const { getCurrentUser, user } = useAuth();
  const {
    sessions,
    messages,
    loading,
    fetchSessions,
    fetchMessages,
    deleteSession,
    updateSessionTitle,
    handleIncomingMessage,
    handleIncomingSession,
  } = useAiChat();

  const aiSession = useAiSession({ deleteSession, updateSessionTitle });
  const { currentSessionId, setCurrentSessionId } = aiSession;

  const wsUrl = user?.id && API_BASE_URL
    ? toWsUrl(`${API_BASE_URL}${WS.aiChat}`)
    : null;

  type AiWsInbound =
    | { type: 'message'; sessionId?: number; [k: string]: unknown }
    | { type: 'session'; id: number; [k: string]: unknown }
    | { type: 'session-deleted'; sessionId: number };

  const { send } = useWebSocketNative({
    url: wsUrl,
    onMessage: (raw) => {
      const data = raw as AiWsInbound;
      if (data.type === 'message') {
        handleIncomingMessage(data as unknown as Parameters<typeof handleIncomingMessage>[0]);
        return;
      }
      if (data.type === 'session') {
        handleIncomingSession(data);
        setCurrentSessionId(data.id);
        return;
      }
      if (data.type === 'session-deleted') {
        if (currentSessionId === data.sessionId) {
          setCurrentSessionId(null);
        }
      }
    },
  });

  const handleSend = useCallback((text: string): void => {
    send({
      type: 'send',
      sessionId: currentSessionId,
      content: text,
      role: 'user',
    });
  }, [send, currentSessionId]);

  // ユーザー情報取得
  useEffect(() => {
    getCurrentUser();
  }, [getCurrentUser]);

  // セッション一覧取得
  useEffect(() => {
    if (user?.id) {
      fetchSessions();
    }
  }, [user?.id, fetchSessions]);

  // URL パラメータのセッション ID 変更で current を切替（無いときは null にリセット）
  useEffect(() => {
    if (urlSessionId) {
      setCurrentSessionId(parseInt(urlSessionId, 10));
    } else {
      setCurrentSessionId(null);
    }
  }, [urlSessionId, setCurrentSessionId]);

  // セッション内のメッセージ履歴取得
  useEffect(() => {
    if (currentSessionId) {
      fetchMessages(currentSessionId);
    }
  }, [currentSessionId, fetchMessages]);

  // メッセージ最下部へスクロール
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const filteredSessions = useMemo(() => {
    if (!sessionSearchQuery) return sessions;
    const query = sessionSearchQuery.toLowerCase();
    return sessions.filter((s) => s.title?.toLowerCase().includes(query));
  }, [sessions, sessionSearchQuery]);

  return {
    sessions,
    filteredSessions,
    messages,
    loading,
    messagesEndRef,
    sessionSearchQuery,
    setSessionSearchQuery,

    ...aiSession,

    handleSend,
  };
}

function toWsUrl(httpUrl: string): string {
  return httpUrl.replace(/^http:/, 'ws:').replace(/^https:/, 'wss:');
}
