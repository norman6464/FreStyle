import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useWebSocketNative } from '../useWebSocketNative';

class MockWebSocket {
  static instances: MockWebSocket[] = [];
  static OPEN = 1;
  static CLOSED = 3;

  url: string;
  readyState = 0;
  onopen: ((ev: Event) => void) | null = null;
  onmessage: ((ev: MessageEvent) => void) | null = null;
  onclose: ((ev: CloseEvent) => void) | null = null;
  onerror: ((ev: Event) => void) | null = null;
  sent: string[] = [];

  constructor(url: string) {
    this.url = url;
    MockWebSocket.instances.push(this);
  }

  open() {
    this.readyState = MockWebSocket.OPEN;
    this.onopen?.(new Event('open'));
  }

  emit(data: unknown) {
    this.onmessage?.(new MessageEvent('message', { data: JSON.stringify(data) }));
  }

  close() {
    this.readyState = MockWebSocket.CLOSED;
    this.onclose?.(new CloseEvent('close'));
  }

  send(data: string) {
    this.sent.push(data);
  }
}

beforeEach(() => {
  MockWebSocket.instances = [];
  vi.stubGlobal('WebSocket', MockWebSocket);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('useWebSocketNative', () => {
  it('url が指定されると WebSocket インスタンスが生成される', () => {
    renderHook(() => useWebSocketNative({ url: 'wss://example.com/ws' }));
    expect(MockWebSocket.instances).toHaveLength(1);
    expect(MockWebSocket.instances[0].url).toBe('wss://example.com/ws');
  });

  it('url が null のときは接続を作らない', () => {
    renderHook(() => useWebSocketNative({ url: null }));
    expect(MockWebSocket.instances).toHaveLength(0);
  });

  it('接続が open になると onOpen が呼ばれる', () => {
    const onOpen = vi.fn();
    renderHook(() => useWebSocketNative({ url: 'wss://example.com', onOpen }));
    act(() => MockWebSocket.instances[0].open());
    expect(onOpen).toHaveBeenCalled();
  });

  it('受信メッセージは JSON parse されて onMessage に渡る', () => {
    const onMessage = vi.fn();
    renderHook(() => useWebSocketNative({ url: 'wss://example.com', onMessage }));
    act(() => {
      MockWebSocket.instances[0].open();
      MockWebSocket.instances[0].emit({ type: 'message', content: 'hi' });
    });
    expect(onMessage).toHaveBeenCalledWith({ type: 'message', content: 'hi' });
  });

  it('send は JSON.stringify して送信する', () => {
    const { result } = renderHook(() => useWebSocketNative({ url: 'wss://example.com' }));
    act(() => MockWebSocket.instances[0].open());
    act(() => result.current.send({ type: 'send', content: 'hello' }));
    expect(MockWebSocket.instances[0].sent).toEqual([JSON.stringify({ type: 'send', content: 'hello' })]);
  });

  it('open 前の send はキューに積まれて open 後に flush される', () => {
    const { result } = renderHook(() => useWebSocketNative({ url: 'wss://example.com' }));
    act(() => result.current.send({ type: 'send', content: 'first' }));
    expect(MockWebSocket.instances[0].sent).toHaveLength(0);
    act(() => MockWebSocket.instances[0].open());
    expect(MockWebSocket.instances[0].sent).toEqual([JSON.stringify({ type: 'send', content: 'first' })]);
  });

  it('disconnect で WebSocket.close が呼ばれる', () => {
    const { result } = renderHook(() => useWebSocketNative({ url: 'wss://example.com' }));
    act(() => MockWebSocket.instances[0].open());
    act(() => result.current.disconnect());
    expect(MockWebSocket.instances[0].readyState).toBe(MockWebSocket.CLOSED);
  });
});
