# コード実行サンドボックス（演習）

学習者が演習エディタ（Monaco）でコードを書いて実行・採点する機構の設計メモ。

## エディタの実行ショートカット

[CodeEditor.tsx](../frontend/src/components/CodeEditor.tsx) は **Ctrl+Enter / Cmd+Enter** でコード実行（`onRun` = `runCode`）をトリガーする。キーバインドは `editor.addCommand` ではなく **`editor.addAction`** で登録する（ESM バンドルの monaco では `addCommand` のキーバインドが発火しないことがあるため）。`addAction` なら右クリックメニュー / コマンドパレットにも載り発火も安定する。`onRun` は最新の `runCode` を ref 経由で参照するためエディタ再生成は不要。

## アーキテクチャ

```text
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
1. 危険な psql メタコマンド（`\!` `\copy` `\c` `\connect` `\g` `\o` `\w` `\i` `\e` `\lo_` `\setenv`）を denylist で拒否
2. CREATEDB ロール `dbadmin`（非 superuser）で使い捨て DB `s_<uuid>` を `CREATE DATABASE`（作成者が owner）
3. その DB の public スキーマに `student`（非 superuser）の CREATE 権限を付与（owner = `dbadmin` が GRANT 可能）
4. 学習者 SQL を `student` で実行: `psql -A -F'|' -P footer=off -v ON_ERROR_STOP=1`、`PGOPTIONS` で `statement_timeout=5s` / `lock_timeout=2s` / `idle_in_transaction_session_timeout=5s`
5. `DROP DATABASE ... WITH (FORCE)`（実行後・timeout 時も別 ctx で必ず実行）

### 多層防御
- `listen_addresses=''` → **unix socket 専用**、TCP/ネットワーク無し
- **superuser(`postgres`) は pg_hba で socket 接続を `reject`** → 学習者が `\c`/`\connect` で superuser へ昇格する経路を塞ぐ。socket に出るのは `dbadmin`(CREATEDB) と `student` のみ
- 実行ロール `student` = 非 superuser / NOCREATEDB / NOCREATEROLE → `COPY ... TO/FROM PROGRAM`・サーバファイル不可（superuser 限定操作）。`dbadmin` も非 superuser なので昇格不能
- `\c`/`\connect` を denylist で弾く（pg_hba reject と二重）
- 危険拡張を入れない（plpgsql のみ）
- `statement_timeout` ほか + 提出ごと throwaway DB（提出間でデータ混在なし）
- psql には `sandboxEnv`/明示 env で秘匿情報を渡さない（PG 接続 env も `isSensitiveEnvKey` で他言語へは遮断）

### ログと incident 監視
- 学習者の SQL ミス（構文/権限エラー）は postgres が `ERROR:` で server log に出すが、`/ecs/frestyle-prod` のサブスクリプションフィルタ `?ERROR ?Exception` が拾って Slack #incident を**誤報で埋める**ため、`postgresql.conf` で **`log_min_messages = log`**（+ `log_min_error_statement = panic`）にして client エラーを server log に出さない。
- 学習者には**クライアント応答としてエラーは返る**ので学習体験は維持される。`LOG:`/`FATAL:` は残すので起動診断・真の障害は見える（どちらも `ERROR`/`Exception` を含まないので incident は鳴らない）。
- 将来の堅牢化（infra）: incident サブスクリプションフィルタを backend の構造化ログ `{ $.level = "ERROR" }` 一致に絞る / coderunner ログを別ロググループへ分離。

### 接続設定（env / 既定はコンテナの socket）
`CODE_PG_HOST`（既定 `/var/run/postgresql`、コンテナでは `/tmp/pgsock`）/ `CODE_PG_PORT` / `CODE_PG_ADMIN`（DB 作成用 CREATEDB ロール・既定 `dbadmin`）/ `CODE_PG_STUDENT`（既定 `student`）/ `CODE_PG_*_PASSWORD`。
`Dockerfile.coderunner` が build 時に `initdb` + 極小チューニング（`shared_buffers=32MB` / `fsync=off` 等、使い捨て前提）+ `dbadmin`/`student` ロール作成 + pg_hba（postgres reject）を済ませ、entrypoint が socket 専用で `pg_ctl start` する。

### 出力フォーマットと出題規約
- 実行結果は `psql -A -F'|' -P footer=off` の**ヘッダ付きパイプ区切り**で stdout に出る:
  ```text
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
- [backend/internal/infra/sandbox/runner_sql_test.go](../backend/internal/infra/sandbox/runner_sql_test.go): `initdb`+`pg_ctl` で使い捨て PG（`dbadmin`/`student` ロール）をローカル起動して executeSQL を実検証。`initdb`/`psql` が無い環境では Skip
  - `Test_ランナー_SQL_単純SELECT` / `Test_ランナー_SQL_複数文と集計` / `Test_ランナー_SQL_文タイムアウトで打ち切る` / `Test_ランナー_SQL_提出間はDBが隔離される` / `Test_ランナー_SQL_COPYPROGRAMは権限拒否` / `Test_ランナー_SQL_危険メタコマンドを拒否`（`\!` / `\c` / `\connect`）

## 言語フィルタ UI（FRESTYLE-101）

演習一覧（/code-editor）の言語絞り込みは、プルダウン（select）ではなく
コース一覧のカテゴリ絞り込みと同じ**チップ型セレクタ**（常時見える一覧）で行う。

- チップ本体は共有コンポーネント `src/components/ui/FilterChip.tsx`
  （コース一覧 FRESTYLE-68 のローカル実装を昇格。`aria-pressed` + `role="group"`）
- 選択肢と有効値の単一情報源は `src/constants/exerciseLanguages.ts`。
  `useExerciseList` の localStorage 復元時の検証セットもここから導出する
  （チップの選択肢と検証セットの二重管理を防ぐ。言語を増やすときはこの定数だけ触る）
- 操作感はコース一覧と統一: 「すべて」+ 各言語、**アクティブなチップの再クリックで「すべて」に戻る**
- 初期言語は localStorage 復元（既定 PHP）、絞り込みはサーバーサイド
  （`GET /api/v2/exercises?language=`）という既存仕様は変更していない
