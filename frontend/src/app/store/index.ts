import { configureStore } from '@reduxjs/toolkit';
import { authReducer } from '@/entities/user';

/*
 * Redux ストア本体。
 *
 * slice の実体は各 entity（`entities/user/model/authSlice` など）が持ち、
 * app レイヤーで各 slice の reducer を集約する。各層は RootState を直接 import せず、
 * `@/shared/lib/store` の useAppSelector / useAppDispatch 経由で状態にアクセスする。
 */
export const store = configureStore({
  reducer: {
    auth: authReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
