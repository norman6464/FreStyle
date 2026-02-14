import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { AxiosError } from 'axios';
import { useForgotPassword } from '../useForgotPassword';

const mockForgotPassword = vi.fn();
const mockNavigate = vi.fn();

vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock('../../repositories/AuthRepository', () => ({
  default: {
    forgotPassword: (...args: unknown[]) => mockForgotPassword(...args),
  },
}));

describe('useForgotPassword', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockForgotPassword.mockResolvedValue({ message: 'コード送信済み' });
  });

  it('初期値が空文字', () => {
    const { result } = renderHook(() => useForgotPassword());
    expect(result.current.email).toBe('');
  });

  it('setEmailでメールアドレスが更新される', () => {
    const { result } = renderHook(() => useForgotPassword());

    act(() => {
      result.current.setEmail('test@example.com');
    });

    expect(result.current.email).toBe('test@example.com');
  });

  it('handleSubmit成功時にナビゲートする', async () => {
    const { result } = renderHook(() => useForgotPassword());

    act(() => {
      result.current.setEmail('test@example.com');
    });

    await act(async () => {
      await result.current.handleSubmit({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(mockForgotPassword).toHaveBeenCalledWith({ email: 'test@example.com' });
    expect(mockNavigate).toHaveBeenCalledWith('/confirm-forgot-password', {
      state: { email: 'test@example.com' },
    });
  });

  it('handleSubmit失敗時にエラーメッセージが表示される', async () => {
    const axiosError = new AxiosError('Bad Request');
    (axiosError as any).response = { data: { error: 'ユーザーが見つかりません' } };
    mockForgotPassword.mockRejectedValue(axiosError);

    const { result } = renderHook(() => useForgotPassword());

    await act(async () => {
      await result.current.handleSubmit({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('ユーザーが見つかりません');
  });

  it('通信エラー時に汎用エラーメッセージが表示される', async () => {
    mockForgotPassword.mockRejectedValue(new Error('Network Error'));

    const { result } = renderHook(() => useForgotPassword());

    await act(async () => {
      await result.current.handleSubmit({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('通信エラーが発生しました。');
  });

  it('初期messageがnull', () => {
    const { result } = renderHook(() => useForgotPassword());
    expect(result.current.message).toBeNull();
  });

  it('handleSubmit成功時にsuccessメッセージが設定される', async () => {
    const { result } = renderHook(() => useForgotPassword());

    act(() => {
      result.current.setEmail('test@example.com');
    });

    await act(async () => {
      await result.current.handleSubmit({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message?.type).toBe('success');
    expect(result.current.message?.text).toBe('コード送信済み');
  });

  it('handleSubmitでpreventDefaultが呼ばれる', async () => {
    const preventDefault = vi.fn();
    const { result } = renderHook(() => useForgotPassword());

    await act(async () => {
      await result.current.handleSubmit({
        preventDefault,
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(preventDefault).toHaveBeenCalled();
  });
});
