// src/store/index.js
import { configureStore } from '@reduxjs/toolkit';
// defaultでexportしたもの別名で定義をしてもいい
import authReducer from './authSlice';

export const store = configureStore({
  reducer: {
    auth: authReducer,
  },
});
// reducerとはアプリケーションの状態を更新する関数意味がある
