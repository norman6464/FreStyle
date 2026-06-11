# コード演習（master_exercises）と コードエディタ

学習者がブラウザのエディタでコードを書いて「実行」「提出（採点）」する演習機能の技術メモ。実行基盤は code-runner サイドカー（[`design/2026年/0004-code-runner-separation.md`](./design/2026年/0004-code-runner-separation.md) / infra docs/36）。

## 1. 演習データモデル

`domain.MasterExercise`（テーブル `master_exercises`）が 1 問を表す。

| 項目 | 意味 |
|---|---|
| `Language` | `php` / `go` / `bash`（実行系）。`sql` / `javascript` 等は表示のみ |
| `Mode` | `execute`=実行して stdout 比較 / `qa`=提出文字列と `ExpectedOutput` を trim 比較 |
| `StarterCode` | エディタ初期表示。学習者が編集する雛形 |
| `ExpectedOutput` | 期待出力（trim 比較） |
| `Category` / `Difficulty` / `OrderIndex` / `IsPublished` | 一覧の分類・並び・公開 |
| `Explanation` | 正解後に表示する解説 |

採点は `MasterExerciseExample`（テーブル `master_exercise_examples`）の各テストケース（`InputText`=stdin → `ExpectedOutput`）でコードを実行し、stdout を比較する。`execute` モードでは sandbox（code-runner）で実際に走らせる。

## 2. 演習を追加する

言語別演習（PHP / Go / Docker / Linux / Git など）の**正本は非公開の教材リポ** [`norman6464/frestyle-teaching-materials`](https://github.com/norman6464/frestyle-teaching-materials) の `exercises/<lang>/*.md`。問題文・期待出力を公開リポに露出させないため、本体（公開リポ）には埋め込まない。

- **追加 / 編集**: 教材リポの `exercises/<lang>/<slug>.md`（YAML frontmatter + 本文）を編集する
- **DB 反映**: 2 つの作業ディレクトリで順に実行する（公開リポでの誤実行を避けるため実行場所を明示）
  1. **教材リポ直下**（`frestyle-teaching-materials/`）で SQL を生成: `python3 exercises/_scripts/seed.py > /tmp/seed-exercises.sql`（slug をキーにした `ON CONFLICT (slug) DO UPDATE` の UPSERT SQL）
  2. **インフラリポ直下**（`frestyle-infrastructure/`）で Supabase に流す: `make apply-migration-supabase FILE=/tmp/seed-exercises.sql DATABASE_URL_SECRET_NAME=frestyle-prod/database-url`（非破壊・冪等）
- **採点**: `master_exercise_examples` が無い演習は `master_exercises.expected_output` を単一の仮想テストケースとして使う（seed.py は examples を作らず expected_output のみ投入する）
- company_admin が UI から作成することも可能。その場合は後で教材リポの `.md` にバックポートして整合を取る

> **例外（本体コードと密結合な演習のみ埋め込み）**: `backend/internal/infra/database/seed_master_exercises.go` は起動時にクリーンアーキテクチャ体験用の Go 演習（`seedCleanArchitectureExercise` / slug `go-clean-arch-greeting`）だけを冪等に投入する。本プロジェクトの層構造そのものを最小例で示す教材のため例外的に埋め込んでいる。PHP 演習の埋め込み seed は教材リポ `.md` への一本化に伴い撤去済み。

### 参照: クリーンアーキテクチャの Go 演習

`seedCleanArchitectureExercise`（slug `go-clean-arch-greeting`）は、**1 ファイルの Go で依存性逆転 (DIP) を体験**する演習。domain（エンティティ）/ port（インターフェース）/ usecase / infra（実装）/ main（wiring）を 1 ファイルに置き、`GreetUseCase.Execute` の未実装部分を学習者が埋める。期待出力 `Hello, FreStyle! (clean architecture)` と一致で正解。「usecase は具体実装でなく port に依存し、実装は main で注入する」という本プロジェクトのクリーンアーキテクチャ（依存方向 handler→usecase→repository/infra→domain）を最小例で示す教材。

> 単一ファイル実行（sandbox は 1 ファイルを `go run`）なので、複数パッケージに分けず**役割をコメントで区切って 1 ファイルに収める**のがコツ。

## 3. コードエディタと実行ショートカット

エディタは `frontend/src/components/CodeEditor.tsx`（Monaco をバンドル直読み）。演習画面は `pages/ExerciseDetailPage.tsx` + `hooks/useExerciseDetail.ts`。

- **実行ショートカット**: エディタ上で **Ctrl+Enter（Mac は Cmd+Enter）** で実行できる。`CodeEditor` の `onRun` prop に `runCode` を渡し、Monaco の `addCommand(KeyMod.CtrlCmd | KeyCode.Enter, ...)` で発火する。`onRun` は ref で保持してエディタ再生成を避ける
- **入場時ウォームアップ**: 演習詳細を開くと `useExerciseDetail` が対象言語で `warmup` を fire-and-forget し、「実行環境 準備完了」を表示（Go のコンパイルキャッシュ温め。code-runner 常駐で原理的に warm）
- 実行（`runCode`）は `detail` 未取得 / 空コード / 実行中はガードされるため、ショートカット連打は安全
