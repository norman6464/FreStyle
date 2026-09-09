#!/usr/bin/env bash
#
# check-auth-config.sh の負例テスト。
#
# 門番は「通す」だけでなく「止める」のが仕事なので、**落ちるべきものが
# 落ちること**まで確かめる。通る側だけ見ていると、条件を 1 つ消しても気づけない。
set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GATE="${HERE}/check-auth-config.sh"

K=AIzaSyTestKeyForCheckingTheGateOnly
D=frestyle-507912.firebaseapp.com
P=frestyle-507912

failures=0

# 期待する終了コード / 説明 / AUTH_MODE / apiKey / authDomain / projectId
check() {
  local want="$1" what="$2" mode="$3" key="$4" domain="$5" project="$6"
  local got
  env AUTH_MODE="${mode}" \
    VITE_FIREBASE_API_KEY="${key}" VITE_FIREBASE_AUTH_DOMAIN="${domain}" \
    VITE_FIREBASE_PROJECT_ID="${project}" \
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

check 0 "設定あり・configured なら通す"            configured   "$K" "$D" "$P"
check 1 "設定なし・configured なら止める"          configured   ""   ""   ""
check 1 "一部だけ欠けても止める"                    configured   "$K" ""   "$P"
check 0 "設定なし・unconfigured なら通す"          unconfigured ""   ""   ""
check 1 "設定ありなのに unconfigured なら止める"   unconfigured "$K" "$D" "$P"
check 1 "知らない AUTH_MODE なら止める"            bogus        "$K" "$D" "$P"

# 「有る」だけでは通さない検査。
check 1 "authDomain に scheme が付いていたら止める" configured "$K" "https://${D}" "$P"
check 1 "authDomain にパスが付いていたら止める"     configured "$K" "${D}/__/auth"  "$P"
check 1 "authDomain と projectId が別プロジェクトなら止める" \
  configured "$K" "frestyle-prod.firebaseapp.com" "$P"
# カスタムの authDomain（.firebaseapp.com 以外）は projectId と綴りが揃わないので照合しない。
check 0 "カスタムの authDomain は projectId と照合しない" configured "$K" "auth.frestyle.dev" "$P"

if [ "${failures}" -ne 0 ]; then
  echo "::error::門番の検査 ${failures} 件が期待どおりに動かなかった"
  exit 1
fi
echo "門番はすべて期待どおりに動いた"
