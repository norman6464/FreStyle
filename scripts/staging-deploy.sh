#!/usr/bin/env bash
#
# ステージング(Lightsail)へデプロイする。SSH してリポジトリを git pull し、ビルド・再起動する。
# 「リリース前確認」のための環境を最新化するのが目的。本番(ECS / S3+CloudFront)には触れない。
#
# 前提（環境に合わせて調整）:
#   - Lightsail インスタンスに SSH でき、リポジトリが clone 済み
#   - インスタンス側に docker-compose.staging.yml がある（無ければ再起動手順を編集）
#
# 使い方:
#   STAGING_HOST=ec2-user@<lightsail-ip> \
#   STAGING_SSH_KEY=~/.ssh/frestyle-staging.pem \
#   STAGING_DIR=/home/ec2-user/FreStyle \
#   ./scripts/staging-deploy.sh [branch]   # branch 既定: main
#
set -euo pipefail

: "${STAGING_HOST:?STAGING_HOST is required (例: ec2-user@<lightsail-ip>)}"
SSH_KEY="${STAGING_SSH_KEY:-}"
STAGING_DIR="${STAGING_DIR:-~/FreStyle}"
BRANCH="${1:-main}"

ssh_opts=(-o StrictHostKeyChecking=accept-new)
[ -n "$SSH_KEY" ] && ssh_opts+=(-i "$SSH_KEY")

echo ">>> staging デプロイ: host=$STAGING_HOST dir=$STAGING_DIR branch=$BRANCH"

ssh "${ssh_opts[@]}" "$STAGING_HOST" bash -se <<EOF
set -euo pipefail
cd "$STAGING_DIR"
echo ">>> [remote] git pull"
git fetch origin --prune
git checkout "$BRANCH"
git pull --ff-only origin "$BRANCH"
echo ">>> [remote] rebuild & restart"
if [ -f docker-compose.staging.yml ]; then
  docker compose -f docker-compose.staging.yml up -d --build
  docker compose -f docker-compose.staging.yml ps
else
  echo "!! docker-compose.staging.yml が無いため再起動はスキップしました。"
  echo "!! Lightsail の実構成（docker / systemd / nginx 等）に合わせてこの部分を編集してください。"
fi
EOF

echo ">>> staging デプロイ完了。次に: make staging-verify STAGING_URL=https://<staging-url>"
