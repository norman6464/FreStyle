# OpenAPI / Swagger UI

## 概要

FreStyle バックエンドは [springdoc-openapi](https://springdoc.org/) を使って実装の Controller / DTO から OpenAPI 3.x 仕様書（JSON）と Swagger UI（HTML）を自動生成する。

| エンドポイント | 用途 |
|---|---|
| `GET /v3/api-docs` | OpenAPI 3.x 仕様書（JSON） |
| `GET /v3/api-docs.yaml` | OpenAPI 3.x 仕様書（YAML） |
| `GET /swagger-ui.html` | ブラウザで閲覧する Swagger UI |

ローカル: `http://localhost:8080/swagger-ui.html`
本番: `https://api.normanblog.com/swagger-ui.html`

## なぜ springdoc を選んだか

- Spring Boot 3 / Java 21 公式サポート（`springdoc-openapi-starter-webmvc-ui` 2.6.0）
- アノテーションは Swagger 2 系（`io.swagger.v3.oas.annotations`）に統一されており追加学習コスト小
- `@RestController` と DTO を自動スキャンするため、最小限のアノテーションで運用できる
- 別途 YAML 手書き運用を持たないので、コードと仕様書のドリフトが起きない

## 設定

### 依存追加（`FreStyle/build.gradle`）

```gradle
implementation 'org.springdoc:springdoc-openapi-starter-webmvc-ui:2.6.0'
```

### `application.properties`

```properties
springdoc.api-docs.path=/v3/api-docs
springdoc.swagger-ui.path=/swagger-ui.html
springdoc.swagger-ui.enabled=true
springdoc.swagger-ui.disable-swagger-default-url=true
springdoc.swagger-ui.tags-sorter=alpha
springdoc.swagger-ui.operations-sorter=alpha
# Actuator はドキュメントから除外
springdoc.paths-to-exclude=/actuator/**
```

本番で UI を非公開にしたい場合は `springdoc.swagger-ui.enabled=false` に切り替える。JSON のみ公開（仕様書配布用）も可能。

### Spring Security 許可

[`SecurityConfig.java`](../../FreStyle/src/main/java/com/example/FreStyle/auth/SecurityConfig.java) で以下を `permitAll()` に追加済み:

```
/v3/api-docs
/v3/api-docs/**
/swagger-ui.html
/swagger-ui/**
```

### メタ情報 / 認証スキーマ

[`OpenApiConfig.java`](../../FreStyle/src/main/java/com/example/FreStyle/config/OpenApiConfig.java) で以下を定義:

- `Info`: タイトル / 説明 / バージョン / Contact
- `Server`: 本番（`https://api.normanblog.com`）/ ローカル（`http://localhost:8080`）
- `SecurityScheme`: `cookieAuth` = `accessToken` HttpOnly Cookie（API Key in Cookie）
- デフォルト `SecurityRequirement` = `cookieAuth`（全エンドポイントが Cookie 認証必須として表現される）

認証不要のエンドポイント（`/api/hello` 等）は `@SecurityRequirements` を空指定して上書きする。

## Controller への注釈ポリシー

最低限のアノテーションだけ付けて、コードを汚さないようにする。

### クラスレベル: `@Tag` で分類

```java
@Tag(name = "AI Chat", description = "AI 対話セッションとメッセージの管理 API")
public class AiChatController { ... }
```

`@Tag` の `name` で Swagger UI のグループが切り替わる。同じ `name` を別 Controller に付けて手動でまとめることもできる。

### メソッドレベル: 必要に応じて `@Operation`

```java
@GetMapping("/hello")
@Operation(summary = "疎通確認", description = "認証不要。固定文字列 'hello' を返す。")
@SecurityRequirements
public String hello() { ... }
```

- `@Operation`: 概要を補足したいときだけ。基本は HTTP メソッド + パス + DTO 名で十分
- `@SecurityRequirements`: 認証不要にしたいときだけ（デフォルトの `cookieAuth` を解除）

### Tag 一覧（推奨）

| Tag | Controller |
|---|---|
| Health | HelloController |
| AI Chat | AiChatController, AiChatWebSocketController |
| Chat | ChatController, ChatWebSocketController |
| Auth | CognitoAuthController |
| Notes | NoteController, NoteImageController |
| Practice | PracticeController, ConversationTemplateController, ScenarioBookmarkController |
| Reports | LearningReportController, ScoreCardController, ScoreGoalController, ScoreTrendController |
| Social | FriendshipController, SharedSessionController |
| Profile | ProfileController, UserStatsController, ReminderSettingController |
| Goals | DailyGoalController, FavoritePhraseController, WeeklyChallengeController |
| Notifications | NotificationController |
| Ranking | RankingController |
| Session Notes | SessionNoteController |

新しい Controller を追加するときは上記表に従って `@Tag` を付けること。

## 動作確認

### ローカル

```bash
cd FreStyle
./gradlew bootRun
# 別ターミナル
curl -s http://localhost:8080/v3/api-docs | jq .info
# 期待: {"title":"FreStyle API","description":"...","version":"v1",...}

open http://localhost:8080/swagger-ui.html
```

### Try it out（Swagger UI から API を叩く）

1. ブラウザで `https://normanblog.com` にログイン → `accessToken` Cookie 発行
2. 同じブラウザで `https://api.normanblog.com/swagger-ui.html` を開く
3. 任意のエンドポイントの「Try it out」→「Execute」
4. ブラウザの `accessToken` Cookie が自動送信される（CORS が同じドメイン構成なら）

クロスオリジンでテストする場合は curl で:

```bash
TOKEN=$(curl -s --cookie-jar /tmp/cookies.txt https://normanblog.com/login ...)
curl -s --cookie /tmp/cookies.txt https://api.normanblog.com/api/chat/ai/sessions | jq .
```

## 仕様書の配布

```bash
# ローカルで static な OpenAPI JSON を吐き出し
curl -s http://localhost:8080/v3/api-docs > openapi.json

# YAML 版
curl -s http://localhost:8080/v3/api-docs.yaml > openapi.yaml
```

これをフロントエンドリポジトリに置く / API 利用者に配布する / Postman に import する 等で活用する。

## CI への組み込み（将来の TODO）

- [ ] `./gradlew openapi` カスタムタスクで OpenAPI JSON を出力 → リポジトリの `frontend/openapi.json` にコミット
- [ ] [openapi-typescript](https://github.com/openapi-ts/openapi-typescript) でフロントエンドの型定義を生成
- [ ] PR で OpenAPI 差分があったら CI に出させる（API 破壊変更レビューを強制）

## 関連

- [springdoc-openapi 公式ドキュメント](https://springdoc.org/)
- [OpenAPI 3.1 Specification](https://spec.openapis.org/oas/v3.1.0)
- [`docs/ARCHITECTURE.md`](../ARCHITECTURE.md) — Controller / DTO の責務
