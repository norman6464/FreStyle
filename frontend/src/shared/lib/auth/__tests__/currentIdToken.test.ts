import { describe, it, expect, vi, beforeEach } from 'vitest';
import { onIdTokenChanged, signOut } from 'firebase/auth';
import {
  resolveAuthMode,
  getCurrentIdToken,
  subscribeAuthState,
  signOutCurrentProvider,
} from '../currentIdToken';
import { readFirebaseAuthConfig } from '../firebaseConfig';
import { readAuthConfig } from '../authConfig';
import { getFirebaseAuth } from '../firebaseApp';
import { getValidDexIdToken, hasDexSession, clearDexSession } from '../dexSession';

vi.mock('firebase/auth', async () => {
  const actual = await vi.importActual<typeof import('firebase/auth')>('firebase/auth');
  return { ...actual, onIdTokenChanged: vi.fn(), signOut: vi.fn() };
});
vi.mock('../firebaseConfig');
vi.mock('../authConfig');
vi.mock('../firebaseApp');
vi.mock('../dexSession');

const firebaseConfigured = { status: 'configured' as const, apiKey: 'k', authDomain: 'd', projectId: 'p' };
const firebaseUnconfigured = { status: 'unconfigured' as const, missing: ['VITE_FIREBASE_API_KEY'] };
const dexConfigured = {
  status: 'configured' as const,
  authorizeUri: 'http://localhost:8081/authorize',
  tokenUri: 'http://localhost:8081/token',
  clientId: 'c',
  redirectUri: 'http://localhost:5173/login/callback',
  scope: 'openid',
};
const dexUnconfigured = { status: 'unconfigured' as const, missing: ['VITE_OIDC_AUTHORIZE_URI'] };

beforeEach(() => {
  vi.clearAllMocks();
});

describe('resolveAuthMode', () => {
  it('Firebase が設定されていれば firebase', () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseConfigured);
    vi.mocked(readAuthConfig).mockReturnValue(dexConfigured);
    expect(resolveAuthMode()).toBe('firebase');
  });

  it('Firebase が無く Dex があれば dex', () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseUnconfigured);
    vi.mocked(readAuthConfig).mockReturnValue(dexConfigured);
    expect(resolveAuthMode()).toBe('dex');
  });

  it('どちらも無ければ unconfigured', () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseUnconfigured);
    vi.mocked(readAuthConfig).mockReturnValue(dexUnconfigured);
    expect(resolveAuthMode()).toBe('unconfigured');
  });
});

describe('getCurrentIdToken', () => {
  it('firebase モード・currentUser が居ればその getIdToken を呼ぶ', async () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseConfigured);
    const getIdToken = vi.fn().mockResolvedValue('firebase-token');
    vi.mocked(getFirebaseAuth).mockReturnValue({ currentUser: { getIdToken } } as never);

    const token = await getCurrentIdToken();

    expect(token).toBe('firebase-token');
    expect(getIdToken).toHaveBeenCalledWith(false);
  });

  it('firebase モード・forceRefresh:true を getIdToken にそのまま渡す', async () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseConfigured);
    const getIdToken = vi.fn().mockResolvedValue('refreshed-token');
    vi.mocked(getFirebaseAuth).mockReturnValue({ currentUser: { getIdToken } } as never);

    await getCurrentIdToken(true);

    expect(getIdToken).toHaveBeenCalledWith(true);
  });

  it('firebase モード・currentUser が居なければ null', async () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseConfigured);
    vi.mocked(getFirebaseAuth).mockReturnValue({ currentUser: null } as never);

    expect(await getCurrentIdToken()).toBeNull();
  });

  it('dex モードなら getValidDexIdToken へ委譲する', async () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseUnconfigured);
    vi.mocked(readAuthConfig).mockReturnValue(dexConfigured);
    vi.mocked(getValidDexIdToken).mockResolvedValue('dex-token');

    const token = await getCurrentIdToken(true);

    expect(token).toBe('dex-token');
    expect(getValidDexIdToken).toHaveBeenCalledWith(dexConfigured, true);
  });

  it('unconfigured なら null', async () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseUnconfigured);
    vi.mocked(readAuthConfig).mockReturnValue(dexUnconfigured);

    expect(await getCurrentIdToken()).toBeNull();
  });
});

describe('subscribeAuthState', () => {
  it('firebase モードは onIdTokenChanged を購読し、ユーザーの有無を渡す', () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseConfigured);
    const fakeAuth = {};
    vi.mocked(getFirebaseAuth).mockReturnValue(fakeAuth as never);
    const unsubscribe = vi.fn();
    vi.mocked(onIdTokenChanged).mockImplementation((_auth, cb) => {
      (cb as (u: unknown) => void)({ uid: 'u1' });
      return unsubscribe;
    });

    const callback = vi.fn();
    const result = subscribeAuthState(callback);

    expect(callback).toHaveBeenCalledWith(true);
    expect(result).toBe(unsubscribe);
  });

  it('firebase モードでユーザーが null なら false を渡す', () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseConfigured);
    vi.mocked(getFirebaseAuth).mockReturnValue({} as never);
    vi.mocked(onIdTokenChanged).mockImplementation((_auth, cb) => {
      (cb as (u: unknown) => void)(null);
      return vi.fn();
    });

    const callback = vi.fn();
    subscribeAuthState(callback);

    expect(callback).toHaveBeenCalledWith(false);
  });

  it('dex モードは hasDexSession の結果を1回だけ渡す', () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseUnconfigured);
    vi.mocked(readAuthConfig).mockReturnValue(dexConfigured);
    vi.mocked(hasDexSession).mockReturnValue(true);

    const callback = vi.fn();
    subscribeAuthState(callback);

    expect(callback).toHaveBeenCalledTimes(1);
    expect(callback).toHaveBeenCalledWith(true);
  });

  it('unconfigured なら false を1回だけ渡す', () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseUnconfigured);
    vi.mocked(readAuthConfig).mockReturnValue(dexUnconfigured);

    const callback = vi.fn();
    subscribeAuthState(callback);

    expect(callback).toHaveBeenCalledTimes(1);
    expect(callback).toHaveBeenCalledWith(false);
  });
});

describe('signOutCurrentProvider', () => {
  it('firebase モードは firebase/auth の signOut を呼ぶ', async () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseConfigured);
    const fakeAuth = {};
    vi.mocked(getFirebaseAuth).mockReturnValue(fakeAuth as never);

    await signOutCurrentProvider();

    expect(signOut).toHaveBeenCalledWith(fakeAuth);
    expect(clearDexSession).not.toHaveBeenCalled();
  });

  it('dex モードは保存したセッションを消す（発行者側のセッションは残る）', async () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseUnconfigured);
    vi.mocked(readAuthConfig).mockReturnValue(dexConfigured);

    await signOutCurrentProvider();

    expect(clearDexSession).toHaveBeenCalled();
    expect(signOut).not.toHaveBeenCalled();
  });

  it('unconfigured なら何もしない（例外を投げない）', async () => {
    vi.mocked(readFirebaseAuthConfig).mockReturnValue(firebaseUnconfigured);
    vi.mocked(readAuthConfig).mockReturnValue(dexUnconfigured);

    await expect(signOutCurrentProvider()).resolves.toBeUndefined();
    expect(signOut).not.toHaveBeenCalled();
    expect(clearDexSession).not.toHaveBeenCalled();
  });
});
