/*
 * features/auth の Public API。
 *
 * 認証（現在ユーザーの取得・ログイン・ログアウト・パスワード再設定）のユーザーシナリオ。
 * 認証状態そのものは entities/user の Redux slice が持ち、ここはそれを操作する feature。
 */
export { useAuth } from './model/useAuth';
