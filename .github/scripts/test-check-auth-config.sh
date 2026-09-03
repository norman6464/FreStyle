#!/usr/bin/env bash
#
# check-auth-config.sh の負例テスト。
#
# 門番は「通す」だけでなく「止める」のが仕事なので、**落ちるべきものが
# 落ちること**まで確かめる。通る側だけ見ていると、条件を 1 つ消しても気づけない。
set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GATE="${HERE}/check-auth-config.sh"

A=https://issuer.test/oauth/v2/authorize
C=test-client-id
R=https://frestyle.jp/login/callback

failures=0

# 期待する終了コード / 説明 / AUTH_MODE / authorize / client_id / redirect
check() {
  local want="$1" what="$2" mode="$3" auth="$4" cid="$5" redirect="$6"
  local got
  env AUTH_MODE="${mode}" SITE_ORIGIN=https://frestyle.jp \
    VITE_OIDC_AUTHORIZE_URI="${auth}" VITE_OIDC_CLIENT_ID="${cid}" \
    VITE_OIDC_REDIRECT_URI="${redirect}" \
    GITHUB_STEP_SUMMARY=/dev/null \
    bash "${GATE}" >/dev/null 2>&1
  got=$?
  if [ "${got}" != "${want}" ]; then
    echo "::error::${what}: 終了コード ${want} を期待したが ${got} だった"
    failures=$((failures + 1))
  else
    echo "  ok  ${what}"
  fi
}

check 0 "設定あり・configured なら通す"            configured   "$A" "$C" "$R"
check 1 "設定なし・configured なら止める"          configured   ""   ""   ""
check 1 "一部だけ欠けても止める"                    configured   "$A" ""   "$R"
check 0 "設定なし・unconfigured なら通す"          unconfigured ""   ""   ""
check 1 "設定ありなのに unconfigured なら止める"   unconfigured "$A" "$C" "$R"
check 1 "戻り先のオリジンが配信先と違えば止める"    configured   "$A" "$C" https://example.com/login/callback
check 1 "知らない AUTH_MODE なら止める"            bogus        "$A" "$C" "$R"

if [ "${failures}" -ne 0 ]; then
  echo "::error::門番の検査 ${failures} 件が期待どおりに動かなかった"
  exit 1
fi
echo "門番はすべて期待どおりに動いた"
