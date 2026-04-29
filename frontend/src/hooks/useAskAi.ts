import { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { useLocation, useParams } from 'react-router-dom';
import { useAuth } from './useAuth';
import { useAiChat } from './useAiChat';
import { useWebSocketNative } from './useWebSocketNative';
import { useAiSession } from './useAiSession';
import { WS } from '../constants/apiRoutes';

/**
 * AskAiPage フック
 *
 * SockJS / STOMP 廃止に伴い、native WebSocket + JSON プロトコルへ移行した。
 * 旧 destination ベースの subscribe は廃止し、受信メッセージ側で `type` を見て分岐する。
 *
 * 送受信プロトコル（暫定）:
 * - 送信: { type: "send", sessionId, content, role, scene?, sessionType?, scenarioId?, fromChatFeedback }
 * - 受信: { type: "message" | "session" | "scorecard" | "session-deleted", ... }
 *
 * Bedrock 連携が backend 側に未実装のため、現時点では echo になっても UI が壊れないように扱う。
 */
export function useAskAi() {
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const [initialPromptSent, setInitialPromptSent] = useState(false);
  const [sessionSearchQuery, setSessionSearchQuery] = useState('');
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  const location = useLocation();
  const { sessionId: urlSessionId } = useParams<{ sessionId: string }>();

  const { getCurrentUser, user } = useAuth();
  const {
    sessions,
    messages,
    scoreCard,
    loading,
    fetchSessions,
    fetchMessages,
    deleteSession,
    updateSessionTitle,
    handleIncomingMessage,
    handleIncomingScoreCard,
    handleIncomingSession,
  } = useAiChat();

  const aiSession = useAiSession({ deleteSession, updateSessionTitle });
  const { currentSessionId, setCurrentSessionId } = aiSession;

  const locationState = location.state as {
    initialPrompt?: string;
    fromChatFeedback?: boolean;
    scene?: string;
    sessionType?: string;
    scenarioId?: number;
    scenarioName?: string;
  } | null;

  const initialPrompt = locationState?.initialPrompt;
  const fromChatFeedback = locationState?.fromChatFeedback || false;
  const scene = locationState?.scene || null;
  const sessionType = locationState?.sessionType || 'normal';
  const scenarioId = locationState?.scenarioId || null;
  const scenarioName = locationState?.scenarioName || null;
  const isPracticeMode = sessionType === 'practice';

  const wsUrl = user?.id && API_BASE_URL
    ? toWsUrl(`${API_BASE_URL}${WS.aiChat}`)
    : null;

  type AiWsInbound =
    | { type: 'message'; sessionId?: number; [k: string]: unknown }
    | { type: 'session'; id: number; [k: string]: unknown }
    | { type: 'scorecard'; sessionId?: number; [k: string]: unknown }
    | { type: 'session-deleted'; sessionId: number };

  const { send } = useWebSocketNative({
    url: wsUrl,
    onOpen: () => {
      if (initialPrompt && !initialPromptSent) {
        sendMessage(initialPrompt);
        setInitialPromptSent(true);
      }
    },
    onMessage: (raw) => {
      const data = raw as AiWsInbound;
      if (data.type === 'message') {
        // backend からの WS 形状は AiMessage と互換性があるため、type フィールドを除いて
        // 既存 handler に渡す。AiMessage は domain.AiMessage と 1:1 (DOP)。
        handleIncomingMessage(data as unknown as Parameters<typeof handleIncomingMessage>[0]);
        return;
      }
      if (data.type === 'session') {
        handleIncomingSession(data);
        setCurrentSessionId(data.id);
        return;
      }
      if (data.type === 'scorecard') {
        handleIncomingScoreCard(data as unknown as Parameters<typeof handleIncomingScoreCard>[0]);
        return;
      }
      if (data.type === 'session-deleted') {
        if (currentSessionId === data.sessionId) {
          setCurrentSessionId(null);
        }
      }
    },
  });

  const sendMessage = useCallback((text: string) => {
    const payload: Record<string, unknown> = {
      type: 'send',
      sessionId: currentSessionId,
      content: text,
      role: 'user',
      fromChatFeedback,
    };
    if (scene) payload.scene = scene;
    if (isPracticeMode && scenarioId) {
      payload.sessionType = 'practice';
      payload.scenarioId = scenarioId;
    }
    send(payload);
  }, [send, currentSessionId, fromChatFeedback, scene, isPracticeMode, scenarioId]);

  const handleSend = useCallback((text: string): void => {
    sendMessage(text);
  }, [sendMessage]);

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

  // URLパラメータのセッションID変更時
  useEffect(() => {
    if (urlSessionId) {
      setCurrentSessionId(parseInt(urlSessionId));
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

  // メッセージ削除処理
  const handleDeleteMessage = (_messageId: string): void => {
    // ローカル状態のみ更新（サーバ側削除は別 issue）
  };

  return {
    sessions,
    filteredSessions,
    messages,
    scoreCard,
    loading,
    messagesEndRef,
    isPracticeMode,
    scenarioId,
    scenarioName,
    sessionSearchQuery,
    setSessionSearchQuery,

    ...aiSession,

    handleSend,
    handleDeleteMessage,
  };
}

function toWsUrl(httpUrl: string): string {
  return httpUrl.replace(/^http:/, 'ws:').replace(/^https:/, 'wss:');
}
