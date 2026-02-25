import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useLoginCallback } from '../useLoginCallback';

const mockNavigate = vi.fn();
const mockDispatch = vi.fn();

vi.mock('react-router-dom', () => ({
  useSearchParams: () => [new URLSearchParams(mockSearchParams)],
  useNavigate: () => mockNavigate,
}));

vi.mock('react-redux', () => ({
  useDispatch: () => mockDispatch,
}));

vi.mock('../../repositories/AuthRepository', () => ({
  default: {
    callback: vi.fn(),
  },
}));

vi.mock('../../store/authSlice', () => ({
  setAuthData: () => ({ type: 'auth/setAuthData' }),
}));

import authRepository from '../../repositories/AuthRepository';

let mockSearchParams = '';

describe('useLoginCallback', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams = '';
  });

  it('codeがある場合にauthRepository.callbackを呼び出す', async () => {
    mockSearchParams = 'code=test-code';
    vi.mocked(authRepository.callback).mockResolvedValue({} as any);

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(authRepository.callback).toHaveBeenCalledWith('test-code');
  });

  it('callback成功時にsetAuthDataをdispatchしホームに遷移する', async () => {
    mockSearchParams = 'code=test-code';
    vi.mocked(authRepository.callback).mockResolvedValue({} as any);

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockDispatch).toHaveBeenCalledWith({ type: 'auth/setAuthData' });
    expect(mockNavigate).toHaveBeenCalledWith('/', { state: { toast: 'ログインしました' } });
  });

  it('callback失敗時にトースト付きでログインページに遷移する', async () => {
    mockSearchParams = 'code=test-code';
    vi.mocked(authRepository.callback).mockRejectedValue(new Error('認証失敗'));

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockNavigate).toHaveBeenCalledWith('/login', { state: { toast: '認証に失敗しました' } });
  });

  it('errorパラメータがある場合にトースト付きでログインページに遷移する', async () => {
    mockSearchParams = 'error=access_denied';

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockNavigate).toHaveBeenCalledWith('/login', { state: { toast: '認証エラーが発生しました' } });
    expect(authRepository.callback).not.toHaveBeenCalled();
  });

  it('codeもerrorもない場合にログインページに遷移する', async () => {
    mockSearchParams = '';

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockNavigate).toHaveBeenCalledWith('/login');
    expect(authRepository.callback).not.toHaveBeenCalled();
  });

  it('errorパラメータがある場合にcallbackを呼ばない', async () => {
    mockSearchParams = 'error=server_error&code=test-code';

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(authRepository.callback).not.toHaveBeenCalled();
  });

  it('callback成功時にトースト付きでホームに遷移する', async () => {
    mockSearchParams = 'code=valid-code';
    vi.mocked(authRepository.callback).mockResolvedValue({} as any);

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockNavigate).toHaveBeenCalledWith('/', { state: { toast: 'ログインしました' } });
  });

  it('callback失敗時にトースト付きで遷移しない', async () => {
    mockSearchParams = 'code=invalid-code';
    vi.mocked(authRepository.callback).mockRejectedValue(new Error('認証失敗'));

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockNavigate).not.toHaveBeenCalledWith('/', expect.objectContaining({ state: expect.objectContaining({ toast: expect.any(String) }) }));
  });

  it('callback成功時にエラートーストなしでホームに遷移する', async () => {
    mockSearchParams = 'code=valid-code';
    vi.mocked(authRepository.callback).mockResolvedValue({} as any);

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockNavigate).not.toHaveBeenCalledWith('/login', expect.objectContaining({ state: expect.objectContaining({ toast: expect.any(String) }) }));
  });

  it('codeもerrorもない場合にdispatchが呼ばれない', async () => {
    mockSearchParams = '';

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockDispatch).not.toHaveBeenCalled();
  });
});
