import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import ConfirmForgotPasswordPage from '../ConfirmForgotPasswordPage';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useLocation: () => ({ state: { email: 'test@example.com' } }),
  };
});

describe('ConfirmForgotPasswordPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('タイトルが表示される', () => {
    render(<BrowserRouter><ConfirmForgotPasswordPage /></BrowserRouter>);

    expect(screen.getByText('パスワードリセット確認')).toBeInTheDocument();
  });

  it('メールアドレスが事前入力される', () => {
    render(<BrowserRouter><ConfirmForgotPasswordPage /></BrowserRouter>);

    const emailInput = screen.getByLabelText('メールアドレス') as HTMLInputElement;
    expect(emailInput.value).toBe('test@example.com');
  });

  it('確認コード入力欄が表示される', () => {
    render(<BrowserRouter><ConfirmForgotPasswordPage /></BrowserRouter>);

    expect(screen.getByLabelText('確認コード')).toBeInTheDocument();
  });

  it('新しいパスワード入力欄が表示される', () => {
    render(<BrowserRouter><ConfirmForgotPasswordPage /></BrowserRouter>);

    expect(screen.getByLabelText('新しいパスワード')).toBeInTheDocument();
  });

  it('リセットボタンが表示される', () => {
    render(<BrowserRouter><ConfirmForgotPasswordPage /></BrowserRouter>);

    expect(screen.getByText('パスワードをリセット')).toBeInTheDocument();
  });

  it('リセット成功時にログインページへ遷移する', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ message: 'リセット成功' }),
    });

    render(<BrowserRouter><ConfirmForgotPasswordPage /></BrowserRouter>);

    fireEvent.change(screen.getByLabelText('確認コード'), {
      target: { value: '123456', name: 'code' },
    });
    fireEvent.change(screen.getByLabelText('新しいパスワード'), {
      target: { value: 'newPassword123', name: 'newPassword' },
    });
    fireEvent.click(screen.getByText('パスワードをリセット'));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login', {
        state: { message: 'パスワードリセットに成功しました。ログインしてください。' },
      });
    });
  });

  it('リセット失敗時にエラーメッセージを表示する', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      json: () => Promise.resolve({ error: '確認コードが無効です' }),
    });

    render(<BrowserRouter><ConfirmForgotPasswordPage /></BrowserRouter>);

    fireEvent.change(screen.getByLabelText('確認コード'), {
      target: { value: '000000', name: 'code' },
    });
    fireEvent.click(screen.getByText('パスワードをリセット'));

    await waitFor(() => {
      expect(screen.getByText('確認コードが無効です')).toBeInTheDocument();
    });
  });
});
