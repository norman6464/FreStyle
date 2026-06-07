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
<img src="https://skillicons.dev/icons?i=java,spring&theme=dark" alt="Backend">
</a>

> Java 21 / Spring Boot（`backend-java/`）。レイヤードアーキテクチャ（インフラ境界のみ DIP）。

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

> GitHub Actions / Gradle / JUnit 5 / Spring Boot Test + H2 / CodeQL / Trivy / gitleaks / Vitest coverage / ESLint / tsc

<h3>Testing</h3>
<a href="https://skillicons.dev">
<img src="https://skillicons.dev/icons?i=vitest,playwright&theme=light" alt="Testing">
</a>

> Vitest + React Testing Library（フロントエンド単体）/ JUnit 5 + Spring Boot Test（バックエンド単体・結合）/ Playwright（本番 E2E スモーク、Chromium）
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
- バックエンドはアーキテクチャを言語非依存に保ち、Go と Java/Spring Boot を行き来して比較検証してきた。現在は **Java 21 / Spring Boot** を本番採用し、**レイヤードアーキテクチャ（インフラ境界のみ DIP）** で構成している（`backend-java/`）。lazy-initialization で起動時間を抑え、ECS Fargate 最小スペックで運用
- 夜間は ECS を停止してコストを抑える運用（詳細は[インフラリポジトリ](https://github.com/norman6464/frestyle-infrastructure)の docs を参照）

| 観点 | 数値 |
|---|---|
| ECS Fargate スペック | 0.25 vCPU / 0.5 GB（最小） |
| ランタイム | Java 21（Temurin）/ Spring Boot |
| スキーマ管理 | Flyway（マイグレーション）|
| 永続化 | Spring Data JPA（PostgreSQL）/ AWS SDK v2（DynamoDB・S3・Bedrock・SQS）|

## AWSアーキテクチャ構成図

![FreStyle AWS アーキテクチャ構成図](./architecture/aws/freestyle-aws-architecture-current.png)

draw.io ソース: [`architecture/aws/freestyle-aws-architecture-current.drawio`](./architecture/aws/freestyle-aws-architecture-current.drawio)

---

## 開発に参加する

開発の規約（ブランチ / コミット / アーキテクチャ / テスト / PR フロー）は **[CONTRIBUTING.md](./CONTRIBUTING.md)** にまとめています。PR / Issue はテンプレートに沿って作成してください。

## ローカル開発環境セットアップ

### バックエンド (Java / Spring Boot) — `backend-java/`

```bash
cd backend-java

# 1) ビルド + テスト（Gradle Wrapper が同梱・Java 21 toolchain）
./gradlew build            # コンパイル + 全テスト
./gradlew test             # テストのみ

# 2) ローカル起動（DB 未指定ならインメモリ H2 で起動する）
./gradlew bootRun
# 本番相当の Supabase に繋ぐ場合は環境変数で注入:
#   export SPRING_DATASOURCE_URL='jdbc:postgresql://...pooler.supabase.com:6543/postgres?sslmode=require'
#   ./gradlew bootRun

# 3) 動作確認
curl http://localhost:8080/api/v2/health
# => {"status":"UP"}
```

Docker でビルドする場合:

```bash
cd backend-java
docker build -t frestyle-backend:dev .
docker run --rm -p 8080:8080 frestyle-backend:dev   # 未指定なら H2 で起動
```

> レイヤ構成・コーディング規約・移植状況の詳細は [`backend-java/README.md`](./backend-java/README.md) を参照。

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

バックエンドは **Java 21 / Spring Boot** で、**レイヤードアーキテクチャ（インフラ境界のみ DIP）** を採用する。

- **依存方向**: `controller → service / usecase → repository → entity` の一方向を厳守。外部サービス（S3 / DynamoDB / Bedrock / SQS / Cognito）との境界 = `infra` のみ **interface で DIP** し、本番実装と stub を差し替え可能にする
- **1 usecase = 1 ビジネスルール**: 新規機能は usecase / service を新規作成し、肥大化させない（採点のように関心が跨る処理は専用 usecase に切り出す）
- **request / response 型は controller 内に local 定義**: 専用の Mapper 層は持たず、`record` で受ける。機密フィールドは隠す方針で扱う
- **テスト必須**: 新規追加コードには必ずテスト（JUnit 5）を付ける。TDD（テスト先行）を基本とする

> レイヤ構成・空行ルール・移植状況の詳細は [`backend-java/README.md`](./backend-java/README.md) を参照。

### テスト戦略（単体 / 結合 / E2E の定義）

このアプリでの各テスト種別の定義・対象・ツール・配置は次のとおり。

| 種別 | このアプリでの定義 | 主な対象 | ツール / 実行 |
|---|---|---|---|
| **単体テスト (unit)** | 1 つのクラス・関数を依存から隔離して検証する。外部依存（DB / HTTP / AWS）は interface 経由で mock / fake / stub に差し替え、実 I/O を行わない | backend: 純粋ロジック（採点等）・service / frontend: component・hook・repository | JUnit 5（+ Mockito / AssertJ）/ Vitest |
| **結合テスト (integration)** | 複数層を実際に繋いで 1 プロセス内で検証する。DB は使い捨ての H2、外部 AWS は stub に差し替える | backend: controller（`@SpringBootTest` + `MockMvc` で HTTP〜JSON）・repository（本物の H2 で SQL / 制約）・Flyway 適用 | JUnit 5 / Spring Boot Test |
| **E2E テスト** | 実ブラウザ（Chromium）で本番 URL に対しユーザー導線を端から端まで検証する | デプロイ済み本番のスモーク（SPA ロード / セキュリティヘッダー / 認証境界） | Playwright |

#### バックエンド (Java / Spring Boot) — `./gradlew test`

JUnit 5（Jupiter）+ Spring Boot Test。テストの詳しい技法はアプリ内コース「テスト徹底入門（Java / JUnit 5）」も参照。

- **純粋ロジック（単体）**: 採点（`ExerciseGrading`）など依存ゼロのロジックは `new` して直接検証（最速・コンテキスト不要）。
- **service（単体 / 結合）**: 外部依存は interface で切り、Mockito（`@MockitoBean`）や手書き stub に差し替えてビジネスルールを検証。
- **controller（結合）**: `@SpringBootTest` + `MockMvc` でルータ〜認証〜ステータス〜JSON を通しで検証（`springSecurity()` で JWT 認証も込み）。
- **repository（結合）**: 本物の **H2（`MODE=PostgreSQL`）** に対して実 SQL を流し、マッピング・並び順・集計・NOT NULL / FK 制約まで検証。スキーマは Flyway が所有（`MigrationTest` が適用を見張る）。
- **infra（境界）**: Bedrock / DynamoDB / S3 / SQS / Cognito は `Stub*` 実装に差し替え、AWS 資格情報の無い環境でも起動・テストできる。

```bash
cd backend-java
./gradlew test                                           # 全テスト
./gradlew test --tests "com.normanblog.frestyle.service.*"   # 層を絞って実行
./gradlew build                                          # コンパイル + テスト
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

デプロイ済み本番（既定 `https://normanblog.com`）に対する **未認証スモーク**（SPA ロード / CloudFront セキュリティヘッダー / CSP meta / `GET /api/v2/health` 200 / 認証必須エンドポイントの 401 / 廃止路の 404・401）。

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
| backend (Java) | `./gradlew test`（JUnit 5 + Spring Boot Test + H2 結合 + Flyway 適用検証）。Gradle ベースの CI を移行に合わせて整備中 |
| `ci-frontend.yml` | `tsc` / ESLint / **Vitest + カバレッジ閾値**（lines/statements/functions/branches）/ build |
| `e2e.yml` | Playwright スモーク（本番）+ ローカルモック認証 E2E |
| `codeql.yml` / `trivy.yml` / `secret-scan.yml` | SAST（Java / TypeScript）/ 依存 CVE・Dockerfile 誤設定 / gitleaks シークレットスキャン |

> **カバレッジ閾値ゲート**: frontend は `vitest.config.js` の `coverage.thresholds` を下回ると CI を fail させる。backend（Java）は JaCoCo で計測し、新規追加コードは 80% 以上を目標にする。

---

## チーム開発の進め方（概要）

複数人での開発を前提に、計画 → 実装 → レビュー → リリースの流れを仕組みで揃えている。

- **カンバン**: GitHub Projects で「Todo / In Progress / Done」を管理し、**1 機能 = 1 Issue = 1 PR** を基本とする。
- **Issue 起票**: テンプレート（`task` / **Design Doc**）から起票する。大きめの機能・アーキテクチャ変更は、実装前に [`docs/design/`](./docs/design/) の Design Doc（背景・目的 / スコープ / 設計 / 代替案 / 運用 / テスト）で合意を取る（提案・レビューは Issue、確定設計は `docs/design/<年>/` に蒸留するハイブリッド運用）。
- **ブランチ**: `feat/*` `fix/*` `refactor/*` `docs/*` `test/*` を main から切る。コミット・PR・Issue は日本語。
- **PR**: 本文は `## 概要 / ## 変更内容 / ## テスト / ## 関連 Issue` を基本形式とし、**対応 Issue に必ず紐付ける**（`Closes #NN`）。レビュアーが設計から辿れるようにする。
- **レビュー**: PR ごとに **CodeRabbit（AI レビュー）** と CI が走り、緑になってから **squash merge**。
- **テスト必須**: 新規コードには必ずテストを付ける（バックエンド JUnit 5 / フロントエンド Vitest）。TDD を基本とする。
- **ドキュメント必須**: 取り組んだ内容・手順は `docs/` に残す（ドキュメント更新の無い PR はマージしない）。

詳細な規約は [CONTRIBUTING.md](./CONTRIBUTING.md) を参照。

## セキュリティ / 本番保護

チーム開発での「誤って本番に影響」「シークレット漏洩」を、人的ミスを前提に多層で防ぐ。

- **ブランチ保護**: `main` は PR に**承認 1 件 + CodeQL green が必須**。自分の PR は self-approve 不可（必ずレビューを経る）。force-push 禁止。
- **本番デプロイの承認ゲート**: マージ即本番反映ではない。**frontend は `production` Environment**（required reviewers = `@norman6464`）の承認待ちで停止する。**backend は素の ECS ローリングデプロイ + 循環ブレーカー（失敗時ロールバック）** で更新する。Code シリーズ（CodePipeline + CodeDeploy ECS Blue/Green）は IaC リポで再設計中。
- **シークレット漏洩の多層防御**:
  - GitHub **Push Protection**（既知パターンを含む push をブロック）
  - **gitleaks** CI（`secret-scan.yml`、PR / main / 週次で履歴含めスキャン）
  - **lefthook + gitleaks** の pre-commit（コミット前に手元で検知）
- **SAST / 依存 / 脆弱性**: CodeQL（Java / TypeScript）、**Trivy**（依存 CVE / Dockerfile 誤設定）、**Dependabot**（依存・Actions の自動更新）。

詳細・セットアップ手順は [CONTRIBUTING.md](./CONTRIBUTING.md) §6〜§8 を参照。

## ライセンス

MIT License
