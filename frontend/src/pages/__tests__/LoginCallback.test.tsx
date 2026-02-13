import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter } from 'react-router-dom';
import LoginCallback from '../LoginCallback';
import authReducer from '../../store/authSlice';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

function renderWithRoute(search: string) {
  const store = configureStore({ reducer: { auth: authReducer } });
  return render(
    <Provider store={store}>
      <MemoryRouter initialEntries={[`/callback${search}`]}>
        <LoginCallback />
      </MemoryRouter>
    </Provider>,
  );
}

describe('LoginCallback', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal('fetch', vi.fn());
    vi.stubGlobal('alert', vi.fn());
  });

  it('ローディング表示がされる', () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({}),
      headers: new Headers(),
    } as Response);

    renderWithRoute('?code=test-code');

    expect(screen.getByText('ログイン中...')).toBeInTheDocument();
  });

  it('codeがない場合はログインページへリダイレクトする', async () => {
    renderWithRoute('');

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login');
    });
  });

  it('errorパラメータがある場合はアラート表示しログインページへリダイレクトする', async () => {
    renderWithRoute('?error=access_denied');

    await waitFor(() => {
      expect(alert).toHaveBeenCalledWith('認証エラーが発生しました。access_denied');
      expect(mockNavigate).toHaveBeenCalledWith('/login');
    });
  });

  it('認証成功時にホームページへリダイレクトする', async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ user: { id: 1, name: 'テスト' } }),
      headers: new Headers(),
    } as Response);

    renderWithRoute('?code=valid-code');

    await waitFor(() => {
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/auth/cognito/callback'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ code: 'valid-code' }),
          credentials: 'include',
        }),
      );
      expect(mockNavigate).toHaveBeenCalledWith('/');
    });
  });

  it('認証失敗時にアラート表示しログインページへリダイレクトする', async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: false,
      status: 401,
      headers: new Headers(),
    } as Response);

    renderWithRoute('?code=invalid-code');

    await waitFor(() => {
      expect(alert).toHaveBeenCalledWith('認証に失敗しました。');
      expect(mockNavigate).toHaveBeenCalledWith('/login');
    });
  });
});
