import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { AuthState } from '@/types';

const initialState: AuthState = {
  isAuthenticated: false,
  loading: true,
  isAdmin: false,
  role: null,
  // 既定 true: /auth/me 未取得や会社未所属でも AI を出す(後方互換)。trainee の会社設定で false になる。
  aiChatEnabledForTrainees: true,
};

type AuthPayload = {
  isAdmin?: boolean;
  role?: string | null;
  aiChatEnabledForTrainees?: boolean;
};

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setAuthData(state, action: PayloadAction<AuthPayload | undefined>) {
      state.isAuthenticated = true;
      state.loading = false;
      // payload の各フィールドが undefined のときは現在の state を維持する
      // (caller が値を明示しない再認証で誤って権限を失わせないため)
      if (action.payload?.isAdmin !== undefined) {
        state.isAdmin = action.payload.isAdmin;
      }
      if (action.payload?.role !== undefined) {
        state.role = action.payload.role;
      }
      if (action.payload?.aiChatEnabledForTrainees !== undefined) {
        state.aiChatEnabledForTrainees = action.payload.aiChatEnabledForTrainees;
      }
    },

    clearAuth(state) {
      state.isAuthenticated = false;
      state.loading = false;
      state.isAdmin = false;
      state.role = null;
      state.aiChatEnabledForTrainees = true;
    },

    // 設定画面で company_admin がトグルした結果を即座に反映する(自分が trainee でなくても整合のため保持)。
    setAiChatEnabledForTrainees(state, action: PayloadAction<boolean>) {
      state.aiChatEnabledForTrainees = action.payload;
    },

    finishLoading(state) {
      state.loading = false;
    },
  },
});

export const {
  setAuthData,
  clearAuth,
  finishLoading,
  setAiChatEnabledForTrainees,
} = authSlice.actions;
export default authSlice.reducer;
