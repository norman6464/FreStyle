import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { ReactNode } from 'react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import authReducer from '../../store/authSlice';
import { useAskAi } from '../useAskAi';

// useAskAi は useAiChatSse / AiChatRepository / fetchSessions / fetchMessages を内部で叩く。
// バックエンド呼び出しはすべてモックして、純粋な state 遷移ロジック（特に
// 「new chat ボタン → currentSessionId=null → messages がクリアされる」） を検証する。

vi.mock('../../repositories/AiChatRepository', () => ({
  default: {
    getSessions: vi.fn().mockResolvedValue([]),
    getMessages: vi.fn().mockResolvedValue([
      {
        id: 'srv-1',
        sessionId: 5,
        role: 'user',
        content: 'session 5 の発話',
      },
    ]),
    deleteSession: vi.fn().mockResolvedValue(undefined),
    updateSessionTitle: vi.fn().mockResolvedValue({}),
  },
}));

vi.mock('../useAuth', () => ({
  useAuth: () => ({
    user: { id: 1, email: 'a@example.com' },
    getCurrentUser: vi.fn(),
  }),
}));

// onEvent コールバックを後から呼べるように capture する。テスト側で
// `lastOnEvent({ type: 'session', id: ... })` 等を投げて SSE をシミュレート可能。
let lastOnEvent: ((ev: unknown) => void) | null = null;
const mockSseSend = vi.fn();
vi.mock('../useAiChatSse', () => ({
  useAiChatSse: (opts: { onEvent: (ev: unknown) => void }) => {
    lastOnEvent = opts.onEvent;
    return { send: mockSseSend, abort: vi.fn() };
  },
}));

function makeWrapper(initialPath: string) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: { isAuthenticated: true, loading: false, isAdmin: false, role: 'trainee' },
    },
  });
  return ({ children }: { children: ReactNode }) => (
    <Provider store={store}>
      <MemoryRouter initialEntries={[initialPath]}>
        <Routes>
          <Route path="/chat/ask-ai" element={<>{children}</>} />
          <Route path="/chat/ask-ai/:sessionId" element={<>{children}</>} />
        </Routes>
      </MemoryRouter>
    </Provider>
  );
}

describe('useAskAi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    lastOnEvent = null;
    mockSseSend.mockReset();
  });

  it('sessionId 付き URL では fetchMessages の結果が messages にロードされる', async () => {
    const { result } = renderHook(() => useAskAi(), {
      wrapper: makeWrapper('/chat/ask-ai/5'),
    });

    await waitFor(() => {
      expect(result.current.messages.length).toBeGreaterThan(0);
    });
    expect(result.current.messages[0].content).toBe('session 5 の発話');
  });

  it('「新しいチャット」を押して session 無し URL に切替えると messages がクリアされる', async () => {
    const { result } = renderHook(() => useAskAi(), {
      wrapper: makeWrapper('/chat/ask-ai/5'),
    });

    // 一度メッセージが入った状態を作る
    await waitFor(() => {
      expect(result.current.messages.length).toBeGreaterThan(0);
    });

    // handleNewSession 経由で sessionId を null に切り替える
    act(() => {
      result.current.handleNewSession();
    });

    // currentSessionId 変更後、useEffect 内の setMessages([]) が走り messages が空配列になる
    await waitFor(() => {
      expect(result.current.messages).toEqual([]);
    });
  });

  // PR-J で導入した「session 切替で messages クリア」のロジックが、PR-G1 から動いていた
  // 「新規セッションを SSE 'session' イベントで受け取って currentSessionId を切替える」flow と
  // race して、streaming 中の placeholder を fetchMessages が上書きしてしまう不具合があった。
  // PR-L で streamingIdRef による guard を追加し、streaming 中は fetchMessages を skip する。
  it('streaming 中に SSE session イベントで currentSessionId が切り替わっても messages を上書きしない', async () => {
    const aiChatRepo = (await import('../../repositories/AiChatRepository')).default;
    vi.mocked(aiChatRepo.getMessages).mockClear();

    const { result } = renderHook(() => useAskAi(), {
      wrapper: makeWrapper('/chat/ask-ai'), // session 無し（新規チャット画面）からスタート
    });

    // 初期: currentSessionId は null、messages は空
    await waitFor(() => {
      expect(result.current.messages).toEqual([]);
    });
    expect(aiChatRepo.getMessages).not.toHaveBeenCalled();

    // 新規送信を発火（streamingIdRef がセットされる）
    act(() => {
      result.current.handleSend('PHPとGo言語の違いは何？');
    });

    // ローカルに userMsg + placeholder の 2 件が入った状態
    expect(result.current.messages).toHaveLength(2);

    // SSE が新規セッション作成イベントを emit したシミュレート
    act(() => {
      lastOnEvent?.({
        type: 'session',
        id: 99,
        title: 'PHPとGo言語の違いは何？',
        sessionType: 'free',
        createdAt: '2026-05-08T11:21:34Z',
      });
    });

    // currentSessionId は 99 に切り替わるが、streaming 中なので fetchMessages は skip される
    // → messages は handleSend で入れた 2 件のまま保持される
    await waitFor(() => {
      expect(aiChatRepo.getMessages).not.toHaveBeenCalled();
    });
    expect(result.current.messages).toHaveLength(2);
  });
});
