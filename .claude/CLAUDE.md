# FreStyle — Claude Code プロジェクト規約

> このファイルは **git 管理されるチーム共通の規約**（`.claude/settings.json` と同様にコミットして共有）。
> 本リポジトリは **PUBLIC** のため、**秘匿情報（DB 接続情報 / Secrets Manager のシークレット名 / 内部運用の詳細）はここに書かない**。それらは private リポ `frestyle-infrastructure` の docs・AWS Secrets Manager・各自の `.env`（gitignore 済）で管理する。個人ローカルのメモはリポ直下の `CLAUDE.md`（`/CLAUDE.md` として gitignore 済）に置ける。
> **アーキテクチャ仕様の一次情報は private リポ `frestyle-pdm` の `docs/ARCHITECTURE.md`**（PR #1633 で本リポから移管）。このファイルは必ず整合を保つこと。

日本語で会話をしてください。

---

## 1. プロジェクト基本情報

- **プロジェクト名**: FreStyle — 新卒 IT エンジニア向け統合研修プラットフォーム（B2B SaaS）
- **本番URL**: https://normanblog.com
- **バックエンド**: Go 1.x / Gin / GORM（`backend/`、ECS Fargate）
- **フロントエンド**: React 19 / TypeScript / Vite / Tailwind CSS（`frontend/`）
- **RDB**: PostgreSQL 17.6（Supabase / Transaction pooler / GORM AutoMigrate、2026-05 に RDS から移行）
- **NoSQL**: DynamoDB（`fre_style_ai_chat` — AI チャットメッセージ）
- **認証**: AWS Cognito（OIDC + JWT HttpOnly Cookie、招待マジックリンク方式）
- **AI**: AWS Bedrock（Claude Sonnet 4.5 / Inference Profile 経由）
- **非同期処理**: SQS（学習レポート生成キュー） / **チャット通信**: SSE on ECS
- **IaC**: Terraform（private リポ `frestyle-infrastructure` で管理）

---

## 2. クリーンアーキテクチャ規約（最重要）

### 2.1 依存方向ルール

```
handler → usecase → repository / infra → domain
```

- **矢印の向き以外の依存は禁止**
- handler は repository / infra を直接呼ばない。必ず usecase を経由する
- usecase は handler を知らない（`*gin.Context` 等を引数で受けない）
- repository / infra は usecase を知らない。domain は他のどの層にも依存しない（標準ライブラリ + GORM tag のみ）

### 2.2 各層の責務

| 層 | パッケージ | 責務 |
|---|---|---|
| handler | `backend/internal/handler` | HTTP / SSE 受付、middleware から認証情報取得、usecase 呼び出し、JSON 返却。ビジネスロジック禁止 |
| middleware | `backend/internal/handler/middleware` | JWT 認証、CORS、current user 注入、CSRF 等の横断的処理 |
| usecase | `backend/internal/usecase` | 1 ユースケース = 1 構造体（単一責任）。repository / infra をオーケストレーション。HTTP 層の型への依存禁止 |
| repository (port) | `backend/internal/usecase/repository` | usecase が依存する repository interface の定義 |
| persistence (adapter) | `backend/internal/adapter/persistence` | GORM / DynamoDB / S3 / SQS 等の repository 実装 |
| infra | `backend/internal/infra/{bedrock,s3,ses,cognito,database,...}` | 外部サービス連携（AWS SDK ラッパ）、DB 接続、設定読み込み |
| domain | `backend/internal/domain` | エンティティ + ビジネス定数。GORM tag は pragmatic に直書き。他層を import しない |

### 2.3 1 構造体 1 責務（usecase）

- usecase 1 つにビジネスルール 1 つ。複数操作をまとめない（❌ `CreateSessionAndSendFirstMessage` → ⭕ usecase を分けて handler 側でオーケストレーション）
- usecase は **struct + `NewXxxUseCase` コンストラクタ + `Execute(ctx, in) (out, error)`** で書く

### 2.4 domain と request / response 型の境界

- handler は domain 構造体をそのまま JSON で返して良い。加工・隠蔽が必要な場合のみ handler 内で response 構造体を定義
- リクエスト入力は handler のファイル内で `xxxRequest` struct + `c.ShouldBindJSON` + `binding:"required"` 等で宣言的にバリデーション
- usecase の入力は `XxxInput` struct、戻り値は `*domain.Xxx` または primitive
- 機密フィールド（パスワード hash、招待 token、BlobData 等）は domain 側で `json:"-"` を付けて除外する

### 2.5 フロントエンドのレイヤー（FSD / Feature-Sliced Design）

`frontend/src/` は **Feature-Sliced Design** で構成する（FRESTYLE-154 で移行完了。旧 `components/hooks/repositories/utils/constants/lib/store/types` は撤去済）。

**レイヤー（上ほど上位・import は下向きの一方通行）**

```
app > pages > widgets > features > entities > shared
```

- **app**: エントリ・Provider・ルーティング・store 組み立て（`app/store`）
- **pages**: 1 画面 = 1 Slice。その画面専用の hook / component は `pages/<slice>/{ui,model,lib,config}` に同居
- **widgets**: 複数機能を組み合わせた自立 UI ブロック（例: `app-shell` = ヘッダ + サイドバー + コマンドパレット）
- **features**: 再利用されるユーザー操作（例: `auth` = ログイン / ログアウト / 認証状態取得）
- **entities**: ビジネス上の「もの」（`course` / `exercise` / `user` / `note` / `ai-chat` など）。`api`(リポジトリ) / `model`(型・slice) / `ui`(単体表示)
- **shared**: ビジネスを知らない再利用資産。UI キット（`shared/ui`）/ axios（`shared/api`）/ 汎用 hook・関数（`shared/lib`）/ 型付き Redux hooks（`shared/lib/store`）/ 定数（`shared/config`）

**ルール（境界 lint `eslint.config.js` が CI で `error` 強制）**

- 自分と同じか上の層は import できない（下向きのみ）。**app と shared のあいだだけ相互 import 可**（公式の例外。typed Redux hooks が RootState を参照するため）
- 各 Slice は **Public API（`index.ts`）経由**で使う。名前付き re-export のみ（`export *` 禁止）。Slice 内部は相対パス（自分の barrel を参照しない）
- entity 同士がどうしても参照し合う場合のみ **`@x` 記法**（`entities/<相手>/@x/<自分>`）。増えたら Slice の切り方を疑う
- **単一画面専用のものは page の model/ui に置く**（features は 2 画面以上で共有される操作に限る）。「どのプロジェクトでも使えるか」で shared か上位かを判断する
- テスト（`__tests__`）は層間ルールの対象外だが、**Slice の自己参照は禁止**（barrel を読むとカバレッジ分母が膨らむため深いパスで mock する）
- 詳細と移行の実績・ハマりどころは `frontend/src/entities/README.md` / `frontend/src/shared/README.md`（設計の一次情報は private リポ `frestyle-pdm`）

### 2.6 Repository 配置規約（物理分離）

- Repository **interface**（port）は `backend/internal/usecase/repository/{entity}.go`、**実装**は `backend/internal/adapter/persistence/{entity}_repository.go` の 2 ファイル構成で書く（依存方向は usecase ← persistence の DIP）
- wiring は `internal/handler/router.go` で `persistence.NewXxxRepository(...)` を組み立てて注入
- 1 boundary = 1 fat interface で良い。単一責務の port（`Presigner` / `Enqueuer` 等）は `-er` 命名で切り出して良い。interface 名は `XxxRepository`
- 旧構造（1 ファイルに interface + 実装同居）は 2026-05 撤去済。戻さない

### 2.7 OpenAPI annotation 規約（swaggo）

新しい HTTP endpoint には handler メソッド直前に swaggo annotation を必ず書く（`@Summary` / `@Description` / `@Tags` / `@Accept`（body があるとき）/ `@Produce` / `@Param` / `@Success` / `@Failure` / `@Router` / 認証必須なら `@Security CookieAuth`）。

- 共通 response 型（`errorResponse` / `messageResponse` 等）は `backend/internal/handler/openapi_types.go` に定義済
- `@Router` の path に `/api/v2` を含めない（`@BasePath` で自動 prefix）
- handler を追加 / 変更する PR では**必ず `make openapi`** を走らせ `backend/docs/`（生成物）の差分も commit する
- SSE / multipart 等 OpenAPI で完全表現できないものはシンプルな POST として表現し `@Description` で補足

---

## 3. コーディング規約

### 3.1 言語

- **日本語**: PR タイトル / チケット / コミットメッセージ / コメント。**英語**: 識別子

### 3.1.1 他社プロダクト名の不使用（重要）

PR / チケット / コミット / コメント / docs に**他社プロダクト・他社サービスの名前を書かない**（「〜風 UI」等の比喩も含む）。❌「（他社サービス名）風 2 カラムレイアウト」→ ⭕「2 カラムレイアウト（本文 + 目次サイドバー）」。機能の中身（何があるか）で説明する。

### 3.2 命名規則

- usecase: `[動詞][目的語]UseCase`（`NewXxxUseCase` / `Execute`）
- repository interface: `[Entity]Repository`。実装は `persistence` の小文字 struct + `NewXxxRepository`
- handler: `[ドメイン]Handler`、メソッドは `(h *XxxHandler) Action(c *gin.Context)`
- domain: 名詞（`User` 等）。GORM tag + JSON tag 直書き、必要なら `TableName()`
- request / response: handler 内 local 定義の `xxxRequest` / `xxxResponse`
- infra: パッケージ名は領域別（`bedrock` / `s3` / `ses` / `cognito` / `database`）
- フロントエンド コンポーネント: PascalCase、1 ファイル 1 コンポーネント

### 3.3 テスト

- **TDD を基本**とする。カバレッジ目標: 新規コード **80% 以上**
- バックエンド: `testing` + `stretchr/testify`（`go test ./...`）— usecase は interface モック（testify/mock）、repository は sqlite メモリ、handler は `httptest` + `gin.New()`、infra は境界で fake / stub 注入
- フロントエンド: Vitest + React Testing Library（`npm test`）— `render` + `screen.getByRole` でアクセシビリティも検証、Hook は `renderHook`

### 3.4 コメント

- 必要最小限。WHY を残し WHAT は書かない
- exported な struct / func には責務を GoDoc コメントで 1〜2 行（識別子から始める）

### 3.5 禁止事項

- ❌ handler から repository / infra への直接呼び出し
- ❌ usecase をまたぐビジネスロジックを 1 つの usecase に詰め込む
- ❌ usecase / repository / infra から `*gin.Context` / `net/http` の参照
- ❌ `db.Begin()` / `Tx` を handler に書く（トランザクションは usecase レベル）
- ❌ `main` への直接コミット / テストのないマージ / domain から他層への依存

### 3.6 スキーマ優先（分類・データ構造はバックエンドで持つ）

**データ構造・分類がスキーマとして不自然なときは、フロントに回避ロジックを足さず、スキーマ（バックエンド）側の変更を第一に検討する。** 「表示のためにフロントで無理やり組み立てる」のは技術的負債になりやすく、複数画面で重複する。分類の正本・派生・集計は原則 DB / backend に置き、フロントは表示に徹する。

- 「フロントだけで頑張れば動く」からといって回避ロジックを足さない。**それはバックエンド（スキーマ）の問題**として扱う。フロントに分岐・整形・集計を積み増す前に「これはスキーマで表現すべきでは？」を問う
- 分類（ジャンル / カテゴリ / 技術タグ等）に **DB 上のメタ情報（説明・表示順・アイコン・多言語）や多対多・ユーザー管理**が必要になったら、enum 文字列カラムから**明示的な型付きテーブル（列を型として持つ + FK / 中間テーブル）**への正規化を検討する。まずスキーマを直す
- パーソナライズ・レコメンド等の**派生データは backend で集計/永続化**し、フロントは受け取って描画するだけにする（クライアントで重い突き合わせをしない）
- **❌ EAV（Entity-Attribute-Value）は禁止**。「スキーマから直す」＝ `user_id, attribute_key, value` のような**汎用属性バッグ**を作ることではない。属性・分類・特性は必ず**意味の分かる型付きカラム / 専用テーブル**で表現する（例: 「直近で好きな技術」は `(user_id, language, activity_count, last_active_at)` のような**明示カラムを持つ集計テーブル**で持つ。任意 key/value にしない）。柔軟さと引き換えに型・制約・クエリ性・可読性を失うため

**現状の分類設計（2026-07 時点の事実）**: コースの学習領域は `courses.category`（`domain.ValidCourseCategories` で検証する **enum 文字列カラム**。別テーブルではない。migration `0007` はカラム値の backfill）。演習の技術は `master_exercises.language`（同じく enum 文字列カラム）。**「ジャンル → 子コース」は DB のエンティティ関係ではなくカラム値による分類**で、カテゴリの表示メタ（ラベル・色・アイコン）は現状フロント（`entities/course/config/courseCategories.ts`）が持つ。将来ここに DB メタや多対多が必要になったら §3.6 に従いスキーマから直す（EAV にはしない）。

---

## 4. PR / チケット / マージフロー

### 4.1 ブランチ運用

1. Jira チケットを起票（または既存を選ぶ）→ **自分にアサイン**してから着手（§4.6）
2. ブランチを切る（`feat/*`, `fix/*`, `refactor/*`, `docs/*`, `test/*`）
3. 作業 → コミット → PR（タイトル・本文とも日本語、チケット番号を含める）
4. CodeRabbit レビューを待つ → 指摘対応 → **squash merge**（マージは必ず PR 経由）

### 4.2 CodeRabbit レビュー待機のルール

- PR 作成後、CodeRabbit の初回レビューが投稿されるまで待機。「Actionable comments」には原則すべて応答（対応 or 意図を説明して reject）
- CodeRabbit が指摘していないセキュリティ・アーキテクチャ違反はセルフレビューで修正。summary コメントは PR 本文に貼らない
- **待つときは必ず `sleep` で実時間を経過させる**（`sleep 270` 等）。「あとで確認します」への逃げは禁止。長い sleep がブロックされる場合は `timeout` を長めにした待機ループで同期的に待つ。例外はユーザーが「待たずに進めて」と明示したときのみ
- **レートリミット特例**: CodeRabbit が `Rate limit exceeded` のとき、(1) ユーザー承認あり (2) セルフレビュー済で Critical / Major なし (3) CI 成功 (4) 破壊的変更でない、をすべて満たせば `gh pr merge <PR#> --squash --admin --delete-branch` で先に進めて良い。マージ後コメントに「レートリミット中のためセルフレビュー済で admin merge」と記録する
- **デザイン専用変更の特例**: ロジックに影響しない見た目だけの変更（CSS / className / 文言 / svg 等）で、`tsc --noEmit` / `vitest run` / `eslint --max-warnings 0` / `npm run build` がすべて成功していれば、CodeRabbit を待たず admin merge して良い。ロジックを伴う変更は分割して通常フローに乗せる

### 4.3 コミットメッセージ / PR 本文

- Prefix: `feat` / `fix` / `refactor` / `docs` / `test` / `chore` / `perf` / `style`。日本語で書く
- コミット末尾に `Co-Authored-By: <使用した Claude モデル名> <noreply@anthropic.com>`（修飾語は付けない）。**`Generated with Claude Code` の文言は入れない**
- PR 本文は `## 概要` / `## 変更内容` / `## テスト` / `## 関連チケット` の 4 セクション（テンプレートは `.github/PULL_REQUEST_TEMPLATE.md`）

### 4.4 シングルタスク運用（最重要・PR を溜めない）

- **PR は必ず 1 つずつ**。「作成 → CodeRabbit/CI 対応 → マージ」まで完了してから次の PR に着手する
- 理由（実際の事故）: PR 4 本同時 open で `backend/docs/swagger.*` 等の生成ファイルが互いにコンフリクトし管理コストが跳ね上がった
- 大きな作業は「1 機能 = 1 PR」で直列に分割。前の PR がマージされ `main` が最新になってから次のブランチを切る
- 並行が避けられない場合のみ、理由を伝えてユーザーの明示承認を得る。次の PR を作る前に open の PR が残っていないか必ず確認する

### 4.5 Jira チケット運用（作業管理は GitHub Issue ではなく Jira）

- **プロジェクト**: `https://frestyle.atlassian.net`。Atlassian Rovo コネクタ（MCP）経由で操作する
- **着手フロー**: チケット作成 or 選択 → **自分にアサイン**（未アサインのまま作業は禁止）→ ブランチ → PR → マージ後にチケットを完了遷移
- **課題タイプ**: `hotfix` / `開発タスク` / `リファクタリング` / `ドキュメント整備` / `Design Doc` / `エピック` から選ぶ。設計判断を伴うものは `Design Doc` で起票し、背景・選択肢・推奨案・承認記録を残す
- **description はテンプレートに沿う**（概要 / ゴール / スコープ外 / 背景・目的 / テスト・検証 / セキュリティ影響 / 影響範囲 / ロールバック方針 / 参考リンク）。報告者は自分、開始日・終了日は原則未記入
- **実在確認した事実のみを書く**: ファイル・挙動・PR のマージ状態は書く前に必ずリポジトリと突き合わせて検証する。検証できないことは書かないか「未確認」と明記する
- **参考リンクは git 管理された文書を指す**。git 管理されていないローカルファイルへの参照は書かない
- **コメント欄**にはステージング検証手順・実行結果・期待値との突き合わせ（合否明記）を書き、チケットだけで検証内容が追える状態にする
- **PR との相互リンク必須**: チケットに PR URL / 番号を書き、PR タイトル / コミットにチケット番号を入れて双方向に辿れるようにする

---

## 5. デプロイ

- **バックエンド**: `gh workflow run "CD - Backend Deploy to ECS" -R norman6464/FreStyle -f confirm=deploy`（ECR build/push + ECS force-update）。ヘルスチェックは本番 API ドメインの `GET /api/v2/health`（CloudFront 配下の SPA パスに叩くと一律 200 になり誤認する）
- **フロントエンド**: `gh workflow run "CD - Frontend Deploy to S3 + CloudFront" -R norman6464/FreStyle -f confirm=deploy`
- **DB マイグレーション**: 新規テーブル / 列追加は GORM AutoMigrate が ECS 起動時に自動適用。列削除 / リネーム / 型変更は `backend/migrations/000X_*.sql` に置き、private リポ `frestyle-infrastructure` の `make apply-migration-supabase` で適用（実引数・手順は同リポ docs 参照）。冪等性（`IF NOT EXISTS` 等）を必ず担保する

---

## 6. ローカル開発環境

- `cp .env.example .env` して接続情報を記入。主な環境変数: `DATABASE_URL`（推奨・`DB_HOST` 等より優先）/ `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` / `AWS_REGION` / `BEDROCK_MODEL_ID` / `DYNAMODB_AI_CHAT_TABLE`（dev は `fre_style_ai_chat_dev`）/ `NOTE_IMAGES_BUCKET`
- **接続情報の実値は git に commit しない**（`.env` または AWS Secrets Manager）。Supabase 接続 URL の取得・pooler の使い分け・runbook は private リポ `frestyle-infrastructure` の docs を参照
- backend は `DATABASE_URL` セット時、pgbouncer 互換のため `PrepareStmt: false` + simple query protocol を自動適用（host 名で自動判定）。GORM AutoMigrate もそのまま動く
- 起動: backend は `cd backend && go run ./cmd/server`（確認は `go build ./... && go test ./...`）、frontend は `cd frontend && npm install && npm run dev`

---

## 7. ドキュメンテーション

**本リポジトリに `docs/` フォルダは置かない**（2026-07 に撤去）。このリポは PUBLIC で、アプリケーション本体を置く場所と位置づける。

- **README はアプリケーションの説明に限定する**（プロダクトの概要 / 機能 / 構成図 / ローカル起動）。開発プロセス・テスト戦略・CI・チーム運用・セキュリティ運用は README に書かない
- 実装・設計ドキュメントは **private リポ `frestyle-pdm` の `docs/`**（アーキテクチャ仕様の一次情報）に置く。インフラ / 運用は private リポ `frestyle-infrastructure` の `docs/`
- 本リポに残す文書は **README.md / CONTRIBUTING.md / 各ディレクトリの README**（`backend/README.md` など）のみ
- 「なぜやったか」「どう動かすか」「何が変だったか」は **Jira チケット**（Design Doc / 各作業チケット）に残す。チケットだけで背景と手順が追える状態を保つ（§4.5）
- 例外: `backend/docs/` は `make openapi` が生成する OpenAPI spec で、`cmd/server/main.go` が import している**ビルド依存**。削除しないこと

---

## 7-bis. 教材コンテンツの管理（重要）

教材本文（Markdown）はアプリ本体のリポジトリには置かず、private リポ **`norman6464/frestyle-teaching-materials`** で管理して本番 DB に同期する（FreStyle 直下に clone する運用・gitignore 済）。

- **構成**: `courses/NN-{course-slug}/` に `course.yaml`（**id（本番コース ID・必須）** / title / description / category / sort_order / is_published）+ `001-{slug}.md` 連番の教材。教材は `# タイトル` から始める（先頭 h1 が DB の title になる）
- **同期フロー**: 教材リポで編集 → `courses/_scripts/seed-courses.py` で章ごとの SQL を生成 → private リポ `frestyle-infrastructure` の `make apply-migration-supabase`（または `make seed-course`）で適用（実引数は同リポ docs 参照）→ `https://normanblog.com/courses` で反映確認
- **差分反映 + 章数検証（FRESTYLE-6 で全削除方式から変更・教材リポ PR https://github.com/norman6464/frestyle-teaching-materials/pull/24）**: `{course}__course.sql` は course を UPSERT 後、リポに現存しない order の章だけを DELETE する差分反映設計（course SQL 単独適用でも現存章は消えない）。章 SQL は UPDATE + INSERT ... WHERE NOT EXISTS の upsert で、既存行の id / created_at を保持する。`{course}__999-verify.sql` が適用後の章数を検証し、不一致なら RAISE EXCEPTION で停止する（order 999 は検証用に予約・教材の連番に使うと生成時エラー）
- **コース ID の解決**: 旧 `SORT_TO_ID`（手書き対応表・未知 sort_order は黙ってスキップ）は撤廃済み（教材リポ PR #19 / FRESTYLE-5）。コース ID は各 `course.yaml` の `id:` 欄で明示し、**未記載はエラーで停止**する（黙って seed 対象から漏れることはなくなった）。新規コース追加時は `id:` に本番のコース ID を書く
- **最新のコース一覧・id 対応は教材リポの `courses/*/course.yaml`（`id:` 欄）を一次情報として参照する**（このファイルに一覧を複製しない — 陳腐化して誤情報の元になるため）
- 編集と DB 同期は同じセッション内で完結させる（push だけして DB に流し忘れる事故を防ぐ）
- **コードサンプルは全文掲載**: 教材内で FreStyle 自身のソースを引用するときは該当ファイルを省略なしで全文掲載する（`// 略` 等の切り詰め禁止）。長い場合は関数単位の複数ブロックに分け、各ブロック内は省略しない。コードブロック直前にファイルパスを明示。ソース更新時は教材側も同じセッション内でバックポートする

---

## 8. Claude Code への指示

1. 新しいユースケースは**必ず usecase struct を新規作成**（既存に詰め込まない）
2. 新しい画面は **`src/components/ui/` の初心者向けコンポーネント**を最大限活用
3. ビジネスロジックを handler に書かない
4. **テストなしでマージしない**（新規コードに `go test` / `vitest` を必ず付ける）
5. PR を作ったら CodeRabbit レビューを待つ。自動マージしない。ただしセッション内で一度でもレートリミットに達したら、以降はセルフレビューの上でマージまで進めて良い（§4.2 特例条件に従う）
6. `main` へ直接コミット・push しない
7. コミット・PR・チケットはすべて**日本語**
8. `xxxRequest` / `xxxResponse` は handler のファイル内で local 定義。機密フィールドは domain 側の `json:"-"` で隠す
9. 層設計の詳細は private リポ `frestyle-pdm` の `docs/ARCHITECTURE.md` を一次情報として参照（本リポの docs/ からは PR #1633 で移管済）
10. **取り組んだ内容は必ず `docs/` に書く**（§7）
11. **チケット・docs には実在確認した事実のみを書く**（ファイルの存在・コードの挙動・PR のマージ状態を検証してから書く。検証できないことは書かない）

### 8-bis. ブランチ衛生ルール

- squash merge 直後のローカルブランチで別タスクを始めない（過去にブランチ名と中身が不一致になる事故あり）
- マージ後は `git checkout main && git pull origin main && git branch -D <merged-branch>` を実行し、新タスクは必ず main から新規ブランチを切る

### 8-ter. メンテナンスページ / useBackendHealth の設計仕様

`frontend/src/hooks/useBackendHealth.ts`: `FAILURE_THRESHOLD = 2`（連続 2 回失敗で unhealthy）/ `POLL_INTERVAL_HEALTHY_MS = 60_000` / `POLL_INTERVAL_AFTER_FAILURE_MS = 2_000` / `POLL_INTERVAL_UNHEALTHY_MS = 15_000` / `TIMEOUT_MS = 5_000`。

`MaintenancePage.tsx` はユーザ要望で**ミニマル**構成: 「定期メンテナンス時間帯」カードなし / 「再試行」ボタンなし / 文言は「サーバーにアクセスできない状態です。 自動的に再接続を試みていますので、しばらくお待ちください。」/ 連絡先は `VITE_SUPPORT_EMAIL` セット時のみ表示。**この仕様を変える場合は必ずユーザ確認を取る**（過去にユーザが明示的に「いらない」と指示したもの）。
