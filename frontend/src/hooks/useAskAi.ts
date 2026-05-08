import { useEffect, useRef, useMemo, useState, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { useAuth } from './useAuth';
import { useAiChat } from './useAiChat';
import { useAiSession } from './useAiSession';
import { useAiChatSse, SseEvent } from './useAiChatSse';
import { AI_CHAT } from '../constants/apiRoutes';
import { AiAttachment, AiMessage } from '../types';

/**
 * AskAiPage フック（SSE ストリーミング版）。
 *
 * メッセージ送信は WebSocket から SSE に切り替え:
 *   1. ユーザー発話を即座に画面に表示
 *   2. 「ストリーミング中のアシスタントメッセージ placeholder」を 1 つ追加
 *   3. token イベントごとに placeholder.content の末尾に append（文字パラパラ）
 *   4. done イベントで placeholder を確定（id / createdAt が backend から戻る）
 *   5. session イベントが来た場合は新規セッションをリストに追加 + currentSessionId 切替
 *
 * 進行中の stream は AbortController で中断可能（次の send 時に自動 abort）。
 */
export function useAskAi() {
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '';
  const STREAM_ENDPOINT = `${API_BASE_URL}${AI_CHAT.stream}`;

  const [sessionSearchQuery, setSessionSearchQuery] = useState('');
  const messagesEndRef = useRef<HTMLDivElement | null>(null);
  // ストリーミング中のアシスタントメッセージ ID（placeholder）
  const streamingIdRef = useRef<string | null>(null);

  const { sessionId: urlSessionId } = useParams<{ sessionId: string }>();

  const { getCurrentUser, user } = useAuth();
  const {
    sessions,
    messages,
    setMessages,
    loading,
    fetchSessions,
    fetchMessages,
    deleteSession,
    updateSessionTitle,
    handleIncomingSession,
  } = useAiChat();

  const aiSession = useAiSession({ deleteSession, updateSessionTitle });
  const { currentSessionId, setCurrentSessionId } = aiSession;

  const handleEvent = useCallback(
    (ev: SseEvent) => {
      switch (ev.type) {
        case 'session':
          handleIncomingSession({
            id: ev.id,
            title: ev.title,
            sessionType: ev.sessionType,
            scenarioId: ev.scenarioId ?? undefined,
            createdAt: ev.createdAt,
          });
          setCurrentSessionId(ev.id);
          return;
        case 'token': {
          const id = streamingIdRef.current;
          if (!id) return;
          setMessages((prev) =>
            prev.map((m) =>
              m.id === id ? { ...m, content: m.content + ev.delta } : m
            )
          );
          return;
        }
        case 'done': {
          const tempId = streamingIdRef.current;
          if (!tempId) return;
          setMessages((prev) =>
            prev.map((m) =>
              m.id === tempId
                ? {
                    ...m,
                    id: ev.id,
                    content: ev.content,
                    createdAt: ev.createdAt,
                  }
                : m
            )
          );
          streamingIdRef.current = null;
          return;
        }
        case 'error':
          // placeholder にエラー本文を残す（ユーザーに状況を伝える）
          if (streamingIdRef.current) {
            setMessages((prev) =>
              prev.map((m) =>
                m.id === streamingIdRef.current
                  ? { ...m, content: `（エラー）${ev.message}` }
                  : m
              )
            );
            streamingIdRef.current = null;
          }
          return;
      }
    },
    [handleIncomingSession, setCurrentSessionId, setMessages]
  );

  const { send: sendSse, abort: abortSse } = useAiChatSse({
    endpoint: STREAM_ENDPOINT,
    onEvent: handleEvent,
  });

  const handleSend = useCallback(
    (text: string, attachments: AiAttachment[] = []): void => {
      if (!text.trim() && attachments.length === 0) return;

      // ユーザー発話 + アシスタント placeholder を即座に追加
      const now = new Date().toISOString();
      const userMsg: AiMessage = {
        id: `local-user-${Date.now()}`,
        sessionId: currentSessionId ?? 0,
        role: 'user',
        content: text,
        attachments: attachments.length > 0 ? attachments : undefined,
        createdAt: now,
      };
      const placeholderId = `streaming-${Date.now()}`;
      streamingIdRef.current = placeholderId;
      const placeholder: AiMessage = {
        id: placeholderId,
        sessionId: currentSessionId ?? 0,
        role: 'assistant',
        content: '',
        createdAt: now,
      };
      setMessages((prev) => [...prev, userMsg, placeholder]);

      void sendSse({
        sessionId: currentSessionId,
        content: text,
        // SSE 送信時は preview URL を含めない（不要 / バックエンドにリーク防止）。
        attachments: attachments.map(({ key, filename, contentType, sizeBytes }) => ({
          key,
          filename,
          contentType,
          sizeBytes,
        })),
      });
    },
    [sendSse, currentSessionId, setMessages]
  );

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

  // URL パラメータのセッション ID 変更で current を切替
  useEffect(() => {
    if (urlSessionId) {
      setCurrentSessionId(parseInt(urlSessionId, 10));
    } else {
      setCurrentSessionId(null);
    }
  }, [urlSessionId, setCurrentSessionId]);

  // セッション内のメッセージ履歴取得。
  //
  // 「新しいチャット」ボタンや、未確定セッション（URL に sessionId 無し）に切替えた場合は
  // currentSessionId が null になるので、前セッションのメッセージを画面から明示的に消去する。
  // クリアしないと前セッションのチャット履歴が残ったまま空打ち状態になり、新規開始
  // の見た目にならない。
  //
  // ただし、ストリーミング中は fetchMessages を skip する。新規セッション作成 flow で
  // SSE が 'session' イベントを emit すると currentSessionId が null → N に切り替わり、
  // この useEffect が fetchMessages(N) を呼んで在 streaming 中の placeholder を
  // バックエンドから取得した（まだ user message も保存されていない可能性のある）配列で
  // 上書きしてしまう race を防ぐ。streaming が終わった次回以降の切替では通常通り fetch する。
  useEffect(() => {
    if (currentSessionId) {
      if (streamingIdRef.current) return;
      fetchMessages(currentSessionId);
    } else {
      setMessages([]);
    }
  }, [currentSessionId, fetchMessages, setMessages]);

  // メッセージ最下部へスクロール
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // unmount で進行中の stream を abort
  useEffect(() => {
    return () => abortSse();
  }, [abortSse]);

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
