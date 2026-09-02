#!/usr/bin/env bash
# ローカルの Zitadel に FreStyle 用のプロジェクトと OIDC クライアントを作る。
#
#   docker compose --profile auth up -d
#   ./backend/scripts/zitadel-setup.sh
#
# 何度流しても同じ結果になる（既にあれば作らず、既存の値を返す）。
# 手作業でコンソールを触ると再現できないので、設定はここに集約する。
#
# 出力はそのまま .env に貼れる形にしてある。
set -euo pipefail

ZITADEL_URL="${ZITADEL_URL:-http://localhost:8081}"
PAT_VOLUME="${ZITADEL_PAT_VOLUME:-frestyle-local-zitadel-pat}"
PROJECT_NAME="${ZITADEL_PROJECT_NAME:-FreStyle}"
APP_NAME="${ZITADEL_APP_NAME:-frestyle-web}"
# 手元のフロントエンド。認可コードの戻り先。
REDIRECT_URI="${ZITADEL_REDIRECT_URI:-http://localhost:5173/auth/callback}"
LOGOUT_URI="${ZITADEL_LOGOUT_URI:-http://localhost:5173/}"

# PAT は zitadel イメージに cat が無いので、alpine を挟んで読む
# （docker compose exec zitadel cat ... は "file not found in $PATH" を
#   そのままトークンとして掴んでしまい、401 の原因が分からなくなる）。
pat() {
  docker run --rm -v "${PAT_VOLUME}:/pat" public.ecr.aws/docker/library/alpine:3.20 \
    cat /pat/frestyle-admin.pat 2>/dev/null | tr -d '\r\n'
}

PAT="$(pat)"
if [ -z "$PAT" ]; then
  echo "PAT を読めない。docker compose --profile auth up -d は済んでいるか？" >&2
  exit 1
fi

api() {
  local method="$1" path="$2" body="${3:-}"
  if [ -n "$body" ]; then
    curl -sS -X "$method" "${ZITADEL_URL}${path}" \
      -H "Authorization: Bearer ${PAT}" -H 'Content-Type: application/json' -d "$body"
  else
    curl -sS -X "$method" "${ZITADEL_URL}${path}" -H "Authorization: Bearer ${PAT}"
  fi
}

jqp() { python3 -c "import json,sys;d=json.load(sys.stdin);print($1)" 2>/dev/null; }

# ---- プロジェクト（既にあれば使い回す） ----
PROJECT_ID="$(
  api POST /management/v1/projects/_search '{"queries":[]}' \
    | python3 -c "
import json,sys
d=json.load(sys.stdin)
for p in d.get('result',[]):
    if p.get('name')=='${PROJECT_NAME}':
        print(p['id']); break
" 2>/dev/null
)"

if [ -z "$PROJECT_ID" ]; then
  PROJECT_ID="$(api POST /management/v1/projects "{\"name\":\"${PROJECT_NAME}\"}" | jqp "d['id']")"
  echo "プロジェクトを作った: ${PROJECT_NAME} (${PROJECT_ID})"
else
  echo "プロジェクトは既にある: ${PROJECT_NAME} (${PROJECT_ID})"
fi

[ -n "$PROJECT_ID" ] || { echo "プロジェクト ID を取れない" >&2; exit 1; }

# ---- OIDC クライアント ----
APP_JSON="$(
  api POST "/management/v1/projects/${PROJECT_ID}/apps/_search" '{"queries":[]}' \
    | python3 -c "
import json,sys
d=json.load(sys.stdin)
for a in d.get('result',[]):
    if a.get('name')=='${APP_NAME}':
        print(json.dumps(a)); break
" 2>/dev/null
)"

if [ -z "$APP_JSON" ]; then
  # 公開クライアント（PKCE）。フロントエンドは秘密を持てないので client secret は使わない。
  CREATED="$(api POST "/management/v1/projects/${PROJECT_ID}/apps/oidc" "$(cat <<JSON
{
  "name": "${APP_NAME}",
  "redirectUris": ["${REDIRECT_URI}"],
  "postLogoutRedirectUris": ["${LOGOUT_URI}"],
  "responseTypes": ["OIDC_RESPONSE_TYPE_CODE"],
  "grantTypes": ["OIDC_GRANT_TYPE_AUTHORIZATION_CODE", "OIDC_GRANT_TYPE_REFRESH_TOKEN"],
  "appType": "OIDC_APP_TYPE_USER_AGENT",
  "authMethodType": "OIDC_AUTH_METHOD_TYPE_NONE",
  "devMode": true,
  "accessTokenType": "OIDC_TOKEN_TYPE_JWT",
  "idTokenUserinfoAssertion": true
}
JSON
)")"
  CLIENT_ID="$(echo "$CREATED" | jqp "d['clientId']")"
  if [ -z "$CLIENT_ID" ]; then
    echo "クライアントを作れなかった:" >&2
    echo "$CREATED" >&2
    exit 1
  fi
  echo "OIDC クライアントを作った: ${APP_NAME}"
else
  CLIENT_ID="$(echo "$APP_JSON" | jqp "d['oidcConfig']['clientId']")"
  echo "OIDC クライアントは既にある: ${APP_NAME}"
fi

# devMode を立てているのは、手元が http で、Zitadel が既定では
# https 以外の redirect_uri を拒むため。**本番では絶対に立てない。**

# ---- ログインのふるまい（Notion に寄せる） ----
#
# Notion の認証は「メールで入る」が中心で、パスワードは主役ではない。
# 具体的には次の 5 つ。ここで作れるのは 1・2・3・5 で、4（招待）はアプリ側の口。
#
#   1. メールアドレスだけでログインできる（コード / パスキー）
#   2. Google などのソーシャルログイン
#   3. ワークスペース（＝ Zitadel の組織）に属する ID
#   4. 招待でメンバーになる
#   5. 段階的に強くできる（2 要素）
#
# ignoreUnknownUsernames を立てるのが要点のひとつ。立てないと、存在しない
# メールアドレスに対してだけ違う応答が返り、**そのアドレスが登録済みかどうかを
# 外から確かめられる**（総当たりでメンバーを割り出せる）。Notion も存在の有無で
# 応答を変えない。
#
# disableLoginWithPhone は、電話番号でのログインを塞ぐ。研修プラットフォームで
# 電話番号を持つ理由が無く、持たない情報は漏れない。
LOGIN_POLICY='{
  "allowUsernamePassword": true,
  "allowRegister": true,
  "allowExternalIdp": true,
  "forceMfa": false,
  "passwordlessType": "PASSWORDLESS_TYPE_ALLOWED",
  "hidePasswordReset": false,
  "ignoreUnknownUsernames": true,
  "disableLoginWithPhone": true,
  "disableLoginWithEmail": false,
  "secondFactors": ["SECOND_FACTOR_TYPE_OTP", "SECOND_FACTOR_TYPE_U2F"],
  "multiFactors": ["MULTI_FACTOR_TYPE_U2F_WITH_VERIFICATION"]
}'

# 組織にポリシーが無ければ作る（無いあいだはインスタンスの既定を継承している）。
# 既にあれば PUT で更新する。
#
# 成否を本文の "details" の有無で見てはいけない — **エラー応答にも details が入る**ので、
# 409（already exists）を成功と読み違える（実際に踏んだ）。HTTP の状態コードで判定する。
POLICY_STATUS="$(
  curl -sS -o /dev/null -w '%{http_code}' -X POST "${ZITADEL_URL}/management/v1/policies/login" \
    -H "Authorization: Bearer ${PAT}" -H 'Content-Type: application/json' -d "$LOGIN_POLICY"
)"
case "$POLICY_STATUS" in
  200|201)
    echo "ログインポリシーを作った（組織独自）" ;;
  409)
    api PUT /management/v1/policies/login "$LOGIN_POLICY" > /dev/null
    echo "ログインポリシーを更新した（既にあった）" ;;
  *)
    echo "ログインポリシーを設定できない（HTTP ${POLICY_STATUS}）" >&2
    exit 1 ;;
esac

cat <<ENV

# ---- .env へ貼る ----
# 認証は Zitadel（OIDC）。issuer は待ち受けと一致していること。
# ここがずれると、発行される issuer と検出文書が食い違い、
# トークン検証が「issuer が違う」で必ず落ちる。
OIDC_ISSUER=${ZITADEL_URL}
OIDC_JWKS_URI=${ZITADEL_URL}/oauth/v2/keys
OIDC_AUTHORIZE_URI=${ZITADEL_URL}/oauth/v2/authorize
OIDC_TOKEN_URI=${ZITADEL_URL}/oauth/v2/token
OIDC_CLIENT_ID=${CLIENT_ID}
OIDC_REDIRECT_URI=${REDIRECT_URI}

# ログイン画面   ${ZITADEL_URL}/ui/v2/login
# コンソール     ${ZITADEL_URL}/ui/console
ENV
