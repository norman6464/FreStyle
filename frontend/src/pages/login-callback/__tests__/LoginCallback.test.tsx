import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter } from 'react-router-dom';
import LoginCallback from '../ui/LoginCallback';
import authReducer from '@/entities/user/model/authSlice';
import authRepository from '@/entities/user/api/authRepository';
import { ToastProvider } from '@/app/providers/ToastProvider';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock('@/entities/user/api/authRepository');

// 認可を始めたときにブラウザが置く値。実装（features/auth/lib/oidcAuthUrl）と同じ鍵を使う。
const FLOW = { state: 'test-state', nonce: 'test-nonce', codeVerifier: 'test-verifier' };

function seedAuthFlow() {
  sessionStorage.setItem('oidc.authFlow', JSON.stringify(FLOW));
}

function renderWithRoute(search: string) {
  const store = configureStore({ reducer: { auth: authReducer } });
  const view = render(
    <Provider store={store}>
      <MemoryRouter initialEntries={[`/login/callback${search}`]}>
        <ToastProvider>
          <LoginCallback />
        </ToastProvider>
      </MemoryRouter>
    </Provider>,
  );
  return { ...view, store };
}

describe('LoginCallback', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal('alert', vi.fn());
    sessionStorage.clear();
    seedAuthFlow();
  });

  it('ローディング表示がされる', () => {
    vi.mocked(authRepository.callback).mockResolvedValue({});

    renderWithRoute('?code=test-code&state=test-state');

    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.getByText('ログイン中...')).toBeInTheDocument();
  });

  it('codeがない場合はログインページへリダイレクトする', async () => {
    renderWithRoute('');

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login');
    });
  });

  it('errorパラメータがある場合はトースト付きでログインページへリダイレクトする', async () => {
    renderWithRoute('?error=access_denied');

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login', { state: { toast: '認証エラーが発生しました' } });
    });
  });

  it('認証成功時にホームページへリダイレクトする', async () => {
    vi.mocked(authRepository.callback).mockResolvedValue({ user: { id: 1, name: 'テスト' } });

    renderWithRoute('?code=valid-code&state=test-state');

    await waitFor(() => {
      // 認可を始めたときに置いた検証値と nonce を添えて交換する。
      expect(authRepository.callback).toHaveBeenCalledWith({
        code: 'valid-code',
        codeVerifier: FLOW.codeVerifier,
        nonce: FLOW.nonce,
      });
      expect(mockNavigate).toHaveBeenCalledWith('/');
    });
  });

  it('認証失敗時にトースト付きでログインページへリダイレクトする', async () => {
    vi.mocked(authRepository.callback).mockRejectedValue(new Error('認証失敗'));

    renderWithRoute('?code=invalid-code&state=test-state');

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login', { state: { toast: '認証に失敗しました' } });
    });
  });

  it('認証成功時に store の isAuthenticated が true になる', async () => {
    vi.mocked(authRepository.callback).mockResolvedValue({});

    const { store } = renderWithRoute('?code=valid-code&state=test-state');

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/');
    });
    expect(store.getState().auth.isAuthenticated).toBe(true);
  });
});
