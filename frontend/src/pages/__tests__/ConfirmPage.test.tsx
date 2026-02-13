import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import ConfirmPage from '../ConfirmPage';
import authRepository from '../../repositories/AuthRepository';
import { AxiosError } from 'axios';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock('../../repositories/AuthRepository');

describe('ConfirmPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('タイトルが表示される', () => {
    render(<BrowserRouter><ConfirmPage /></BrowserRouter>);

    expect(screen.getByText('確認コードの入力')).toBeInTheDocument();
  });

  it('メールアドレス入力欄が表示される', () => {
    render(<BrowserRouter><ConfirmPage /></BrowserRouter>);

    expect(screen.getByLabelText('メールアドレス')).toBeInTheDocument();
  });

  it('確認コード入力欄が表示される', () => {
    render(<BrowserRouter><ConfirmPage /></BrowserRouter>);

    expect(screen.getByLabelText('確認コード')).toBeInTheDocument();
  });

  it('確認ボタンが表示される', () => {
    render(<BrowserRouter><ConfirmPage /></BrowserRouter>);

    expect(screen.getByText('確認する')).toBeInTheDocument();
  });

  it('アカウント作成リンクが表示される', () => {
    render(<BrowserRouter><ConfirmPage /></BrowserRouter>);

    expect(screen.getByText('アカウント作成に戻る')).toBeInTheDocument();
  });

  it('確認成功時にログインページへ遷移する', async () => {
    vi.mocked(authRepository.confirmSignup).mockResolvedValue({ message: '確認成功' });

    render(<BrowserRouter><ConfirmPage /></BrowserRouter>);

    fireEvent.change(screen.getByLabelText('メールアドレス'), {
      target: { value: 'test@example.com', name: 'email' },
    });
    fireEvent.change(screen.getByLabelText('確認コード'), {
      target: { value: '123456', name: 'code' },
    });
    fireEvent.click(screen.getByText('確認する'));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login', {
        state: { message: '確認に成功しました。ログインしてください。' },
      });
    });
  });

  it('確認失敗時にエラーメッセージを表示する', async () => {
    const axiosError = new AxiosError('Request failed');
    (axiosError as any).response = { data: { error: '確認コードが無効です' } };
    vi.mocked(authRepository.confirmSignup).mockRejectedValue(axiosError);

    render(<BrowserRouter><ConfirmPage /></BrowserRouter>);

    fireEvent.change(screen.getByLabelText('メールアドレス'), {
      target: { value: 'test@example.com', name: 'email' },
    });
    fireEvent.change(screen.getByLabelText('確認コード'), {
      target: { value: '000000', name: 'code' },
    });
    fireEvent.click(screen.getByText('確認する'));

    await waitFor(() => {
      expect(screen.getByText('確認コードが無効です')).toBeInTheDocument();
    });
  });

  it('通信エラー時にエラーメッセージを表示する', async () => {
    vi.mocked(authRepository.confirmSignup).mockRejectedValue(new Error('Network error'));

    render(<BrowserRouter><ConfirmPage /></BrowserRouter>);

    fireEvent.click(screen.getByText('確認する'));

    await waitFor(() => {
      expect(screen.getByText('通信エラーが発生しました。')).toBeInTheDocument();
    });
  });
});
