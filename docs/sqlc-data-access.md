# sqlc によるデータアクセス（生 SQL 直書きへの段階移行）

バックエンドのデータアクセスを、GORM のクエリビルダから **sqlc（生 SQL から型安全な Go を生成）** へ段階的に移行する。背景・全体方針は [`design/2026年/0002-backend-java-to-go-revert.md`](./design/2026年/0002-backend-java-to-go-revert.md) を参照。

## なぜ sqlc か

- **生 SQL 直書き**: 実際に走る SQL がそのまま `.sql` に見える（AI / 人がレビュー・EXPLAIN・最適化しやすい）
- **コンパイル時型安全**: `.sql` と `schema.sql` から型付き Go を生成。タイポ・列不一致を**ビルド時**に検出（`db.Raw` の実行時エラーを避ける）
- **依存が薄い**: 生成コードは `database/sql` のみに依存（ORM の巨大な依存を持たない）

## 全体方針（段階移行）

```
GORM は当面「接続 + AutoMigrate」に残す
  └ クエリは repository 層の中だけで sqlc に順次置換（usecase は不変・archlint 維持）
  └ 接続は GORM の *gorm.DB.DB() で得た *sql.DB を sqlc に共有（別 pool を持たない）
最終的に GORM を撤去し、接続/Migrate も pgx + マイグレーションツールへ寄せる
```

## ファイル構成

| 役割 | パス |
|---|---|
| sqlc 設定 | `backend/sqlc.yaml` |
| 型付け用スキーマ（実体は AutoMigrate） | `backend/internal/adapter/persistence/queries/schema.sql` |
| クエリ（`.sql`、生 SQL を直書き） | `backend/internal/adapter/persistence/queries/*.sql` |
| 生成コード（コミット対象・編集禁止） | `backend/internal/adapter/persistence/sqlcgen/` |

## 使い方

1. `queries/*.sql` に `-- name: Xxx :many|:one|:exec` 付きで生 SQL を書く
2. 参照するテーブルの `CREATE TABLE` を `queries/schema.sql` に追記（`docs/schema.sql` と列を一致させる）
3. `cd backend && make sqlc` で生成（`internal/adapter/persistence/sqlcgen/` が更新される）
4. repository 実装で `*gorm.DB.DB()` から `*sql.DB` を取り、`sqlcgen.New(sqlDB).Xxx(ctx, ...)` を呼ぶ
5. 生成モデル → `domain` への詰め替え関数（例: `toDomainExample`）でドメイン型に変換して返す

参照実装: `internal/adapter/persistence/master_exercise_example_repository.go` の `ListByExerciseID`。

## 既知の制約

- **IN 句のスライス展開（`sqlc.slice`）は PostgreSQL × `database/sql` モードでは正しく生成されない**。バッチ取得（`WHERE id IN (...)`）は当面 GORM のまま残し、GORM 撤去で接続を **pgx** に寄せる際に `= ANY($1)` で書き直す
- repository テストは `//go:build integration` の実 Postgres（`testsupport.OpenTestDB`）。`FILTER` / `BOOL_OR` など Postgres 固有構文は CI の `integration tests (postgres)` ジョブで担保する
