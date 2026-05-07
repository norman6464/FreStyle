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
    role: null,
  };

  it('初期状態はisAuthenticated=false, loading=true, isAdmin=false, onboarded=false, role=null', () => {
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
    const adminState = {
      isAuthenticated: true,
      loading: false,
      isAdmin: true,
      onboarded: true,
      role: 'super_admin',
    };
    const state = authReducer(adminState, setAuthData());
    expect(state.isAdmin).toBe(true);
    expect(state.onboarded).toBe(true);
    expect(state.role).toBe('super_admin');
  });

  it('setAuthDataでpayload.onboarded=trueを渡すとonboardedがtrueになる', () => {
    const state = authReducer(initialState, setAuthData({ onboarded: true }));
    expect(state.onboarded).toBe(true);
  });

  it('setAuthDataでpayload.roleを渡すとroleが反映される', () => {
    const state = authReducer(initialState, setAuthData({ role: 'super_admin' }));
    expect(state.role).toBe('super_admin');
  });

  it('setAuthenticatedでisAuthenticated=true, loading=falseになる', () => {
    const state = authReducer(initialState, setAuthenticated());
    expect(state.isAuthenticated).toBe(true);
    expect(state.loading).toBe(false);
    expect(state.isAdmin).toBe(false);
  });

  it('setAuthenticatedでpayload未指定の場合は既存のisAdminを保持する', () => {
    const adminState = {
      isAuthenticated: true,
      loading: false,
      isAdmin: true,
      onboarded: true,
      role: 'company_admin',
    };
    const state = authReducer(adminState, setAuthenticated());
    expect(state.isAdmin).toBe(true);
    expect(state.role).toBe('company_admin');
  });

  it('markOnboardedでonboardedがtrueになる', () => {
    const state = authReducer(initialState, markOnboarded());
    expect(state.onboarded).toBe(true);
  });

  it('clearAuthで全フィールドが初期値にリセットされる', () => {
    const authenticatedState = {
      isAuthenticated: true,
      loading: false,
      isAdmin: true,
      onboarded: true,
      role: 'super_admin',
    };
    const state = authReducer(authenticatedState, clearAuth());
    expect(state.isAuthenticated).toBe(false);
    expect(state.loading).toBe(false);
    expect(state.isAdmin).toBe(false);
    expect(state.onboarded).toBe(false);
    expect(state.role).toBeNull();
  });

  it('finishLoadingでloading=falseになりisAuthenticatedは変わらない', () => {
    const state = authReducer(initialState, finishLoading());
    expect(state.isAuthenticated).toBe(false);
    expect(state.loading).toBe(false);
  });
});
