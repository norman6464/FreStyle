import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useWebSocket } from '../useWebSocket';

const mockActivate = vi.fn();
const mockDeactivate = vi.fn();
const mockSubscribe = vi.fn();
const mockPublish = vi.fn();

let capturedOnConnect: (() => void) | undefined;
let capturedOnDisconnect: (() => void) | undefined;

vi.mock('@stomp/stompjs', () => {
  return {
    Client: function (this: any, config: any) {
      capturedOnConnect = config.onConnect;
      capturedOnDisconnect = config.onDisconnect;
      this.activate = mockActivate;
      this.deactivate = mockDeactivate;
      this.subscribe = mockSubscribe;
      this.publish = mockPublish;
      this.connected = false;
    },
  };
});

vi.mock('sockjs-client', () => ({
  default: vi.fn(),
}));

describe('useWebSocket', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    capturedOnConnect = undefined;
    capturedOnDisconnect = undefined;
  });

  it('userIdがnullの場合は接続しない', () => {
    renderHook(() =>
      useWebSocket({ url: 'http://localhost/ws', userId: null })
    );

    expect(mockActivate).not.toHaveBeenCalled();
  });

  it('userIdがある場合にactivateが呼ばれる', () => {
    renderHook(() =>
      useWebSocket({ url: 'http://localhost/ws', userId: 1 })
    );

    expect(mockActivate).toHaveBeenCalledOnce();
  });

  it('onConnectコールバックが呼ばれる', () => {
    const onConnect = vi.fn();

    renderHook(() =>
      useWebSocket({ url: 'http://localhost/ws', userId: 1, onConnect })
    );

    act(() => {
      capturedOnConnect?.();
    });

    expect(onConnect).toHaveBeenCalledOnce();
  });

  it('onDisconnectコールバックが呼ばれる', () => {
    const onDisconnect = vi.fn();

    renderHook(() =>
      useWebSocket({ url: 'http://localhost/ws', userId: 1, onDisconnect })
    );

    act(() => {
      capturedOnDisconnect?.();
    });

    expect(onDisconnect).toHaveBeenCalledOnce();
  });

  it('アンマウント時にdeactivateが呼ばれる', () => {
    const { unmount } = renderHook(() =>
      useWebSocket({ url: 'http://localhost/ws', userId: 1 })
    );

    unmount();

    expect(mockDeactivate).toHaveBeenCalledOnce();
  });
});
