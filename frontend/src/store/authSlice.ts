import { createSlice } from '@reduxjs/toolkit';
import type { AuthState } from '../types';

const initialState: AuthState = {
  isAuthenticated: false,
  loading: true,
};

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setAuthData(state) {
      state.isAuthenticated = true;
      state.loading = false;
    },

    setAuthenticated(state) {
      state.isAuthenticated = true;
      state.loading = false;
    },

    clearAuth(state) {
      state.isAuthenticated = false;
      state.loading = false;
    },

    finishLoading(state) {
      state.loading = false;
    },
  },
});

export const { setAuthData, setAuthenticated, clearAuth, finishLoading } = authSlice.actions;
export default authSlice.reducer;
