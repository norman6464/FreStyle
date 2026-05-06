import { describe, it, expect } from 'vitest';
import authReducer, {
  setAuthData,
  setAuthenticated,
  markOnboarded,
  clearAuth,
  finishLoading,
} from '../authSlice';

describe('authSlice', () => {
  const initialState = {
    isAuthenticated: false,
    loading: true,
    isAdmin: false,
    onboarded: false,
  };

  it('初期状態はisAuthenticated=false, loading=true, isAdmin=false, onboarded=false', () => {
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
    const adminState = { isAuthenticated: true, loading: false, isAdmin: true, onboarded: true };
    const state = authReducer(adminState, setAuthData());
    expect(state.isAdmin).toBe(true);
    expect(state.onboarded).toBe(true);
  });

  it('setAuthDataでpayload.onboarded=trueを渡すとonboardedがtrueになる', () => {
    const state = authReducer(initialState, setAuthData({ onboarded: true }));
    expect(state.onboarded).toBe(true);
  });

  it('setAuthenticatedでisAuthenticated=true, loading=falseになる', () => {
    const state = authReducer(initialState, setAuthenticated());
    expect(state.isAuthenticated).toBe(true);
    expect(state.loading).toBe(false);
    expect(state.isAdmin).toBe(false);
  });

  it('setAuthenticatedでpayload未指定の場合は既存のisAdminを保持する', () => {
    const adminState = { isAuthenticated: true, loading: false, isAdmin: true, onboarded: true };
    const state = authReducer(adminState, setAuthenticated());
    expect(state.isAdmin).toBe(true);
  });

  it('markOnboardedでonboardedがtrueになる', () => {
    const state = authReducer(initialState, markOnboarded());
    expect(state.onboarded).toBe(true);
  });

  it('clearAuthでisAuthenticated=false, loading=false, isAdmin=false, onboarded=falseになる', () => {
    const authenticatedState = {
      isAuthenticated: true,
      loading: false,
      isAdmin: true,
      onboarded: true,
    };
    const state = authReducer(authenticatedState, clearAuth());
    expect(state.isAuthenticated).toBe(false);
    expect(state.loading).toBe(false);
    expect(state.isAdmin).toBe(false);
    expect(state.onboarded).toBe(false);
  });

  it('finishLoadingでloading=falseになりisAuthenticatedは変わらない', () => {
    const state = authReducer(initialState, finishLoading());
    expect(state.isAuthenticated).toBe(false);
    expect(state.loading).toBe(false);
  });
});
