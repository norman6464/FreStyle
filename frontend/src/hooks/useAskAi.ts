import { useEffect, useRef, useMemo, useState, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { useAuth } from './useAuth';
import { useAiChat } from './useAiChat';
import { useAiSession } from './useAiSession';
import { useAiChatSse, SseEvent } from './useAiChatSse';
import { AI_CHAT } from '@/shared/config/apiRoutes';
import type { AiAttachment, AiMessage } from '@/entities/ai-chat';

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
        case 'error': {
          // placeholder にエラー本文を残す（ユーザーに状況を伝える）。
          // id を非 streaming 値へ差し替えて「ストリーミング中」状態を閉じる。 これをしないと
          // isStreaming が true のまま固まり、 句読点を含まないエラー文が useSmoothReveal の
          // 未完チャンク保留で永久に非表示になる（FRESTYLE-146 レビュー指摘）。 clientId は
          // spread で保持されるので key は安定し、 バブルは remount しない。
          const errId = streamingIdRef.current;
          if (errId) {
            setMessages((prev) =>
              prev.map((m) =>
                m.id === errId
                  ? { ...m, id: `error-${errId}`, content: `（エラー）${ev.message}` }
                  : m
              )
            );
            streamingIdRef.current = null;
          }
          return;
        }
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
        // Date.now() は同一ミリ秒の連投で衝突し得る（clientId=key が重複する）ため UUID を使う。
        id: `local-user-${crypto.randomUUID()}`,
        sessionId: currentSessionId ?? 0,
        role: 'user',
        content: text,
        attachments: attachments.length > 0 ? attachments : undefined,
        createdAt: now,
      };
      // 連投時は useAiChatSse が前の stream を abort する（done/error は届かない）。
      // その旧 placeholder が `streaming-` id のまま残ると active が閉じず、 受信済みの末尾が
      // 未完チャンク保留で欠けたり「考え中」で固まる（FRESTYLE-146 レビュー指摘）。 新 placeholder を
      // 積む前に旧 placeholder を非 streaming id へ確定させ、 受信済みぶんを流し切らせる。
      const prevStreamingId = streamingIdRef.current;
      const placeholderId = `streaming-${crypto.randomUUID()}`;
      streamingIdRef.current = placeholderId;
      const placeholder: AiMessage = {
        id: placeholderId,
        sessionId: currentSessionId ?? 0,
        role: 'assistant',
        content: '',
        // done で id がサーバ確定値へ差し替わっても React の key を安定させるための
        // クライアント側 ID(FRESTYLE-146)。done handler の spread で自動的に保持される。
        clientId: placeholderId,
        createdAt: now,
      };
      setMessages((prev) => {
        const base = prevStreamingId
          ? prev.map((m) =>
              m.id === prevStreamingId ? { ...m, id: `aborted-${prevStreamingId}` } : m,
            )
          : prev;
        return [...base, userMsg, placeholder];
      });

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

  // 自動スクロールは AskAiPage 側で「ユーザーが下端付近にいる時のみ」の条件付きで実装する。
  // 旧実装は messages 変化のたびに無条件で scrollIntoView していたため、ストリーミング中
  // にユーザーが上にスクロールしても強制的に底へ戻されてしまう不具合があった。

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
    sessionSearchQuery,
    setSessionSearchQuery,

    ...aiSession,

    handleSend,
  };
}
