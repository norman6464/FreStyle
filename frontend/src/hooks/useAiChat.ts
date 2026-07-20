import { useState, useCallback } from 'react';
import { classifyApiError } from '@/shared/lib/classifyApiError';
import AiChatRepository, {
  CreateSessionRequest,
  UpdateSessionTitleRequest,
} from '@/entities/ai-chat/api/aiChatRepository';
import type { AiSession, AiMessage } from '@/entities/ai-chat';

/**
 * AI チャットのセッション・メッセージ管理フック。
 *
 * 旧版にあった「scoreCard / rephrase / scoreHistory / addMessage（HTTP 経由）」は
 * すべて廃止して、純粋な「セッション CRUD + WebSocket 受信時のローカル state 反映」だけに
 * 限定する。AI 応答は WebSocket（PR-C で SSE に置換予定）経由でリアルタイムに追加される。
 */
export const useAiChat = () => {
  const [sessions, setSessions] = useState<AiSession[]>([]);
  const [currentSession, setCurrentSession] = useState<AiSession | null>(null);
  const [messages, setMessages] = useState<AiMessage[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchSessions = useCallback(async (): Promise<void> => {
    setLoading(true);
    setError(null);
    try {
      const data = await AiChatRepository.getSessions();
      setSessions(data);
    } catch (err) {
      setError(classifyApiError(err, 'セッション一覧の取得に失敗しました。'));
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchSession = useCallback(async (sessionId: number): Promise<void> => {
    setLoading(true);
    setError(null);
    try {
      const data = await AiChatRepository.getSession(sessionId);
      setCurrentSession(data);
    } catch (err) {
      setError(classifyApiError(err, 'セッション詳細の取得に失敗しました。'));
    } finally {
      setLoading(false);
    }
  }, []);

  const createSession = useCallback(
    async (request: CreateSessionRequest): Promise<AiSession | null> => {
      setLoading(true);
      setError(null);
      try {
        const data = await AiChatRepository.createSession(request);
        setSessions((prev) => [data, ...prev]);
        setCurrentSession(data);
        return data;
      } catch (err) {
        setError(classifyApiError(err, 'セッションの作成に失敗しました。'));
        return null;
      } finally {
        setLoading(false);
      }
    },
    []
  );

  const updateSessionTitle = useCallback(
    async (sessionId: number, request: UpdateSessionTitleRequest): Promise<boolean> => {
      setLoading(true);
      setError(null);
      try {
        const updatedSession = await AiChatRepository.updateSessionTitle(sessionId, request);
        setSessions((prev) =>
          prev.map((session) => (session.id === sessionId ? updatedSession : session))
        );
        if (currentSession?.id === sessionId) {
          setCurrentSession(updatedSession);
        }
        return true;
      } catch (err) {
        setError(classifyApiError(err, 'セッションタイトルの更新に失敗しました。'));
        return false;
      } finally {
        setLoading(false);
      }
    },
    [currentSession]
  );

  const deleteSession = useCallback(async (sessionId: number): Promise<boolean> => {
    setLoading(true);
    setError(null);
    try {
      await AiChatRepository.deleteSession(sessionId);
      setSessions((prev) => prev.filter((session) => session.id !== sessionId));
      if (currentSession?.id === sessionId) {
        setCurrentSession(null);
      }
      return true;
    } catch (err) {
      setError(classifyApiError(err, 'セッションの削除に失敗しました。'));
      return false;
    } finally {
      setLoading(false);
    }
  }, [currentSession]);

  const fetchMessages = useCallback(async (sessionId: number): Promise<void> => {
    setLoading(true);
    setError(null);
    try {
      const data = await AiChatRepository.getMessages(sessionId);
      setMessages(data);
    } catch (err) {
      setError(classifyApiError(err, 'メッセージ一覧の取得に失敗しました。'));
    } finally {
      setLoading(false);
    }
  }, []);

  const handleIncomingMessage = useCallback((message: AiMessage) => {
    setMessages((prev) => {
      if (prev.some((m) => m.id === message.id)) return prev;
      return [...prev, message];
    });
  }, []);

  const handleIncomingSession = useCallback((session: AiSession) => {
    setSessions((prev) => {
      if (prev.some((s) => s.id === session.id)) return prev;
      return [session, ...prev];
    });
  }, []);

  return {
    sessions,
    currentSession,
    messages,
    /** ストリーミング受信中の token append に使う。useAskAi 経由の SSE 取り込みで利用。 */
    setMessages,
    loading,
    error,
    fetchSessions,
    fetchSession,
    createSession,
    updateSessionTitle,
    deleteSession,
    fetchMessages,
    handleIncomingMessage,
    handleIncomingSession,
  };
};
