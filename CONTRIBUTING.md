# コントリビューションガイド

FreStyle（新卒 IT エンジニア向け統合研修プラットフォーム）の開発に参加するための規約をまとめます。
チーム全員が参照できるよう、このファイルはリポジトリにコミットしています。

- バックエンド: Go / Gin / GORM（`backend/`）— 詳細は [`backend/README.md`](./backend/README.md)
- フロントエンド: React 19 / TypeScript / Vite / Tailwind（`frontend/`）
- インフラ / CI の設計判断: 別リポ [`frestyle-infrastructure`](https://github.com/norman6464/frestyle-infrastructure) の `docs/`

---

## 1. セットアップ

```bash
# バックエンド
cd backend
go mod download
go run ./cmd/server        # DATABASE_URL 等は .env で設定

# フロントエンド
cd frontend
npm install
npm run dev
```

DB 接続情報・環境変数は `.env`（gitignore 済）に置き、**絶対にコミットしない**。

---

## 2. ブランチ / コミット / PR

### ブランチ

- `main` への直接コミットは禁止（ブランチ保護済み）。必ず PR 経由。
- ブランチ名は Prefix を付ける: `feat/*` / `fix/*` / `refactor/*` / `docs/*` / `test/*` / `chore/*`
- マージ後はローカルブランチを削除し、新しい作業は `main` から切り直す。

### コミットメッセージ

- 日本語で書く。先頭に Prefix: `feat` / `fix` / `refactor` / `docs` / `test` / `chore` / `perf` / `style`
- 例: `feat: 企業申請フォームを追加`

### 言語

- **日本語**: PR タイトル / 本文 / Issue / コミットメッセージ / コメント
- **英語**: 識別子（型・変数・関数名）
- 他社プロダクト名（Zenn / Qiita / Slack / Notion 等）は PR / Issue / コード / docs に書かない。機能の中身で説明する。

### PR フロー

1. Issue を起票（テンプレートあり）
2. ブランチを切って作業 → コミット
3. PR を作成（テンプレートの「概要 / 変更内容 / テスト / 関連 Issue」を埋める）
4. **CodeRabbit のレビューを待ち**、指摘に対応（対応 or 意図を説明）
5. **Code Owner（`@norman6464`）の承認**後に **squash merge**

---

## 3. アーキテクチャ規約（クリーンアーキテクチャ）

依存方向を厳守する。**矢印の向き以外の依存は禁止**。

```
handler  →  usecase  →  repository(port) / infra  →  domain
(Gin)      (Application)   (Persistence / External)    (Entity)
```

- handler は repository / infra を直接呼ばず、必ず usecase を経由する（wiring の `router.go` / `routes_*.go` は例外）
- usecase は `*gin.Context` / `net/http` を参照しない
- domain は他層に依存しない（標準ライブラリ + GORM tag のみ）
- repository は **interface（`usecase/repository/`）** と **実装（`adapter/persistence/`）** を分離
- 1 usecase = 1 ビジネスルール（`struct + New...UseCase + Execute`）。集約系の例外は許容
- 新 endpoint は handler メソッド直前に **swaggo annotation** を書き、`make openapi` で `docs/` を更新

これらは **自作 linter** で CI 検証される（違反すると CI が落ちる）:

| linter | 検証 |
|---|---|
| `archlint` | 依存方向（層をまたぐ禁止 import） |
| `apispec-lint` | Gin ルート ↔ swaggo `@Router`（method + path 一致） |
| `naminglint` | usecase の命名・構造（`XxxUseCase` + `NewXxxUseCase` + `Execute`） |

意図的な例外は `//archlint:allow` / `//apispec:allow` / `//naminglint:allow` で抑制できる。

---

## 4. テスト

新規・変更コードには必ずテストを付ける。

```bash
# バックエンド
cd backend
make verify          # gofmt / vet / build / test / 3 linter を一括
make test-integration  # docker-compose で本物の PostgreSQL に対する結合テスト

# フロントエンド
cd frontend
npm run test:run     # Vitest
npm run e2e          # Playwright 本番スモーク
npm run e2e:local    # ローカルビルド + API モックの認証導線 E2E（要 build）
```

- **単体**: 依存を interface で mock 化（backend usecase / handler、frontend component / hook / repository）
- **結合**: handler（httptest）/ repository（`//go:build integration` で本物の Postgres）
- **E2E**: 本番スモーク + ローカルモック（`/auth/me` のレスポンスで認証状態を制御。本番 Cognito/DB に触れない）
- **カバレッジゲート**: frontend は閾値（lines 85 等）、backend は総計 floor（`COVERAGE_MIN`）を下回ると CI が落ちる。テスト追加に合わせて floor を引き上げる

テスト戦略の詳細は `frestyle-infrastructure/docs/25` / `26`、[トップ README のテスト節](./README.md) を参照。

---

## 5. CI / 品質ゲート

PR では次が走る（詳細は `frestyle-infrastructure/docs/23` / `24`）:

- backend: gofmt / vet / staticcheck / go mod tidy / **race + coverage(floor)** / govulncheck(advisory) / **OpenAPI drift** / **archlint・apispec-lint・naminglint** / build / 結合テスト(Postgres)
- frontend: tsc / ESLint(max-warnings=0) / **Vitest + coverage 閾値** / build
- 全体: **CodeQL**（SAST）/ E2E（Playwright スモーク + ローカルモック）

**ドキュメント更新の無い PR はマージしない**（軽微なタイポ除く）。取り組んだ内容・手順は `docs/` か該当 README に残す。
