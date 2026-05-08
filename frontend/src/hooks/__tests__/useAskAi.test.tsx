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

vi.mock('../useAiChatSse', () => ({
  useAiChatSse: () => ({
    send: vi.fn(),
    abort: vi.fn(),
  }),
}));

function makeWrapper(initialPath: string) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: { isAuthenticated: true, loading: false, isAdmin: false, onboarded: true, role: 'trainee' },
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
});
