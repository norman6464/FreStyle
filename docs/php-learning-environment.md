# PHP コード実行環境

## 概要

FreStyle に組み込まれたブラウザ上で PHP コードを記述・実行できる学習環境です。
Monaco Editor（VS Code と同じエンジン）を使用し、PHP 演習問題を選択してその場でコードを実行・検証できます。

対象: 株式会社freestyle の新卒エンジニア向け PHP 基礎教材

---

## アーキテクチャ

```
ブラウザ (Monaco Editor)
  → POST /api/v2/php/execute (JSON: { code, language })
  → Go バックエンド (ExecuteCodeUseCase)
  → php CLI (子プロセス、制限付きサンドボックス)
  → 実行結果を JSON で返却
```

### コンポーネント構成

| 層 | ファイル | 役割 |
|---|---|---|
| Page | `frontend/src/pages/CodeEditorPage.tsx` | 演習選択・エディタ・実行結果の UI |
| Component | `frontend/src/components/CodeEditor.tsx` | Monaco Editor ラッパー |
| Hook | `frontend/src/hooks/usePhpEditor.ts` | 状態管理・API 呼び出し |
| Repository | `frontend/src/repositories/PhpRepository.ts` | API ラッパー |
| Handler | `backend/internal/handler/php_handler.go` | HTTP エンドポイント |
| UseCase | `backend/internal/usecase/execute_code_usecase.go` | PHP 実行サンドボックス |
| UseCase | `backend/internal/usecase/list_php_exercises_usecase.go` | 演習一覧取得 |
| Domain | `backend/internal/domain/php_exercise.go` | 演習問題エンティティ |

---

## API エンドポイント

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/api/v2/php/exercises` | 演習問題一覧 |
| GET | `/api/v2/php/exercises/:id` | 演習問題詳細 |
| POST | `/api/v2/php/execute` | PHP コードを実行 |

### POST /api/v2/php/execute

**リクエスト:**
```json
{
  "code": "<?php echo 'Hello, World!\\n';",
  "language": "php"
}
```

**レスポンス:**
```json
{
  "stdout": "Hello, World!\n",
  "stderr": "",
  "exitCode": 0
}
```

---

## セキュリティ設計

Go バックエンドが PHP コードをサンドボックス実行する際、以下の制限を適用しています:

| 制限 | 設定値 | 目的 |
|---|---|---|
| 実行タイムアウト | 5秒 | 無限ループ防止 |
| メモリ制限 | 32 MB | リソース枯渇防止 |
| open_basedir | `/tmp` のみ | ファイルアクセス制限 |
| disable_functions | exec, system, shell_exec 等 35 関数 | OS 操作・ネットワーク通信禁止 |
| コードサイズ上限 | 64 KB | 過大なペイロード防止 |
| 出力サイズ上限 | 64 KB | 大量出力の切り詰め |

**注意:** 教材内の制御された演習問題用途を前提とした制限です。任意コード実行プラットフォームとして公開する場合は、さらに Docker による完全分離を推奨します。

---

## 演習問題一覧 (12 問)

| # | カテゴリ | タイトル |
|---|---|---|
| 1 | 基礎 | こんにちは世界 |
| 2 | 基礎 | 変数と文字列 |
| 3 | 基礎 | 数値計算 |
| 4 | 制御構文 | 条件分岐（if文） |
| 5 | 配列 | 配列の基本 |
| 6 | 制御構文 | for文 |
| 7 | 制御構文 | while文 |
| 8 | 配列 | foreach文と連想配列 |
| 9 | 関数 | 関数の定義と呼び出し |
| 10 | 文字列 | 文字列操作 |
| 11 | OOP | クラスの基本 |
| 12 | OOP | 継承 |

---

## 本番環境への反映手順

### 演習問題のシード投入

初回のみ、EC2 踏み台サーバー経由で RDS にシード SQL を流し込む:

```bash
# EC2 踏み台サーバーで実行
mysql -h $RDS_HOST -u $RDS_USER -p$RDS_PASSWORD $RDS_DATABASE < migrations/php_exercises_seed.sql
```

### バックエンドのデプロイ

```bash
# GitHub Actions: cd-backend.yml を手動トリガー
gh workflow run cd-backend.yml -f confirm=deploy
```

Dockerfile が `debian:bookworm-slim + php-cli` に変更されているため、新しいイメージが ECR に push される。

### フロントエンドのデプロイ

```bash
gh workflow run cd-frontend.yml -f confirm=deploy
```

---

## ローカル開発

```bash
# バックエンド (php が必要)
brew install php   # macOS の場合

# フロントエンド
cd frontend && npm run dev
```

`php` CLI が PATH に無い場合、`ExecuteCodeUseCase` の実行テストはスキップされます (`t.Skip`)。
