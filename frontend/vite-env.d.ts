/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
  readonly VITE_WEB_SOCKET_URL_AI_CHAT: string;
  readonly VITE_COGNITO_DOMAIN: string;
  readonly VITE_CLIENT_ID: string;
  readonly VITE_REDIRECT_URI: string;
  readonly VITE_RESPONSE_TYPE: string;
  readonly VITE_SCOPE: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
