# 夜間コスト節約スケジュール

## 概要

毎日 22:00 JST に ECS + ALB を destroy、翌日 07:00 JST に再作成して Cloudflare DNS を新 ALB に更新する自動スケジュールジョブです。

## 効果

| リソース | 24/7 月額目安 | 9 時間/日停止での節約 |
|---|---|---|
| ECS Fargate (2 vCPU/4GB) | ~$50 | ~$19 (37.5%) |
| ALB (時間課金) | ~$22 | ~$8 |
| **合計** | **~$72** | **~$27/月** |

RDS / DynamoDB / S3 / SQS / IAM / SG / ECR / Cognito は対象外（データ保護 or 課金影響小）。

## 仕組み

```
┌──────────────────────────────────────────────────────────────┐
│ GitHub Actions Scheduled Workflow                            │
│                                                              │
│ 22:00 JST (13:00 UTC) ─┐  scheduled-stop.yml                │
│                        ▼                                     │
│   1. CFn delete-stack  frestyle-prod-ecs                     │
│   2. CFn delete-stack  frestyle-prod-alb                     │
│                                                              │
│ 07:00 JST (22:00 UTC) ─┐  scheduled-start.yml               │
│                        ▼                                     │
│   1. CFn deploy        frestyle-prod-alb（新 DNS 払い出し）   │
│   2. CFn deploy        frestyle-prod-ecs                     │
│   3. Cloudflare API    api.normanblog.com → 新 ALB DNS       │
│   4. ECS services-stable wait + ヘルスチェック                │
└──────────────────────────────────────────────────────────────┘
```

## ワークフロー

| ファイル | トリガー | 処理 |
|---|---|---|
| [`.github/workflows/scheduled-stop.yml`](../../.github/workflows/scheduled-stop.yml) | `cron: '0 13 * * *'` (22:00 JST) | ECS + ALB destroy |
| [`.github/workflows/scheduled-start.yml`](../../.github/workflows/scheduled-start.yml) | `cron: '0 22 * * *'` (07:00 JST) | ALB + ECS deploy + Cloudflare DNS 更新 |

両方とも `workflow_dispatch` で手動起動も可能（`stop` / `start` を confirm 入力）。

## 認証

GitHub OIDC で `frestyle-prod-github-actions-role` を AssumeRole（Step 3 で導入）。`AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` 等の長寿命 Secret は不要。

## 必要な GitHub Secrets

`scheduled-start.yml` が Cloudflare DNS を更新するため、以下が必要:

| Secret | 用途 |
|---|---|
| `IAC_REPO_TOKEN` | frestyle-infrastructure (private) を clone する PAT |
| `CF_API_TOKEN` | Cloudflare API トークン（Zone:DNS:Edit 権限） |
| `CF_ZONE_ID` | normanblog.com の Zone ID |
| `COGNITO_CLIENT_ID` | ECS Task Definition のパラメータ用 |
| `COGNITO_REDIRECT_URI` / `COGNITO_TOKEN_URI` / `COGNITO_JWK_SET_URI` | 同上 |

CF_API_TOKEN / CF_ZONE_ID 未登録の場合の登録方法:

```bash
# .env から読み込む例
source .env
echo "$CF_API_TOKEN" | gh secret set CF_API_TOKEN -R norman6464/FreStyle
echo "$CF_ZONE_ID"   | gh secret set CF_ZONE_ID   -R norman6464/FreStyle
```

## 手動実行

スケジュールを待たず即時に実行する:

```bash
# 即時 stop
gh workflow run scheduled-stop.yml --ref main -f confirm=stop

# 即時 start
gh workflow run scheduled-start.yml --ref main -f confirm=start

# 進捗確認
gh run watch
```

## 一時的に停止スケジュールを無効にする

旅行中など、ずっと起動状態にしておきたい場合:

```bash
# scheduled-stop.yml を無効化
gh workflow disable "Scheduled - Stop (destroy ECS + ALB at night)" -R norman6464/FreStyle

# 再有効化
gh workflow enable "Scheduled - Stop (destroy ECS + ALB at night)" -R norman6464/FreStyle
```

## ロールバック / 緊急対応

### 朝の自動起動が失敗した

```bash
gh run list --workflow=scheduled-start.yml --limit 3
gh run view <run-id> --log-failed

# 手動で再実行
gh workflow run scheduled-start.yml --ref main -f confirm=start
```

### スケジュールごと止めたい（廃止）

両方の workflow ファイルを `git rm` して PR → マージ。

## 制約と注意

1. **RDS は停止しない**: データ保護のため。RDS インスタンスは 24/7 課金（db.t4g.micro なら ~$13/月）
2. **Cloudflare DNS 伝播**: 1〜数分は新 ALB DNS が浸透するまでアクセス不可になる場合あり（proxied=true でキャッシュ駆動）
3. **GitHub Actions 課金**: 1 日 2 実行 × 平均 5 分 = 月 ~5 時間。Free 枠（月 2,000 分）の範囲内
4. **OIDC Role の Trust Policy**: `refs/heads/main` のみ許可。スケジュール実行も main から発火するので OK
5. **ECR fre-style:latest**: ECS は常に最新タグを pull するため、毎朝 fresh なイメージで起動。古いタグに固定したい場合は `cd-backend.yml` で `:${sha}` タグを使う

## 観測

CloudWatch Logs `/ecs/frestyle-prod` で毎朝の起動ログを確認:

```bash
make logs   # frestyle-infrastructure リポで実行
```

GitHub Actions の実行履歴:

```bash
gh run list --workflow=scheduled-stop.yml -R norman6464/FreStyle
gh run list --workflow=scheduled-start.yml -R norman6464/FreStyle
```

## 関連

- [`docs/security/02-github-oidc.md`](../../../frestyle-infrastructure/docs/security/02-github-oidc.md) — OIDC 認証の設計
- frestyle-infrastructure リポ `Makefile` — 同等の `make destroy-ecs` / `make destroy-alb` / `make deploy-alb` / `make deploy-ecs` で手元から同じ操作可能
