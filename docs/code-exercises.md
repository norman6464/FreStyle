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
| `javascript` | `node`（単一ファイル） | コンパイル不要で起動が速い。最小 env + 独立 process group（下記） |
| `typescript` | `node --experimental-transform-types`（単一ファイル） | Node 組み込みの型除去で実行（tsc 等の別工程なし）。下記 |

共通制約: コード 64KB / 出力 64KB 上限、言語別 timeout、`sandboxEnv` が AWS/DB/Cognito 等の機密 env を子プロセスから除去。

## JavaScript / TypeScript 実行系（Node.js）— FRESTYLE-110

Go 演習は `go run` のコンパイルを伴い初回実行が重い。起動が速いスクリプト言語の選択肢として
Node.js による `javascript` / `typescript` 実行を追加した（[`executeNode`](../backend/internal/infra/sandbox/runner.go)）。

- **TypeScript は Node 組み込みの型除去（type stripping）でそのまま実行**する。`main.ts` の拡張子 +
  `--experimental-transform-types`（enum 等、型除去だけでは動かない構文も変換）で起動し、
  **`--disable-warning=ExperimentalWarning`** で実験機能の警告が学習者の stderr に混ざるのを防ぐ
  - 型チェックはしない（型エラーでも動く構文なら実行される）。採点は他言語と同じ stdout 比較
- bash と同じ隔離方針: **os.Environ を継承せず最小 env のみ**（PATH/HOME/PWD/LANG。`NODE_OPTIONS`
  注入も防げる）、HOME/PWD はリクエストごとの temp dir、独立 process group で timeout 時に
  `child_process` の子孫ごと SIGKILL、timeout 8s
- `--max-old-space-size=128` でヒープ上限を絞る（512MB タスクに PG 等と同居するため）
- stderr のスタックトレースに出る temp dir の絶対パスは `./` に整形して内部パスを隠す
- ランタイムは `Dockerfile.coderunner` が **公式 node:24-bookworm イメージから `node` バイナリだけ** を
  取り込む。npm は同梱しない（演習実行に不要・サンドボックス内の攻撃面を増やさない）。
  backend 本体イメージには Node を追加しない

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
- [backend/internal/infra/sandbox/runner_test.go](../backend/internal/infra/sandbox/runner_test.go): php / go / bash / javascript / typescript の実実行テスト。ランタイムが PATH に無い環境では Skip（TS は `--experimental-transform-types` 未対応の古い node でも Skip）
  - JS/TS 分: HelloWorld / 構文エラーで内部パス隠蔽 / stdin / 終了コード伝播 / 親 env 遮断 / 無限ループ timeout / 型注釈 / enum（transform-types の regression）/ ExperimentalWarning 非混入
- [backend/internal/infra/sandbox/runner_sql_test.go](../backend/internal/infra/sandbox/runner_sql_test.go): `initdb`+`pg_ctl` で使い捨て PG（`dbadmin`/`student` ロール）をローカル起動して executeSQL を実検証。`initdb`/`psql` が無い環境では Skip
  - `Test_ランナー_SQL_単純SELECT` / `Test_ランナー_SQL_複数文と集計` / `Test_ランナー_SQL_文タイムアウトで打ち切る` / `Test_ランナー_SQL_提出間はDBが隔離される` / `Test_ランナー_SQL_COPYPROGRAMは権限拒否` / `Test_ランナー_SQL_危険メタコマンドを拒否`（`\!` / `\c` / `\connect`）

## 実行結果プレビューの 3 状態表示（FRESTYLE-111）

「コード実行」の結果ステータスは、exit 0 なら常に緑だと「エラーなく動いた ＝ 正解」と色の印象で誤解されやすい。
[`ExecutionResultTable`](../frontend/src/components/exercise/ExecutionResultTable.tsx) は stdout をサーバ採点と同じ正規化
（改行コード統一・行末空白/末尾改行の除去）で期待出力とプレビュー比較し、**緑を「一致したときだけ」に予約**する:

- exit 0 + 出力一致 → 緑 ✓「実行成功・期待する出力と一致」
- exit 0 + 出力不一致 → 琥珀 ⚠「実行成功（エラーなし）・期待する出力とはまだ一致していません」
- exit != 0 → 赤 ✗「実行エラー（exit N）」
- 期待出力が空の演習は比較不能なので中立の「実行成功（エラーなし）」にフォールバック

正誤の**確定**は従来どおり提出時のサーバ側採点（複数テストケース）で、この比較は画面に表示中の期待出力 1 件とのプレビュー。

## 実行エラーの行マーカーとエラーメッセージ行（FRESTYLE-117）

エラーになったとき「どの行が原因か」を自力で stderr から読み解かなくて済むようにする。

- [`utils/executionErrors.ts`](../frontend/src/utils/executionErrors.ts) が stderr から行番号を言語別に抽出する
  （go: `main.go:N` / javascript・typescript: `main.js|ts:N` / php: `on line N` / bash: `script.sh: line N`。
  実行系が内部パスを `./main.go` 等に整形しているため行番号はエディタとそのまま一致する）。
  node のスタックトレース形式は例外行（`SyntaxError: ...`）をホバー用メッセージに添える
- `CodeEditor` の `errorMarkers` prop が monaco の `setModelMarkers`（赤波線 + ホバーでメッセージ）と
  `createDecorationsCollection`（ガターの赤丸 ✕ + 行の薄い赤ハイライト。スタイルは index.css）を適用。
  ガター余白はマーカーがあるときだけ確保する。行番号は 1〜行数にクランプ
- `ExecutionResultTable` は stderr を「エラーメッセージ」行に分離（Go のコンパイル失敗 =
  stderr に `main.go:N` があるときはラベルを「コンパイル時エラーメッセージ」に）。
  テーブルは「実行結果ステータス / エラーメッセージ / 提出コードのアウトプット / 期待する出力」の構成

## 演習一覧の視認性（FRESTYLE-112）

いずれも「バッジ類が淡くて見えにくい」というユーザー要望への対応:

- **カテゴリ見出し**: 一時カテゴリ名ハッシュの色付きバッジにしたが（FRESTYLE-112）、
  **FRESTYLE-121 でユーザー要望により撤回**し、無色のテキスト見出しに戻した
- **ステータスバッジ**: 解いた = `bg-emerald-600` / 取り組み中 = `bg-amber-600`（塗り + 白文字。淡色 /15 は白背景で薄すぎた）
- **言語バッジ**: 背景 /15 → /25・枠 /30 → /50 にコントラスト強化（色相は FRESTYLE-109 のまま）。表記は全大文字をやめ先頭のみ大文字（FRESTYLE-121）

## 言語選択カード → 問題一覧（FRESTYLE-152。旧: 言語フィルタチップ FRESTYLE-101）

全言語の問題を 1 画面に縦積みしていて見通しが悪かったため、**「言語を選ぶ → その言語の問題一覧」**の
2 段構成にした（学習サービスで一般的な入口）。

### 画面と URL

| URL | 画面 | 中身 |
|---|---|---|
| `/code-editor` | [ExerciseLanguageSelectPage](../frontend/src/pages/ExerciseLanguageSelectPage.tsx) | 言語カードのグリッド（ロゴ + 問題数 + 進捗バー + はじめる/続きから） |
| `/code-editor/lang/:language` | [ExerciseListPage](../frontend/src/pages/ExerciseListPage.tsx) | その言語の問題一覧（無限スクロール） |
| `/code-editor/:slug` | ExerciseDetailPage | 問題 + エディタ |

`/code-editor/lang/:language` は 2 セグメントなので 1 セグメントの `:slug` とルート衝突しない。

### 進捗の集計 API

`GET /api/v2/exercises/summary` が公開済み問題を言語ごとに集計し `{ language, total, solved }` を返す
（`solved` は current user が 1 回でも正解した問題数。未ログインは 0）。問題本文を含まないので一覧 API より軽い。

- repository: `SummaryByLanguage`（1 クエリ。`master_exercises` ⟕ 自分の提出の `BOOL_OR(is_correct)`）
- usecase: `GetExerciseLanguageSummaryUseCase` / handler: `MasterExerciseHandler.Summary`
- frontend: [useExerciseLanguageSummary](../frontend/src/hooks/useExerciseLanguageSummary.ts) が取得し、
  表示名を `exerciseLanguages.ts` から解決する

### 並び順・未知言語の扱い

- カードの並びは `EXERCISE_LANGUAGES` の**定義順**（学習の入口として見せたい順）。定数に無い言語は
  後ろに言語名順で続き、表示名は key をそのまま出す（教材が先に増えても壊れない）
- 問題が 0 件の言語は API が返さないのでカードにも出ない

### 言語の選択状態

対象言語は **URL が単一の正**。旧実装の「チップ + localStorage 復元」は撤去した
（戻る / 共有 / リロードで状態がぶれないため）。`useExerciseList(language)` は引数の言語で取得するだけ。

### アイコン

言語ロゴは Devicon（MIT）の SVG を vendoring。詳細は [exercise-language-icons.md](./exercise-language-icons.md)。
