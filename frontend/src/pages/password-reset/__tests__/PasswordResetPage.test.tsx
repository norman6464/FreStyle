import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { sendPasswordResetEmail } from 'firebase/auth';
import PasswordResetPage from '../ui/PasswordResetPage';

vi.mock('firebase/auth', async () => {
  const actual = await vi.importActual<typeof import('firebase/auth')>('firebase/auth');
  return {
    ...actual,
    sendPasswordResetEmail: vi.fn(),
  };
});

vi.mock('@/shared/lib/auth/firebaseApp', () => ({
  getFirebaseAuth: vi.fn(() => ({ /* フェイクの Auth インスタンス */ })),
}));

function renderPage() {
  return render(
    <MemoryRouter>
      <PasswordResetPage />
    </MemoryRouter>,
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

describe('PasswordResetPage（Dex モード・既定）', () => {
  it('Dex はパスワード再設定に対応していない旨を案内する', () => {
    renderPage();
    expect(screen.getByRole('status')).toHaveTextContent('対応していません');
    expect(screen.queryByLabelText('メールアドレス')).not.toBeInTheDocument();
  });

  it('ログインへ戻る導線がある', () => {
    renderPage();
    expect(screen.getByRole('link', { name: 'ログインへ戻る' })).toHaveAttribute('href', '/login');
  });
});

describe('PasswordResetPage（Firebase モード）', () => {
  beforeEach(() => {
    stubFirebaseEnv();
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it('メールアドレスの入力欄と送信ボタンを置く', () => {
    renderPage();
    expect(screen.getByRole('form', { name: 'パスワード再設定フォーム' })).toBeInTheDocument();
    expect(screen.getByLabelText('メールアドレス')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '再設定メールを送信' })).toBeInTheDocument();
  });

  it('送信すると sendPasswordResetEmail が呼ばれ、確認の案内に切り替わる', async () => {
    vi.mocked(sendPasswordResetEmail).mockResolvedValue(undefined);

    renderPage();
    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'user@example.com' } });
    fireEvent.click(screen.getByRole('button', { name: '再設定メールを送信' }));

    await waitFor(() => {
      expect(screen.getByRole('status')).toHaveTextContent('該当するアカウントが存在する場合');
    });
    expect(sendPasswordResetEmail).toHaveBeenCalledWith(expect.anything(), 'user@example.com');
    expect(screen.queryByLabelText('メールアドレス')).not.toBeInTheDocument();
  });

  // **この画面の要**。存在しないメールアドレスでも同じ案内を出す。表示を変えると、
  // それ自体が「登録されているかどうか」を外部から探る手がかりになる。
  it('該当アカウントが無くて失敗しても、同じ確認の案内を出す(表示だけでは成否が分からない)', async () => {
    vi.mocked(sendPasswordResetEmail).mockRejectedValue(new Error('auth/user-not-found'));

    renderPage();
    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'nobody@example.com' } });
    fireEvent.click(screen.getByRole('button', { name: '再設定メールを送信' }));

    await waitFor(() => {
      expect(screen.getByRole('status')).toHaveTextContent('該当するアカウントが存在する場合');
    });
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });
});

describe('PasswordResetPage（認可の設定が欠けているとき）', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_OIDC_AUTHORIZE_URI', '');
    vi.stubEnv('VITE_OIDC_TOKEN_URI', '');
    vi.stubEnv('VITE_OIDC_CLIENT_ID', '');
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it('押せない理由を画面に出す', () => {
    renderPage();
    expect(screen.getByRole('status')).toHaveTextContent('現在ログインを受け付けていません');
  });
});
