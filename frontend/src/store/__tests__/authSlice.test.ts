import { describe, it, expect } from 'vitest';
import authReducer, {
  setAuthData,
  setAuthenticated,
  clearAuth,
  finishLoading,
  setAiChatEnabledForTrainees,
} from '../authSlice';

describe('authSlice', () => {
  const initialState = {
    isAuthenticated: false,
    loading: true,
    isAdmin: false,
    role: null,
    aiChatEnabledForTrainees: true,
  };

  it('初期状態はisAuthenticated=false, loading=true, isAdmin=false, role=null, AI 有効=true', () => {
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
      role: 'super_admin',
    };
    const state = authReducer(adminState, setAuthData());
    expect(state.isAdmin).toBe(true);
    expect(state.role).toBe('super_admin');
  });

  it('setAuthDataでpayload.roleを渡すとroleが反映される', () => {
    const state = authReducer(initialState, setAuthData({ role: 'super_admin' }));
    expect(state.role).toBe('super_admin');
  });

  it('setAuthDataでaiChatEnabledForTrainees=falseを渡すと反映される', () => {
    const state = authReducer(initialState, setAuthData({ aiChatEnabledForTrainees: false }));
    expect(state.aiChatEnabledForTrainees).toBe(false);
  });

  it('setAuthDataでaiChatEnabledForTrainees未指定なら既存値を保持する', () => {
    const disabled = { ...initialState, aiChatEnabledForTrainees: false };
    const state = authReducer(disabled, setAuthData({ isAdmin: true }));
    expect(state.aiChatEnabledForTrainees).toBe(false);
  });

  it('setAiChatEnabledForTraineesでフラグだけ更新できる', () => {
    const state = authReducer(initialState, setAiChatEnabledForTrainees(false));
    expect(state.aiChatEnabledForTrainees).toBe(false);
    const back = authReducer(state, setAiChatEnabledForTrainees(true));
    expect(back.aiChatEnabledForTrainees).toBe(true);
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
      role: 'company_admin',
    };
    const state = authReducer(adminState, setAuthenticated());
    expect(state.isAdmin).toBe(true);
    expect(state.role).toBe('company_admin');
  });

  it('clearAuthで全フィールドが初期値にリセットされる', () => {
    const authenticatedState = {
      isAuthenticated: true,
      loading: false,
      isAdmin: true,
      role: 'super_admin',
    };
    const state = authReducer(authenticatedState, clearAuth());
    expect(state.isAuthenticated).toBe(false);
    expect(state.loading).toBe(false);
    expect(state.isAdmin).toBe(false);
    expect(state.role).toBeNull();
    expect(state.aiChatEnabledForTrainees).toBe(true);
  });

  it('finishLoadingでloading=falseになりisAuthenticatedは変わらない', () => {
    const state = authReducer(initialState, finishLoading());
    expect(state.isAuthenticated).toBe(false);
    expect(state.loading).toBe(false);
  });
});
