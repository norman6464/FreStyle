## FreStyle とは

新卒エンジニア向けに作成したプロダクとであり、主に研修用のソフトウェアです。
このソフトウェアは研修用の資料が散在していてさまざまなツールを使用することに慣れていない新卒エンジニアの「探す」という余計な脳のリソースを
割くことなく本来の会社に必要な知識を吸収するのに最適化したプロダクトになっている

## なぜ作るのか（プロダクトの意義）
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
  なのでその性質上さまざまな情報源が散財していること自体もUIと同様日本人向けではないのではないかと思いこのプロダクト作成に至りました。


### 技術面：
- JWT を HttpOnly Cookie に保存（XSS 対策）
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

### テスト

#### バックエンド (Go)

```bash
cd backend
go vet ./...
go test ./...
```

- 標準 `testing` パッケージ + `github.com/stretchr/testify`（順次導入）
- usecase: 依存をインターフェイスとしてモック化した単体テスト
- repository: テスト用 PostgreSQL コンテナまたは sqlite による統合テスト
- handler: `httptest.NewRecorder` + `gin.New()` でルータごと検証

#### フロントエンド

```bash
cd frontend
npm test
```

- Vitest + React Testing Library
- 慣習: `vi.stubGlobal('localStorage', createMockStorage())` で localStorage をスタブ、`fireEvent` でイベント発火

#### E2E (Playwright)

```bash
cd frontend
npm run e2e:install   # 初回のみ（Chromium + OS deps）
npm run e2e            # 本番に対してスモーク 6 ケース
npm run e2e:ui         # UI モードでデバッグ
PLAYWRIGHT_BASE_URL=http://localhost:5173 npm run e2e   # ローカル dev server 経由
```

---

## ライセンス

MIT License
