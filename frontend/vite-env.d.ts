/// <reference types="vite/client" />

/**
 * `strictImportMetaEnv` を宣言すると、`ImportMetaEnv` が継承している
 * `Record<string, any>` の逃げ道が閉じる。
 *
 * 閉じないと、**綴りを間違えた `VITE_*` も `any` として型検査を通る**。
 * 実際このファイルは Cognito 時代の 5 つを宣言したまま残っていて、
 * いま読んでいる `VITE_OIDC_*` は 1 つも宣言が無いのに型検査が通っていた。
 */
interface ViteTypeOptions {
  strictImportMetaEnv: unknown;
}

interface ImportMetaEnv {
  /**
   * **すべて optional。** 値はビルド時に注入されるので、注入されなければ無い。
   *
   * 必須と宣言すると「必ずある」という嘘を型が保証してしまい、欠けた設定で
   * 認可 URL を組み立てるコードが素通りする（クリックして初めて例外になる）。
   */
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_OIDC_AUTHORIZE_URI?: string;
  readonly VITE_OIDC_CLIENT_ID?: string;
  readonly VITE_OIDC_REDIRECT_URI?: string;
  readonly VITE_OIDC_SCOPE?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
