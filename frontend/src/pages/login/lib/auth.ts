/**
 * Cognito Hosted UI の認可 URL を組み立てる。
 *
 * provider を渡すと特定の IdP(例: Google)へ直行する。省略すると identity_provider を付けず、
 * Cognito Hosted UI のログイン画面(メール/パスワード + ソーシャルを選べる)へ遷移する。
 */
export function getCognitoAuthUrl(provider?: string): string {
  const url = new URL(
    `https://${import.meta.env.VITE_COGNITO_DOMAIN}/oauth2/authorize`
  );

  url.searchParams.append('client_id', import.meta.env.VITE_CLIENT_ID);
  url.searchParams.append('redirect_uri', import.meta.env.VITE_REDIRECT_URI);
  url.searchParams.append('response_type', import.meta.env.VITE_RESPONSE_TYPE);
  url.searchParams.append('scope', import.meta.env.VITE_SCOPE);
  url.searchParams.append('state', generateRandomState());
  if (provider) {
    url.searchParams.append('identity_provider', provider);
  }

  return url.toString();
}

function generateRandomState(length = 32): string {
  const array = new Uint8Array(length);
  window.crypto.getRandomValues(array);
  return Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join(
    ''
  );
}
