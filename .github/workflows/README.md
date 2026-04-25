# GitHub Actions ワークフロー構成

CI と CD を **完全に分離** しています。テスト・ビルド検証は自動、デプロイは明示的トリガーのみです。

## ワークフロー一覧

| ファイル | 種別 | トリガー | やること |
|---|---|---|---|
| `ci-backend.yml` | CI | PR / push to main（`FreStyle/**`） | `./gradlew test` + Docker image ビルド検証 |
| `ci-frontend.yml` | CI | PR / push to main（`frontend/**`） | `npm test` + `npm run build` |
| `cd-backend.yml` | CD | **workflow_dispatch のみ** + tag `release/v*` | ECR push + ECS deploy |
| `cd-frontend.yml` | CD | **workflow_dispatch のみ** + tag `release/v*` | S3 sync + CloudFront invalidation |

## 設計方針

### 1. CI と CD を分離
- 旧: `back-deploy.yml` / `front-deploy.yml` が `push to main` で test → build → deploy を一気に実行
- 新: CI（テスト・検証）と CD（デプロイ）を別ファイルに分離。**CD は AWS リソースを触る**ため、明示的なトリガーでのみ動かす

### 2. 通常 push では CD は動かない
- main にマージしただけではデプロイされない
- ドキュメント更新やリファクタなど「動作に影響しない変更」で誤って本番デプロイされることがない
- デプロイしたいときは **手動で workflow_dispatch を起動** するか **`release/v*` タグを push**

### 3. 手動実行時の二重確認
- `workflow_dispatch` の入力欄に `deploy` と入力しないと先に進まない
- 誤クリックでデプロイされない

## デプロイ手順

### A. 手動デプロイ（普段はこちら）

1. GitHub UI: Actions → 対象ワークフロー（`CD - Backend Deploy to ECS` 等）を開く
2. 「Run workflow」ボタン → ブランチ選択（通常 `main`）
3. `confirm` 欄に **`deploy`** と入力 → Run
4. 完了を待つ

### B. リリースタグ付与でのデプロイ

```bash
# main を最新に
git checkout main
git pull

# リリースタグ付与
git tag release/v1.2.3
git push origin release/v1.2.3
```

タグ push をフックに `cd-backend.yml` / `cd-frontend.yml` が自動実行される。

### C. CLI でのデプロイ（`gh` 使用）

```bash
# Backend
gh workflow run cd-backend.yml --ref main -f confirm=deploy

# Frontend
gh workflow run cd-frontend.yml --ref main -f confirm=deploy

# 進捗確認
gh run list --workflow=cd-backend.yml --limit 5
```

## CI と CD のスコープまとめ

| ワークフロー | テスト | Docker ビルド | ECR push | ECS deploy | S3 sync | CF invalidation |
|---|:-:|:-:|:-:|:-:|:-:|:-:|
| ci-backend | ✅ | ✅ (verify) | ❌ | ❌ | – | – |
| ci-frontend | ✅ | – | – | – | ❌ | ❌ |
| cd-backend | – | ✅ | ✅ | ✅ | – | – |
| cd-frontend | – | – | – | – | ✅ | ✅ |

## トラブル: CD が古い image を取って来てしまう

ECS service は task-definition で参照されている image tag を動的に解決するので、`:latest` を使い回しているとロールバックが面倒。`cd-backend.yml` は **`:${{ github.sha }}` 付きでも push** しているので、`task-definition.json` を SHA 指定に書き換えれば immutable deploy が可能。
