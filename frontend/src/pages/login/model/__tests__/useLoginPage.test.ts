import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { AxiosError, AxiosHeaders } from 'axios';
import { useLoginPage } from '../useLoginPage';
import authRepository from '@/entities/user/api/authRepository';

const mockLocationState: { message: string } = { message: '' };

vi.mock('react-router-dom', () => ({
  useLocation: () => ({ state: mockLocationState }),
}));

vi.mock('@/entities/user/api/authRepository', () => ({
  default: { loginWithChallenge: vi.fn(), submitNewPassword: vi.fn() },
}));

// 本物の AxiosError を組み立てる（getApiError は instanceof AxiosError で判定するため）。
function axiosError(status: number, data: Record<string, unknown> = {}): AxiosError {
  return new AxiosError('Request failed', 'ERR_BAD_REQUEST', undefined, undefined, {
    status,
    statusText: '',
    headers: {},
    config: { headers: new AxiosHeaders() },
    data,
  });
}

const loginMock = authRepository.loginWithChallenge as ReturnType<typeof vi.fn>;
const newPasswordMock = authRepository.submitNewPassword as ReturnType<typeof vi.fn>;
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
    loginMock.mockResolvedValueOnce({ kind: 'success' });
    const { result } = renderHook(() => useLoginPage());

    await act(async () => {
      await result.current.handleLogin(submitEvent());
    });

    expect(loginMock).toHaveBeenCalledWith({ email: '', password: '' });
    expect(assignMock).toHaveBeenCalledWith('/');
  });

  it('資格情報誤り（非 403）はエラーメッセージを表示し遷移しない', async () => {
    loginMock.mockRejectedValueOnce(axiosError(401));
    const { result } = renderHook(() => useLoginPage());

    await act(async () => {
      await result.current.handleLogin(submitEvent());
    });

    expect(result.current.loginMessage?.text).toContain('正しくありません');
    expect(assignMock).not.toHaveBeenCalled();
  });

  it('招待なし（403）は招待文言を表示する', async () => {
    loginMock.mockRejectedValueOnce(
      axiosError(403, { message: 'FreStyle のご利用には管理者からの招待が必要です。' }),
    );
    const { result } = renderHook(() => useLoginPage());

    await act(async () => {
      await result.current.handleLogin(submitEvent());
    });

    expect(result.current.loginMessage?.text).toBe('FreStyle のご利用には管理者からの招待が必要です。');
  });

  it('NEW_PASSWORD_REQUIRED で新パスワード設定フェーズに入り、遷移しない', async () => {
    loginMock.mockResolvedValueOnce({ kind: 'new_password_required', session: 'sess-1' });
    const { result } = renderHook(() => useLoginPage());

    await act(async () => {
      await result.current.handleLogin(submitEvent());
    });

    expect(result.current.newPasswordPhase).not.toBeNull();
    expect(result.current.loginMessage?.type).toBe('success');
    expect(assignMock).not.toHaveBeenCalled();
  });

  it('新パスワード設定成功でトップへフル遷移する', async () => {
    loginMock.mockResolvedValueOnce({ kind: 'new_password_required', session: 'sess-1' });
    newPasswordMock.mockResolvedValueOnce(undefined);
    const { result } = renderHook(() => useLoginPage());

    await act(async () => {
      await result.current.handleLogin(submitEvent());
    });
    act(() => {
      result.current.setNewPassword('New-Pass-1');
      result.current.setNewPasswordConfirm('New-Pass-1');
    });
    await act(async () => {
      await result.current.handleNewPassword(submitEvent());
    });

    expect(newPasswordMock).toHaveBeenCalledWith({
      email: '',
      session: 'sess-1',
      newPassword: 'New-Pass-1',
    });
    expect(assignMock).toHaveBeenCalledWith('/');
  });

  it('新パスワードの確認不一致はエラーで送信しない', async () => {
    loginMock.mockResolvedValueOnce({ kind: 'new_password_required', session: 'sess-1' });
    const { result } = renderHook(() => useLoginPage());
    await act(async () => {
      await result.current.handleLogin(submitEvent());
    });
    act(() => {
      result.current.setNewPassword('a');
      result.current.setNewPasswordConfirm('b');
    });
    await act(async () => {
      await result.current.handleNewPassword(submitEvent());
    });

    expect(newPasswordMock).not.toHaveBeenCalled();
    expect(result.current.loginMessage?.text).toContain('一致しません');
  });

  it('新パスワード設定の session 失効(401)は最初からやり直させる', async () => {
    loginMock.mockResolvedValueOnce({ kind: 'new_password_required', session: 'sess-1' });
    newPasswordMock.mockRejectedValueOnce(axiosError(401));
    const { result } = renderHook(() => useLoginPage());
    await act(async () => {
      await result.current.handleLogin(submitEvent());
    });
    act(() => {
      result.current.setNewPassword('New-Pass-1');
      result.current.setNewPasswordConfirm('New-Pass-1');
    });
    await act(async () => {
      await result.current.handleNewPassword(submitEvent());
    });

    expect(result.current.newPasswordPhase).toBeNull();
    expect(result.current.loginMessage?.text).toContain('有効期限');
    expect(assignMock).not.toHaveBeenCalled();
  });
});
