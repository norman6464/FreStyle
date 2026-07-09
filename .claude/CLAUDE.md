# FreStyle — Claude Code プロジェクト規約

> このファイルは **git 管理されるチーム共通の規約**です（`.claude/settings.json` と同様にコミットして共有する）。
>
> 本リポジトリは **PUBLIC** のため、**秘匿情報（DB 接続情報 / AWS Secrets Manager のシークレット名 / 内部運用の詳細）はここに書かない**。それらは private リポ `frestyle-infrastructure` の docs・AWS Secrets Manager・各自の `.env`（gitignore 済）で管理する。個人ローカルのメモはリポ直下の `CLAUDE.md`（`/CLAUDE.md` として gitignore 済）に置ける。
>
> **コミット対象のアーキテクチャ仕様**の一次情報は [`docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md)。このファイルのルールは必ず `docs/ARCHITECTURE.md` と整合性を保つこと。

日本語で会話をしてください。

---

## 1. プロジェクト基本情報

- **プロジェクト名**: FreStyle — 新卒 IT エンジニア向け統合研修プラットフォーム（B2B SaaS）
- **本番URL**: https://normanblog.com
- **バックエンド**: Go 1.x / Gin / GORM（`backend/` 配下、ECS Fargate でコンテナ実行）
- **フロントエンド**: React 19 / TypeScript / Vite / Tailwind CSS（`frontend/` 配下）
- **RDB**: PostgreSQL 17.6（Supabase / Transaction pooler / GORM AutoMigrate で運用、2026-05 に RDS から移行）
- **NoSQL**: DynamoDB（`fre_style_ai_chat` — AI チャットメッセージ）
- **認証**: AWS Cognito（OIDC + JWT HttpOnly Cookie、招待マジックリンク方式）
- **AI**: AWS Bedrock（Claude Sonnet 4.5 / Inference Profile 経由）
- **非同期処理**: SQS（学習レポート生成キュー）
- **チャット通信**: SSE on ECS（旧 API Gateway + Lambda + WebSocket は撤去済）
- **IaC**: Terraform（private リポ `frestyle-infrastructure` で管理。CloudFormation は 2026-07 に完全撤去）

---

## 2. クリーンアーキテクチャ規約（最重要）

本プロジェクトは Spring Boot → Go 移行（2026-04 完了）後も **クリーンアーキテクチャ** を維持しています。以降のすべてのコード追加・変更は、以下のルールに**必ず**従ってください。

### 2.1 依存方向ルール

```
handler  →  usecase  →  repository / infra  →  domain
(Gin)      (Application)  (Persistence / External) (Entity)
```

- **矢印の向き以外の依存は禁止**
- handler は repository / infra を **直接** 呼び出してはならない。必ず usecase を経由する
- usecase は handler のことを知ってはならない（`*gin.Context` などを引数で受けない）
- repository / infra は usecase のことを知ってはならない
- domain は他のどの層にも依存しない（標準ライブラリ + GORM tag のみ許容）

### 2.2 各層の責務

| 層 | パッケージ | 責務 | 禁止事項 |
|---|---|---|---|
| **handler** | `backend/internal/handler` | Gin で HTTP / SSE リクエストを受け付け、middleware から認証情報を取得し、usecase を呼び出して JSON 返却 | ビジネスロジックの記述、repository / infra の直接呼び出し |
| **middleware** | `backend/internal/handler/middleware` | JWT 認証、CORS、current user 注入、CSRF など横断的処理 | ビジネスロジック、DB アクセス |
| **usecase** | `backend/internal/usecase` | 1 ユースケース = 1 構造体（単一責任）。複数の repository / infra をオーケストレーションしてビジネスロジックを実行 | HTTP 層の型（`*gin.Context` など）への依存、Gin / `net/http` の参照 |
| **repository (port)** | `backend/internal/usecase/repository` | usecase が依存する repository interface の定義 | 実装の記述 |
| **persistence (adapter)** | `backend/internal/adapter/persistence` | RDB（GORM）/ DynamoDB / S3 / SQS 等へアクセスする repository 実装 | ビジネスロジック、外部サービスのオーケストレーション |
| **infra** | `backend/internal/infra/{bedrock,s3,ses,cognito,database,...}` | 外部サービス連携（AWS SDK ラッパ）、DB 接続・マイグレーション、設定読み込み | usecase / handler の参照、ビジネスルール判定 |
| **domain** | `backend/internal/domain` | エンティティ構造体 + ビジネスルール上の定数（role / status enum 等）。GORM tag は **pragmatic に** 直書きする | repository / infra / usecase / handler への依存、副作用関数 |

### 2.3 1 構造体 1 責務（usecase）

- usecase 1 つに対してビジネスルール 1 つ。複数操作をまとめない。
  - ❌ `AiChatUseCase.CreateSessionAndSendFirstMessage`
  - ⭕ `CreateAiChatSessionUseCase` + `SendAiMessageStreamUseCase` を handler 側でオーケストレーション
- usecase は **struct + コンストラクタ + `Execute` メソッド** で書く:

  ```go
  type CreateAiChatSessionUseCase struct {
      sessions repository.AiChatSessionRepository
  }

  func NewCreateAiChatSessionUseCase(r repository.AiChatSessionRepository) *CreateAiChatSessionUseCase {
      return &CreateAiChatSessionUseCase{sessions: r}
  }

  func (u *CreateAiChatSessionUseCase) Execute(ctx context.Context, in CreateAiChatSessionInput) (*domain.AiChatSession, error) {
      // ...
  }
  ```

### 2.4 domain と request / response 型の境界

- handler は **domain 構造体をそのまま JSON で返して良い**（GORM tag + JSON tag を直書きする方針のため、DTO 詰め替えは不要）。
- ただし「クライアント向けに値を加工する / 機密フィールドを隠す」必要があれば、handler 内で response 構造体を定義して詰め替える。
- リクエスト用の入力は handler のファイル内で `xxxRequest` struct を定義し、`c.ShouldBindJSON(&body)` でバリデーション。`binding:"required"` などの Gin タグで宣言的に書く。
- usecase の入力は `XxxInput` struct（domain 型を含んで OK）、戻り値は `*domain.Xxx` または primitive。
- 機密フィールド（パスワード hash、招待 token、attachment の BlobData など）は domain 構造体側で `json:"-"` を付けてシリアライズから除外する。

### 2.5 フロントエンドのレイヤー

フロントエンドにも類似の層構造を適用します。

```
Page  →  Hook  →  Repository  →  HTTP クライアント
 (UI)     (Application)  (Infra)       (axios)
```

- **Page**（`src/pages/`）: 画面コンポーネント。ビジネスロジックは書かない
- **Hook**（`src/hooks/`）: 画面固有の状態管理・API 呼び出しを Hook にまとめる
- **Repository**（`src/repositories/`）: HTTP API 呼び出しのラップ。axios の直接利用はここに集約
- **Component**（`src/components/`）: プレゼンテーショナル。副作用を持たない
- **Store**（`src/store/`）: Redux Toolkit slice。グローバル状態のみ（auth, flash message など）

### 2.6 Repository 配置規約（Clean Architecture 物理分離）

新規 / 既存を問わず、backend の repository は以下の 2 階層構成で配置すること。Use Case 層 (port) ↔ Interface Adapters 層 (adapter) を Go のディレクトリにそのまま写したもの。

| 役割 | 配置 | パッケージ名 |
|---|---|---|
| Repository **interface**（Use Case が依存する port） | `backend/internal/usecase/repository/` | `repository` |
| Repository **実装**（GORM / DynamoDB / S3 / SQS 等） | `backend/internal/adapter/persistence/` | `persistence` |

- 依存方向は **usecase ← persistence**（DIP）。usecase 側は interface だけを import する
- wiring（`internal/handler/router.go`）は `persistence.NewXxxRepository(...)` で実装を組み立て、各 usecase に注入する
- **1 boundary = 1 fat interface** で良い。単一メソッド / 単一責務の port（例: `Presigner` / `Enqueuer` / `Publisher`）は Effective Go 流の `-er` 命名で別 interface に切り出して良い
- interface 名は DDD 用語そのままの `XxxRepository` で OK
- 新規 repository を追加する場合は **必ず** `usecase/repository/{entity}.go`（interface）+ `adapter/persistence/{entity}_repository.go`（実装）の 2 ファイル構成で書く
- 旧構造（1 ファイルに interface + 実装同居の `internal/legacyrepository/`）は 2026-05 に完全撤去済。この流儀に戻さない

### 2.7 OpenAPI annotation 規約（swaggo）

新しい HTTP endpoint を追加する際は、handler メソッドの直前に **swaggo annotation** を必ず書くこと。これにより `make openapi` で OpenAPI spec（`backend/docs/swagger.{json,yaml}`）と Swagger UI（`/swagger/index.html`）が自動生成 / 更新され、frontend は `openapi-typescript` で型を自動取得できる。

必須 annotation:

```go
// MethodName は ... ハンドラの役割を 1 行で。
//
//	@Summary      短い概要 (1 行)
//	@Description  詳細説明 (複数行可)
//	@Tags         tag-name      // (health / auth / profile / notes 等)
//	@Accept       json          // (リクエスト body があるときのみ)
//	@Produce      json
//	@Param        body  body    requestType  true  "説明"
//	@Param        userId  path  string  true  "数字 ID または 'me'"
//	@Success      200  {object}  responseType
//	@Failure      401  {object}  errorResponse  "未認証"
//	@Failure      500  {object}  errorResponse  "内部エラー"
//	@Router       /path/here [post]
//	@Security     CookieAuth    // 認証必須 endpoint のみ
func (h *Handler) MethodName(c *gin.Context) { ... }
```

ルール:

- `errorResponse` / `messageResponse` 等の**共通 response 型**は `backend/internal/handler/openapi_types.go` に定義済。新規 response 型はここか handler ファイルに local 定義
- annotation は **handler メソッドの直前**に書く（struct の上ではない）
- `@Router` の path は `/api/v2` を**含めない**（main.go の `@BasePath /api/v2` で自動 prefix）
- 認証必須 endpoint には **`@Security CookieAuth`** を付ける
- handler を追加 / 変更する PR で**必ず `make openapi`** を走らせて `docs/` の差分も commit する
- SSE / WebSocket / multipart-upload は OpenAPI で完全表現できないので、spec ではシンプルな POST として表現し詳細は `@Description` で補足する

---

## 3. コーディング規約

### 3.1 言語

- **日本語**: PR タイトル / チケット / コミットメッセージ / コメント
- **英語**: 識別子（クラス名・変数名・関数名）

### 3.1.1 他社プロダクト名の不使用（重要）

PR タイトル / 本文 / チケット / コミットメッセージ / コード内コメント / docs に**他社プロダクト・他社サービスの名前を書かない**（「〜風 UI」「〜のようなレイアウト」といった比喩を含む）。

- ❌ 「（他社サービス名）風 2 カラムレイアウト」
- ❌ 「（他社サービス名）のような UI に寄せる」
- ⭕ 「2 カラムレイアウト（本文 + 目次サイドバー）」
- ⭕ 「目次サイドバー付きの記事レイアウト」

理由: 他社プロダクト名を持ち込むと「真似た」ような表現になる / 商標やイメージを巻き込むリスクがある。機能の中身（何があるか）で説明する。

### 3.2 命名規則

- **usecase struct**: `[動詞][目的語]UseCase` 例: `CreateAiChatSessionUseCase`
  - コンストラクタは `NewXxxUseCase`、メソッドは `Execute(ctx, in) (out, error)`
- **repository interface**: `[Entity]Repository` 例: `UserRepository` / `AiChatSessionRepository`
  - 実装は `persistence` パッケージの小文字 struct + `NewXxxRepository(...)` で公開
- **handler struct**: `[ドメイン]Handler` 例: `AiChatHandler` / `MasterExerciseHandler`
  - メソッドは `(h *XxxHandler) Action(c *gin.Context)`
- **domain 構造体**: `[名詞]` 例: `User` / `AiChatSession` / `MasterExercise`
  - GORM tag + JSON tag を直書き、`TableName()` メソッドを必要に応じて定義
- **request / response 構造体**: handler 内に local 定義。`xxxRequest` / `xxxResponse`
- **infra クライアント**: パッケージ名は領域別（`bedrock` / `s3` / `ses` / `cognito` / `database`）
- **フロントエンド コンポーネント**: PascalCase、1 ファイル 1 コンポーネント

### 3.3 テスト

- **TDD**（テスト駆動開発）を基本とする
- バックエンド: **`testing` + `github.com/stretchr/testify`**（`go test ./...`）
  - usecase: 依存を interface でモック化した単体テスト（testify/mock）
  - repository: `*gorm.DB` を sqlite メモリで差し替えた統合テスト
  - handler: `httptest.NewRecorder` + `gin.New()` でルータごと検証（`gin.SetMode(gin.TestMode)`）
  - infra: 外部サービス呼び出しはインターフェイス境界で fake / stub 注入
- フロントエンド: **Vitest + React Testing Library**（`npm test`）
  - コンポーネントは `render` + `screen.getByRole` でアクセシビリティも検証
  - Hook は `renderHook` で状態遷移テスト
- **カバレッジ目標**: 新規追加コードは **80% 以上**

### 3.4 コメント

- コメントは必要最小限。コード自身が「何をするか」を語るように書く
- WHY（なぜこう書いたか）をコメントに残す。WHAT は書かない
- 各 exported な struct / func の先頭には「責務」を **GoDoc コメント**で 1〜2 行書く（`// XxxHandler は ... を扱う。` のように識別子から始める）
- 命名衝突を避けるためのプライベート小道具には `// ...` で軽く意図を残す

### 3.5 禁止事項

- ❌ handler から repository / infra への直接呼び出し
- ❌ usecase をまたぐビジネスロジックを 1 つの usecase に詰め込む
- ❌ usecase / repository / infra から `*gin.Context` や `net/http` の型を参照する
- ❌ `db.Begin()` / `Tx` を handler に書く（**GORM トランザクションは usecase レベル**で開始する）
- ❌ `main` ブランチへの直接コミット（ブランチ保護設定済み）
- ❌ テストのない状態でのマージ
- ❌ domain 構造体に repository / infra への依存を持ち込む（domain は他層を import しない）

---

## 4. PR / チケット / マージフロー

### 4.1 ブランチ運用

1. **Jira チケットを起票**（または既存チケットを選ぶ）→ **自分にアサイン**してから着手する（詳細は §4.6。作業管理は GitHub Issue ではなく Jira）
2. ブランチを切る（`feat/*`, `fix/*`, `refactor/*`, `docs/*`, `test/*`）
3. 作業 → コミット（コミットメッセージは日本語）
4. PR 作成（タイトル・本文とも日本語）
5. **CodeRabbit によるコードレビューを待つ**
6. CodeRabbit 指摘に対応
7. **squash merge**（直接マージ禁止、マージは必ず PR 経由）

### 4.2 CodeRabbit レビュー待機のルール

- PR 作成後、**CodeRabbit の初回レビューが投稿されるまで待機**
- レビューの「Actionable comments」に対しては、原則**すべて応答**（対応 or 意図を説明して reject）
- CodeRabbit が指摘していないセキュリティ・アーキテクチャ違反があれば、セルフレビューで修正
- CodeRabbit からの summary コメントは PR 本文に貼り付けない（ノイズになるため）

#### CodeRabbit を待つときは必ず `sleep` コマンドを打つ（絶対ルール）

- Claude Code が CodeRabbit のレビューを待つときは、**`bash sleep <秒数>` を実行して実時間を確実に経過させる**こと
- 「あとで確認します」やバックグラウンド化への逃げは禁止。ユーザーから見て「ちゃんと同期的に待っている」状態を保つ
- その場で `sleep 270`（約 4.5 分）または `sleep 300〜600`（数分〜10 分）で実際に待つ
- 例外: ユーザーが「待たずに進めて」と明示している場合のみ skip 可
- ランタイムが長い `sleep` をブロックする場合は、`timeout` を長めに指定した待機ループ（`end=$(($(date +%s)+270)); while [ $(date +%s) -lt $end ]; do sleep 30; done`）で同期的に待つ

#### CodeRabbit レートリミット時の特例（admin merge 許可）

CodeRabbit が `Rate limit exceeded` を返してレビュー不能になった場合、以下の条件をすべて満たすときは **admin merge で先に進めて良い**:

1. PR 作成者またはユーザーが「レートリミット中なので admin merge で進めて良い」と明示的に承認している
2. PR の差分をセルフレビュー済（Critical / Major レベルの問題なし）
3. 該当 PR で Status checks（CI / Tests / Lint）が成功している
4. 破壊的変更ではない（DB schema breaking 変更 / 認証フロー全断など）

該当する場合のコマンド:

```bash
gh pr merge <PR#> -R norman6464/FreStyle --squash --admin --delete-branch
```

admin merge した PR はマージ後コメントで「CodeRabbit レートリミット中のため、セルフレビュー済で admin merge」と記録する（監査目的）。

#### デザイン専用変更の特例（admin merge 許可）

見た目（カラー / レイアウト / typo / 文言 / ロゴ等）のみを変更する PR は、**CodeRabbit レビューを待たずに admin merge で進めて良い**。

適用条件（すべて満たすこと）:

1. 変更が `frontend/src/index.css` / Tailwind class / コンポーネントの className / public/*.svg / 文言など、**ロジックに影響しない見た目だけ**であること
2. ロジック / API 契約 / DB スキーマ / 認証フロー / セキュリティ設定への変更を含まないこと
3. `tsc --noEmit` / `vitest run` / `eslint --max-warnings 0` / `npm run build` がすべて成功していること

ロジックを伴う変更（例: 新しい hook / API 呼び出しの追加 / 状態管理の修正）はデザイン PR でも分割し、ロジック側は通常通り CodeRabbit レビューを待つ。

### 4.3 コミットメッセージ

- Prefix: `feat` / `fix` / `refactor` / `docs` / `test` / `chore` / `perf` / `style`
- 例: `feat: 初心者向けUIコンポーネント群を追加`
- 末尾に `Co-Authored-By: <使用した Claude モデル名> <noreply@anthropic.com>` を含める（「1M context」などの修飾語は付けない）
- **`Generated with Claude Code` という文言は入れない**（既存運用に合わせる）

### 4.4 PR 本文

- 「Generated with Claude Code」の文言は含めない
- `## 概要` / `## 変更内容` / `## テスト` / `## 関連チケット` の 4 セクションを基本形式とする
- **PR タイトル・本文は必ず日本語**（技術用語・識別子・パス・conventional-commit の prefix は英語のまま可）
- テンプレートは `.github/PULL_REQUEST_TEMPLATE.md`（日本語ベース）。新規 PR はこの構成に従う

### 4.5 シングルタスク運用（最重要・PR を溜めない）

**PR は必ず 1 つずつ進める。複数の PR を開いたまま放置しない。**

- ❌ 複数の機能ブランチを並行で起こし、PR を 2 本以上 open のまま積み増す
- ⭕ **1 PR を「作成 → CodeRabbit/CI 対応 → マージ」まで完了させてから、次の PR に着手する**
- 理由（実際に起きた事故）: 4 本の PR を同時に open にした結果、`docs/swagger.*` 等の生成ファイルが互いにコンフリクトし、マージ順・rebase・再生成の管理コストが跳ね上がった
- 大きな作業は「1 機能 = 1 PR」で**直列**に分割する。前の PR がマージされて `main` が最新になってから次のブランチを `main` から切る
- どうしても並行が避けられない依存関係がある場合のみ、ユーザーに理由を伝えて明示承認を得る
- Claude Code は「次の PR を作る前に、今 open の PR が残っていないか」を必ず確認する

### 4.6 Jira チケット運用（作業管理は GitHub Issue ではなく Jira）

作業の起票・追跡は **Jira** に一本化する（GitHub Issue は使わない）。

- **プロジェクト**: `https://frestyle.atlassian.net`（例: `https://frestyle.atlassian.net/browse/FRESTYLE-3`）
- **MCP 経由で操作する**: Atlassian Rovo コネクタ（`createJiraIssue` / `editJiraIssue` / `addCommentToJiraIssue` / `transitionJiraIssue` 等）を使う

#### 着手フロー（必須）

1. **Jira チケットを作成**（新規）または**既存チケットを選ぶ**
2. **必ず自分（報告者と同じ担当者）にアサイン**してから作業に取り掛かる（「未アサインのまま作業」は禁止）
3. ブランチを切って実装 → PR（§4.1）。PR タイトル / コミットにチケット番号を紐付ける
4. マージ後、チケットを完了状態に遷移する

#### チケット作成時のルール

- **開発タイプ（課題タイプ）**を内容に応じて選ぶ: `hotfix` / `開発タスク` / `リファクタリング` / `ドキュメント整備` / `Design Doc` / `エピック`
- **description はプロジェクトのテンプレートに沿う**（概要 / ゴール・受け入れ条件 / スコープ外 / 背景・目的 / テスト・検証 / ドキュメント / セキュリティ影響 / 影響範囲 / ロールバック方針 / 参考リンク）。バグ系は再現手順入りのバグ用テンプレを使う
- **報告者は自分**。**開始日・終了日（スケジュール）は原則未記入**
- スキーマ変更や新機能追加など設計判断を伴うものは **`Design Doc` 種別**で起票し、背景・選択肢・推奨案・**承認記録**を残す
- **実在確認した事実のみを書く**（ファイル・挙動・PR のマージ状態は、書く前に必ずリポジトリと突き合わせて検証する。検証できないことは書かないか「未確認」と明記する）
- **参考リンクは git 管理された文書（リポジトリ内の docs / 各リポの README 等）を指す**。git 管理されていないローカルファイルへの参照は書かない

#### コメント欄に書くこと（最重要）

作業の検証は **Jira チケットのコメント欄**に記録する。主に次を書く:

- **ステージング検証手順**: ステージング環境でどう検証したか（叩いた画面 / エンドポイント / 操作手順）
- **その実行結果**: 実際の出力・レスポンス・スクリーンショット等（事実）
- **期待値の突き合わせ**: 「期待する値」と「実際の値」を並べ、**期待どおりか（合否）を明記**する。違っていれば原因と次アクションも残す

> 目的: チケットページだけ見れば「何を・どう検証し・期待どおりだったか」が追える状態にする。実装の証跡を PR ではなく **Jira チケットのコメントに集約**する。

#### PR との相互リンク（必須）

- **Jira の description / コメントには、関連する PR の URL もしくは PR 番号を必ず書く**。チケット ↔ PR を**双方向に辿れる**ようにする
- 逆方向（PR → チケット）は **PR タイトル / コミットにチケット番号**（例: `FRESTYLE-55`）を入れて担保する（§4.1）
- 着手時（PR 作成）・マージ時のコメントに、その都度**対象 PR を明記**する（複数 PR に分かれた場合は全部書く）

---

## 5. デプロイ

### 5.1 バックエンド

- GitHub Actions の **`CD - Backend Deploy to ECS`** ワークフロー（`workflow_dispatch` 手動 + `confirm=deploy` 入力で起動）が ECR への build/push と ECS Service の force-update を行う
- 起動コマンド: `gh workflow run "CD - Backend Deploy to ECS" -R norman6464/FreStyle -f confirm=deploy`
- ヘルスチェック: `GET /api/v2/health`（本番の API ドメインに対して確認する。CloudFront 配下の SPA パスに対して叩くと一律 200 になり誤認するため）

### 5.2 フロントエンド

- GitHub Actions **`CD - Frontend Deploy to S3 + CloudFront`** が S3 アップロード + CloudFront キャッシュ無効化（`confirm=deploy` 入力必須）

### 5.3 DB マイグレーション

- AutoMigrate（GORM）で**新規テーブル / 列追加**は ECS 起動時に自動適用される
- **列削除 / リネーム / 型変更**は GORM AutoMigrate が処理しないので、明示的な SQL を `backend/migrations/000X_*.sql` に置き、private リポ `frestyle-infrastructure` の `make apply-migration-supabase` 経由で Supabase に適用する（シークレット名などの実引数・詳細手順は同リポの docs 参照）
- 冪等性（`IF NOT EXISTS` / `DO $$ ... $$` の存在チェック）を必ず担保
- 適用手順は migration ごとに記録を残す（private リポ側の docs）

---

## 6. ローカル開発環境

### 6.1 環境変数

```bash
cp .env.example .env
# .env に Cognito / DB / Bedrock のローカル接続情報を記入
```

主な環境変数:
- `DATABASE_URL` — PostgreSQL の完全接続文字列（推奨）。セットされていると `DB_HOST` 等より優先
- `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` — PostgreSQL 個別接続（`DATABASE_URL` 未設定時のフォールバック）
- `AWS_REGION` / `BEDROCK_MODEL_ID`
- `DYNAMODB_AI_CHAT_TABLE`（dev は `fre_style_ai_chat_dev`）
- `NOTE_IMAGES_BUCKET` — note 画像 / AI チャット添付の S3 prefix を共有

**接続情報の実値は git に commit しない。** `.env`（gitignore 済）または AWS Secrets Manager に保存する。Supabase の接続 URL の取得方法・pooler の使い分け・migration runbook は private リポ `frestyle-infrastructure` の docs を参照。

backend は `DATABASE_URL` がセットされていると、pgbouncer 互換のため `PrepareStmt: false` + simple query protocol を自動適用する（host 名で pgbouncer を自動判定）。GORM AutoMigrate もそのまま動く。

### 6.2 起動手順

```bash
# バックエンド (Go)
cd backend
go mod download
go run ./cmd/server         # 開発時は直接実行
# 本番ビルドの確認: go build ./... && go test ./...

# フロントエンド
cd frontend
npm install
npm run dev
```

---

## 7. ドキュメンテーション（最重要ルール）

### 7.1 ドキュメント無しの変更は禁止

**取り組んだもの・実装した内容・手順は必ず `docs/` に書く。**

- 新しい機能・画面を追加した → `docs/` のどれかに必ず使い方 / 設計判断を書く
- 新しい usecase / repository / infra / handler / Component を追加した → `docs/ARCHITECTURE.md` の依存関係図とチェックリストを更新
- 失敗から学んだ知見 → `docs/` のトラブルシューティング系 docs に追記
- 新しい運用フロー（CI/CD, デプロイ手順）を作った → `.github/workflows/README.md` または専用 docs を更新
- 既存仕様を変えた → 該当 docs を更新

**ドキュメント更新が無い PR はマージ禁止**（軽微なタイポ修正など明らかに不要な場合を除く）。

### 7.2 docs を書くタイミング

実装が終わってから書くのではなく、**実装と同じ PR 内で書く**。

- ❌ 実装 PR → マージ → 後日 docs PR
- ✅ 実装 + docs を同じ PR に
- ✅ 仕様検討段階で docs を先に書いてから実装（Documentation-First）も歓迎

### 7.3 取り組み履歴の記録

PR で「やったこと」を docs に書くのは将来の自分・他のエンジニアへの引き継ぎ。「なぜやったか」「どう動かすか」「何が変だったか」を残すことで、半年後に同じ作業を再現できる状態を保つ。

---

## 7-bis. 教材コンテンツの管理（重要）

教材本文（Markdown）は**アプリ本体のリポジトリには置かない**。専用のプライベートリポジトリで管理し、そこから本番 DB に同期する。

### リポジトリ

**`norman6464/frestyle-teaching-materials`**（private）

ローカルには FreStyle 直下に clone する運用（`.gitignore` で除外済）。

### ディレクトリ構成（このルールを守る）

```
frestyle-teaching-materials/
└── courses/
    ├── NN-{course-slug}/                # 例: 02-git, 03-docker, 04-linux-cli
    │   ├── course.yaml                  # title / description / category / sort_order / is_published
    │   ├── 001-{material-slug}.md       # 教材 1
    │   ├── 002-{material-slug}.md       # 教材 2
    │   └── ...
    └── ...
```

- **`NN-`** は数字 prefix。コース・教材ともに並び順を git 上で確定するため必須
- **`course.yaml`** はコースのメタ情報（title / description / category / sort_order / is_published）。**コース ID は含まれない**
- **教材は `# タイトル` から始める**（先頭 h1 が DB の `title` カラムに入る）

### 同期フロー（教材 → 本番 DB）

1. 教材リポジトリで Markdown を編集 → `git push`
2. `courses/_scripts/seed-courses.py` で章ごとに分割した SQL を生成する
3. private リポ `frestyle-infrastructure` の `make apply-migration-supabase`（または `make seed-course`）で章ごとに順次適用する（実引数・詳細手順は同リポの docs 参照）
4. `https://normanblog.com/courses` で UI 反映確認

#### 重要な落とし穴 1: `course.sql` は既存教材を**全削除**する

`seed-courses.py` が出力する `{course-slug}__course.sql` は course 行を UPSERT したあと、その course の `teaching_materials` を **`DELETE` してから**章 SQL を入れ直す冪等設計になっている。そのため:

- **1 つだけ章を追加**する場合でも、そのコースの `course.sql` + 章 SQL **全件**を再 apply する必要がある
- `course.sql` だけ流して章 SQL を流し忘れると DB から**全章が消えた状態**になる
- 適用後は必ず本番 UI（`https://normanblog.com/courses/{id}`）で章数を確認する

#### 重要な落とし穴 2: 未知の sort_order は**エラーにならず黙ってスキップされる**

`seed-courses.py` の `SORT_TO_ID` マップ（sort_order → 本番コース ID の手書き対応表）に無い sort_order のコースは、エラーで止まるのではなく `skip {course_dir} (unknown sort_order N)` の警告を出して**スキップされ、SQL が生成されないまま正常終了する**（終了コード 0）。新規コース追加時に `SORT_TO_ID` への追記を忘れると、seed されないことに気づきにくい。生成後は出力 SQL のファイル数を必ず確認する。

### 教材の作成 / 更新ルール

- 新規コース: `courses/NN-{slug}/` を作成して `course.yaml` + `001-*.md` から書く。`SORT_TO_ID` への id 追記も同時に行う
- 既存コースに教材追加: `courses/NN-{course}/00X-{slug}.md` を追加（番号は連番）
- 編集と DB 同期は**同じセッション内で完結**させる（リポジトリだけ push して DB に流し忘れる事故を防ぐ）
- **最新のコース一覧・id 対応は、教材リポの `courses/` ディレクトリと `seed-courses.py` の `SORT_TO_ID` を一次情報として参照する**（このファイルに一覧を複製しない — 陳腐化して誤情報の元になるため）

### コードサンプルの掲載ルール（重要）

- **教材内で FreStyle 自身のソースコードを引用するときは、該当ファイルを「全文」掲載する**
- 「`// ... 他メソッド`」「`// 略`」「`...`」のような**省略 / 切り詰めをしてはいけない**
  - 学習者は教材だけで完結して読みたいので、「リポジトリを開いて続きを読んでください」では不親切
  - 半年後に同じ教材を読む人にとっても、当時のスナップショットが残っている方が価値が高い
- ファイルが 200 行を超えるなど本当に長い場合は、関数単位で分割して**複数の code block** に分け、各ブロック内は省略しない
- 「いま見ているファイルパス」を**必ずコードブロック直前に明示**する
- ソースが更新されたら**教材側も同じ PR / 同じセッション内でバックポート**する（古い教材が残ると新人が混乱する）

---

## 8. Claude Code への指示

Claude Code がこのプロジェクトで作業するときは、以下を**必ず**遵守してください。

1. 新しいユースケースを追加するときは、**必ず usecase struct を新規作成**する（既存の usecase / repository に詰め込まない）
2. 新しい画面を追加するときは、**`src/components/ui/` の初心者向けコンポーネント**を最大限活用する
3. ビジネスロジックを handler に書かない（handler は middleware から user を取り、usecase に丸投げするだけ）
4. **テストなしでマージしない**（新規コードには `go test` / `vitest` の単体テストを必ず付ける）
5. PR を作ったら **CodeRabbit レビューを待つ**。自動マージしない。ただしセッション内で一度でもレートリミットに達した場合は、以降そのセッションではセルフレビューの上でマージまで進めて良い（§4.2 の特例条件に従う）
6. `main` ブランチへ直接コミット・push しない
7. すべてのコミットメッセージ・PR・チケットは**日本語**で書く
8. リクエスト / レスポンス用の専用型（`xxxRequest` / `xxxResponse`）は handler のファイル内で local 定義する。**機密フィールドは domain 構造体側で `json:"-"` で隠す**こと
9. 詳細な層設計・パッケージ依存関係は [`docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md) を一次情報として参照する
10. **取り組んだもの・実装した内容・手順は必ず `docs/` に書く**（ドキュメント更新無しの PR はマージしない）
11. **チケット・docs には実在確認した事実のみを書く**（ファイルの存在・コードの挙動・PR のマージ状態を検証してから書く。検証できないことは書かない）

### 8-bis. ブランチ衛生ルール

PR マージ後の取り扱い:

- **PR が squash merge された直後に、そのローカルブランチで別タスクの作業を始めない**
  - 過去事故: マージ済ブランチにそのまま README 修正を commit してしまい、ブランチ名と中身が不一致になった
- マージ後は次を実行:
  ```bash
  git checkout main && git pull origin main
  git branch -D <merged-branch>     # ローカルブランチ削除
  ```
- 新しいタスクを始めるときは**必ず main から新規ブランチを切る**

### 8-ter. メンテナンスページ / useBackendHealth の設計仕様

`frontend/src/hooks/useBackendHealth.ts` の仕様:

- `FAILURE_THRESHOLD = 2`: 連続 2 回失敗で `'unhealthy'` に落とす（single transient error の誤判定防止）
- `POLL_INTERVAL_HEALTHY_MS = 60_000`: 正常時は 60 秒間隔
- `POLL_INTERVAL_AFTER_FAILURE_MS = 2_000`: 1 回失敗後は 2 秒で再 poll（メンテナンス判定を高速化）
- `POLL_INTERVAL_UNHEALTHY_MS = 15_000`: unhealthy 中は 15 秒間隔で復旧検知
- `TIMEOUT_MS = 5_000`: 1 回の health check の上限

`MaintenancePage.tsx` はユーザ要望で**ミニマル**構成:
- 「定期メンテナンス時間帯」案内カードなし
- 「再試行」ボタンなし
- 文言は「サーバーにアクセスできない状態です。 自動的に再接続を試みていますので、しばらくお待ちください。」
- 連絡先は `VITE_SUPPORT_EMAIL` がセットされていれば末尾に表示

これらの仕様を変える場合は**必ずユーザ確認**を取る（過去にユーザが明示的に「いらない」と指示したもの）。
