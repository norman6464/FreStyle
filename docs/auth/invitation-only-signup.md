# 招待限定サインアップ（OIDC ゲート）

## 概要

FreStyle のアカウント作成は **必ず上位ロールからの招待経由** に限定する。
具体的には:

- **trainee** は CompanyAdmin から招待されて初めてログインできる
- **CompanyAdmin** は SuperAdmin から招待されて初めてログインできる
- **SuperAdmin** だけは Cognito 側のグループ `admin` 所属で初期作成できる（運営側が直接作成）

---

## 背景

PR-B（SES マジックリンク方式）導入前は、Google OIDC 経由で誰でもログインすると `users` 行が自動作成され（role=`trainee`）、社外の任意ユーザーがアプリにアクセスできてしまう状態だった。

招待マジックリンクを採用しただけでは、メールリンクを踏まずに `/login` から直接 OIDC ログインされた場合に DB に行が作られてしまう。**バックエンド側で「招待なし新規作成」を拒否する必要がある**ため、本 PR で `auth_handler.upsertUserFromIDToken` にゲートを追加した。

---

## 実装

### `Callback` の挙動

`POST /api/v2/auth/cognito/callback` の処理フロー:

1. 認可コードをトークンに交換 → access_token / refresh_token / id_token を取得
2. HttpOnly Cookie に access_token / refresh_token を設定
3. `upsertUserFromIDToken` で users 行を確認・作成
4. **作成不可（=招待なし & Cognito admin でもない）の場合**:
   - Cookie をクリア（`middleware.ClearAuthCookies(c)`）
   - `403 Forbidden` を返す（`{"error": "invitation_required", "message": "..."}`）
5. それ以外は `200 OK` を返す

### `upsertUserFromIDToken` の判定ルール

| 状況 | 判定 |
|---|---|
| DB に sub 既存 | 許可（必要なら role を Cognito group に同期） |
| DB に未登録 + `cognito:groups` に `admin` を含む | 許可（role=super_admin で作成） |
| DB に未登録 + email に対する pending invitation あり | 許可（role / company_id / displayName を invitation から反映） |
| 上記いずれもなし | **拒否（false 返却）** |

`Refresh` (`POST /auth/cognito/refresh-token`) では戻り値を無視する。refresh は既存ユーザーが対象なので、一般的にはここで拒否する必要はない（DB から削除済みのユーザーが refresh を試みる稀ケースは別 issue で扱う）。

---

## フロントエンドの想定動作

`POST /auth/cognito/callback` が `403 invitation_required` を返した場合、フロントは:

1. ユーザーに「招待が必要です」とメッセージを表示
2. `/login` に戻して再試行を促す
3. 招待メールに記載の `/invitations/accept?token=...` リンクから入り直してもらう（PR-D で実装予定）

---

## テスト（`auth_handler_test.go`）

| テストケース | 期待 |
|---|---|
| 招待なし & Cognito admin でもない新規ユーザー | 拒否（allowed=false / users.Create が呼ばれない） |
| Cognito group `admin` 新規ユーザー（招待なし） | 許可（super_admin で作成） |
| 招待あり新規ユーザー | 許可（招待の role / company_id / displayName が反映 / invitation が accepted にマーク） |
| 既存ユーザー（招待無関係） | 常に許可 |
| 既存ユーザー + Cognito group `admin` | 許可（DB role が super_admin に昇格） |
| 壊れた id_token | 拒否（false） |

---

## 既知の制約

- 既存 SuperAdmin（Cognito group `admin` 所属）が DB から削除されると、再ログイン時に「Cognito admin」判定で再作成される（=他人にアカウントを乗っ取られる懸念は無いが、ID は新採番）。
- Cognito group `admin` の付け外しは AWS Console から実施するため、運用ドキュメントで管理する必要がある。
- 招待メールリンクの token を `sessionStorage` に渡し、callback 側で読むフローは PR-D で実装予定。本 PR の段階では「招待 email」と「ログインしてくる Cognito ユーザーの email」が一致する前提。
