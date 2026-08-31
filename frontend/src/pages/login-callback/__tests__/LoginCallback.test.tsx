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
    // 既定は「認可コード交換に成功し、/auth/me も引ける」状態。
    vi.mocked(authRepository.probeCurrentUser).mockResolvedValue({
      id: 1,
      isAdmin: true,
      role: 'super_admin',
    });
  });

  it('ローディング表示がされる', () => {
    vi.mocked(authRepository.callback).mockResolvedValue({});

    renderWithRoute('?code=test-code');

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

    renderWithRoute('?code=valid-code');

    await waitFor(() => {
      // 第 2 引数 invitationToken は sessionStorage に何も無いとき null。
      expect(authRepository.callback).toHaveBeenCalledWith('valid-code', null);
      expect(mockNavigate).toHaveBeenCalledWith('/');
    });
  });

  it('認証失敗時にトースト付きでログインページへリダイレクトする', async () => {
    vi.mocked(authRepository.callback).mockRejectedValue(new Error('認証失敗'));

    renderWithRoute('?code=invalid-code');

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login', { state: { toast: '認証に失敗しました' } });
    });
  });

  // 遷移前にロールを確定させないと、ダッシュボードが「管理者ではない」と誤判定して
  // 学習者向け画面を一瞬描画してしまう（FRESTYLE-233）。
  describe('ロールの確定', () => {
    it('遷移前に /auth/me を引いてロールを反映する', async () => {
      vi.mocked(authRepository.callback).mockResolvedValue({});

      const { store } = renderWithRoute('?code=valid-code');

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/');
      });
      expect(authRepository.probeCurrentUser).toHaveBeenCalled();
      expect(store.getState().auth.role).toBe('super_admin');
      expect(store.getState().auth.isAdmin).toBe(true);
    });

    it('ロールが反映されてから遷移する（順序）', async () => {
      vi.mocked(authRepository.callback).mockResolvedValue({});
      let roleAtNavigate: string | null = 'まだ呼ばれていない';
      const { store } = renderWithRoute('?code=valid-code');
      mockNavigate.mockImplementation(() => {
        roleAtNavigate = store.getState().auth.role;
      });

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/');
      });
      expect(roleAtNavigate).toBe('super_admin');
    });

    it('trainee のロールもそのまま反映される', async () => {
      vi.mocked(authRepository.callback).mockResolvedValue({});
      vi.mocked(authRepository.probeCurrentUser).mockResolvedValue({
        id: 2,
        isAdmin: false,
        role: 'trainee',
      });

      const { store } = renderWithRoute('?code=valid-code');

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/');
      });
      expect(store.getState().auth.role).toBe('trainee');
    });

    // 認可コードの交換は成功しているので認証は成立している。ここでログイン画面へ
    // 戻すと「ログインできたのに戻される」挙動になるため、フル読み込みで復帰する。
    it('ロールの取得だけ失敗したときはログイン画面へ戻さずフル読み込みする', async () => {
      // 使用済みの認可コードを含む URL を履歴に残さないため replace で遷移する。
      const replace = vi.fn();
      vi.stubGlobal('location', { ...window.location, replace });
      vi.mocked(authRepository.callback).mockResolvedValue({});
      vi.mocked(authRepository.probeCurrentUser).mockRejectedValue(new Error('一時的な通信エラー'));

      renderWithRoute('?code=valid-code');

      await waitFor(() => {
        expect(replace).toHaveBeenCalledWith('/');
      });
      expect(mockNavigate).not.toHaveBeenCalledWith('/login', expect.anything());
    });
  });
});
