import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import AddUserPage from '../AddUserPage';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock('react-redux', () => ({
  useDispatch: () => vi.fn(),
}));

describe('AddUserPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ users: [] }),
    });
  });

  it('検索ボックスが表示される', () => {
    render(<BrowserRouter><AddUserPage /></BrowserRouter>);

    expect(screen.getByPlaceholderText('ユーザー名またはメールアドレスで検索...')).toBeInTheDocument();
  });

  it('初期状態で検索案内を表示する', () => {
    render(<BrowserRouter><AddUserPage /></BrowserRouter>);

    expect(screen.getByText('ユーザーを検索してみましょう')).toBeInTheDocument();
  });

  it('ユーザー取得後にユーザー一覧を表示する', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({
        users: [
          { id: 1, name: '山田太郎', email: 'yamada@example.com', roomId: null },
        ],
      }),
    });

    render(<BrowserRouter><AddUserPage /></BrowserRouter>);

    await waitFor(() => {
      expect(screen.getByText('山田太郎')).toBeInTheDocument();
    });
  });

  it('fetch失敗時にエラーメッセージを表示する', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      json: () => Promise.resolve({}),
    });

    render(<BrowserRouter><AddUserPage /></BrowserRouter>);

    await waitFor(() => {
      expect(screen.getByText('ユーザー取得に失敗しました')).toBeInTheDocument();
    });
  });

  it('401レスポンス時にログインページへリダイレクトする', async () => {
    global.fetch = vi.fn()
      .mockResolvedValueOnce({ status: 401, ok: false })
      .mockResolvedValueOnce({ ok: false, status: 401 });

    render(<BrowserRouter><AddUserPage /></BrowserRouter>);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login');
    });
  });

  it('ユーザー数を表示する', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({
        users: [
          { id: 1, name: 'ユーザー1', email: 'u1@example.com', roomId: null },
          { id: 2, name: 'ユーザー2', email: 'u2@example.com', roomId: null },
        ],
      }),
    });

    render(<BrowserRouter><AddUserPage /></BrowserRouter>);

    await waitFor(() => {
      expect(screen.getByText('2人のユーザーが見つかりました')).toBeInTheDocument();
    });
  });
});
