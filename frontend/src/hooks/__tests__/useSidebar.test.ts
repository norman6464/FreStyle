import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useSidebar } from '../useSidebar';

const mockDispatch = vi.fn();
const mockNavigate = vi.fn();

vi.mock('react-redux', () => ({
  useDispatch: () => mockDispatch,
}));

vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock('../../store/authSlice', () => ({
  clearAuth: () => ({ type: 'auth/clearAuth' }),
}));

const mockFetchChatUsers = vi.fn();
const mockLogout = vi.fn();

vi.mock('../../repositories/ChatRepository', () => ({
  default: {
    fetchChatUsers: (...args: unknown[]) => mockFetchChatUsers(...args),
  },
}));

vi.mock('../../repositories/AuthRepository', () => ({
  default: {
    logout: (...args: unknown[]) => mockLogout(...args),
  },
}));

describe('useSidebar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockFetchChatUsers.mockResolvedValue([]);
    mockLogout.mockResolvedValue(undefined);
  });

  it('未読数を取得して合計する', async () => {
    mockFetchChatUsers.mockResolvedValue([
      { roomId: 1, unreadCount: 3, partnerName: 'User1', partnerId: 1 },
      { roomId: 2, unreadCount: 2, partnerName: 'User2', partnerId: 2 },
    ]);

    const { result } = renderHook(() => useSidebar());

    await waitFor(() => {
      expect(result.current.totalUnread).toBe(5);
    });
    expect(mockFetchChatUsers).toHaveBeenCalledOnce();
  });

  it('ログアウトでdispatch(clearAuth)とnavigate(/login)を呼ぶ', async () => {
    const { result } = renderHook(() => useSidebar());

    await act(async () => {
      await result.current.handleLogout();
    });

    expect(mockLogout).toHaveBeenCalledOnce();
    expect(mockDispatch).toHaveBeenCalledWith({ type: 'auth/clearAuth' });
    expect(mockNavigate).toHaveBeenCalledWith('/login');
  });

  it('ログアウト失敗時はnavigate呼ばない', async () => {
    mockLogout.mockRejectedValue(new Error('Network Error'));

    const { result } = renderHook(() => useSidebar());

    await act(async () => {
      await result.current.handleLogout();
    });

    expect(mockNavigate).not.toHaveBeenCalled();
  });
});
