import { useCallback, useRef } from 'react';

/**
 * Server-Sent Events ベースの AI チャット送受信フック。
 *
 * 標準の EventSource は GET しかサポートしないため、本実装は fetch + ReadableStream で
 * line-based に SSE フレームを読む（多くの汎用 AI チャット製品が採用するパターン）。
 *
 * バックエンドの送信フォーマット（backend/internal/handler/ai_chat_sse_handler.go）:
 *
 *   event: session   data: {...}
 *   event: token     data: {"delta": "..."}
 *   event: done      data: {"sessionId":..., "id":..., "content":...}
 *   event: error     data: {"message": "..."}
 *
 * 各イベントは `event:` 行 + `data:` 行 + 空行 で 1 単位。本フックは line buffer を
 * 持って 1 イベントが揃った時点でハンドラを呼ぶ。
 */

export interface SseSessionEvent {
  type: 'session';
  id: number;
  title?: string;
  sessionType?: string;
  scenarioId?: number | null;
  createdAt?: string;
}

export interface SseTokenEvent {
  type: 'token';
  delta: string;
}

export interface SseDoneEvent {
  type: 'done';
  sessionId: number;
  id: string;
  role: 'assistant';
  content: string;
  createdAt: string;
}

export interface SseErrorEvent {
  type: 'error';
  message: string;
}

export type SseEvent = SseSessionEvent | SseTokenEvent | SseDoneEvent | SseErrorEvent;

export interface SendStreamRequest {
  sessionId: number | null;
  content: string;
}

export interface UseAiChatSseOptions {
  endpoint: string;
  onEvent: (ev: SseEvent) => void;
  /** stream 終了（done / error / abort）で呼ぶ。クライアント側のリセットに使う */
  onClose?: () => void;
}

export function useAiChatSse({ endpoint, onEvent, onClose }: UseAiChatSseOptions) {
  // 進行中のリクエストを中断するための AbortController を保持。
  // 同じセッションで連投したときに前の stream を捨てるため。
  const abortRef = useRef<AbortController | null>(null);

  const send = useCallback(
    async (req: SendStreamRequest): Promise<void> => {
      // 既存の stream があればキャンセル
      abortRef.current?.abort();
      const controller = new AbortController();
      abortRef.current = controller;

      try {
        const res = await fetch(endpoint, {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
            Accept: 'text/event-stream',
          },
          body: JSON.stringify({
            sessionId: req.sessionId ?? 0,
            content: req.content,
          }),
          signal: controller.signal,
        });

        if (!res.ok || !res.body) {
          onEvent({
            type: 'error',
            message:
              res.status === 401
                ? 'ログインが必要です'
                : res.status === 503
                ? 'AI チャットが一時的に利用できません'
                : 'メッセージの送信に失敗しました',
          });
          onClose?.();
          return;
        }

        const reader = res.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';

        while (true) {
          const { value, done } = await reader.read();
          if (done) break;
          buffer += decoder.decode(value, { stream: true });

          // SSE は \n\n で 1 イベントを区切る。一度に複数イベントが届く可能性もある。
          let sep: number;
          while ((sep = buffer.indexOf('\n\n')) !== -1) {
            const raw = buffer.slice(0, sep);
            buffer = buffer.slice(sep + 2);
            const ev = parseSseFrame(raw);
            if (ev) onEvent(ev);
          }
        }
      } catch (err) {
        // AbortController.abort() は AbortError を投げる。意図的なキャンセルなので無視。
        if ((err as Error).name === 'AbortError') return;
        onEvent({ type: 'error', message: 'ネットワークエラーが発生しました' });
      } finally {
        if (abortRef.current === controller) abortRef.current = null;
        onClose?.();
      }
    },
    [endpoint, onEvent, onClose]
  );

  const abort = useCallback(() => {
    abortRef.current?.abort();
  }, []);

  return { send, abort };
}

/**
 * 1 つの SSE フレームをパースする。
 * 形式:
 *   event: <name>
 *   data: <json>
 * data 行が複数になる可能性もある（仕様上は \n で連結される）。
 */
function parseSseFrame(raw: string): SseEvent | null {
  const lines = raw.split('\n');
  let event = '';
  let data = '';
  for (const line of lines) {
    if (line.startsWith(':')) continue; // コメント行（keepalive など）
    if (line.startsWith('event:')) {
      event = line.slice('event:'.length).trim();
    } else if (line.startsWith('data:')) {
      data += (data ? '\n' : '') + line.slice('data:'.length).trim();
    }
  }
  if (!event || !data) return null;

  try {
    const payload = JSON.parse(data);
    switch (event) {
      case 'session':
        return { type: 'session', ...payload };
      case 'token':
        return { type: 'token', delta: String(payload.delta ?? '') };
      case 'done':
        return {
          type: 'done',
          sessionId: payload.sessionId,
          id: payload.id,
          role: 'assistant',
          content: payload.content,
          createdAt: payload.createdAt,
        };
      case 'error':
        return { type: 'error', message: String(payload.message ?? '') };
      default:
        return null;
    }
  } catch {
    return null;
  }
}
