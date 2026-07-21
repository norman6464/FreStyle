/*
 * entities/user の Public API。
 *
 * 外から使ってよいものだけを名前付きで re-export する（FSD 公式仕様。`export *` は使わない）。
 * この Slice の内部ファイルを直接 import してはいけない。
 */

export { default as AuthRepository } from './api/authRepository';
export type { LoginRequest } from './api/authRepository';
export type { ForgotPasswordRequest } from './api/authRepository';
export type { ConfirmForgotPasswordRequest } from './api/authRepository';
export type { UserInfo } from './api/authRepository';
export { default as ProfileRepository } from './api/profileRepository';
export type { UpdateProfileRequest } from './api/profileRepository';
export { default as ProfileStatsRepository } from './api/profileStatsRepository';
export type { ProfileStats } from './api/profileStatsRepository';
export { default as ImageUploadRepository } from './api/imageUploadRepository';
export { default as DashboardRepository } from './api/dashboardRepository';

export type {
  Profile,
  User,
  AuthState,
  SnsProvider,
  UserDailyActivity,
  UserDashboard,
} from './model/types';

export { default as ProfileStatsSection } from './ui/ProfileStatsSection';

// 認証状態の Redux slice。reducer は app 側の configureStore が組み立てる。
export { default as authReducer } from './model/authSlice';
export { setAuthData, clearAuth, finishLoading, setAiChatEnabledForTrainees } from './model/authSlice';
