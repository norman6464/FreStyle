import { useState, useCallback } from 'react';
import AiChatRepository, {
  CreateSessionRequest,
  UpdateSessionTitleRequest,
  AddMessageRequest,
  RephraseRequest,
} from '../repositories/AiChatRepository';
import { AiSession, AiMessage, ScoreCard, ScoreHistoryItem } from '../types';

/**
 * AI Chatフック
 *
 * <p>役割:</p>
 * <ul>
 *   <li>AI Chatセッション・メッセージ管理</li>
 *   <li>AiChatRepositoryを使用してAPI呼び出し</li>
 * </ul>
 *
 * <p>Hooks層（Presentation Layer - Business Logic）:</p>
 * <ul>
 *   <li>コンポーネントからビジネスロジックを分離</li>
 *   <li>Repository層を使用してAPI呼び出し</li>
 * </ul>
 */
export const useAiChat = () => {
  const [sessions, setSessions] = useState<AiSession[]>([]);
  const [currentSession, setCurrentSession] = useState<AiSession | null>(null);
  const [messages, setMessages] = useState<AiMessage[]>([]);
  const [scoreCard, setScoreCard] = useState<ScoreCard | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  /**
   * セッション一覧を取得
   */
  const fetchSessions = useCallback(async (): Promise<void> => {
    setLoading(true);
    setError(null);

    try {
      const data = await AiChatRepository.getSessions();
      setSessions(data);
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'セッション一覧の取得に失敗しました。';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * セッション詳細を取得
   */
  const fetchSession = useCallback(async (sessionId: number): Promise<void> => {
    setLoading(true);
    setError(null);

    try {
      const data = await AiChatRepository.getSession(sessionId);
      setCurrentSession(data);
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'セッション詳細の取得に失敗しました。';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * 新規セッションを作成
   */
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
        const errorMessage =
          err instanceof Error ? err.message : 'セッションの作成に失敗しました。';
        setError(errorMessage);
        return null;
      } finally {
        setLoading(false);
      }
    },
    []
  );

  /**
   * セッションタイトルを更新
   */
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
        const errorMessage =
          err instanceof Error ? err.message : 'セッションタイトルの更新に失敗しました。';
        setError(errorMessage);
        return false;
      } finally {
        setLoading(false);
      }
    },
    [currentSession]
  );

  /**
   * セッションを削除
   */
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
      const errorMessage =
        err instanceof Error ? err.message : 'セッションの削除に失敗しました。';
      setError(errorMessage);
      return false;
    } finally {
      setLoading(false);
    }
  }, [currentSession]);

  /**
   * セッション内のメッセージ一覧を取得
   */
  const fetchMessages = useCallback(async (sessionId: number): Promise<void> => {
    setLoading(true);
    setError(null);

    try {
      const data = await AiChatRepository.getMessages(sessionId);
      setMessages(data);
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'メッセージ一覧の取得に失敗しました。';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * メッセージを追加
   */
  const addMessage = useCallback(
    async (sessionId: number, request: AddMessageRequest): Promise<AiMessage | null> => {
      setLoading(true);
      setError(null);

      try {
        const data = await AiChatRepository.addMessage(sessionId, request);
        setMessages((prev) => [...prev, data]);
        return data;
      } catch (err) {
        const errorMessage =
          err instanceof Error ? err.message : 'メッセージの追加に失敗しました。';
        setError(errorMessage);
        return null;
      } finally {
        setLoading(false);
      }
    },
    []
  );

  /**
   * メッセージの言い換え提案を取得
   */
  const rephrase = useCallback(
    async (request: RephraseRequest): Promise<string | null> => {
      setLoading(true);
      setError(null);

      try {
        const data = await AiChatRepository.rephrase(request);
        return data.result;
      } catch (err) {
        const errorMessage =
          err instanceof Error ? err.message : '言い換え提案の取得に失敗しました。';
        setError(errorMessage);
        return null;
      } finally {
        setLoading(false);
      }
    },
    []
  );

  /**
   * セッションのスコアカードを取得
   */
  const fetchScoreCard = useCallback(async (sessionId: number): Promise<void> => {
    setLoading(true);
    setError(null);

    try {
      const data = await AiChatRepository.getScoreCard(sessionId);
      setScoreCard({
        ...data,
        scores: Array.isArray(data.scores) ? data.scores : [],
      });
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'スコアカードの取得に失敗しました。';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * スコア履歴を取得
   */
  const fetchScoreHistory = useCallback(async (): Promise<ScoreHistoryItem[]> => {
    setLoading(true);
    setError(null);

    try {
      const data = await AiChatRepository.getScoreHistory();
      return data;
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'スコア履歴の取得に失敗しました。';
      setError(errorMessage);
      return [];
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    sessions,
    currentSession,
    messages,
    scoreCard,
    loading,
    error,
    fetchSessions,
    fetchSession,
    createSession,
    updateSessionTitle,
    deleteSession,
    fetchMessages,
    addMessage,
    rephrase,
    fetchScoreCard,
    fetchScoreHistory,
  };
};
