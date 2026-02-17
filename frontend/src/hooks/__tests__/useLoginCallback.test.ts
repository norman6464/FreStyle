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

const mockShowToast = vi.fn();
vi.mock('../useToast', () => ({
  useToast: () => ({ showToast: mockShowToast, toasts: [], removeToast: vi.fn() }),
}));

import authRepository from '../../repositories/AuthRepository';

let mockSearchParams = '';

describe('useLoginCallback', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams = '';
    vi.spyOn(window, 'alert').mockImplementation(() => {});
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
    expect(mockNavigate).toHaveBeenCalledWith('/');
  });

  it('callback失敗時にアラートを表示しログインページに遷移する', async () => {
    mockSearchParams = 'code=test-code';
    vi.mocked(authRepository.callback).mockRejectedValue(new Error('認証失敗'));

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(window.alert).toHaveBeenCalledWith('認証に失敗しました。');
    expect(mockNavigate).toHaveBeenCalledWith('/login');
  });

  it('errorパラメータがある場合にアラートを表示しログインページに遷移する', async () => {
    mockSearchParams = 'error=access_denied';

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(window.alert).toHaveBeenCalledWith('認証エラーが発生しました。access_denied');
    expect(mockNavigate).toHaveBeenCalledWith('/login');
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

  it('callback成功時にログインしましたトーストを表示する', async () => {
    mockSearchParams = 'code=valid-code';
    vi.mocked(authRepository.callback).mockResolvedValue({} as any);

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockShowToast).toHaveBeenCalledWith('success', 'ログインしました');
  });

  it('callback失敗時にトーストを表示しない', async () => {
    mockSearchParams = 'code=invalid-code';
    vi.mocked(authRepository.callback).mockRejectedValue(new Error('認証失敗'));

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockShowToast).not.toHaveBeenCalled();
  });

  it('callback成功時にalertが呼ばれない', async () => {
    mockSearchParams = 'code=valid-code';
    vi.mocked(authRepository.callback).mockResolvedValue({} as any);

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(window.alert).not.toHaveBeenCalled();
  });

  it('codeもerrorもない場合にdispatchが呼ばれない', async () => {
    mockSearchParams = '';

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockDispatch).not.toHaveBeenCalled();
  });
});
