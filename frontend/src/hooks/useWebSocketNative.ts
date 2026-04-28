import { useEffect, useRef, useCallback, useLayoutEffect } from 'react';

/**
 * useWebSocketNative — ブラウザ標準 WebSocket API ベースの接続フック。
 * SockJS / STOMP は廃止し、Go raw WebSocket と JSON ペイロードで直接やり取りする。
 *
 * - url が null のときは接続を作らない（認証前など、まだ接続条件が揃わないケース用）
 * - send() は OPEN 前でも呼べる（キューに積み、open 直後に flush）
 * - onMessage は受信した JSON を parse 済の object として呼び出す
 */

interface UseWebSocketNativeOptions {
  url: string | null;
  onOpen?: () => void;
  onMessage?: (payload: unknown) => void;
  onClose?: () => void;
  onError?: (err: Event) => void;
}

export function useWebSocketNative({
  url,
  onOpen,
  onMessage,
  onClose,
  onError,
}: UseWebSocketNativeOptions) {
  const socketRef = useRef<WebSocket | null>(null);
  const queueRef = useRef<string[]>([]);

  // コールバックは ref に保持して、ハンドラ参照変更による再接続を防ぐ。
  const onOpenRef = useRef(onOpen);
  const onMessageRef = useRef(onMessage);
  const onCloseRef = useRef(onClose);
  const onErrorRef = useRef(onError);
  useLayoutEffect(() => { onOpenRef.current = onOpen; }, [onOpen]);
  useLayoutEffect(() => { onMessageRef.current = onMessage; }, [onMessage]);
  useLayoutEffect(() => { onCloseRef.current = onClose; }, [onClose]);
  useLayoutEffect(() => { onErrorRef.current = onError; }, [onError]);

  const flushQueue = useCallback(() => {
    const sock = socketRef.current;
    if (!sock || sock.readyState !== WebSocket.OPEN) return;
    while (queueRef.current.length > 0) {
      const next = queueRef.current.shift();
      if (next !== undefined) sock.send(next);
    }
  }, []);

  const send = useCallback((payload: unknown) => {
    const raw = JSON.stringify(payload);
    const sock = socketRef.current;
    if (sock && sock.readyState === WebSocket.OPEN) {
      sock.send(raw);
      return;
    }
    queueRef.current.push(raw);
  }, []);

  const disconnect = useCallback(() => {
    const sock = socketRef.current;
    if (sock) {
      sock.close();
    }
    socketRef.current = null;
    queueRef.current = [];
  }, []);

  useEffect(() => {
    if (!url) return;
    const sock = new WebSocket(url);
    socketRef.current = sock;

    sock.onopen = () => {
      flushQueue();
      onOpenRef.current?.();
    };
    sock.onmessage = (ev) => {
      try {
        const parsed = typeof ev.data === 'string' ? JSON.parse(ev.data) : ev.data;
        onMessageRef.current?.(parsed);
      } catch {
        // バックエンドが JSON 以外を返したらドロップする（プロトコル外）。
      }
    };
    sock.onclose = () => {
      onCloseRef.current?.();
    };
    sock.onerror = (ev) => {
      onErrorRef.current?.(ev);
    };

    return () => {
      sock.close();
      socketRef.current = null;
      queueRef.current = [];
    };
  }, [url, flushQueue]);

  return { send, disconnect };
}
