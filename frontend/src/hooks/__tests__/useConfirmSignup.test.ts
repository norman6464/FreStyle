import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { AxiosError } from 'axios';
import { useConfirmSignup } from '../useConfirmSignup';

const mockConfirmSignup = vi.fn();
const mockNavigate = vi.fn();

vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock('../../repositories/AuthRepository', () => ({
  default: {
    confirmSignup: (...args: unknown[]) => mockConfirmSignup(...args),
  },
}));

describe('useConfirmSignup', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockConfirmSignup.mockResolvedValue({ message: '確認成功' });
  });

  it('初期フォーム値が空文字', () => {
    const { result } = renderHook(() => useConfirmSignup());
    expect(result.current.form.email).toBe('');
    expect(result.current.form.code).toBe('');
  });

  it('handleChangeでフォームの値が更新される', () => {
    const { result } = renderHook(() => useConfirmSignup());

    act(() => {
      result.current.handleChange({
        target: { name: 'email', value: 'test@example.com' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    expect(result.current.form.email).toBe('test@example.com');
  });

  it('handleConfirm成功時にログインページへナビゲートする', async () => {
    const { result } = renderHook(() => useConfirmSignup());

    act(() => {
      result.current.handleChange({
        target: { name: 'email', value: 'test@example.com' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'code', value: '123456' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleConfirm({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(mockConfirmSignup).toHaveBeenCalledWith({
      email: 'test@example.com',
      code: '123456',
    });
    expect(mockNavigate).toHaveBeenCalledWith('/login', {
      state: { message: '確認に成功しました。ログインしてください。' },
    });
  });

  it('handleConfirm失敗時にエラーメッセージが表示される', async () => {
    const axiosError = new AxiosError('Bad Request');
    (axiosError as any).response = { data: { error: '確認コードが無効です' } };
    mockConfirmSignup.mockRejectedValue(axiosError);

    const { result } = renderHook(() => useConfirmSignup());

    await act(async () => {
      await result.current.handleConfirm({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('確認コードが無効です');
  });

  it('通信エラー時に汎用エラーメッセージが表示される', async () => {
    mockConfirmSignup.mockRejectedValue(new Error('Network Error'));

    const { result } = renderHook(() => useConfirmSignup());

    await act(async () => {
      await result.current.handleConfirm({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('通信エラーが発生しました。');
  });

  it('初期messageがnull', () => {
    const { result } = renderHook(() => useConfirmSignup());
    expect(result.current.message).toBeNull();
  });

  it('codeフィールドのhandleChangeで値が更新される', () => {
    const { result } = renderHook(() => useConfirmSignup());

    act(() => {
      result.current.handleChange({
        target: { name: 'code', value: '654321' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    expect(result.current.form.code).toBe('654321');
    expect(result.current.form.email).toBe('');
  });

  it('handleConfirmでpreventDefaultが呼ばれる', async () => {
    const preventDefault = vi.fn();
    const { result } = renderHook(() => useConfirmSignup());

    await act(async () => {
      await result.current.handleConfirm({
        preventDefault,
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(preventDefault).toHaveBeenCalledOnce();
  });
});
