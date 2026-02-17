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
    expect(mockNavigate).toHaveBeenCalledWith('/login', { state: { toast: 'ログアウトしました' } });
  });

  it('ログアウト失敗時はnavigate呼ばない', async () => {
    mockLogout.mockRejectedValue(new Error('Network Error'));

    const { result } = renderHook(() => useSidebar());

    await act(async () => {
      await result.current.handleLogout();
    });

    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it('チャットユーザーがいない場合はtotalUnreadが0', async () => {
    mockFetchChatUsers.mockResolvedValue([]);

    const { result } = renderHook(() => useSidebar());

    await waitFor(() => {
      expect(result.current.totalUnread).toBe(0);
    });
  });

  it('fetchChatUsers失敗時にtotalUnreadが0のまま', async () => {
    mockFetchChatUsers.mockRejectedValue(new Error('Network Error'));

    const { result } = renderHook(() => useSidebar());

    await waitFor(() => {
      expect(mockFetchChatUsers).toHaveBeenCalledOnce();
    });

    expect(result.current.totalUnread).toBe(0);
  });

  it('unreadCountがundefinedの場合は0として扱われる', async () => {
    mockFetchChatUsers.mockResolvedValue([
      { roomId: 1, unreadCount: undefined, partnerName: 'User1', partnerId: 1 },
      { roomId: 2, unreadCount: 5, partnerName: 'User2', partnerId: 2 },
    ]);

    const { result } = renderHook(() => useSidebar());

    await waitFor(() => {
      expect(result.current.totalUnread).toBe(5);
    });
  });

  it('初期状態でtotalUnreadが0である', () => {
    const { result } = renderHook(() => useSidebar());
    expect(result.current.totalUnread).toBe(0);
  });

  it('ログアウト成功時にトースト付きでログインページに遷移する', async () => {
    const { result } = renderHook(() => useSidebar());

    await act(async () => {
      await result.current.handleLogout();
    });

    expect(mockNavigate).toHaveBeenCalledWith('/login', { state: { toast: 'ログアウトしました' } });
  });

  it('ログアウト失敗時にトースト付きで遷移しない', async () => {
    mockLogout.mockRejectedValue(new Error('Network Error'));

    const { result } = renderHook(() => useSidebar());

    await act(async () => {
      await result.current.handleLogout();
    });

    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it('ログアウト失敗時にdispatchも呼ばれない', async () => {
    mockLogout.mockRejectedValue(new Error('Network Error'));

    const { result } = renderHook(() => useSidebar());

    await act(async () => {
      await result.current.handleLogout();
    });

    expect(mockDispatch).not.toHaveBeenCalled();
  });

  it('初期状態でloggingOutがfalseである', () => {
    const { result } = renderHook(() => useSidebar());
    expect(result.current.loggingOut).toBe(false);
  });

  it('ログアウト開始時にloggingOutがtrueになる', async () => {
    let resolveLogout: () => void;
    mockLogout.mockImplementation(() => new Promise<void>(r => { resolveLogout = r; }));

    const { result } = renderHook(() => useSidebar());

    act(() => {
      result.current.handleLogout();
    });

    expect(result.current.loggingOut).toBe(true);

    await act(async () => {
      resolveLogout!();
    });
  });

  it('ログアウト失敗時にloggingOutがfalseに戻る', async () => {
    mockLogout.mockRejectedValue(new Error('Network Error'));

    const { result } = renderHook(() => useSidebar());

    await act(async () => {
      await result.current.handleLogout();
    });

    expect(result.current.loggingOut).toBe(false);
  });
});
