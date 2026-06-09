# ログイン（認証）フロー

FreStyle のログイン画面（`/login`）は **メール / パスワード** と **Google（Hosted UI）** の 2 経路を提供する。
どちらも最終的に Cognito が発行した token を **HttpOnly Cookie**（access / refresh）に格納し、以降の
API は Cookie 認証で動く。新規ユーザーは「招待」または Cognito admin group が無いと弾かれる（招待ゲート）。

## 経路

| 経路 | フロント | backend エンドポイント | Cognito |
|---|---|---|---|
| メール / パスワード | ログインフォーム送信 | `POST /api/v2/auth/cognito/login` | InitiateAuth `USER_PASSWORD_AUTH` |
| Google | 「Google でログイン」→ Hosted UI | `POST /api/v2/auth/login`（認可コード交換） | OAuth2 code → token |
| トークン更新 | 自動 | `POST /api/v2/auth/refresh` | refresh_token |

> 旧 `/auth/login` は Hosted UI の認可コード交換、アプリ内フォームは `/auth/cognito/login`。
> フロントの `src/constants/apiRoutes.ts` の `AUTH.login` / `AUTH.callback` と一致させている。

## メール / パスワードログインの実装

### backend（Go）

- `internal/infra/cognito/password_authenticator.go`: `PasswordAuthenticator`。
  Cognito `InitiateAuth(USER_PASSWORD_AUTH)` を呼ぶ。client secret 付きクライアントなので
  `SECRET_HASH = Base64(HMAC-SHA256(username + clientId, clientSecret))` を計算して渡す。
  資格情報誤り（`NotAuthorized` / `UserNotFound` / `UserNotConfirmed`）は `ErrInvalidCredentials`
  に丸めてユーザー列挙を防ぐ。
- `internal/handler/auth_handler.go` の `Login`: フォームを受けて `Authenticate` → Cookie 発行 →
  `upsertUserFromIDToken`（Callback と共通の招待ゲート）。拒否時は Cookie をクリアして 403。
- `internal/handler/routes_auth.go`: `POST /auth/cognito/login`。パスワード総当たり面なので
  per-IP のレート制限を callback より厳しく（10 req/min, burst 5）。

### frontend

- `src/pages/LoginPage.tsx`: 公開ヘッダー（`PublicHeader`）+ メール / パスワードフォーム +
  「または」+ Google ボタン + 招待 / 利用申請の案内。
- `src/hooks/useLoginPage.ts`: フォーム状態 + `handleLogin`。成功時は **フル再読み込み**
  （`window.location.assign('/')`）でトップへ。SPA 内 navigate では `AuthInitializer` が
  `/auth/me` を引き直さず role / isAdmin が確定しないため、あえてフルロードする。
  403（招待なし）は専用文言、それ以外の失敗は「メールアドレスまたはパスワードが正しくありません。」。

## Cognito 前提

`USER_PASSWORD_AUTH` を使うには App Client の `ExplicitAuthFlows` に
**`ALLOW_USER_PASSWORD_AUTH`** が必要（既存の SRP / REFRESH / Hosted UI code フローは維持）。
本番クライアントは手動管理のため、有効化手順は infra リポの
`make enable-cognito-user-password-auth`（docs/18 系）に集約している。

## OpenAPI

`/auth/cognito/login` は swaggo annotation 付き。handler 変更時は `make openapi` で
`backend/docs/swagger.{json,yaml}` を再生成してコミットする。
