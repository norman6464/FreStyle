import type { Page } from '@playwright/test';

/**
 * ローカル E2E ビルド（`e2e.yml` の `VITE_OIDC_*`）を認証済みとして描画するための共通 mock。
 *
 * AuthInitializer はもう backend の Cookie セッションを見ない。ビルド時の env で決まる
 * 発行者（ここでは Dex）のクライアント側状態を見てから `POST /auth/login` でセッションを
 * 確立する（`shared/lib/auth/currentIdToken.ts` の `subscribeAuthState`）。Dex は
 * `localStorage` の `dex.session`（`shared/lib/auth/dexSession.ts` 参照）の有無を見るだけの
 * 一発確認なので、有効に見える値を先に入れておけば「サインイン済み」として扱われる。
 *
 * `page.addInitScript` は `page.goto` より前に呼ぶこと（登録した navigation から効く）。
 */
export async function mockAuthenticated(page: Page, overrides: Record<string, unknown> = {}) {
  await page.addInitScript(() => {
    localStorage.setItem(
      'dex.session',
      JSON.stringify({
        idToken: 'e2e-fake-id-token',
        refreshToken: 'e2e-fake-refresh-token',
        // 十分先の期限にして、getValidDexIdToken が実在しない Dex への更新を試みないようにする。
        expiresAt: Date.now() + 60 * 60 * 1000,
      }),
    );
  });

  // 既定: 未指定の API は空配列で 200(リスト/オブジェクトどちらの消費側も undefined 安全)。
  // これには POST /auth/login（AuthInitializer がセッション確立に呼ぶ）も含まれる——
  // 呼び出しが成功で終わることだけが要点で、応答の中身は見ていない。
  await page.route('**/api/v2/**', (route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
  );
  for (const [pattern, body] of Object.entries(overrides)) {
    await page.route(pattern, (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(body),
      })
    );
  }
}
