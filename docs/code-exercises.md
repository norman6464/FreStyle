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

初期演習は `backend/internal/infra/database/seed_master_exercises.go` が起動時に投入する（冪等）。

- **PHP 演習**: `exercises []domain.MasterExercise` に追加。末尾の一括ループが `Language=php` / `Slug=php-{ID}` を自動設定する
- **他言語（go / bash）演習**: 一括ループは php 固定なので、**独立した seed 関数**で `Language` を明示して入れる。参照実装が `seedCleanArchitectureExercise`（slug 存在チェックで冪等 → autoIncrement で採番 → その ID で example を 1 件作成）
- company_admin が UI から作成することも可能。その場合は後で seed にバックポートして整合を取る

### 参照: クリーンアーキテクチャの Go 演習

`seedCleanArchitectureExercise`（slug `go-clean-arch-greeting`）は、**1 ファイルの Go で依存性逆転 (DIP) を体験**する演習。domain（エンティティ）/ port（インターフェース）/ usecase / infra（実装）/ main（wiring）を 1 ファイルに置き、`GreetUseCase.Execute` の未実装部分を学習者が埋める。期待出力 `Hello, FreStyle! (clean architecture)` と一致で正解。「usecase は具体実装でなく port に依存し、実装は main で注入する」という本プロジェクトのクリーンアーキテクチャ（依存方向 handler→usecase→repository/infra→domain）を最小例で示す教材。

> 単一ファイル実行（sandbox は 1 ファイルを `go run`）なので、複数パッケージに分けず**役割をコメントで区切って 1 ファイルに収める**のがコツ。

## 3. コードエディタと実行ショートカット

エディタは `frontend/src/components/CodeEditor.tsx`（Monaco をバンドル直読み）。演習画面は `pages/ExerciseDetailPage.tsx` + `hooks/useExerciseDetail.ts`。

- **実行ショートカット**: エディタ上で **Ctrl+Enter（Mac は Cmd+Enter）** で実行できる。`CodeEditor` の `onRun` prop に `runCode` を渡し、Monaco の `addCommand(KeyMod.CtrlCmd | KeyCode.Enter, ...)` で発火する。`onRun` は ref で保持してエディタ再生成を避ける
- **入場時ウォームアップ**: 演習詳細を開くと `useExerciseDetail` が対象言語で `warmup` を fire-and-forget し、「実行環境 準備完了」を表示（Go のコンパイルキャッシュ温め。code-runner 常駐で原理的に warm）
- 実行（`runCode`）は `detail` 未取得 / 空コード / 実行中はガードされるため、ショートカット連打は安全
