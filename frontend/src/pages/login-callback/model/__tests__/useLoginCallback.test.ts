import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useLoginCallback } from '../useLoginCallback';

const mockNavigate = vi.fn();
const mockDispatch = vi.fn();

vi.mock('react-router-dom', () => ({
  useSearchParams: () => [new URLSearchParams(mockSearchParams)],
  useNavigate: () => mockNavigate,
}));

vi.mock('@/shared/lib/store', () => ({
  useAppDispatch: () => mockDispatch,
}));

vi.mock('@/entities/user/api/authRepository', () => ({
  default: {
    callback: vi.fn(),
  },
}));

vi.mock('@/entities/user/model/authSlice', () => ({
  setAuthData: () => ({ type: 'auth/setAuthData' }),
}));

// 認可を始めたときに置いた値を取り出す側。テストごとに中身を差し替える。
vi.mock('@/features/auth', () => ({
  consumeAuthFlowState: () => mockFlow,
}));

import authRepository from '@/entities/user/api/authRepository';

let mockSearchParams = '';
let mockFlow: { state: string; nonce: string; codeVerifier: string } | null = null;

const FLOW = { state: 'my-state', nonce: 'my-nonce', codeVerifier: 'my-verifier' };

describe('useLoginCallback', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams = '';
    mockFlow = { ...FLOW };
  });

  it('state が一致すれば、検証値と nonce を添えて交換する', async () => {
    mockSearchParams = 'code=test-code&state=my-state';
    vi.mocked(authRepository.callback).mockResolvedValue({} as never);

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(authRepository.callback).toHaveBeenCalledWith({
      code: 'test-code',
      codeVerifier: 'my-verifier',
      nonce: 'my-nonce',
    });
  });

  // **この PR の要のひとつ。**
  // state を確かめないと、攻撃者が自分の認可コードを他人のブラウザに踏ませて、
  // 被害者を攻撃者のアカウントでログインさせられる。
  it('state が一致しなければ交換しない', async () => {
    mockSearchParams = 'code=test-code&state=attacker-state';

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(authRepository.callback).not.toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith('/login', {
      state: { toast: 'ログインの検証に失敗しました。もう一度お試しください。' },
    });
  });

  it('state が返ってこなければ交換しない', async () => {
    mockSearchParams = 'code=test-code';

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(authRepository.callback).not.toHaveBeenCalled();
  });

  // この端末で始めていない認可の戻り（別タブ・別端末で始めた、あるいは仕込まれた URL）。
  it('手元に手続きが残っていなければ交換しない', async () => {
    mockSearchParams = 'code=test-code&state=my-state';
    mockFlow = null;

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(authRepository.callback).not.toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith('/login', {
      state: { toast: 'ログインの手続きが見つかりませんでした。もう一度お試しください。' },
    });
  });

  it('交換に成功したら認証状態を確定させてホームへ遷移する', async () => {
    mockSearchParams = 'code=test-code&state=my-state';
    vi.mocked(authRepository.callback).mockResolvedValue({} as never);

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockDispatch).toHaveBeenCalledWith({ type: 'auth/setAuthData' });
    expect(mockNavigate).toHaveBeenCalledWith('/');
  });

  it('交換に失敗したら案内つきでログイン画面へ戻す', async () => {
    mockSearchParams = 'code=test-code&state=my-state';
    vi.mocked(authRepository.callback).mockRejectedValue(new Error('認証失敗'));

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockNavigate).toHaveBeenCalledWith('/login', { state: { toast: '認証に失敗しました' } });
  });

  it('error が返っていれば交換しない', async () => {
    mockSearchParams = 'error=access_denied&code=test-code&state=my-state';

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockNavigate).toHaveBeenCalledWith('/login', {
      state: { toast: '認証エラーが発生しました' },
    });
    expect(authRepository.callback).not.toHaveBeenCalled();
  });

  it('code も error も無ければログイン画面へ戻す', async () => {
    mockSearchParams = '';

    await act(async () => {
      renderHook(() => useLoginCallback());
    });

    expect(mockNavigate).toHaveBeenCalledWith('/login');
    expect(authRepository.callback).not.toHaveBeenCalled();
    expect(mockDispatch).not.toHaveBeenCalled();
  });
});
