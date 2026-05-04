import { describe, it, expect } from 'vitest';
import authReducer, { setAuthData, setAuthenticated, clearAuth, finishLoading } from '../authSlice';

describe('authSlice', () => {
  const initialState = { isAuthenticated: false, loading: true, isAdmin: false };

  it('初期状態はisAuthenticated=false, loading=true, isAdmin=false', () => {
    const state = authReducer(undefined, { type: 'unknown' });
    expect(state).toEqual(initialState);
  });

  it('setAuthDataでisAuthenticated=true, loading=falseになる', () => {
    const state = authReducer(initialState, setAuthData());
    expect(state.isAuthenticated).toBe(true);
    expect(state.loading).toBe(false);
    expect(state.isAdmin).toBe(false);
  });

  it('setAuthDataでpayload.isAdmin=trueを渡すとisAdminがtrueになる', () => {
    const state = authReducer(initialState, setAuthData({ isAdmin: true }));
    expect(state.isAdmin).toBe(true);
  });

  it('setAuthDataでpayload未指定の場合は既存のisAdminを保持する', () => {
    const adminState = { isAuthenticated: true, loading: false, isAdmin: true };
    const state = authReducer(adminState, setAuthData());
    expect(state.isAdmin).toBe(true);
  });

  it('setAuthenticatedでisAuthenticated=true, loading=falseになる', () => {
    const state = authReducer(initialState, setAuthenticated());
    expect(state.isAuthenticated).toBe(true);
    expect(state.loading).toBe(false);
    expect(state.isAdmin).toBe(false);
  });

  it('setAuthenticatedでpayload未指定の場合は既存のisAdminを保持する', () => {
    const adminState = { isAuthenticated: true, loading: false, isAdmin: true };
    const state = authReducer(adminState, setAuthenticated());
    expect(state.isAdmin).toBe(true);
  });

  it('clearAuthでisAuthenticated=false, loading=false, isAdmin=falseになる', () => {
    const authenticatedState = { isAuthenticated: true, loading: false, isAdmin: true };
    const state = authReducer(authenticatedState, clearAuth());
    expect(state.isAuthenticated).toBe(false);
    expect(state.loading).toBe(false);
    expect(state.isAdmin).toBe(false);
  });

  it('finishLoadingでloading=falseになりisAuthenticatedは変わらない', () => {
    const state = authReducer(initialState, finishLoading());
    expect(state.isAuthenticated).toBe(false);
    expect(state.loading).toBe(false);
  });
});
