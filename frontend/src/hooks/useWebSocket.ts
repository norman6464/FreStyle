import { useEffect, useRef, useCallback } from 'react';
import SockJS from 'sockjs-client';
import { Client, IMessage } from '@stomp/stompjs';

/**
 * WebSocketフック
 *
 * <p>役割:</p>
 * <ul>
 *   <li>WebSocket接続管理（SockJS + STOMP）</li>
 *   <li>購読管理、メッセージ送受信</li>
 * </ul>
 *
 * <p>Hooks層（Presentation Layer - Business Logic）:</p>
 * <ul>
 *   <li>コンポーネントからWebSocketロジックを分離</li>
 *   <li>STOMP Clientの抽象化</li>
 * </ul>
 */

interface UseWebSocketOptions {
  url: string;
  userId: number | null;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: any) => void;
}

interface Subscription {
  destination: string;
  callback: (message: IMessage) => void;
}

export const useWebSocket = ({ url, userId, onConnect, onDisconnect, onError }: UseWebSocketOptions) => {
  const clientRef = useRef<Client | null>(null);
  const subscriptionsRef = useRef<Subscription[]>([]);
  const isConnectedRef = useRef(false);

  /**
   * WebSocket接続
   */
  const connect = useCallback(() => {
    if (!userId) return;

    const client = new Client({
      webSocketFactory: () => new SockJS(url, undefined, { withCredentials: true }),
      reconnectDelay: 5000,

      onConnect: () => {
        clientRef.current = client;
        isConnectedRef.current = true;

        // 接続後に既存の購読を再登録
        subscriptionsRef.current.forEach((sub) => {
          client.subscribe(sub.destination, sub.callback);
        });

        onConnect?.();
      },

      onDisconnect: () => {
        isConnectedRef.current = false;
        onDisconnect?.();
      },

      onStompError: (frame) => {
        onError?.(frame);
      },
    });

    client.activate();
    clientRef.current = client;
  }, [url, userId, onConnect, onDisconnect, onError]);

  /**
   * WebSocket切断
   */
  const disconnect = useCallback(() => {
    if (clientRef.current) {
      clientRef.current.deactivate();
      clientRef.current = null;
      isConnectedRef.current = false;
      subscriptionsRef.current = [];
    }
  }, []);

  /**
   * トピック購読
   */
  const subscribe = useCallback((destination: string, callback: (message: IMessage) => void) => {
    const subscription: Subscription = { destination, callback };
    subscriptionsRef.current.push(subscription);

    if (clientRef.current?.connected) {
      clientRef.current.subscribe(destination, callback);
    }

    // 購読解除関数を返す
    return () => {
      subscriptionsRef.current = subscriptionsRef.current.filter((sub) => sub !== subscription);
    };
  }, []);

  /**
   * メッセージ送信
   */
  const publish = useCallback((destination: string, body: any) => {
    if (clientRef.current?.connected) {
      clientRef.current.publish({
        destination,
        body: JSON.stringify(body),
      });
    }
  }, []);

  /**
   * 接続状態の取得
   */
  const isConnected = useCallback(() => {
    return isConnectedRef.current;
  }, []);

  /**
   * useEffect: 接続・切断
   */
  useEffect(() => {
    connect();

    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    client: clientRef.current,
    isConnected,
    subscribe,
    publish,
    disconnect,
  };
};
