## FreStyle とは

新卒エンジニア向けに作成したプロダクトであり、主に新卒エンジニア向けの研修用のソフトウェアです。
このソフトウェアが解決することは研修用の資料が散在していてさまざまなツールを使用することに慣れていない新卒エンジニアの「探す」という余計な脳のリソースを
割くことなく本来の会社に必要な知識を吸収するのに最適化したプロダクトになっている

## なぜこのプロダクトを作成したのか？
- ツール散乱を解消: 教材・解答・進捗が 1 ヶ所にまとまる
- メンターのナレッジを資産化: 元々スライドの資料などに書き留めていてDriveにアップロードするなどの手間はなくなりこのアプリだけで完結をする
- シングルタスク化: 新卒は「アプリを切り替える」「資料を探す」コストをゼロにし、開発・問題演習・写経・実行を 1 画面で完結

### コア機能（実装済 / 維持）

| 機能 | 概要 |
|---|---|
| **AI チャット** | Bedrock を介した汎用 AI チャット。質問・要約・コードレビュー依頼など自由対話。Markdown レンダリング + ストリーミング表示 |
| **コード学習** | 演習問題を解きながら手を動かしてプログラミングを学ぶ。Monaco Editor + 言語サンドボックス（PHP から拡充予定） |
| **ノート** | 学習ログ・振り返りメモを残せる。画像添付に対応 |
| **レポート** | 月次の学習サマリー |
| **通知** | システム通知（招待・案内など） |
| **プロフィール** | 表示名・アイコン・所属の確認 / 編集 |
| **管理（SuperAdmin / CompanyAdmin）** | 会社一覧 / 招待管理（CompanyAdmin から trainee を招待） |


## デプロイURL

[https://normanblog.com](https://normanblog.com)


## 使用技術

<h3>Frontend</h3>
<a href="https://skillicons.dev">
<img src="https://skillicons.dev/icons?i=react,ts,tailwind,vite&theme=dark" alt="Frontend">
</a>

<h3>Backend</h3>
<a href="https://skillicons.dev">
<img src="https://skillicons.dev/icons?i=go&theme=dark" alt="Backend">
</a>

<h3>Infrastructure</h3>
<a href="https://skillicons.dev">
<img src="https://skillicons.dev/icons?i=aws,cloudflare,docker&theme=light" alt="Infrastructure">
</a>

<h3>Database</h3>
<a href="https://skillicons.dev">
<img src="https://skillicons.dev/icons?i=postgres,dynamodb&theme=light" alt="Database">
</a>

<h3>CI/CD</h3>
<a href="https://skillicons.dev">
<img src="https://skillicons.dev/icons?i=githubactions&theme=light" alt="CI/CD">
</a>

> GitHub Actions / gofmt / go vet / staticcheck / govulncheck / CodeQL / race detector / go test -cover / OpenAPI drift check / Vitest coverage / ESLint / tsc

<h3>Testing</h3>
<a href="https://skillicons.dev">
<img src="https://skillicons.dev/icons?i=vitest,playwright&theme=light" alt="Testing">
</a>

> Vitest + React Testing Library（フロントエンド単体）/ `go test`（バックエンド単体）/ Playwright（本番 E2E スモーク、Chromium）
---

## 工夫した点

### プロダクト面
- 今までの経験で自分が苦に感じたことをベースに何をしたら一番早く新卒のエンジニアが現場に入るまでの必要最低限な知識が身につけられるのかを考えました。
  まず情報設計について調べており、日本人が一覧性のUIを日頃みることが多いらしくスターバックスのUIで日本とアメリカとの違いでありました。
  なのでその性質上さまざまな情報源が散財していることもUIと同様日本人向けではないのではないかと思いこのプロダクト作成に至りました。


### 技術面：
- Cognito発行のJWT を HttpOnly Cookie に保存（XSS 対策）
- アクセストークンの有効期間を短くしリフレッシュトークンで再発行
- OIDC & JWK を活用した堅牢な認証フロー（Google認証をできるようにした）
- NoSQL、RDBのユースケースによって適した使い分け
- 最初はSpring Bootを採用していたがECSタスク定義のスペックを上げないとJVMの起動時にメモリの消費が激しいため料金がかかっていたのでそこをGoに置き換えた

| 観点 | 数値 |
|---|---|
| ECS Fargate スペック | 0.25 vCPU / 0.5 GB（最小） |
| Fargate コスト | ~$0.30/日 |
| 起動時間 | サブ秒 |
| バイナリサイズ | distroless + static binary 約 30 MB |
| 並行処理 | goroutine（軽量） |

## AWSアーキテクチャ構成図

![FreStyle AWS アーキテクチャ構成図](./architecture/aws/freestyle-aws-architecture-current.png)

draw.io ソース: [`architecture/aws/freestyle-aws-architecture-current.drawio`](./architecture/aws/freestyle-aws-architecture-current.drawio)

---

## 開発に参加する

開発の規約（ブランチ / コミット / アーキテクチャ / テスト / PR フロー）は **[CONTRIBUTING.md](./CONTRIBUTING.md)** にまとめています。PR / Issue はテンプレートに沿って作成してください。

## ローカル開発環境セットアップ

### バックエンド (Go) — `backend/`

```bash
cd backend

# 1) 依存解決
go mod tidy

# 2) DB 接続情報を環境変数に
# 通常は Supabase Transaction pooler URL を使う:
export DATABASE_URL='postgresql://postgres.xxxxx:PASSWORD@aws-N-ap-northeast-1.pooler.supabase.com:6543/postgres?sslmode=require'
export PORT=8080
# (ローカル docker postgres で開発する場合は DB_HOST 等を使うフォールバックも config.go に残してある)

# 3) 起動
go run ./cmd/server

# 4) 動作確認
curl http://localhost:8080/
# => {"message":"FreStyle Go backend"}
```

Docker でビルドする場合:

```bash
cd backend
docker build -t frestyle-backend:dev .
docker run --rm -p 8080:8080 \
  -e DATABASE_URL='postgresql://postgres.xxxxx:PASSWORD@aws-N-ap-northeast-1.pooler.supabase.com:6543/postgres?sslmode=require' \
  frestyle-backend:dev
```

### フロントエンド

```bash
# 1. リポジトリをクローンして frontend ディレクトリに移動
cd frontend

# 2. 依存パッケージをインストール
npm install

# 3. 動作確認
npm run dev

# 4. Tailwind CSS が動作しない場合
npm uninstall tailwindcss
npm install -D tailwindcss@バージョン指定
npx tailwindcss init -p
```

---


### コーディング規約（要点）

- **クリーンアーキテクチャ**: handler → usecase → repository / infra → domain（Go / Gin / GORM）の依存方向を厳守
- **1 usecase = 1 ビジネスルール**: 新規機能追加時は usecase struct を新規作成し、肥大化させない
- **request / response 型は handler 内 local 定義**: DTO / Mapper 層は持たず、機密フィールドは domain 構造体側で `json:"-"` で隠す
- **テスト必須**: 新規追加コードには必ず単体テストを付ける

### テスト戦略（単体 / 結合 / E2E の定義）

このアプリでの各テスト種別の定義・対象・ツール・配置は次のとおり。

| 種別 | このアプリでの定義 | 主な対象 | ツール / 実行 |
|---|---|---|---|
| **単体テスト (unit)** | 1 つの関数・構造体を依存から隔離して検証する。依存は interface 経由で mock / fake / stub に差し替え、DB / HTTP / AWS への実 I/O を行わない | backend: usecase・middleware・infra クライアント・自作 linter / frontend: component・hook・repository | `go test` (testify) / Vitest |
| **結合テスト (integration)** | 複数層を実際に繋いで 1 プロセス内で検証する。外部サービスはテスト用実体（sqlite / httptest）に差し替える | backend: handler（router〜JSON）・repository（GORM × sqlite）・linter の走査 run | `go test` |
| **E2E テスト** | 実ブラウザ（Chromium）で本番 URL に対しユーザー導線を端から端まで検証する | デプロイ済み本番のスモーク（SPA ロード / セキュリティヘッダー / 認証境界） | Playwright |

#### バックエンド (Go) — `go test ./...`

標準 `testing` + `github.com/stretchr/testify`。

- **usecase（単体, 22 ファイル）**: 依存する repository / infra を `testify/mock` の `mock.Mock` で差し替え、ビジネスロジックだけを隔離検証する。例: `internal/usecase/*_test.go`。
- **handler（結合, 9 ファイル）**: `gin.SetMode(gin.TestMode)` + `httptest.NewRecorder()` でルータに `ServeHTTP` し、ルーティング〜ステータス〜JSON を通しで検証する。例: `internal/handler/profile_handler_test.go`。
- **middleware（単体, 4 ファイル）**: JWT 認証・CurrentUser 注入・レートリミットなど横断処理を個別に検証する。
- **infra（単体 / 境界, 4 ファイル）**: 外部 I/O を境界で差し替える。Cognito トークン交換・JWKS 検証・oEmbed 取得・SES メールを `httptest` サーバや fake で検証する。例: `internal/infra/cognito/jwt_verifier_test.go`。
- **repository（結合）**: `adapter/persistence` 層を **本物の PostgreSQL**（`docker-compose.integration.yml` の `postgres-integration-test`、本番と同系の 17.x）に対して検証する。`//go:build integration` で隔離し、`make test-integration`（compose 起動 → `go test -tags=integration -run Integration` → teardown）で実行する。通常の `go test ./...`（単体）からは独立。対象例: `company_application`（Create / ListAll 降順 / UpdateStatus）/ `note`（`WHERE user_id` 所有権スコープ・`updated_at DESC`）/ `course`（company 絞り・`is_published` フィルタ・`sort_order` 昇順）の各 `*_repository_integration_test.go`。実 SQL の WHERE / ORDER / 制約を本物の Postgres で検証する。
- **自作 linter（単体 + 結合）**: `cmd/archlint` / `cmd/apispec-lint` / `cmd/naminglint` は、分類・ルール評価の純粋関数（単体）と、一時ディレクトリツリーを走査する `run` / CLI（結合）の両方を検証する（各カバレッジ 87〜89%）。

```bash
cd backend
go test ./...                                   # 全テスト
go test -race -covermode=atomic ./...           # CI と同条件（競合検出 + カバレッジ）
go test ./internal/usecase/...                  # 層を絞って実行
```

#### フロントエンド (Vitest + React Testing Library) — `npm test`

セットアップは `frontend/src/test/setup.ts`。

- **component（単体）**: `render` + `screen.getByRole` でアクセシビリティ込みに描画検証、`fireEvent` でイベント発火。`localStorage` 等は `vi.stubGlobal` でスタブ。
- **hook（単体）**: `renderHook` で状態遷移・副作用を検証。例: `src/hooks/__tests__/*.test.ts`。
- **repository（単体）**: axios を `vi.mock` で差し替え、API 呼び出しの URL / payload / レスポンス整形を検証。例: `src/repositories/__tests__/AuthRepository.test.ts`。

```bash
cd frontend
npm test                 # watch
npm run test:run         # 1 回実行（CI と同じ）
npm run test:run -- --coverage
```

E2E は **2 系統**に分かれる。

**(a) 本番スモーク（外形監視）** — `frontend/e2e/smoke.spec.ts` / 設定 `playwright.config.ts`

デプロイ済み本番（既定 `https://normanblog.com`）に対する **未認証スモーク 6 ケース**: ① SPA ロード + ロゴ / ログイン誘導表示 ② CloudFront セキュリティヘッダー配信 ③ `index.html` の CSP meta ④ `GET /api/v2/health` が 200 ⑤ 認証必須エンドポイントが Cookie 無で 401 ⑥ 廃止済み SockJS フォールバック路が 404 / 401。

**(b) ローカルビルド + API モック（認証付き導線）** — `frontend/e2e/local/*.spec.ts` / 設定 `playwright.local.config.ts`

`vite preview` で配信したビルド済み SPA に対し、Playwright の `page.route` で `/api/v2/**` をモックして認証付きの主要画面を検証する。`GET /auth/me` のレスポンスで認証状態を制御する（401→未認証で `/login` リダイレクト、200+`role`→認証済みで AppShell 描画）。**本番 Cognito / DB に一切触れない**ため CI で毎回安全に回せる。ビルドは `VITE_API_BASE_URL=''`（同一オリジン相対）で行い、`index.html` の CSP `connect-src 'self'` に収めてモックを差し込める状態にするのが要点。

```bash
cd frontend
npm run e2e:install                                       # 初回のみ（Chromium + OS deps）
npm run e2e                                                # (a) 本番スモーク
npm run e2e:local                                         # (b) ローカルビルド + API モック（要 build）
npm run e2e:ui                                             # UI モードでデバッグ
PLAYWRIGHT_BASE_URL=http://localhost:5173 npm run e2e      # (a) をローカル dev server に向ける
```

CI（`e2e.yml`）は **smoke（本番）** と **local-mocked（ローカル + モック）** の 2 ジョブで実行する。

#### CI でのテスト実行

| ワークフロー | 実行内容 |
|---|---|
| `ci-backend-go.yml` | `go test -race` + **カバレッジ閾値ゲート**（総計 floor、回帰防止のラチェット）/ 3 linter（archlint・apispec-lint・naminglint）/ OpenAPI drift |
| `ci-frontend.yml` | `tsc` / ESLint / **Vitest + カバレッジ閾値**（lines/statements/functions/branches）/ build |
| `e2e.yml` | Playwright スモーク（本番）+ ローカルモック認証 E2E |
| `ci-backend-go.yml`（integration ジョブ） | docker-compose の本物 PostgreSQL に対する結合テスト |

> **カバレッジ閾値ゲート**: frontend は `vitest.config.js` の `coverage.thresholds`（lines 85 / statements 85 / functions 80 / branches 78）で、現状（lines 88.6% 等）を下回ると CI を fail させる。backend は総計 floor（`COVERAGE_MIN`、現状 50.2% に対し 49%）。backend は domain / 生成 docs / `cmd/server` / 薄い infra など 0% 群を含む総計のため低めの floor で、テスト追加に合わせてラチェット的に引き上げる（handler テスト追加で 39.2%→50.2%、floor 38→49）。
| `e2e.yml` | Playwright スモーク（Chromium） |

---

## セキュリティ / 本番保護

チーム開発での「誤って本番に影響」「シークレット漏洩」を、人的ミスを前提に多層で防ぐ。

- **ブランチ保護**: `main` は PR に**承認 1 件 + CodeQL green が必須**。自分の PR は self-approve 不可（必ずレビューを経る）。force-push 禁止。
- **本番デプロイの承認ゲート**: CD の本番反映ジョブは `production` Environment（required reviewers = `@norman6464`）に紐付き、**起動しても承認待ちで停止**する。マージ即本番反映ではなく、手動 `confirm=deploy` + 承認の二段。
- **シークレット漏洩の多層防御**:
  - GitHub **Push Protection**（既知パターンを含む push をブロック）
  - **gitleaks** CI（`secret-scan.yml`、PR / main / 週次で履歴含めスキャン）
  - **lefthook + gitleaks** の pre-commit（コミット前に手元で検知）
- **SAST / 依存**: CodeQL（Go / TypeScript）、govulncheck（advisory）。

詳細・セットアップ手順は [CONTRIBUTING.md](./CONTRIBUTING.md) §6〜§8 を参照。

## ライセンス

MIT License
