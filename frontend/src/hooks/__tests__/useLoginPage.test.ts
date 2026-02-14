import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useLoginPage } from '../useLoginPage';

const mockLogin = vi.fn();
const mockNavigate = vi.fn();
const mockLocationState = { message: '' };

vi.mock('react-router-dom', () => ({
  useLocation: () => ({ state: mockLocationState }),
  useNavigate: () => mockNavigate,
}));

vi.mock('../useAuth', () => ({
  useAuth: () => ({ login: mockLogin }),
}));

describe('useLoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockLocationState.message = '';
  });

  it('初期状態でformが空文字列である', () => {
    const { result } = renderHook(() => useLoginPage());
    expect(result.current.form).toEqual({ email: '', password: '' });
  });

  it('初期状態でloginMessageがnullである', () => {
    const { result } = renderHook(() => useLoginPage());
    expect(result.current.loginMessage).toBeNull();
  });

  it('handleChangeでフォーム値が更新される', () => {
    const { result } = renderHook(() => useLoginPage());
    act(() => {
      result.current.handleChange({
        target: { name: 'email', value: 'test@example.com' },
      } as React.ChangeEvent<HTMLInputElement>);
    });
    expect(result.current.form.email).toBe('test@example.com');
  });

  it('handleChangeでパスワードが更新される', () => {
    const { result } = renderHook(() => useLoginPage());
    act(() => {
      result.current.handleChange({
        target: { name: 'password', value: 'secret123' },
      } as React.ChangeEvent<HTMLInputElement>);
    });
    expect(result.current.form.password).toBe('secret123');
  });

  it('ログイン成功時にナビゲートされる', async () => {
    mockLogin.mockResolvedValue(true);
    const { result } = renderHook(() => useLoginPage());

    await act(async () => {
      await result.current.handleLogin({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(mockLogin).toHaveBeenCalledWith({ email: '', password: '' });
    expect(mockNavigate).toHaveBeenCalledWith('/');
  });

  it('ログイン失敗時にエラーメッセージが設定される', async () => {
    mockLogin.mockResolvedValue(false);
    const { result } = renderHook(() => useLoginPage());

    await act(async () => {
      await result.current.handleLogin({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(result.current.loginMessage).toEqual({
      type: 'error',
      text: 'ログインに失敗しました。',
    });
    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it('handleLoginでpreventDefaultが呼ばれる', async () => {
    mockLogin.mockResolvedValue(true);
    const { result } = renderHook(() => useLoginPage());
    const preventDefault = vi.fn();

    await act(async () => {
      await result.current.handleLogin({
        preventDefault,
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(preventDefault).toHaveBeenCalled();
  });

  it('locationStateのメッセージが取得される', () => {
    mockLocationState.message = '確認メールを送信しました';
    const { result } = renderHook(() => useLoginPage());
    expect(result.current.flashMessage).toBe('確認メールを送信しました');
  });

  it('locationStateにメッセージがない場合はundefined', () => {
    mockLocationState.message = '';
    const { result } = renderHook(() => useLoginPage());
    expect(result.current.flashMessage).toBe('');
  });

  it('ログイン時にフォームの値が送信される', async () => {
    mockLogin.mockResolvedValue(true);
    const { result } = renderHook(() => useLoginPage());

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

    await act(async () => {
      await result.current.handleLogin({
        preventDefault: vi.fn(),
      } as unknown as React.FormEvent<HTMLFormElement>);
    });

    expect(mockLogin).toHaveBeenCalledWith({ email: 'user@test.com', password: 'pass123' });
  });
});
