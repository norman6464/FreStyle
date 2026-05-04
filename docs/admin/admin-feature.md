# 管理者機能（Admin）

## 概要

Cognito の `cognito:groups` クレームに `admin` を含むユーザーのみアクセス可能な管理機能。今回はまず **練習シナリオの CRUD** を提供する。

## 認可フロー

```
ブラウザ
  └─ Cookie (Cognito JWT)
      └─ Spring Security / JwtCookieFilter
          └─ JwtAuthenticationConverter
              setAuthoritiesClaimName("cognito:groups")
              setAuthorityPrefix("ROLE_")
                ↓
              ユーザーが Cognito group "admin" 所属なら
              Spring Security に "ROLE_admin" 権限が付与される
                ↓
              SecurityConfig.requestMatchers("/api/admin/**").hasRole("admin")
              で /api/admin/** へのアクセスを制御
```

## バックエンド

### `SecurityConfig`

```java
.authorizeHttpRequests(auth -> auth
    .requestMatchers(...).permitAll()
    .requestMatchers("/api/admin/**").hasRole("admin")  // ← これ
    .anyRequest().authenticated())
```

### `/api/admin/scenarios`（[`AdminScenarioController`](../../FreStyle/src/main/java/com/example/FreStyle/controller/AdminScenarioController.java)）

| メソッド | パス | 用途 |
|---|---|---|
| GET | `/api/admin/scenarios` | 全シナリオ一覧 |
| POST | `/api/admin/scenarios` | 新規作成 |
| PUT | `/api/admin/scenarios/{id}` | 更新 |
| DELETE | `/api/admin/scenarios/{id}` | 削除 |

UseCase（クリーンアーキテクチャ）:
- [`CreatePracticeScenarioUseCase`](../../FreStyle/src/main/java/com/example/FreStyle/usecase/CreatePracticeScenarioUseCase.java)
- [`UpdatePracticeScenarioUseCase`](../../FreStyle/src/main/java/com/example/FreStyle/usecase/UpdatePracticeScenarioUseCase.java)
- [`DeletePracticeScenarioUseCase`](../../FreStyle/src/main/java/com/example/FreStyle/usecase/DeletePracticeScenarioUseCase.java)

### `/api/auth/cognito/me` の拡張

レスポンスに `groups: string[]` と `isAdmin: boolean` を追加。フロントエンドの ServerInitializer がこれを Redux store の `auth.isAdmin` に伝播し、サイドバーの管理メニューを出し分ける。

## フロントエンド

| 場所 | 変更点 |
|---|---|
| `src/types/index.ts` | `AuthState.isAdmin: boolean` を追加 |
| `src/store/authSlice.ts` | `setAuthData` / `setAuthenticated` が `{ isAdmin?: boolean }` を受け取る |
| `src/utils/AuthInitializer.tsx` | `/me` の `isAdmin` を store に伝播 |
| `src/components/layout/Sidebar.tsx` | `isAdmin` のときだけ「管理: シナリオ」メニュー表示 |
| `src/pages/AdminScenariosPage.tsx` | 管理画面本体（一覧・作成・編集・削除フォーム） |
| `src/repositories/AdminScenarioRepository.ts` | `/api/admin/scenarios` クライアント |
| `src/App.tsx` | `/admin/scenarios` ルート追加 |

非 admin が `/admin/scenarios` を直接叩いても `<Navigate to="/" replace />` でホームへリダイレクト + バックエンド側で 403。

## Cognito で admin グループを作る手順

```bash
USER_POOL_ID=ap-northeast-1_TkRen4lyD   # oauth-cource-user-pool（本番）

# 1) admin グループを作成（一度だけ）
aws cognito-idp create-group --region ap-northeast-1 \
  --user-pool-id $USER_POOL_ID \
  --group-name admin \
  --description "FreStyle の管理者グループ"

# 2) 自分のユーザー名を確認
aws cognito-idp list-users --region ap-northeast-1 --user-pool-id $USER_POOL_ID \
  --query 'Users[].[Username,Attributes[?Name==`email`].Value | [0]]' --output table

# 3) 自分を admin グループに追加
aws cognito-idp admin-add-user-to-group --region ap-northeast-1 \
  --user-pool-id $USER_POOL_ID \
  --username <ステップ 2 で確認した username> \
  --group-name admin

# 4) いったんログアウト → 再ログインで JWT に cognito:groups: ["admin"] が乗る
```

## 動作確認

1. ログイン
2. サイドバーに「管理: シナリオ」メニューが表示される
3. クリックして `/admin/scenarios` に遷移、シナリオ一覧が表示される
4. 「新規シナリオを作成」フォームから作成・編集・削除できる

非 admin ユーザーで:
- サイドバーには「管理: シナリオ」が**表示されない**
- 直接 URL `/admin/scenarios` に行くと **/ にリダイレクト**
- 万一 API を直接叩いても **403 Forbidden**

## メンバー招待機能（AdminInvitation）

### 概要

Super Admin が会社・ロール・メールアドレスを指定すると、Cognito `AdminCreateUser` API が一時パスワード付きの招待メールを送信する。  
受信者は Cognito Hosted UI で初回ログイン時にパスワード変更を要求される。

### バックエンド

| 層 | ファイル |
|---|---|
| Handler | `internal/handler/admin_invitation_handler.go` |
| UseCase | `internal/usecase/create_admin_invitation_usecase.go` |
| Repository | `internal/repository/admin_invitation_repository.go` |
| Cognito client | `internal/infra/cognito/admin_client.go` |

`CognitoAdminClient` インターフェースを実装した `AdminClient` が `AdminCreateUser` を呼び出す。  
環境変数 `COGNITO_USER_POOL_ID` が未設定の場合はスタブにフォールバックし、メールは送信されない（ローカル開発用）。

### 必要な環境変数

| 変数 | 値（本番） | 備考 |
|---|---|---|
| `COGNITO_USER_POOL_ID` | `ap-northeast-1_TkRen4lyD` | AdminCreateUser API に必要 |

ECS タスク定義（`ecs.yml`）に追加済み。

### フロントエンド

- `src/pages/AdminInvitationsPage.tsx`: 会社セレクター・メール・ロール・表示名のフォーム
- `src/repositories/AdminInvitationRepository.ts`: POST に `companyId` を含める
- `src/repositories/CompanyRepository.ts`: `/admin/companies` から会社一覧を取得

### API

| メソッド | パス | 用途 |
|---|---|---|
| GET | `/api/v2/admin/invitations` | 未承諾招待一覧 |
| POST | `/api/v2/admin/invitations` | 招待メール送信 |
| DELETE | `/api/v2/admin/invitations/:id` | 招待取り消し |

## 拡張ポイント

- ユーザー管理（一覧・凍結・admin 付与）
- シナリオの公開フラグ管理
- レポートの集計（システム全体）
- 監査ログ
