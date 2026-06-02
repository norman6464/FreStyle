# FreStyle backend (Go)

Spring Boot から Go (Gin + GORM) に段階的に置き換える Go 版バックエンド。
Phase 0 では基盤のみ。Phase 1 以降で `FreStyle/` 配下の Spring Boot コントローラーを機能ごとに移植する。

## ディレクトリ構造（クリーンアーキテクチャ）

```
backend/
├── cmd/server/          エントリーポイント (main.go)
├── internal/
│   ├── handler/         HTTP ハンドラ層 (Spring Boot Controller 相当)
│   ├── usecase/         ユースケース層 (Spring Boot UseCase 相当)
│   ├── repository/      リポジトリ層 (Spring Boot Repository 相当)
│   ├── domain/          ドメインモデル (Spring Boot Entity 相当)
│   └── infra/
│       ├── config/      環境変数ロード
│       └── database/    GORM セットアップ
├── Dockerfile           multi-stage / distroless / static binary
└── go.mod
```

### 依存方向ルール

```
handler  →  usecase  →  repository  →  domain
(外側)     (Application)  (Infra)        (Domain)
```

- handler は usecase のみを呼び出す（repository を直接呼ばない）
- usecase は handler のことを知らない
- repository は usecase のことを知らない
- domain は他のどの層にも依存しない

### archlint — 依存方向の静的検証

上記ルールを人手のレビューだけに頼らず CI で機械的に弾くため、自作 linter `cmd/archlint` を用意している（Go 標準の `go/parser` / `go/ast` のみ・外部依存なし）。

```bash
make archlint        # = go run ./cmd/archlint .
```

`internal/` 配下の `.go`（`_test.go` は除外）の import を解析し、層をまたぐ禁止依存を検出する。違反は `path:line: メッセージ` 形式で出力し exit 1。

| ソース層 | 禁止する import |
|---|---|
| `domain` | 他の内部層すべて / `gin` / `net/http`（標準ライブラリ + GORM tag のみ） |
| `usecase/repository`（port） | `domain` 以外の内部層 / `gin` / `net/http` |
| `usecase` | `handler` / `adapter/persistence`（DIP: port に依存）/ `gin` / `net/http` |
| `adapter/persistence` | `handler` / `usecase` 本体（依存先は port のみ）/ `gin` |
| `infra` | `handler` / `usecase` / `usecase/repository` / `adapter/persistence` / `gin` |
| `handler` | `adapter/persistence` の直接 import（**wiring の `router.go` / `routes_*.go` は例外**） |

- `usecase → infra`、`infra → net/http`（HTTP クライアント用途）、`handler → infra`（auth / embed の pragmatic 例外）は **許容**。
- 意図的に例外を通したいときは、import 行末に `//archlint:allow <理由>`、ファイル全体なら先頭コメントに `//archlint:ignore-file <理由>` を付ける。
- ルールを追加・変更したら `cmd/archlint/main.go` の `rules` を編集し、`cmd/archlint/main_test.go` にケースを足す。

### apispec-lint — ルート ↔ swaggo 注釈の整合検証（strict path 照合）

CLAUDE.md §2.7「新しい HTTP endpoint には handler メソッドの直前に swaggo annotation を必ず書く」を機械化した自作 linter（`go/ast` のみ）。ルートと `@Router` 宣言の **HTTP method + path を完全一致で照合**し、注釈漏れ・path 相違・method 相違を CI で弾く。

```bash
make apispec-lint    # = go run ./cmd/apispec-lint .
```

- `internal/handler` の `g.GET("/path", ..., h.Method)` 形式のルート登録を AST で抽出する（最後の引数が handler）。
- レシーバ付き func の doc コメントの `@Router /path [method]` 宣言を `(method, path)` として全ファイル横断で収集する（1 メソッドが複数 `@Router` を持つ場合は複数宣言）。
- gin の `:id` / `*filepath` を swaggo の `{id}` / `{filepath}` に正規化し、ルートと一致する `@Router` 宣言が無ければ `path:line` で報告し exit 1。
- SSE / WebSocket / multipart など OpenAPI で表現しない endpoint や、フロント互換の別 path / 別 method エイリアスは、ルート登録行の行末に `//apispec:allow <理由>`、ファイル全体なら先頭コメントに `//apispec:ignore-file <理由>` で抑制する。

> 生成 spec そのものの正しさ（schema 等）は `make openapi` の drift check（CI）が担い、apispec-lint は「ルートに対応する @Router 宣言が正しい method/path で存在するか」を担う。両者は補完関係。

### naminglint — usecase 命名・構造規約の検証

CLAUDE.md §2.3「1 usecase = struct + コンストラクタ + Execute メソッド」を機械化した自作 linter（`go/ast` のみ）。対象は `internal/usecase` 直下（`repository` サブパッケージは除外）。

```bash
make naminglint    # = go run ./cmd/naminglint .
```

- 公開 struct `XxxUseCase` に対し、コンストラクタ `NewXxxUseCase` と メソッド `Execute` の存在を検査し、欠けていれば `path:line` で報告し exit 1。
- 複数 CRUD を束ねる集約 usecase（`CourseUseCase` / `TeachingMaterialUseCase` / `ListAdminInvitationsUseCase`）など `Execute` 単一メソッドにしない正当な例外は、struct の doc コメントに `//naminglint:allow <理由>` を付けて Execute 必須を免除する。ファイル全体は先頭コメントに `//naminglint:ignore-file <理由>`。

## ローカル開発

```bash
cd backend

# 依存解決
go mod tidy

# 環境変数 (最低限)
# 本番 / staging は Supabase Transaction pooler URL を使う (RDS は 2026-05 廃止)。
# ローカル docker postgres で開発する場合のみ DB_HOST 系を使う。
export DATABASE_URL='postgresql://postgres.xxxxx:PASSWORD@aws-N-ap-northeast-1.pooler.supabase.com:6543/postgres?sslmode=require'
export PORT=8080

# 起動
go run ./cmd/server
```

DATABASE_URL の取得方法:
- **ローカル / dev 開発**: 各自で Supabase project を作成し、 Dashboard → 上部 `Connect` → `Direct` タブ → `Transaction pooler` の URI を取得。 個人 dev 用 secret に保存する場合は `frestyle-dev/database-url` のように環境 prefix を分けることを推奨（本番 secret `frestyle-prod/database-url` をローカルから参照すると誤更新リスクあり）
- **CI / staging**: AWS Secrets Manager の対応する `<env>/database-url` から取得:
  ```bash
  aws secretsmanager get-secret-value --region ap-northeast-1 \
    --secret-id <env>/database-url --query SecretString --output text
  ```

確認:

```bash
curl http://localhost:8080/
# => {"message":"FreStyle Go backend"}
```

## ビルド（Docker）

```bash
docker build -t frestyle-backend:latest .
# 完成イメージは distroless ベースで ~30MB 前後
```

## テスト / Lint

```bash
go vet ./...
go test ./...        # 単体テスト（DB 不要）
make archlint      # クリーンアーキテクチャ依存方向チェック
make apispec-lint  # ルート ↔ swaggo @Router 注釈チェック（strict path 照合）
make naminglint    # usecase 命名・構造規約チェック
make verify        # gofmt / vet / build / test / 3 linter を一括実行
```

### 結合テスト（本物の PostgreSQL）

`adapter/persistence` 層は、GORM の実 SQL を **本物の PostgreSQL** に対して検証する結合テストを持つ。単体テスト（usecase の mock）とは独立し、`//go:build integration` タグで隔離しているため通常の `go test ./...` には含まれない。

```bash
make test-integration
# = docker compose -f docker-compose.integration.yml up -d --wait
#   → TEST_DATABASE_URL=... go test -tags=integration -run Integration ./...
#   → docker compose ... down -v（テスト失敗でも teardown）
```

- **DB コンテナ**: `docker-compose.integration.yml` の `postgres-integration-test`（`postgres:17.6-alpine`、host 側 5433、`tmpfs` で毎回まっさら）。
- **接続先**: `TEST_DATABASE_URL`（既定 `postgres://frestyle:frestyle@localhost:5433/frestyle_integration?sslmode=disable`）。未設定 / 未起動なら結合テストは `t.Skip`。本番が使う `DATABASE_URL` とは**別の env**で、Supabase / 本番には**接続しない**。
- **安全弁（本番 Supabase 保護）**: `OpenTestDB` は接続先が `supabase.com` / `pooler.supabase` を含む場合、接続前に `t.Fatal` で**必ず落とす**。結合テストは `TruncateAll`（`TRUNCATE ... CASCADE`）でテーブルを破壊するため、誤って `TEST_DATABASE_URL` に本番を入れてもデータを消さない。
- **スキーマ**: `testsupport.OpenTestDB(t)` が `database.AutoMigrateAll(db)`（seed なしの全 domain AutoMigrate）でスキーマを構築。テスト間は `testsupport.TruncateAll(t, db, ...)` で独立性を確保。
- **命名規約**: 結合テストの関数名には `Integration` を含める（CI / make は `-run Integration` で選別実行する。`TEST_DATABASE_URL` を env に持つこのジョブで env 依存の単体テストを巻き込まないため）。
- **CI**: `ci-backend-go.yml` の `integration` ジョブが docker compose で Postgres を起動して実行する（単体 `test` ジョブとは別ジョブ）。
- 新しい repository を足したら `internal/adapter/persistence/{entity}_repository_integration_test.go` を追加する。

## ログ方針（CloudWatch コスト対策）

- アクセスログは `gin.LoggerWithConfig` で出力し、`SkipPaths` に
  `/api/v2/health`（ALB が 30 秒間隔で叩くヘルスチェック）と `/` を指定して**除外**する。
  大量の health ログが CloudWatch の取り込み課金を押し上げるのを防ぐため。
- 認証・業務ロジックの WARN / ERROR は `log.Printf` で従来どおり出力する。
- 本番（`APP_ENV` が `"local"` 以外）は gin を **release モード**にして、起動時の
  `[GIN-debug]` ルート登録ログ・warning を抑止する（ローカルは debug のまま）。
- ロググループ `/ecs/frestyle-prod` は CFn 側で retention 14 日。Container Insights は
  コスト削減のため無効（詳細は frestyle-infrastructure の `docs/06-maintenance-cookbook.md` §15）。

## レートリミット（公開エンドポイント）

`middleware.RateLimitPerMinute(perMinute, burst)`（per-IP トークンバケット）を公開ルートに適用する。

| エンドポイント | 制限 | 目的 |
|---|---|---|
| `POST /company-applications` | 5/分・burst 5 | 公開フォームのスパム対策 |
| `GET /invitations/accept/:token` | 30/分・burst 10 | 招待 token の総当たり緩和 |
| `POST /auth/cognito/callback` | 30/分・burst 10 | コールバック総当たり緩和 |
| `POST /auth/cognito/refresh-token` | 60/分・burst 30 | 正規利用は高頻度なので緩め（NAT 共有 IP 考慮） |

- 超過時は `429 Too Many Requests` + `Retry-After: 60`。
- **単一 ECS タスク前提の in-memory 実装**。スケールアウト時はインスタンスごとに別カウントになるため共有ストア（Redis 等）が必要。
- キーは `c.ClientIP()`（ALB / CloudFront の X-Forwarded-For 由来）。詐称を完全には防げないため、軽い総当たり / スパムの緩和が目的。

## デプロイ方針（後続 Phase で確立）

- ALB の path-based routing で `/api/v2/*` を Go ECS Service に振り、
  既存 `/api/*`（Spring Boot）と並行運用する
- 全機能の移植が完了したら `/api/v2/*` から `/api/*` に切替、Spring Boot 側を destroy
- ECS Fargate のタスクスペックは Spring Boot の `2 vCPU / 4GB` から
  Go 版では `0.25 vCPU / 0.5GB` に下げる（コスト約 80% 削減見込み）
