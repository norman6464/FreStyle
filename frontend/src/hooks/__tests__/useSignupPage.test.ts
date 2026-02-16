import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useSignupPage } from '../useSignupPage';

const mockSignup = vi.fn();
const mockNavigate = vi.fn();

vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock('../useAuth', () => ({
  useAuth: () => ({ signup: mockSignup, loading: false }),
}));

describe('useSignupPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
  });

  it('初期状態でformが空文字列である', () => {
    const { result } = renderHook(() => useSignupPage());
    expect(result.current.form).toEqual({ email: '', password: '', name: '' });
  });

  it('初期状態でmessageがnullである', () => {
    const { result } = renderHook(() => useSignupPage());
    expect(result.current.message).toBeNull();
  });

  it('handleChangeでフォーム値が更新される', () => {
    const { result } = renderHook(() => useSignupPage());
    act(() => {
      result.current.handleChange({
        target: { name: 'email', value: 'test@example.com' },
      } as React.ChangeEvent<HTMLInputElement>);
    });
    expect(result.current.form.email).toBe('test@example.com');
  });

  it('handleChangeで名前が更新される', () => {
    const { result } = renderHook(() => useSignupPage());
    act(() => {
      result.current.handleChange({
        target: { name: 'name', value: 'テスト太郎' },
      } as React.ChangeEvent<HTMLInputElement>);
    });
    expect(result.current.form.name).toBe('テスト太郎');
  });

  it('サインアップ成功時にメッセージが設定される', async () => {
    mockSignup.mockResolvedValue(true);
    const { result } = renderHook(() => useSignupPage());

    act(() => {
      result.current.handleChange({ target: { name: 'name', value: 'テスト' } } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({ target: { name: 'email', value: 'test@example.com' } } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({ target: { name: 'password', value: 'pass123' } } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleSignup({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message).toEqual({
      type: 'success',
      text: 'サインアップに成功しました！',
    });
  });

  it('サインアップ成功後にナビゲートされる', async () => {
    mockSignup.mockResolvedValue(true);
    const { result } = renderHook(() => useSignupPage());

    act(() => {
      result.current.handleChange({ target: { name: 'name', value: 'テスト' } } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({ target: { name: 'email', value: 'test@example.com' } } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({ target: { name: 'password', value: 'pass123' } } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleSignup({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    act(() => {
      vi.advanceTimersByTime(1500);
    });

    expect(mockNavigate).toHaveBeenCalledWith('/confirm', {
      state: { message: 'サインアップに成功しました！メール確認をお願いします。' },
    });
  });

  it('サインアップ失敗時にエラーメッセージが設定される', async () => {
    mockSignup.mockResolvedValue(false);
    const { result } = renderHook(() => useSignupPage());

    act(() => {
      result.current.handleChange({ target: { name: 'name', value: 'テスト' } } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({ target: { name: 'email', value: 'test@example.com' } } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({ target: { name: 'password', value: 'pass123' } } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleSignup({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message).toEqual({
      type: 'error',
      text: '登録に失敗しました。',
    });
    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it('handleSignupでpreventDefaultが呼ばれる', async () => {
    mockSignup.mockResolvedValue(true);
    const { result } = renderHook(() => useSignupPage());
    const preventDefault = vi.fn();

    await act(async () => {
      await result.current.handleSignup({
        preventDefault,
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(preventDefault).toHaveBeenCalled();
  });

  it('サインアップ時にフォームの値が送信される', async () => {
    mockSignup.mockResolvedValue(true);
    const { result } = renderHook(() => useSignupPage());

    act(() => {
      result.current.handleChange({
        target: { name: 'email', value: 'user@test.com' },
      } as React.ChangeEvent<HTMLInputElement>);
    });
    act(() => {
      result.current.handleChange({
        target: { name: 'password', value: 'pass123' },
      } as React.ChangeEvent<HTMLInputElement>);
    });
    act(() => {
      result.current.handleChange({
        target: { name: 'name', value: 'テスト' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleSignup({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(mockSignup).toHaveBeenCalledWith({
      email: 'user@test.com',
      password: 'pass123',
      name: 'テスト',
    });
  });

  it('loadingがuseAuthから取得される', () => {
    const { result } = renderHook(() => useSignupPage());
    expect(result.current.loading).toBe(false);
  });

  it('ニックネームが空の場合エラーメッセージが表示されsignupが呼ばれない', async () => {
    const { result } = renderHook(() => useSignupPage());

    act(() => {
      result.current.handleChange({
        target: { name: 'email', value: 'test@example.com' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'password', value: 'password123' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleSignup({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('すべてのフィールドを入力してください。');
    expect(mockSignup).not.toHaveBeenCalled();
  });

  it('メールアドレスが空の場合エラーメッセージが表示されsignupが呼ばれない', async () => {
    const { result } = renderHook(() => useSignupPage());

    act(() => {
      result.current.handleChange({
        target: { name: 'name', value: 'テスト' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'password', value: 'password123' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleSignup({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('すべてのフィールドを入力してください。');
    expect(mockSignup).not.toHaveBeenCalled();
  });

  it('パスワードが空の場合エラーメッセージが表示されsignupが呼ばれない', async () => {
    const { result } = renderHook(() => useSignupPage());

    act(() => {
      result.current.handleChange({
        target: { name: 'name', value: 'テスト' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'email', value: 'test@example.com' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleSignup({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('すべてのフィールドを入力してください。');
    expect(mockSignup).not.toHaveBeenCalled();
  });

  it('メールアドレス形式が無効の場合エラーメッセージが表示されsignupが呼ばれない', async () => {
    const { result } = renderHook(() => useSignupPage());

    act(() => {
      result.current.handleChange({
        target: { name: 'name', value: 'テスト' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'email', value: 'invalid-email' },
      } as React.ChangeEvent<HTMLInputElement>);
      result.current.handleChange({
        target: { name: 'password', value: 'password123' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    await act(async () => {
      await result.current.handleSignup({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('有効なメールアドレスを入力してください。');
    expect(mockSignup).not.toHaveBeenCalled();
  });
});
