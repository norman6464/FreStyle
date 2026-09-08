import { describe, it, expect, beforeEach, vi } from 'vitest';
import { readFirebaseAuthConfig } from '../firebaseConfig';

const complete = {
  VITE_FIREBASE_API_KEY: 'test-api-key',
  VITE_FIREBASE_AUTH_DOMAIN: 'frestyle-507912.firebaseapp.com',
  VITE_FIREBASE_PROJECT_ID: 'frestyle-507912',
};

function stub(overrides: Record<string, string> = {}) {
  for (const [k, v] of Object.entries({ ...complete, ...overrides })) {
    vi.stubEnv(k, v);
  }
}

beforeEach(() => {
  vi.unstubAllEnvs();
});

describe('readFirebaseAuthConfig', () => {
  it('揃っていれば configured を返す', () => {
    stub();
    const config = readFirebaseAuthConfig();
    expect(config).toEqual({
      status: 'configured',
      apiKey: complete.VITE_FIREBASE_API_KEY,
      authDomain: complete.VITE_FIREBASE_AUTH_DOMAIN,
      projectId: complete.VITE_FIREBASE_PROJECT_ID,
    });
  });

  it.each([
    ['VITE_FIREBASE_API_KEY'],
    ['VITE_FIREBASE_AUTH_DOMAIN'],
    ['VITE_FIREBASE_PROJECT_ID'],
  ])('%s が空なら unconfigured で、その名前を挙げる', (key) => {
    stub({ [key]: '' });
    const config = readFirebaseAuthConfig();
    expect(config.status).toBe('unconfigured');
    if (config.status === 'unconfigured') {
      expect(config.missing).toEqual([key]);
    }
  });

  it('複数欠けていれば全部挙げる', () => {
    stub({ VITE_FIREBASE_API_KEY: '', VITE_FIREBASE_AUTH_DOMAIN: '' });
    const config = readFirebaseAuthConfig();
    expect(config.status).toBe('unconfigured');
    if (config.status === 'unconfigured') {
      expect(config.missing).toEqual(['VITE_FIREBASE_API_KEY', 'VITE_FIREBASE_AUTH_DOMAIN']);
    }
  });

  it('何も設定していなければ3つとも挙げる', () => {
    stub({ VITE_FIREBASE_API_KEY: '', VITE_FIREBASE_AUTH_DOMAIN: '', VITE_FIREBASE_PROJECT_ID: '' });
    const config = readFirebaseAuthConfig();
    expect(config.status).toBe('unconfigured');
    if (config.status === 'unconfigured') {
      expect(config.missing).toEqual([
        'VITE_FIREBASE_API_KEY',
        'VITE_FIREBASE_AUTH_DOMAIN',
        'VITE_FIREBASE_PROJECT_ID',
      ]);
    }
  });
});
