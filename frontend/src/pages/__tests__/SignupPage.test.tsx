import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter } from 'react-router-dom';
import SignupPage from '../SignupPage';
import authReducer from '../../store/authSlice';

const mockSignup = vi.fn();

vi.mock('../../hooks/useAuth', () => ({
  useAuth: () => ({
    signup: mockSignup,
    loading: false,
    error: null,
  }),
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => vi.fn(),
  };
});

function renderSignupPage() {
  const store = configureStore({ reducer: { auth: authReducer } });
  return render(
    <Provider store={store}>
      <MemoryRouter>
        <SignupPage />
      </MemoryRouter>
    </Provider>
  );
}

describe('SignupPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('サインアップフォームが表示される', () => {
    renderSignupPage();

    expect(screen.getByRole('heading', { name: 'アカウント作成' })).toBeInTheDocument();
    expect(screen.getByLabelText('ニックネーム')).toBeInTheDocument();
    expect(screen.getByLabelText('メールアドレス')).toBeInTheDocument();
    expect(screen.getByLabelText('パスワード')).toBeInTheDocument();
  });

  it('ログインリンクが表示される', () => {
    renderSignupPage();

    expect(screen.getByText('ログインする')).toBeInTheDocument();
  });

  it('サインアップ失敗時にエラーメッセージが表示される', async () => {
    mockSignup.mockResolvedValue(false);
    renderSignupPage();

    fireEvent.change(screen.getByLabelText('ニックネーム'), { target: { value: 'テスト' } });
    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'test@example.com' } });
    fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'pass123' } });

    const submitButtons = screen.getAllByText('アカウント作成');
    const submitButton = submitButtons.find(el => el.tagName === 'BUTTON');
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(screen.getByText(/登録に失敗しました/)).toBeInTheDocument();
    });
  });

  it('サインアップ成功時に成功メッセージが表示される', async () => {
    mockSignup.mockResolvedValue(true);
    renderSignupPage();

    fireEvent.change(screen.getByLabelText('ニックネーム'), { target: { value: 'テスト' } });
    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'test@example.com' } });
    fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'pass123' } });

    const submitButtons = screen.getAllByText('アカウント作成');
    const submitButton = submitButtons.find(el => el.tagName === 'BUTTON');
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(screen.getByText(/サインアップに成功しました/)).toBeInTheDocument();
    });
  });

  it('未入力で送信するとバリデーションエラーが表示される', async () => {
    renderSignupPage();

    const submitButtons = screen.getAllByText('アカウント作成');
    const submitButton = submitButtons.find(el => el.tagName === 'BUTTON');
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(screen.getByText('すべてのフィールドを入力してください。')).toBeInTheDocument();
    });
    expect(mockSignup).not.toHaveBeenCalled();
  });
});
