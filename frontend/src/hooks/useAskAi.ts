import { useState, useEffect, useRef, useMemo } from 'react';
import { useLocation, useParams } from 'react-router-dom';
import { useAuth } from './useAuth';
import { useAiChat } from './useAiChat';
import { useWebSocket } from './useWebSocket';
import { useAiSession } from './useAiSession';

/**
 * AskAiPageフック
 *
 * <p>役割:</p>
 * <ul>
 *   <li>AskAiページのビジネスロジックを管理</li>
 *   <li>WebSocket購読・メッセージ送受信・セッション管理</li>
 * </ul>
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

  // メッセージ送信
  const handleSend = (text: string): void => {
    const payload: Record<string, unknown> = {
      userId: user?.id,
      sessionId: currentSessionId,
      content: text,
      role: 'user',
      fromChatFeedback: fromChatFeedback,
    };

    if (scene) {
      payload.scene = scene;
    }

    if (isPracticeMode && scenarioId) {
      payload.sessionType = 'practice';
      payload.scenarioId = scenarioId;
    }

    publish('/app/ai-chat/send', payload);
  };

  // WebSocket接続
  const { subscribe, publish } = useWebSocket({
    url: `${API_BASE_URL}/ws/ai-chat`,
    userId: user?.id || null,
    onConnect: () => {
      if (currentSessionId) {
        subscribeToSession(currentSessionId);
      }

      if (initialPrompt && !initialPromptSent) {
        handleSend(initialPrompt);
        setInitialPromptSent(true);
      }
    },
  });

  // セッション購読
  const subscribeToSession = (sessionId: number): void => {
    subscribe(`/topic/ai-chat/session/${sessionId}`, (message) => {
      const newMessage = JSON.parse(message.body);
      handleIncomingMessage(newMessage);
    });
  };

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
  }, [urlSessionId]);

  // セッション内のメッセージ履歴取得
  useEffect(() => {
    if (currentSessionId) {
      fetchMessages(currentSessionId);
    }
  }, [currentSessionId, fetchMessages]);

  // WebSocket購読設定
  useEffect(() => {
    if (!user?.id) return;

    subscribe(`/topic/ai-chat/user/${user.id}/session`, (message) => {
      const newSession = JSON.parse(message.body);
      handleIncomingSession(newSession);
      setCurrentSessionId(newSession.id);
    });

    subscribe(`/topic/ai-chat/user/${user.id}/scorecard`, (message) => {
      const data = JSON.parse(message.body);
      handleIncomingScoreCard(data);
    });

    subscribe(`/topic/ai-chat/user/${user.id}/session-deleted`, (message) => {
      const data = JSON.parse(message.body);
      if (currentSessionId === data.sessionId) {
        setCurrentSessionId(null);
      }
    });
  }, [user?.id, subscribe, currentSessionId]);

  // セッション変更時の購読
  useEffect(() => {
    if (currentSessionId) {
      subscribeToSession(currentSessionId);
    }
  }, [currentSessionId]);

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
  const handleDeleteMessage = (_messageId: number): void => {
    // メッセージ削除はローカル状態のみ更新（サーバー側の削除は未実装）
  };

  return {
    // データ
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

    // セッション管理
    ...aiSession,

    // アクション
    handleSend,
    handleDeleteMessage,
  };
}
