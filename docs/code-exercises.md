# コード実行サンドボックス（演習）

学習者が演習エディタ（Monaco）でコードを書いて実行・採点する機構の設計メモ。

## アーキテクチャ

```
frontend ExerciseDetailPage (Monaco / monacoLanguageOf)
  └─ POST /api/v2/exercises/:slug/submit  または  /code/execute
       └─ usecase.SubmitMasterExercise / ExecuteCode
            └─ runner (port: usecase.CodeRunner)
                 ├─ CODE_RUNNER_URL あり → coderunner サイドカーへ HTTP 委譲
                 └─ 未設定          → in-process sandbox.Runner
                      └─ 言語別に os/exec でサンドボックス実行
```

- 実装: [backend/internal/infra/sandbox/runner.go](../backend/internal/infra/sandbox/runner.go)（in-process 本体。サイドカー `cmd/coderunner` も中身は同じ）
- 本番は **code-runner サイドカー**（[backend/Dockerfile.coderunner](../backend/Dockerfile.coderunner)）に委譲。backend 本体イメージは distroless でランタイムを持たない
- 採点は言語非依存: 実行 stdout を `normalizeOutput` して `expected_output`（または examples）と比較

## 対応言語

| language | 実行系 | 備考 |
|---|---|---|
| `php` | `php` CLI | `disable_functions` でファイル/OS/ネットワークを封じる。`<?php` 必須 |
| `go` | `go run`（単一ファイル） | `package main` 必須。GOCACHE 共有で cold compile を短縮 |
| `bash` | `/bin/bash` | 独立 process group で timeout 時に子孫まで kill |
| `sql` | 同居 PostgreSQL + `psql` | **使い捨て DB を提出ごとに作成し非 superuser で実行**（下記） |

共通制約: コード 64KB / 出力 64KB 上限、言語別 timeout、`sandboxEnv` が AWS/DB/Cognito 等の機密 env を子プロセスから除去。

## SQL 実行系（PostgreSQL）

コース「PostgreSQL 徹底入門」の方言を正確に再現するため、SQLite ではなく**本物の PostgreSQL** を使う。
本番 Supabase / `DATABASE_URL` には一切到達しない（runner 同居の使い捨て PG・socket 専用）。

### 実行ライフサイクル（[`executeSQL`](../backend/internal/infra/sandbox/runner.go)）
1. 危険な psql メタコマンド（`\!` `\copy` `\g` `\o` `\w` `\i` `\e` `\lo_` `\setenv`）を denylist で拒否
2. superuser で使い捨て DB `s_<uuid>` を `CREATE DATABASE`
3. その DB の public スキーマに `student`（非 superuser）の CREATE 権限を付与
4. 学習者 SQL を `student` で実行: `psql -A -F'|' -P footer=off -v ON_ERROR_STOP=1`、`PGOPTIONS` で `statement_timeout=5s` / `lock_timeout=2s` / `idle_in_transaction_session_timeout=5s`
5. `DROP DATABASE ... WITH (FORCE)`（実行後・timeout 時も別 ctx で必ず実行）

### 多層防御
- `listen_addresses=''` → **unix socket 専用**、TCP/ネットワーク無し
- 実行ロール `student` = 非 superuser / NOCREATEDB / NOCREATEROLE → `COPY ... TO/FROM PROGRAM`・サーバファイル不可（superuser 限定操作）
- 危険拡張を入れない（plpgsql のみ）
- `statement_timeout` ほか + 提出ごと throwaway DB（提出間でデータ混在なし）
- psql には `sandboxEnv`/明示 env で秘匿情報を渡さない（PG 接続 env も `isSensitiveEnvKey` で他言語へは遮断）

### 接続設定（env / 既定はコンテナの socket）
`CODE_PG_HOST`（既定 `/var/run/postgresql`、コンテナでは `/tmp/pgsock`）/ `CODE_PG_PORT` / `CODE_PG_SUPERUSER`（既定 `postgres`）/ `CODE_PG_STUDENT`（既定 `student`）/ `CODE_PG_*_PASSWORD`。
`Dockerfile.coderunner` が build 時に `initdb` + 極小チューニング（`shared_buffers=32MB` / `fsync=off` 等、使い捨て前提）+ `student` ロール作成を済ませ、entrypoint が socket 専用で `pg_ctl start` する。

### 出力フォーマットと出題規約
- 実行結果は `psql -A -F'|' -P footer=off` の**ヘッダ付きパイプ区切り**で stdout に出る:
  ```
  name|salary
  suzuki|550
  kato|600
  ```
- `expected_output` はこの形式で書く
- **`SELECT` は ORDER BY 必須**（無指定は行順未定義で採点が不安定）
- MVP は**自己完結 SQL**（学習者が CREATE/INSERT も書く）。事前投入テーブル前提の出題は将来 `setup_sql` 拡張で対応

## 演習の追加（教材リポが正本）
問題文・期待出力は公開リポに置かず、非公開の教材リポ `frestyle-teaching-materials/exercises/<lang>/*.md` を唯一の正本とする。
`exercises/_scripts/seed.py` が UPSERT SQL を生成 → infra リポ `make apply-migration-supabase` で Supabase に投入（詳細は CLAUDE.md §7-bis）。

## テスト
- [backend/internal/infra/sandbox/runner_sql_test.go](../backend/internal/infra/sandbox/runner_sql_test.go): `initdb`+`pg_ctl` で使い捨て PG をローカル起動して executeSQL を実検証（単純 SELECT / 複数文集計 / statement_timeout / 提出間隔離 / COPY PROGRAM 権限拒否 / denylist）。`initdb`/`psql` が無い環境では Skip
