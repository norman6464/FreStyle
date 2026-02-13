import { describe, it, expect } from 'vitest';
import authReducer, { setAuthData, setAuthenticated, clearAuth, finishLoading } from '../authSlice';

describe('authSlice', () => {
  const initialState = { isAuthenticated: false, loading: true };

  it('初期状態はisAuthenticated=false, loading=true', () => {
    const state = authReducer(undefined, { type: 'unknown' });
    expect(state).toEqual(initialState);
  });

  it('setAuthDataでisAuthenticated=true, loading=falseになる', () => {
    const state = authReducer(initialState, setAuthData());
    expect(state.isAuthenticated).toBe(true);
    expect(state.loading).toBe(false);
  });

  it('setAuthenticatedでisAuthenticated=true, loading=falseになる', () => {
    const state = authReducer(initialState, setAuthenticated());
    expect(state.isAuthenticated).toBe(true);
    expect(state.loading).toBe(false);
  });

  it('clearAuthでisAuthenticated=false, loading=falseになる', () => {
    const authenticatedState = { isAuthenticated: true, loading: false };
    const state = authReducer(authenticatedState, clearAuth());
    expect(state.isAuthenticated).toBe(false);
    expect(state.loading).toBe(false);
  });

  it('finishLoadingでloading=falseになりisAuthenticatedは変わらない', () => {
    const state = authReducer(initialState, finishLoading());
    expect(state.isAuthenticated).toBe(false);
    expect(state.loading).toBe(false);
  });
});
