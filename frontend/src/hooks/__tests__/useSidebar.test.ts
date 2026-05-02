import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
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

const mockLogout = vi.fn();

vi.mock('../../repositories/AuthRepository', () => ({
  default: {
    logout: (...args: unknown[]) => mockLogout(...args),
  },
}));

describe('useSidebar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockLogout.mockResolvedValue(undefined);
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

  it('ログアウト成功時にトースト付きでログインページに遷移する', async () => {
    const { result } = renderHook(() => useSidebar());

    await act(async () => {
      await result.current.handleLogout();
    });

    expect(mockNavigate).toHaveBeenCalledWith('/login', { state: { toast: 'ログアウトしました' } });
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
