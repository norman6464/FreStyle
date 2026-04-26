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
      state.isAdmin = action.payload?.isAdmin ?? false;
    },

    setAuthenticated(state, action: PayloadAction<{ isAdmin?: boolean } | undefined>) {
      state.isAuthenticated = true;
      state.loading = false;
      state.isAdmin = action.payload?.isAdmin ?? false;
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
