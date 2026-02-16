import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useAskAi } from '../useAskAi';

const mockGetCurrentUser = vi.fn();
const mockFetchSessions = vi.fn();
const mockFetchMessages = vi.fn();
const mockDeleteSession = vi.fn();
const mockUpdateSessionTitle = vi.fn();
const mockSubscribe = vi.fn(() => vi.fn());
const mockPublish = vi.fn();
const mockNavigate = vi.fn();

vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
  useParams: () => ({ sessionId: undefined }),
  useLocation: () => ({ state: null }),
}));

vi.mock('../useAuth', () => ({
  useAuth: () => ({
    getCurrentUser: mockGetCurrentUser,
    user: { id: 1 },
  }),
}));

vi.mock('../useAiChat', () => ({
  useAiChat: () => ({
    sessions: [],
    messages: [],
    scoreCard: null,
    fetchSessions: mockFetchSessions,
    fetchMessages: mockFetchMessages,
    deleteSession: mockDeleteSession,
    updateSessionTitle: mockUpdateSessionTitle,
  }),
}));

vi.mock('../useWebSocket', () => ({
  useWebSocket: (opts: any) => {
    return {
      subscribe: mockSubscribe,
      publish: mockPublish,
    };
  },
}));

vi.mock('../useAiSession', () => ({
  useAiSession: (opts: any) => ({
    currentSessionId: null,
    setCurrentSessionId: vi.fn(),
    deleteModal: { isOpen: false, sessionId: null },
    editingSessionId: null,
    editingTitle: '',
    setEditingTitle: vi.fn(),
    handleNewSession: vi.fn(),
    handleSelectSession: vi.fn(),
    handleDeleteSession: vi.fn(),
    confirmDeleteSession: vi.fn(),
    cancelDeleteSession: vi.fn(),
    handleStartEditTitle: vi.fn(),
    handleSaveTitle: vi.fn(),
    handleCancelEditTitle: vi.fn(),
  }),
}));

describe('useAskAi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('初期化時にgetCurrentUserが呼ばれる', () => {
    renderHook(() => useAskAi());
    expect(mockGetCurrentUser).toHaveBeenCalled();
  });

  it('user取得後にfetchSessionsが呼ばれる', () => {
    renderHook(() => useAskAi());
    expect(mockFetchSessions).toHaveBeenCalled();
  });

  it('messagesEndRefが存在する', () => {
    const { result } = renderHook(() => useAskAi());
    expect(result.current.messagesEndRef).toBeDefined();
  });

  it('isPracticeModeの初期値がfalse', () => {
    const { result } = renderHook(() => useAskAi());
    expect(result.current.isPracticeMode).toBe(false);
  });

  it('scenarioNameの初期値がnull', () => {
    const { result } = renderHook(() => useAskAi());
    expect(result.current.scenarioName).toBeNull();
  });

  it('handleSendがpublishを呼ぶ', () => {
    const { result } = renderHook(() => useAskAi());
    act(() => {
      result.current.handleSend('テストメッセージ');
    });
    expect(mockPublish).toHaveBeenCalledWith('/app/ai-chat/send', expect.objectContaining({
      content: 'テストメッセージ',
      role: 'user',
    }));
  });

  it('sessionsが空配列で返される', () => {
    const { result } = renderHook(() => useAskAi());
    expect(result.current.sessions).toEqual([]);
  });

  it('messagesが空配列で返される', () => {
    const { result } = renderHook(() => useAskAi());
    expect(result.current.messages).toEqual([]);
  });

  it('scoreCardがnullで返される', () => {
    const { result } = renderHook(() => useAskAi());
    expect(result.current.scoreCard).toBeNull();
  });

  it('deleteModalが閉じた状態で返される', () => {
    const { result } = renderHook(() => useAskAi());
    expect(result.current.deleteModal.isOpen).toBe(false);
  });

  it('sessionSearchQueryの初期値が空文字である', () => {
    const { result } = renderHook(() => useAskAi());
    expect(result.current.sessionSearchQuery).toBe('');
  });

  it('setSessionSearchQueryで検索クエリを変更できる', () => {
    const { result } = renderHook(() => useAskAi());
    act(() => { result.current.setSessionSearchQuery('テスト'); });
    expect(result.current.sessionSearchQuery).toBe('テスト');
  });

  it('filteredSessionsが返される', () => {
    const { result } = renderHook(() => useAskAi());
    expect(result.current.filteredSessions).toEqual([]);
  });
});
