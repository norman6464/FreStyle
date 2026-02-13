import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import ForgotPasswordPage from '../ForgotPasswordPage';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

describe('ForgotPasswordPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('タイトルが表示される', () => {
    render(<BrowserRouter><ForgotPasswordPage /></BrowserRouter>);

    expect(screen.getByText('パスワードリセット')).toBeInTheDocument();
  });

  it('メールアドレス入力欄が表示される', () => {
    render(<BrowserRouter><ForgotPasswordPage /></BrowserRouter>);

    expect(screen.getByLabelText('メールアドレス')).toBeInTheDocument();
  });

  it('送信ボタンが表示される', () => {
    render(<BrowserRouter><ForgotPasswordPage /></BrowserRouter>);

    expect(screen.getByText('確認コードを送信')).toBeInTheDocument();
  });

  it('送信成功時にリセット確認ページへ遷移する', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ message: '確認コードを送信しました' }),
    });

    render(<BrowserRouter><ForgotPasswordPage /></BrowserRouter>);

    fireEvent.change(screen.getByLabelText('メールアドレス'), {
      target: { value: 'test@example.com' },
    });
    fireEvent.click(screen.getByText('確認コードを送信'));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/confirm-forgot-password', {
        state: { email: 'test@example.com' },
      });
    });
  });

  it('送信失敗時にエラーメッセージを表示する', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      json: () => Promise.resolve({ error: 'メールアドレスが見つかりません' }),
    });

    render(<BrowserRouter><ForgotPasswordPage /></BrowserRouter>);

    fireEvent.change(screen.getByLabelText('メールアドレス'), {
      target: { value: 'invalid@example.com' },
    });
    fireEvent.click(screen.getByText('確認コードを送信'));

    await waitFor(() => {
      expect(screen.getByText('メールアドレスが見つかりません')).toBeInTheDocument();
    });
  });

  it('通信エラー時にエラーメッセージを表示する', async () => {
    global.fetch = vi.fn().mockRejectedValue(new Error('Network error'));

    render(<BrowserRouter><ForgotPasswordPage /></BrowserRouter>);

    fireEvent.change(screen.getByLabelText('メールアドレス'), {
      target: { value: 'test@example.com' },
    });
    fireEvent.click(screen.getByText('確認コードを送信'));

    await waitFor(() => {
      expect(screen.getByText('通信エラーが発生しました。')).toBeInTheDocument();
    });
  });
});
