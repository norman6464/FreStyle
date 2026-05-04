import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { AuthState } from '../types';

const initialState: AuthState = {
  isAuthenticated: false,
  loading: true,
  isAdmin: false,
};

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setAuthData(state, action: PayloadAction<{ isAdmin?: boolean } | undefined>) {
      state.isAuthenticated = true;
      state.loading = false;
      // payload.isAdmin が undefined のときは現在の state を維持する
      // (caller が isAdmin を明示しない再認証で誤って権限を失わせないため)
      if (action.payload?.isAdmin !== undefined) {
        state.isAdmin = action.payload.isAdmin;
      }
    },

    setAuthenticated(state, action: PayloadAction<{ isAdmin?: boolean } | undefined>) {
      state.isAuthenticated = true;
      state.loading = false;
      if (action.payload?.isAdmin !== undefined) {
        state.isAdmin = action.payload.isAdmin;
      }
    },

    clearAuth(state) {
      state.isAuthenticated = false;
      state.loading = false;
      state.isAdmin = false;
    },

    finishLoading(state) {
      state.loading = false;
    },
  },
});

export const { setAuthData, setAuthenticated, clearAuth, finishLoading } = authSlice.actions;
export default authSlice.reducer;
