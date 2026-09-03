/*
 * entities/user の Public API。
 *
 * 外から使ってよいものだけを名前付きで re-export する（FSD 公式仕様。`export *` は使わない）。
 * この Slice の内部ファイルを直接 import してはいけない。
 */

export { default as AuthRepository } from './api/authRepository';
export type { UserInfo } from './api/authRepository';
export { default as ProfileRepository } from './api/profileRepository';
export type { UpdateProfileRequest } from './api/profileRepository';
export { default as ImageUploadRepository } from './api/imageUploadRepository';
export { default as ChapterViewRepository } from './api/chapterViewRepository';

export type { Profile, User, AuthState, UserRole } from './model/types';

// 認証状態の Redux slice。reducer は app 側の configureStore が組み立てる。
export { default as authReducer } from './model/authSlice';
export { setAuthData, clearAuth, finishLoading } from './model/authSlice';
