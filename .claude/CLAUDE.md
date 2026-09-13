# FreStyle — Claude Code プロジェクト規約
---

## 1. プロジェクト基本情報
- **本番URL**: https://frestyle.dev
- **バックエンド**: Go 1.x / Gin / sqlc（`backend/`）
- **フロントエンド**: React 19 / TypeScript / Vite / Tailwind CSS（`frontend/`）
- **RDB**: PostgreSQL 17.6。データアクセスは **sqlc**（SQL から型付き Go を生成）
- **本番はすべて GCP**（ECS は使っていない）
  - **バックエンド**: Cloud Run サービス `frestyle-prod-backend`（プロジェクト `frestyle-prod` / `asia-northeast1`）。イメージは Artifact Registry の `asia-northeast1-docker.pkg.dev/frestyle-prod/frestyle-prod-backend/fre-style`。公開先は https://api.frestyle.dev
  - **フロントエンド**: Firebase Hosting（プロジェクト `frestyle-507912`。バックエンドとは**別プロジェクト**で表示名がどちらも「FreStyle」なので必ず ID で指定する）。公開先は https://frestyle.dev
  - インフラ定義（Cloud Run / Artifact Registry / Firebase Hosting / WIF）は private リポ `frestyle-infrastructure` の Terraform が正。CD はイメージの差し替えと配信だけを担い、インフラ定義には触れない
- **デプロイは手動トリガーのみ**。`cd-backend.yml` / `cd-frontend.yml` を `confirm` に `deploy` と入れて `workflow_dispatch` で起動する（main への push では動かない）。認証は GitHub OIDC + WIF で、長寿命のサービスアカウントキーは発行しない
  - WIF の binding は `refs/heads/main` の実行に限定されている。**principalSet はリポジトリ名を大文字小文字まで含めて突き合わせる** — リポジトリを改名したら infra 側の binding も直すこと（追従し忘れると、認証は通るのに直後のサービスアカウントへのなりすましが `iam.serviceAccounts.getAccessToken` の 403 になり、backend・frontend とも全部止まる。実際に踏んだ）
- **デプロイの順序は「本番 DB のスキーマ適用（`make schema-apply`）→ backend → frontend」**。スキーマが古いまま backend を出すと、存在しない列を引く問い合わせが本番で 500 になる（実際に踏んだ）。フロントが新しい API に依存するときも backend を先に出す。本番スキーマの現状は Supabase CLI（`supabase db query --linked`・読み取りのみ）で確かめる
- 開発そのものはローカル環境（`docker compose up`）で進める
---

## 2. クリーンアーキテクチャ規約（最重要）

### 2.1 依存方向ルール

```
handler → usecase → repository / infra → domain
```

- **矢印の向き以外の依存は禁止**
- handler は repository / infra を直接呼ばない。必ず usecase を経由する
- usecase は handler を知らない（`*gin.Context` 等を引数で受けない）
- repository / infra は usecase を知らない。domain は他のどの層にも依存しない（標準ライブラリのみ）


### 2.2 構造体 1 責務（usecase）

- usecase 1 つにビジネスルール 1 つ。複数操作をまとめない
- usecase は **struct + `NewXxxUseCase` コンストラクタ + `Execute(ctx, in) (out, error)`** で書く
- 新しい usecase は `internal/usecase/<domain>/` に置く。直下（`internal/usecase/*.go`）には置かない
- usecase サブパッケージ同士は import しない（相互依存が要るなら責務の切り方を疑う）。
  handler 側でパッケージ名（`user` / `exercise` / `kb` 等）と同じ名前のローカル変数を宣言しない
  （パッケージ参照が隠れてコンパイルエラーになる）

### 2.4 domain と request / response 型の境界

- handler は domain 構造体をそのまま JSON で返して良い。加工・隠蔽が必要な場合のみ handler 内で response 構造体を定義
- リクエスト入力は handler のファイル内で `xxxRequest` struct + `c.ShouldBindJSON` + `binding:"required"` 等で宣言的にバリデーション
- usecase の入力は `XxxInput` struct、戻り値は `*domain.Xxx` または primitive
- 機密フィールド（パスワード hash、招待 token、BlobData 等）は domain 側で `json:"-"` を付けて除外する

### 2.5 フロントエンドのレイヤー（FSD / Feature-Sliced Design）

`frontend/src/` は **Feature-Sliced Design** で構成する

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

### 3.3 テスト

- **TDD を基本**とする。カバレッジ目標: 新規コード **80% 以上**
- **バックエンド（単体）**: `testing` + `stretchr/testify`（`go test ./...`）— usecase は interface モック（testify/mock）、handler は `httptest` + `gin.New()`、infra は境界で fake / stub 注入。**DB を必要としないものだけ**をここに置く
- **バックエンド（結合）**: repository は **本物の PostgreSQL** で検証する（sqlite は使わない。依存も入れていない）。ファイル先頭に `//go:build integration`、テスト関数名に `Integration` を含める。ローカルは `make test-integration`（docker で postgres 起動 → 実行 → 必ず破棄）、CI は専用ジョブ `integration tests (postgres)` が `-tags=integration` で実行する
- 結合テストの接続は `internal/testsupport.OpenTestDB`。`TruncateAll` が TRUNCATE CASCADE するため、**DSN が Supabase / 本番 pooler を指す場合は接続前に落とす安全弁**が入っている（誤設定で本番データを消さないため）
- フロントエンド: Vitest + React Testing Library（`pnpm test`）。**`vitest` / `@vitest/browser-playwright` / `@vitest/coverage-v8` は同じ版に固定する**（`^` を付けない）。本体とブラウザ側でプロトコルが一致している必要があり、ずれると story のテストが「ブラウザセッションに接続できない」で丸ごと止まる（実際に踏んだ）— `render` + `screen.getByRole` でアクセシビリティも検証、Hook は `renderHook`

---

## Claude Code への指示
- 新しい画面は **`src/shared/ui/` の再利用コンポーネント**を最大限活用
- `main` へ直接コミット・push しない
- `xxxRequest` / `xxxResponse` は handler のファイル内で local 定義。機密フィールドは domain 側の `json:"-"` で隠す
