import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { AuthState } from '../types';

const initialState: AuthState = {
  isAuthenticated: false,
  loading: true,
  isAdmin: false,
  onboarded: false,
};

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setAuthData(
      state,
      action: PayloadAction<{ isAdmin?: boolean; onboarded?: boolean } | undefined>
    ) {
      state.isAuthenticated = true;
      state.loading = false;
      // payload.isAdmin / onboarded が undefined のときは現在の state を維持する
      // (caller が値を明示しない再認証で誤って権限を失わせないため)
      if (action.payload?.isAdmin !== undefined) {
        state.isAdmin = action.payload.isAdmin;
      }
      if (action.payload?.onboarded !== undefined) {
        state.onboarded = action.payload.onboarded;
      }
    },

    setAuthenticated(
      state,
      action: PayloadAction<{ isAdmin?: boolean; onboarded?: boolean } | undefined>
    ) {
      state.isAuthenticated = true;
      state.loading = false;
      if (action.payload?.isAdmin !== undefined) {
        state.isAdmin = action.payload.isAdmin;
      }
      if (action.payload?.onboarded !== undefined) {
        state.onboarded = action.payload.onboarded;
      }
    },

    /** Welcome 画面で「はじめる」を押した直後に dispatch する。store を即時更新して
     *  Protected の /welcome リダイレクトループを回避する（API 側は IS NULL ガードで冪等）。 */
    markOnboarded(state) {
      state.onboarded = true;
    },

    clearAuth(state) {
      state.isAuthenticated = false;
      state.loading = false;
      state.isAdmin = false;
      state.onboarded = false;
    },

    finishLoading(state) {
      state.loading = false;
    },
  },
});

export const { setAuthData, setAuthenticated, markOnboarded, clearAuth, finishLoading } =
  authSlice.actions;
export default authSlice.reducer;
