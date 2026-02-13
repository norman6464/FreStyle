import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter } from 'react-router-dom';
import LoginPage from '../LoginPage';
import authReducer from '../../store/authSlice';

const mockLogin = vi.fn();
const mockNavigate = vi.fn();

vi.mock('../../hooks/useAuth', () => ({
  useAuth: () => ({
    login: mockLogin,
    loading: false,
    error: null,
    isAuthenticated: false,
  }),
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

function renderLoginPage() {
  const store = configureStore({ reducer: { auth: authReducer } });
  return render(
    <Provider store={store}>
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    </Provider>
  );
}

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('ログインフォームが表示される', () => {
    renderLoginPage();

    expect(screen.getByRole('heading', { name: 'ログイン' })).toBeInTheDocument();
    expect(screen.getByLabelText('メールアドレス')).toBeInTheDocument();
    expect(screen.getByLabelText('パスワード')).toBeInTheDocument();
  });

  it('リンクテキストが表示される', () => {
    renderLoginPage();

    expect(screen.getByText('パスワードをお忘れですか？')).toBeInTheDocument();
    expect(screen.getByText('アカウント作成')).toBeInTheDocument();
  });

  it('ログイン成功時にホームに遷移する', async () => {
    mockLogin.mockResolvedValue(true);
    renderLoginPage();

    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'test@example.com', name: 'email' } });
    fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'password123', name: 'password' } });

    const loginButtons = screen.getAllByText('ログイン');
    const submitButton = loginButtons.find(el => el.tagName === 'BUTTON' && el.getAttribute('type') === 'submit');
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({ email: 'test@example.com', password: 'password123' });
      expect(mockNavigate).toHaveBeenCalledWith('/');
    });
  });

  it('ログイン失敗時にエラーメッセージが表示される', async () => {
    mockLogin.mockResolvedValue(false);
    renderLoginPage();

    const loginButtons = screen.getAllByText('ログイン');
    const submitButton = loginButtons.find(el => el.tagName === 'BUTTON' && el.getAttribute('type') === 'submit');
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(screen.getByText(/ログインに失敗しました/)).toBeInTheDocument();
    });
  });
});
