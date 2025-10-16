// Cognitoのプレフィックスドメインで取得
const COGNITO_DOMAIN =
  'ap-northeast-1tkren4lyd.auth.ap-northeast-1.amazoncognito.com';
const CLIENT_ID = '2tpsf29tjjehqa9lsm2l2p7qbf';
const REDIRECT_URI = 'http://localhost:5173/login/callback'; // Cognitoのredirect_uriと一致する必要がある
const RESPONSE_TYPE = 'code';
const SCOPE = 'openid profile email';

export function getCognitoAuthUrl(provider) {
  const url = new URL(`https://${COGNITO_DOMAIN}/oauth2/authorize`); // Cognitoの認可コードフローURL
  url.searchParams.append('client_id', CLIENT_ID);
  url.searchParams.append('redirect_uri', REDIRECT_URI);
  url.searchParams.append('response_type', RESPONSE_TYPE);
  url.searchParams.append('scope', SCOPE);
  url.searchParams.append('state', generateRandomState());
  url.searchParams.append('identity_provider', provider);

  return url.toString();
}

function generateRandomState(length = 32) {
  const array = new Uint8Array(length);
  window.crypto.getRandomValues(array);
  return Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join(
    ''
  );
}
