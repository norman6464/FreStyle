# cd-backend.yml の OIDC Role 認証移行

## 概要

`cd-backend.yml` が AWS にアクセスするとき、長寿命の `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` を GitHub Secrets に保管していた構成から、GitHub OIDC + IAM Role に切り替える。`scheduled-stop.yml` / `scheduled-start.yml` と同じパターン。

## なぜ

- 長寿命キーの漏洩リスク
- ローテーションが手動で運用負荷が高い
- Step 3 で OIDC Role (`frestyle-prod-github-actions-role`) を導入済みなのに `cd-backend.yml` だけ取り残されていた
- 旧 `GitHubActions` IAM User には `secretsmanager:DescribeSecret` が無く、`Resolve dependent stack outputs` ステップで `AccessDeniedException` が出てデプロイが失敗していた

## 変更内容

`.github/workflows/cd-backend.yml`:

```yaml
env:
  AWS_ROLE_ARN: arn:aws:iam::010928196665:role/frestyle-prod-github-actions-role

permissions:
  id-token: write     # OIDC token の取得
  contents: read

# Build & Push Image / CFn Deploy 両方の job で:
- name: Configure AWS credentials (OIDC)
  uses: aws-actions/configure-aws-credentials@v4
  with:
    role-to-assume: ${{ env.AWS_ROLE_ARN }}
    role-session-name: frestyle-cd-backend-build  # / deploy
    aws-region: ${{ env.AWS_REGION }}
```

## 廃止する GitHub Secrets

| Secret | 旧用途 | 移行後 |
|---|---|---|
| `AWS_ACCESS_KEY_ID` | cd-backend が AWS にアクセスする際の credential | **不要**（OIDC Role） |
| `AWS_SECRET_ACCESS_KEY` | 同上 | **不要** |

`scheduled-*.yml` も同じ Role を使うため、両 workflow から同様に削除済み。

```bash
# 削除コマンド（移行確認後に実施）
gh secret delete AWS_ACCESS_KEY_ID -R norman6464/FreStyle
gh secret delete AWS_SECRET_ACCESS_KEY -R norman6464/FreStyle
```

## OIDC Role の権限スコープ

`frestyle-prod-github-actions-role` の trust policy は **main ブランチ** と **`refs/tags/release/*` タグ** からの AssumeRole のみ許可する。PR ブランチや別リポからは AssumeRole 不可。

許可される操作（Inline Policies）:
- `cloudformation:*` on `frestyle-*` stacks のみ
- `ecs:UpdateService` / `ecs:DescribeServices` 等
- `ecr:Get*` / `ecr:Put*`（PowerUser 経由）
- `secretsmanager:DescribeSecret` / `secretsmanager:GetSecretValue` on `frestyle-prod/*` のみ
- `iam:PassRole` on `frestyle-prod-ecs-task-execution-role` / `frestyle-prod-ecs-task-role` のみ

詳細: [`frestyle-infrastructure/docs/security/02-github-oidc.md`](../../frestyle-infrastructure/docs/security/02-github-oidc.md)

## OIDC Role のデプロイ確認

```bash
aws iam get-role --role-name frestyle-prod-github-actions-role --query 'Role.Arn' --output text
# arn:aws:iam::010928196665:role/frestyle-prod-github-actions-role
```

Role が存在しない場合は IaC リポで以下を実行:

```bash
cd ~/Desktop/FreStyle/frestyle-infrastructure
# OIDC Provider が既存なら CreateOidcProvider=false が必要
aws cloudformation deploy --region ap-northeast-1 \
  --stack-name frestyle-prod-github-oidc \
  --template-file infrastructure/cloudformation/templates/iam/github-oidc.yml \
  --parameter-overrides Environment=prod CreateOidcProvider=false \
  --capabilities CAPABILITY_NAMED_IAM --no-fail-on-empty-changeset
```

## 動作確認

```bash
gh workflow run cd-backend.yml --ref main -R norman6464/FreStyle -f confirm=deploy
gh run watch $(gh run list --workflow=cd-backend.yml -R norman6464/FreStyle --limit 1 --json databaseId -q '.[0].databaseId') -R norman6464/FreStyle
```

期待:
1. `Configure AWS credentials (OIDC)` で `Successfully assumed role` ログが出る
2. `Resolve dependent stack outputs` が DescribeSecret で AccessDenied を出さない
3. `Deploy ECS stack` → `Wait for service to stabilize` → `Verify health` がグリーン

## ロールバック

問題が起きた場合は当該コミットを revert すれば旧 IAM User 認証に戻せる。
ただし旧 IAM User `GitHubActions` には `secretsmanager:DescribeSecret` が無いため、戻すなら先に IAM User へ inline policy を追加すること。

## 関連

- [`docs/operations/scheduled-cost-savings.md`](./scheduled-cost-savings.md) — 同じ OIDC Role を使っている
- [`frestyle-infrastructure/docs/security/02-github-oidc.md`](../../frestyle-infrastructure/docs/security/02-github-oidc.md) — Role / Trust Policy の設計
