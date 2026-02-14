import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useAiSession } from '../useAiSession';

const mockNavigate = vi.fn();
const mockDeleteSession = vi.fn();
const mockUpdateSessionTitle = vi.fn();

vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

describe('useAiSession', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockDeleteSession.mockResolvedValue(true);
    mockUpdateSessionTitle.mockResolvedValue(true);
  });

  it('セッション選択でcurrentSessionIdが更新される', () => {
    const { result } = renderHook(() =>
      useAiSession({ deleteSession: mockDeleteSession, updateSessionTitle: mockUpdateSessionTitle })
    );

    act(() => {
      result.current.handleSelectSession(42);
    });

    expect(result.current.currentSessionId).toBe(42);
    expect(mockNavigate).toHaveBeenCalledWith('/chat/ask-ai/42');
  });

  it('新規セッション作成でcurrentSessionIdがnullになる', () => {
    const { result } = renderHook(() =>
      useAiSession({ deleteSession: mockDeleteSession, updateSessionTitle: mockUpdateSessionTitle })
    );

    act(() => {
      result.current.handleSelectSession(42);
    });

    act(() => {
      result.current.handleNewSession();
    });

    expect(result.current.currentSessionId).toBeNull();
    expect(mockNavigate).toHaveBeenCalledWith('/chat/ask-ai');
  });

  it('セッション削除確認モーダルの開閉ができる', () => {
    const { result } = renderHook(() =>
      useAiSession({ deleteSession: mockDeleteSession, updateSessionTitle: mockUpdateSessionTitle })
    );

    act(() => {
      result.current.handleDeleteSession(5);
    });

    expect(result.current.deleteModal.isOpen).toBe(true);
    expect(result.current.deleteModal.sessionId).toBe(5);

    act(() => {
      result.current.cancelDeleteSession();
    });

    expect(result.current.deleteModal.isOpen).toBe(false);
  });

  it('セッション削除確定でdeleteSessionが呼ばれる', async () => {
    const { result } = renderHook(() =>
      useAiSession({ deleteSession: mockDeleteSession, updateSessionTitle: mockUpdateSessionTitle })
    );

    act(() => {
      result.current.handleSelectSession(5);
    });

    act(() => {
      result.current.handleDeleteSession(5);
    });

    await act(async () => {
      await result.current.confirmDeleteSession();
    });

    expect(mockDeleteSession).toHaveBeenCalledWith(5);
    expect(result.current.currentSessionId).toBeNull();
  });

  it('タイトル編集の開始・保存・キャンセルができる', async () => {
    const session = { id: 10, title: '既存タイトル', createdAt: '2026-01-01' };

    const { result } = renderHook(() =>
      useAiSession({ deleteSession: mockDeleteSession, updateSessionTitle: mockUpdateSessionTitle })
    );

    act(() => {
      result.current.handleStartEditTitle(session);
    });

    expect(result.current.editingSessionId).toBe(10);
    expect(result.current.editingTitle).toBe('既存タイトル');

    act(() => {
      result.current.setEditingTitle('新しいタイトル');
    });

    await act(async () => {
      await result.current.handleSaveTitle(10);
    });

    expect(mockUpdateSessionTitle).toHaveBeenCalledWith(10, { title: '新しいタイトル' });
    expect(result.current.editingSessionId).toBeNull();
  });

  it('deleteModal初期値がclosed状態', () => {
    const { result } = renderHook(() =>
      useAiSession({ deleteSession: mockDeleteSession, updateSessionTitle: mockUpdateSessionTitle })
    );

    expect(result.current.deleteModal.isOpen).toBe(false);
    expect(result.current.deleteModal.sessionId).toBeNull();
  });

  it('空タイトル保存時にeditingSessionIdがnullに戻る', async () => {
    const { result } = renderHook(() =>
      useAiSession({ deleteSession: mockDeleteSession, updateSessionTitle: mockUpdateSessionTitle })
    );

    act(() => {
      result.current.handleStartEditTitle({ id: 10, title: 'テスト' });
    });

    act(() => {
      result.current.setEditingTitle('   ');
    });

    await act(async () => {
      await result.current.handleSaveTitle(10);
    });

    expect(mockUpdateSessionTitle).not.toHaveBeenCalled();
    expect(result.current.editingSessionId).toBeNull();
  });

  it('handleCancelEditTitleでeditingTitleが空に戻る', () => {
    const { result } = renderHook(() =>
      useAiSession({ deleteSession: mockDeleteSession, updateSessionTitle: mockUpdateSessionTitle })
    );

    act(() => {
      result.current.handleStartEditTitle({ id: 10, title: '既存' });
    });

    expect(result.current.editingTitle).toBe('既存');

    act(() => {
      result.current.handleCancelEditTitle();
    });

    expect(result.current.editingSessionId).toBeNull();
    expect(result.current.editingTitle).toBe('');
  });
});
