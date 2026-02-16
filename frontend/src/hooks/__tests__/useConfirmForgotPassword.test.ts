import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { AxiosError } from 'axios';
import { useConfirmForgotPassword } from '../useConfirmForgotPassword';

const mockConfirmForgotPassword = vi.fn();
const mockNavigate = vi.fn();

vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
  useLocation: () => ({ state: { email: 'test@example.com' } }),
}));

vi.mock('../../repositories/AuthRepository', () => ({
  default: {
    confirmForgotPassword: (...args: unknown[]) => mockConfirmForgotPassword(...args),
  },
}));

describe('useConfirmForgotPassword', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockConfirmForgotPassword.mockResolvedValue({ message: '成功' });
  });

  it('location.stateからemailが初期値にセットされる', () => {
    const { result } = renderHook(() => useConfirmForgotPassword());
    expect(result.current.form.email).toBe('test@example.com');
  });

  it('handleChangeでフォームの値が更新される', () => {
    const { result } = renderHook(() => useConfirmForgotPassword());

    act(() => {
      result.current.handleChange({
        target: { name: 'code', value: '123456' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    expect(result.current.form.code).toBe('123456');
  });

  it('handleConfirm成功時にログインページへナビゲートする', async () => {
    const { result } = renderHook(() => useConfirmForgotPassword());

    act(() => {
      result.current.handleChange({
        target: { name: 'code', value: '123456' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'newPassword', value: 'newpass123' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'confirmPassword', value: 'newpass123' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleConfirm({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(mockConfirmForgotPassword).toHaveBeenCalledWith({
      email: 'test@example.com',
      confirmationCode: '123456',
      newPassword: 'newpass123',
    });
    expect(mockNavigate).toHaveBeenCalledWith('/login', {
      state: { message: 'パスワードリセットに成功しました。ログインしてください。' },
    });
  });

  it('handleConfirm失敗時にエラーメッセージが表示される', async () => {
    const axiosError = new AxiosError('Bad Request');
    (axiosError as any).response = { data: { error: '確認コードが無効です' } };
    mockConfirmForgotPassword.mockRejectedValue(axiosError);

    const { result } = renderHook(() => useConfirmForgotPassword());

    act(() => {
      result.current.handleChange({
        target: { name: 'code', value: '123456' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'newPassword', value: 'newpass123' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'confirmPassword', value: 'newpass123' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleConfirm({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('確認コードが無効です');
  });

  it('通信エラー時に汎用エラーメッセージが表示される', async () => {
    mockConfirmForgotPassword.mockRejectedValue(new Error('Network Error'));

    const { result } = renderHook(() => useConfirmForgotPassword());

    act(() => {
      result.current.handleChange({
        target: { name: 'code', value: '123456' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'newPassword', value: 'newpass123' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'confirmPassword', value: 'newpass123' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleConfirm({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('通信エラーが発生しました。');
  });

  it('初期messageがnull', () => {
    const { result } = renderHook(() => useConfirmForgotPassword());
    expect(result.current.message).toBeNull();
  });

  it('code・newPasswordの初期値が空文字', () => {
    const { result } = renderHook(() => useConfirmForgotPassword());
    expect(result.current.form.code).toBe('');
    expect(result.current.form.newPassword).toBe('');
  });

  it('handleConfirmでpreventDefaultが呼ばれる', async () => {
    const preventDefault = vi.fn();
    const { result } = renderHook(() => useConfirmForgotPassword());

    await act(async () => {
      await result.current.handleConfirm({
        preventDefault,
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(preventDefault).toHaveBeenCalled();
  });

  it('初期状態でloadingがfalseである', () => {
    const { result } = renderHook(() => useConfirmForgotPassword());
    expect(result.current.loading).toBe(false);
  });

  it('handleConfirm実行中にloadingがtrueになる', async () => {
    let resolvePromise: (value: unknown) => void;
    mockConfirmForgotPassword.mockReturnValue(new Promise((resolve) => { resolvePromise = resolve; }));
    const { result } = renderHook(() => useConfirmForgotPassword());

    act(() => {
      result.current.handleChange({
        target: { name: 'code', value: '123456' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'newPassword', value: 'newpass123' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'confirmPassword', value: 'newpass123' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    let confirmPromise: Promise<void>;
    act(() => {
      confirmPromise = result.current.handleConfirm({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.loading).toBe(true);

    await act(async () => {
      resolvePromise!({ message: 'ok' });
      await confirmPromise;
    });

    expect(result.current.loading).toBe(false);
  });

  it('confirmPasswordの初期値が空文字', () => {
    const { result } = renderHook(() => useConfirmForgotPassword());
    expect(result.current.form.confirmPassword).toBe('');
  });

  it('パスワード不一致時にエラーメッセージが表示される', async () => {
    const { result } = renderHook(() => useConfirmForgotPassword());

    act(() => {
      result.current.handleChange({
        target: { name: 'code', value: '123456' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'newPassword', value: 'password1' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'confirmPassword', value: 'password2' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleConfirm({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('パスワードが一致しません。');
    expect(mockConfirmForgotPassword).not.toHaveBeenCalled();
  });

  it('パスワード一致時にAPIが呼ばれる', async () => {
    const { result } = renderHook(() => useConfirmForgotPassword());

    act(() => {
      result.current.handleChange({
        target: { name: 'code', value: '123456' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'newPassword', value: 'samePassword' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'confirmPassword', value: 'samePassword' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleConfirm({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(mockConfirmForgotPassword).toHaveBeenCalled();
  });

  it('確認コードが空の場合エラーメッセージが表示されAPIが呼ばれない', async () => {
    const { result } = renderHook(() => useConfirmForgotPassword());

    act(() => {
      result.current.handleChange({
        target: { name: 'newPassword', value: 'password123' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'confirmPassword', value: 'password123' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleConfirm({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('すべてのフィールドを入力してください。');
    expect(mockConfirmForgotPassword).not.toHaveBeenCalled();
  });

  it('新パスワードが空の場合エラーメッセージが表示されAPIが呼ばれない', async () => {
    const { result } = renderHook(() => useConfirmForgotPassword());

    act(() => {
      result.current.handleChange({
        target: { name: 'code', value: '123456' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleConfirm({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('すべてのフィールドを入力してください。');
    expect(mockConfirmForgotPassword).not.toHaveBeenCalled();
  });
});
