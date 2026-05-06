import { useEffect, useRef, useCallback, useLayoutEffect } from 'react';

/**
 * useWebSocketNative — ブラウザ標準 WebSocket API ベースの接続フック。
 * SockJS / STOMP は廃止し、Go raw WebSocket と JSON ペイロードで直接やり取りする。
 *
 * - url が null のときは接続を作らない（認証前など、まだ接続条件が揃わないケース用）
 * - send() は OPEN 前でも呼べる（キューに積み、open 直後に flush）
 * - onMessage は受信した JSON を parse 済の object として呼び出す
 * - onclose 時は自動再接続する（指数バックオフ: 1s → 2s → 4s → 8s → 16s → 30s 上限）
 *   フロント・サーバ・ALB のいずれが切断しても、ユーザー操作なしで復旧する
 */

interface UseWebSocketNativeOptions {
  url: string | null;
  onOpen?: () => void;
  onMessage?: (payload: unknown) => void;
  onClose?: () => void;
  onError?: (err: Event) => void;
}

const RECONNECT_BASE_MS = 1000;
const RECONNECT_MAX_MS = 30000;

export function useWebSocketNative({
  url,
  onOpen,
  onMessage,
  onClose,
  onError,
}: UseWebSocketNativeOptions) {
  const socketRef = useRef<WebSocket | null>(null);
  const queueRef = useRef<string[]>([]);
  // 再接続まわりの状態。effect cleanup から touch するため ref に保持。
  const reconnectAttemptsRef = useRef(0);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const cancelledRef = useRef(false);

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
    // OPEN 前 / 切断中 → キューに積む。再接続後に flushQueue で送信される。
    queueRef.current.push(raw);
  }, []);

  const disconnect = useCallback(() => {
    cancelledRef.current = true;
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }
    const sock = socketRef.current;
    if (sock) {
      sock.close();
    }
    socketRef.current = null;
    queueRef.current = [];
  }, []);

  useEffect(() => {
    if (!url) return;
    cancelledRef.current = false;
    reconnectAttemptsRef.current = 0;

    const connect = () => {
      if (cancelledRef.current) return;
      const sock = new WebSocket(url);
      socketRef.current = sock;

      sock.onopen = () => {
        // 再接続成功でカウンタをリセットする。次回の切断は再び 1s から開始。
        reconnectAttemptsRef.current = 0;
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
        if (cancelledRef.current) return;
        // ALB のアイドルタイムアウト / サーバ再起動 / ネットワーク断のいずれでも、
        // ユーザー操作なしで復旧したい。指数バックオフで再接続を試みる。
        const attempt = reconnectAttemptsRef.current;
        const delay = Math.min(RECONNECT_MAX_MS, RECONNECT_BASE_MS * 2 ** attempt);
        reconnectAttemptsRef.current = attempt + 1;
        reconnectTimerRef.current = setTimeout(connect, delay);
      };
      sock.onerror = (ev) => {
        onErrorRef.current?.(ev);
      };
    };

    connect();

    return () => {
      cancelledRef.current = true;
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
      const sock = socketRef.current;
      if (sock) sock.close();
      socketRef.current = null;
      queueRef.current = [];
    };
  }, [url, flushQueue]);

  return { send, disconnect };
}
