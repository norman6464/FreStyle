# 運営 SQL コンソール（super_admin 専用・読み取り専用）

super_admin（運営管理者）が、アプリ画面から本番 DB に対して **read-only の SQL** を直接実行できる管理ツール。外部 BI を導入せず、調査・確認用のアドホッククエリをその場で叩くための最小機能。

- 画面: `/admin/sql`（サイドバー「管理 → SQL コンソール」。super_admin にのみ表示）
- API: `POST /api/v2/admin/sql` body `{ "query": "SELECT ..." }`

## セキュリティ（多層防御）

任意 SQL を本番 DB に投げる高リスク機能のため、以下を**多層**で担保する。

1. **認可**: handler で `super_admin` のみ許可（`isSuperAdmin`）。company_admin / trainee / 未認証は `403`。フロントも `role !== 'super_admin'` を `/` にリダイレクト（UI 二重化）。
2. **DB レベルの読み取り専用トランザクション（最終防壁）**: repository は `BEGIN; SET TRANSACTION READ ONLY; ...` 内でクエリを実行する。万一バリデーションをすり抜けた書き込み（INSERT/UPDATE/DELETE/DDL や、関数経由の副作用）も **Postgres が "cannot execute ... in a read-only transaction" で拒否**する。
3. **クエリ検証（前段）**: usecase が「単一の `SELECT` / `WITH` で始まる」クエリのみ許可。文中セミコロン（複文）は拒否。末尾の単一セミコロンのみ許容。
4. **DoS / OOM 抑止**: `SET LOCAL statement_timeout = 5000`（5 秒）+ 行数上限 **1000 行**（超過は `truncated=true` で打ち切り）。
5. **監査ログ**: 実行のたびに `AUDIT sql-console: super_admin(id, email) query=...` を出力（誰がいつ何を実行したか追跡可能）。

> 補足: super_admin は最も信頼されたロールのため、DB エラー（構文エラー / relation does not exist 等）は調査の利便性を優先してそのまま画面に返す。文中セミコロンを含む文字列リテラル（例 `LIKE '%;%'`）は複文判定で弾かれる既知の制限がある。

## レイヤー構成（クリーンアーキテクチャ）

```
AdminSqlPage / useAdminSql / AdminSqlRepository (frontend)
        │  POST /api/v2/admin/sql
        ▼
AdminSQLHandler (super_admin チェック + 監査ログ)
        ▼
ExecuteReadOnlySQLUseCase (SELECT/WITH 検証)
        ▼
SQLConsoleRepository(port) ← persistence.sqlConsoleRepository(impl)
        ▼
read-only トランザクション + statement_timeout + 行数上限
```

| 層 | ファイル |
|---|---|
| handler | `backend/internal/handler/admin_sql_handler.go` |
| usecase | `backend/internal/usecase/execute_readonly_sql_usecase.go` |
| repository(port) | `backend/internal/usecase/repository/sql_console.go` |
| repository(impl) | `backend/internal/adapter/persistence/sql_console_repository.go` |
| frontend | `frontend/src/pages/AdminSqlPage.tsx` / `hooks/useAdminSql.ts` / `repositories/AdminSqlRepository.ts` |

## 使い方

1. super_admin でログイン → サイドバー「管理 → SQL コンソール」
2. テキストエリアに `SELECT` / `WITH` クエリを入力（`⌘ / Ctrl + Enter` でも実行）
3. 結果はテーブル表示。`NULL` は `NULL`、boolean は `true/false` を明示。1000 行超は打ち切り表示。

## テスト

- usecase: 空 / 非 SELECT（DELETE・UPDATE・複文等）の拒否、SELECT/WITH の許可（`execute_readonly_sql_usecase_test.go`）
- handler: super_admin のみ 200・他ロール/未認証は 403・query 欠落 400・write は 400（`admin_sql_handler_test.go`）
- repository(integration): 実 Postgres で SELECT 取得・`maxRows` 打ち切り・**write が read-only tx で拒否される**ことを検証（`sql_console_repository_integration_test.go`）
- frontend: `AdminSqlRepository.test.ts`
