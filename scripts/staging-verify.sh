#!/usr/bin/env bash
#
# ステージングの「リリース前確認」スモーク。最低限の死活・認証境界を検査し、
# どれか失敗したら非ゼロ終了する（= 本番反映に進ませない）。
#
# 使い方:
#   STAGING_URL=https://staging.example.com \
#   [STAGING_API_URL=https://api-staging.example.com] \
#   ./scripts/staging-verify.sh
#
set -euo pipefail

: "${STAGING_URL:?STAGING_URL is required (例: https://staging.example.com)}"
API="${STAGING_API_URL:-$STAGING_URL}"

fail=0
check() { # 説明, 期待コード, URL
  local name="$1" want="$2" url="$3" code
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "$url" || echo 000)
  if [ "$code" = "$want" ]; then
    printf 'OK   %-28s (%s)\n' "$name" "$code"
  else
    printf 'NG   %-28s (got %s, want %s) %s\n' "$name" "$code" "$want" "$url"
    fail=1
  fi
}

echo ">>> staging リリース前確認: $STAGING_URL"
check "トップ (SPA)"            200 "$STAGING_URL/"
check "health"                 200 "$API/api/v2/health"
check "認証必須は Cookie 無で 401" 401 "$API/api/v2/auth/me"

if [ "$fail" -ne 0 ]; then
  echo ">>> 検証 NG。本番反映は中止してください。"
  exit 1
fi
echo ">>> 検証 OK。本番反映してよい状態です（make prod-deploy-* で起動 → GitHub の production 承認ゲートで停止）。"
