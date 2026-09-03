# FreStyle — Claude Code プロジェクト規約
---

## 1. プロジェクト基本情報
- **本番URL**: https://frestyle.jp
- **バックエンド**: Go 1.x / Gin / sqlc（`backend/`）
- **フロントエンド**: React 19 / TypeScript / Vite / Tailwind CSS（`frontend/`）
- **RDB**: PostgreSQL 17.6。データアクセスは **sqlc**（SQL から型付き Go を生成）
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

### 2.2 各層の責務

| 層 | パッケージ | 責務 |
|---|---|---|
| handler | `backend/internal/handler` | HTTP 受付、middleware から認証情報取得、usecase 呼び出し、JSON 返却。ビジネスロジック禁止 |
| middleware | `backend/internal/handler/middleware` | JWT 認証、CORS、current user 注入、CSRF 等の横断的処理 |
| usecase | `backend/internal/usecase` | 1 ユースケース = 1 構造体（単一責任）。repository / infra をオーケストレーション。HTTP 層の型への依存禁止 |
| repository (port) | `backend/internal/usecase/repository` | usecase が依存する repository interface の定義 |
| persistence (adapter) | `backend/internal/adapter/persistence` | sqlc 生成コード / S3 等の repository 実装 |
| infra | `backend/internal/infra/{oidc,s3,ses,database,sandbox,...}` | 外部サービス連携（AWS SDK ラッパ）、DB 接続、設定読み込み |
| domain | `backend/internal/domain` | エンティティ + ビジネス定数。JSON tag のみ直書き（永続化の都合は持ち込まない）。他層を import しない |

### 2.3 1 構造体 1 責務（usecase）

- usecase 1 つにビジネスルール 1 つ。複数操作をまとめない
- usecase は **struct + `NewXxxUseCase` コンストラクタ + `Execute(ctx, in) (out, error)`** で書く

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
- **チケット・docs には実在確認した事実のみを書く**（ファイルの存在・コードの挙動・PR のマージ状態を検証してから書く。検証できないことは書かない）

