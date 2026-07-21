import { configureStore } from '@reduxjs/toolkit';
import { authReducer } from '@/entities/user';

/*
 * Redux ストア本体。
 *
 * slice の実体は各 entity（`entities/user/model/authSlice` など）が持ち、
 * ここは組み立てだけを担う。FSD 移行の Phase 6 で `app/store` へ移す予定。
 */
export const store = configureStore({
  reducer: {
    auth: authReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
