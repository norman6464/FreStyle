import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useLoginPage } from '../useLoginPage';
import authRepository from '../../repositories/AuthRepository';

const mockLocationState: { message: string } = { message: '' };

vi.mock('react-router-dom', () => ({
  useLocation: () => ({ state: mockLocationState }),
}));

vi.mock('../../repositories/AuthRepository', () => ({
  default: { login: vi.fn() },
}));

vi.mock('axios', () => ({
  default: { isAxiosError: (e: unknown) => !!(e as { isAxiosError?: boolean })?.isAxiosError },
}));

const loginMock = authRepository.login as ReturnType<typeof vi.fn>;
const assignMock = vi.fn();

beforeEach(() => {
  vi.clearAllMocks();
  mockLocationState.message = '';
  Object.defineProperty(window, 'location', {
    value: { assign: assignMock },
    writable: true,
  });
});

function submitEvent() {
  return { preventDefault: vi.fn() } as unknown as React.FormEvent<HTMLFormElement>;
}

describe('useLoginPage', () => {
  it('flashMessage を location state から取得する', () => {
    mockLocationState.message = '確認メールを送信しました';
    const { result } = renderHook(() => useLoginPage());
    expect(result.current.flashMessage).toBe('確認メールを送信しました');
  });

  it('ログイン成功でトップへフル遷移する', async () => {
    loginMock.mockResolvedValueOnce({});
    const { result } = renderHook(() => useLoginPage());

    await act(async () => {
      await result.current.handleLogin(submitEvent());
    });

    expect(loginMock).toHaveBeenCalledWith({ email: '', password: '' });
    expect(assignMock).toHaveBeenCalledWith('/');
  });

  it('資格情報誤り（非 403）はエラーメッセージを表示し遷移しない', async () => {
    loginMock.mockRejectedValueOnce({ response: { status: 401 } });
    const { result } = renderHook(() => useLoginPage());

    await act(async () => {
      await result.current.handleLogin(submitEvent());
    });

    expect(result.current.loginMessage?.text).toContain('正しくありません');
    expect(assignMock).not.toHaveBeenCalled();
  });

  it('招待なし（403）は招待文言を表示する', async () => {
    loginMock.mockRejectedValueOnce({
      isAxiosError: true,
      response: { status: 403, data: { message: 'FreStyle のご利用には管理者からの招待が必要です。' } },
    });
    const { result } = renderHook(() => useLoginPage());

    await act(async () => {
      await result.current.handleLogin(submitEvent());
    });

    expect(result.current.loginMessage?.text).toBe('FreStyle のご利用には管理者からの招待が必要です。');
  });
});
