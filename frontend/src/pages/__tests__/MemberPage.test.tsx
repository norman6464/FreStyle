import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import MemberPage from '../MemberPage';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

describe('MemberPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ users: [] }),
    });
  });

  it('ヘッダーが表示される', () => {
    render(<BrowserRouter><MemberPage /></BrowserRouter>);

    expect(screen.getByText('チャットメンバー')).toBeInTheDocument();
    expect(screen.getByText('メンバーを検索または選択')).toBeInTheDocument();
  });

  it('検索ボックスが表示される', () => {
    render(<BrowserRouter><MemberPage /></BrowserRouter>);

    expect(screen.getByPlaceholderText('メンバーを検索...')).toBeInTheDocument();
  });

  it('メンバーがいない場合にメッセージを表示する', () => {
    render(<BrowserRouter><MemberPage /></BrowserRouter>);

    expect(screen.getByText('メンバーがまだいません')).toBeInTheDocument();
  });

  it('メンバー取得後に一覧を表示する', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({
        users: [
          { id: 1, name: '鈴木一郎', email: 'suzuki@example.com', roomId: 10 },
        ],
      }),
    });

    render(<BrowserRouter><MemberPage /></BrowserRouter>);

    await waitFor(() => {
      expect(screen.getByText('鈴木一郎')).toBeInTheDocument();
    });
  });

  it('401レスポンス時にログインページへリダイレクトする', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 401,
    });

    render(<BrowserRouter><MemberPage /></BrowserRouter>);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login');
    });
  });

  it('500レスポンス時にトップページへリダイレクトする', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
    });

    render(<BrowserRouter><MemberPage /></BrowserRouter>);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/');
    });
  });
});
