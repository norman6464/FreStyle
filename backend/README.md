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
go test ./...
```

## デプロイ方針（後続 Phase で確立）

- ALB の path-based routing で `/api/v2/*` を Go ECS Service に振り、
  既存 `/api/*`（Spring Boot）と並行運用する
- 全機能の移植が完了したら `/api/v2/*` から `/api/*` に切替、Spring Boot 側を destroy
- ECS Fargate のタスクスペックは Spring Boot の `2 vCPU / 4GB` から
  Go 版では `0.25 vCPU / 0.5GB` に下げる（コスト約 80% 削減見込み）
