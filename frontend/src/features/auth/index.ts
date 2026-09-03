/*
 * features/auth の Public API。
 *
 * 認証（現在ユーザーの取得・ログイン・ログアウト）のユーザーシナリオ。
 * 認証状態そのものは entities/user の Redux slice が持ち、ここはそれを操作する feature。
 */
export { useAuth } from './model/useAuth';
export { useOidcLogin } from './model/useOidcLogin';
export type { OidcLogin } from './model/useOidcLogin';
export { default as AuthUnavailableNotice } from './ui/AuthUnavailableNotice';
export { buildAuthorizeUrl, consumeAuthFlowState } from './lib/oidcAuthUrl';
export { readAuthConfig, DEFAULT_SCOPE } from './lib/authConfig';
export type { AuthConfig, ConfiguredAuth, UnconfiguredAuth } from './lib/authConfig';
