import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import AuthInitializer from '../AuthInitializer';
import authReducer from '@/entities/user/model/authSlice';
import authRepository from '@/entities/user/api/authRepository';
import { createMockStorage } from '@/test/mockStorage';

vi.mock('@/entities/user/api/authRepository');

/** subscribe を「1回だけ signedIn を通知して、それ以降は何もしない」形にする。 */
function stubSubscribe(signedIn: boolean) {
  return vi.fn((cb: (signedIn: boolean) => void) => {
    cb(signedIn);
    return () => {};
  });
}

function renderWithStore(subscribe: ReturnType<typeof stubSubscribe>, initialLoading = true) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: { isAuthenticated: false, loading: initialLoading },
    },
  });
  return {
    store,
    ...render(
      <Provider store={store}>
        <AuthInitializer subscribe={subscribe}>
          <div>認証済みコンテンツ</div>
        </AuthInitializer>
      </Provider>,
    ),
  };
}

describe('AuthInitializer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('ローディング中はLoadingコンポーネントを表示する', () => {
    // signedIn の通知そのものを起こさない = ローディングのまま。
    const subscribe = vi.fn(() => () => {});

    renderWithStore(subscribe);

    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('サインイン確認時にセッションを確立してchildrenを表示する', async () => {
    vi.mocked(authRepository.login).mockResolvedValue({ message: 'ログインしました。' });

    renderWithStore(stubSubscribe(true));

    await waitFor(() => {
      expect(screen.getByText('認証済みコンテンツ')).toBeInTheDocument();
    });
    expect(authRepository.login).toHaveBeenCalled();
  });

  it('未サインイン確認時もchildrenを表示する（login は呼ばない）', async () => {
    renderWithStore(stubSubscribe(false));

    await waitFor(() => {
      expect(screen.getByText('認証済みコンテンツ')).toBeInTheDocument();
    });
    expect(authRepository.login).not.toHaveBeenCalled();
  });

  it('セッション確立に失敗してもchildrenを表示する', async () => {
    vi.mocked(authRepository.login).mockRejectedValue(new Error('Unauthorized'));

    renderWithStore(stubSubscribe(true));

    await waitFor(() => {
      expect(screen.getByText('認証済みコンテンツ')).toBeInTheDocument();
    });
  });

  it('サインイン確認 + セッション確立成功時にisAuthenticatedがtrueになる', async () => {
    vi.mocked(authRepository.login).mockResolvedValue({ message: 'ログインしました。' });

    const { store } = renderWithStore(stubSubscribe(true));

    await waitFor(() => {
      expect(store.getState().auth.isAuthenticated).toBe(true);
      expect(store.getState().auth.loading).toBe(false);
    });
  });

  it('未サインインならisAuthenticatedがfalseになる', async () => {
    const { store } = renderWithStore(stubSubscribe(false));

    await waitFor(() => {
      expect(store.getState().auth.isAuthenticated).toBe(false);
      expect(store.getState().auth.loading).toBe(false);
    });
  });

  it('セッション確立に失敗したらisAuthenticatedがfalseになる', async () => {
    vi.mocked(authRepository.login).mockRejectedValue(new Error('Unauthorized'));

    const { store } = renderWithStore(stubSubscribe(true));

    await waitFor(() => {
      expect(store.getState().auth.isAuthenticated).toBe(false);
      expect(store.getState().auth.loading).toBe(false);
    });
  });

  it('アンマウント時に購読解除する', () => {
    const unsubscribe = vi.fn();
    const subscribe = vi.fn(() => unsubscribe);

    const { unmount } = renderWithStore(subscribe);
    unmount();

    expect(unsubscribe).toHaveBeenCalled();
  });

  it('subscribe を省略すると本物の subscribeAuthState が使われる（Dex モード・セッション無しでは false 通知）', async () => {
    // テスト環境の既定（src/test/setup.ts）では Dex 設定が揃っており、
    // resolveAuthMode は 'dex' になる。hasDexSession が localStorage を読むため、
    // 他のテスト同様スタブする(FreStyle の既定方針。src/test/mockStorage.ts 参照)。
    // セッションを保存していないので、subscribeAuthState は同期的に false を1回通知する。
    vi.stubGlobal('localStorage', createMockStorage());

    const store = configureStore({
      reducer: { auth: authReducer },
      preloadedState: { auth: { isAuthenticated: false, loading: true } },
    });
    render(
      <Provider store={store}>
        <AuthInitializer>
          <div>認証済みコンテンツ</div>
        </AuthInitializer>
      </Provider>,
    );

    await waitFor(() => {
      expect(screen.getByText('認証済みコンテンツ')).toBeInTheDocument();
    });
    expect(store.getState().auth.isAuthenticated).toBe(false);
  });
});
