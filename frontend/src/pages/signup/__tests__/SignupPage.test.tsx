import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter } from 'react-router-dom';
import { createUserWithEmailAndPassword } from 'firebase/auth';
import SignupPage from '../ui/SignupPage';
import authReducer from '@/entities/user/model/authSlice';
import { ToastProvider } from '@/app/providers/ToastProvider';

vi.mock('firebase/auth', async () => {
  const actual = await vi.importActual<typeof import('firebase/auth')>('firebase/auth');
  return {
    ...actual,
    createUserWithEmailAndPassword: vi.fn(),
    signInWithPopup: vi.fn(),
    sendEmailVerification: vi.fn(),
  };
});

vi.mock('@/shared/lib/auth/firebaseApp', () => ({
  getFirebaseAuth: vi.fn(() => ({ /* フェイクの Auth インスタンス */ })),
}));

function renderSignupPage() {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: { auth: { isAuthenticated: false, loading: false } },
  });
  return render(
    <Provider store={store}>
      <ToastProvider>
        <MemoryRouter>
          <SignupPage />
        </MemoryRouter>
      </ToastProvider>
    </Provider>,
  );
}

const FIREBASE_ENVS = {
  VITE_FIREBASE_API_KEY: 'test-api-key',
  VITE_FIREBASE_AUTH_DOMAIN: 'test.firebaseapp.com',
  VITE_FIREBASE_PROJECT_ID: 'test-project',
} as const;

function stubFirebaseEnv() {
  Object.entries(FIREBASE_ENVS).forEach(([key, value]) => vi.stubEnv(key, value));
}

// テストの既定値は Dex 設定が揃っている（src/test/setup.ts）。
describe('SignupPage（Dex モード・既定）', () => {
  it('見出しとメールで始めるボタンが表示される', () => {
    renderSignupPage();

    expect(screen.getByRole('heading', { name: 'アカウントを作成' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'メールで始める' })).toBeInTheDocument();
  });

  it('Google での開始導線がある', () => {
    renderSignupPage();
    expect(screen.getByRole('button', { name: /Google/ })).toBeInTheDocument();
  });

  it('ログインへの導線がある', () => {
    renderSignupPage();
    const loginLink = screen.getByRole('link', { name: 'ログイン' });
    expect(loginLink).toHaveAttribute('href', '/login');
  });
});

describe('SignupPage（Firebase モード）', () => {
  beforeEach(() => {
    stubFirebaseEnv();
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it('メールとパスワードの入力欄を置く', () => {
    renderSignupPage();

    expect(screen.getByRole('form', { name: 'アカウント作成フォーム' })).toBeInTheDocument();
    expect(screen.getByLabelText('メールアドレス')).toBeInTheDocument();
    expect(screen.getByLabelText('パスワード')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'アカウントを作成' })).toBeInTheDocument();
  });

  it('Google での開始導線がある', () => {
    renderSignupPage();
    expect(screen.getByRole('button', { name: /Google/ })).toBeInTheDocument();
  });

  it('メールとパスワードを入力して送信すると createUserWithEmailAndPassword が呼ばれる', () => {
    vi.mocked(createUserWithEmailAndPassword).mockImplementation(() => new Promise(() => {}));

    renderSignupPage();

    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'new@example.com' } });
    fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: 'アカウントを作成' }));

    expect(createUserWithEmailAndPassword).toHaveBeenCalledWith(expect.anything(), 'new@example.com', 'password123');
  });
});

describe('SignupPage（認可の設定が欠けているとき）', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_OIDC_AUTHORIZE_URI', '');
    vi.stubEnv('VITE_OIDC_TOKEN_URI', '');
    vi.stubEnv('VITE_OIDC_CLIENT_ID', '');
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it('フォームもボタンも出さず、押せない理由を画面に出す', () => {
    renderSignupPage();
    expect(screen.queryByLabelText('メールアドレス')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'メールで始める' })).not.toBeInTheDocument();
    expect(screen.getByRole('status')).toHaveTextContent('現在ログインを受け付けていません');
  });
});
