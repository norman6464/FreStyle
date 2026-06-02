import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { AuthState } from '../types';

const initialState: AuthState = {
  isAuthenticated: false,
  loading: true,
  isAdmin: false,
  role: null,
};

type AuthPayload = {
  isAdmin?: boolean;
  role?: string | null;
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
    },

    setAuthenticated(state, action: PayloadAction<AuthPayload | undefined>) {
      state.isAuthenticated = true;
      state.loading = false;
      if (action.payload?.isAdmin !== undefined) {
        state.isAdmin = action.payload.isAdmin;
      }
      if (action.payload?.role !== undefined) {
        state.role = action.payload.role;
      }
    },

    clearAuth(state) {
      state.isAuthenticated = false;
      state.loading = false;
      state.isAdmin = false;
      state.role = null;
    },

    finishLoading(state) {
      state.loading = false;
    },
  },
});

export const { setAuthData, setAuthenticated, clearAuth, finishLoading } =
  authSlice.actions;
export default authSlice.reducer;
