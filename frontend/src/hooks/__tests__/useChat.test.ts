import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useChat } from '../useChat';
import ChatRepository from '../../repositories/ChatRepository';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useParams: () => ({ roomId: '1' }),
  useNavigate: () => mockNavigate,
}));

vi.mock('../../repositories/ChatRepository');
const mockShowToast = vi.fn();
vi.mock('../useToast', () => ({
  useToast: () => ({ showToast: mockShowToast, toasts: [], removeToast: vi.fn() }),
}));
vi.mock('sockjs-client', () => ({ default: vi.fn() }));
vi.mock('@stomp/stompjs', () => ({
  Client: class MockClient {
    activate = vi.fn();
    deactivate = vi.fn();
    subscribe = vi.fn();
    publish = vi.fn();
    connected = true;
    constructor(_config: any) {}
  },
}));

const mockedRepo = vi.mocked(ChatRepository);

describe('useChat', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedRepo.fetchCurrentUser.mockResolvedValue({ id: 1, name: 'テスト' });
    mockedRepo.fetchHistory.mockResolvedValue([]);
    mockedRepo.markAsRead.mockResolvedValue(undefined);
  });

  it('初期ロード時にsenderIdを取得する', async () => {
    const { result } = renderHook(() => useChat());

    await waitFor(() => {
      expect(result.current.senderId).toBe(1);
    });
  });

  it('messagesの初期値が空配列', () => {
    const { result } = renderHook(() => useChat());
    expect(result.current.messages).toEqual([]);
  });

  it('deleteModalの初期値が閉じた状態', () => {
    const { result } = renderHook(() => useChat());
    expect(result.current.deleteModal.isOpen).toBe(false);
    expect(result.current.deleteModal.messageId).toBeNull();
  });

  it('selectionModeの初期値がfalse', () => {
    const { result } = renderHook(() => useChat());
    expect(result.current.selectionMode).toBe(false);
  });

  it('handleDeleteMessageでdeleteModalが開く', async () => {
    const { result } = renderHook(() => useChat());

    act(() => {
      result.current.handleDeleteMessage('msg-42');
    });

    expect(result.current.deleteModal.isOpen).toBe(true);
    expect(result.current.deleteModal.messageId).toBe('msg-42');
  });

  it('cancelDeleteでdeleteModalが閉じる', async () => {
    const { result } = renderHook(() => useChat());

    act(() => {
      result.current.handleDeleteMessage('msg-42');
    });
    expect(result.current.deleteModal.isOpen).toBe(true);

    act(() => {
      result.current.cancelDelete();
    });
    expect(result.current.deleteModal.isOpen).toBe(false);
  });

  it('handleAiFeedbackでselectionModeがtrueになる', () => {
    const { result } = renderHook(() => useChat());

    act(() => {
      result.current.handleAiFeedback();
    });

    expect(result.current.selectionMode).toBe(true);
  });

  it('handleCancelSelectionでselectionModeがfalseに戻る', () => {
    const { result } = renderHook(() => useChat());

    act(() => {
      result.current.handleAiFeedback();
    });
    expect(result.current.selectionMode).toBe(true);

    act(() => {
      result.current.handleCancelSelection();
    });
    expect(result.current.selectionMode).toBe(false);
  });

  it('showRephraseModalの初期値がfalse', () => {
    const { result } = renderHook(() => useChat());
    expect(result.current.showRephraseModal).toBe(false);
  });

  it('fetchCurrentUser失敗時にログインページにナビゲートする', async () => {
    mockedRepo.fetchCurrentUser.mockRejectedValue(new Error('Unauthorized'));

    renderHook(() => useChat());

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login');
    });
  });
});
