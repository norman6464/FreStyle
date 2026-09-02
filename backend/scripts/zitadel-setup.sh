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

ZITADEL_URL="${ZITADEL_URL:-http://zitadel.localhost:8081}"
PAT_VOLUME="${ZITADEL_PAT_VOLUME:-frestyle-local-zitadel-pat}"
PROJECT_NAME="${ZITADEL_PROJECT_NAME:-FreStyle}"
APP_NAME="${ZITADEL_APP_NAME:-frestyle-web}"
# 手元のフロントエンド。認可コードの戻り先。
# **SPA のルートと一致していること**（frontend/src/app/App.tsx の /login/callback）。
# ここがずれると、発行者は redirect_uri_mismatch で交換を拒む。
REDIRECT_URI="${ZITADEL_REDIRECT_URI:-http://localhost:5173/login/callback}"
LOGOUT_URI="${ZITADEL_LOGOUT_URI:-http://localhost:5173/}"
# docker-compose の初期化で作られる管理者。ここに admin の役割を与える。
ADMIN_LOGIN_NAME="${ZITADEL_ADMIN_LOGIN_NAME:-admin@frestyle.local}"

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
  "idTokenUserinfoAssertion": true,
  "idTokenRoleAssertion": true,
  "accessTokenRoleAssertion": true
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
#
# roleAssertion を 2 つ立てているのが要点。Zitadel は既定では**役割をトークンに載せない**。
# 立て忘れると、アプリ側でクレーム名を正しく読んでいても役割は常に空になり、
# 管理者判定が黙って false のままになる。エラーもログも出ないので気づけない。

# ---- プロジェクトロール ----
#
# アプリは「発行者側で admin の役割を持っているか」を見て運営管理者へ昇格させる。
# その役割がそもそも存在しないと、誰にも付けられないので昇格の経路が丸ごと死ぬ。
#
# 既にあれば 409 が返る。作成の成否は HTTP の状態コードで見る
# （本文の details を見てはいけない — エラー応答にも入っているので取り違える）。
ROLE_STATUS="$(
  curl -sS -o /dev/null -w '%{http_code}' -X POST "${ZITADEL_URL}/management/v1/projects/${PROJECT_ID}/roles" \
    -H "Authorization: Bearer ${PAT}" -H 'Content-Type: application/json' \
    -d '{"roleKey":"admin","displayName":"FreStyle 運営管理者","group":"frestyle"}'
)"
case "$ROLE_STATUS" in
  200|201) echo "プロジェクトロール admin を作った" ;;
  409)     echo "プロジェクトロール admin は既にある" ;;
  *)       echo "プロジェクトロールを作れない（HTTP ${ROLE_STATUS}）" >&2; exit 1 ;;
esac

# このプロジェクトへの認可を持つ相手にだけトークンを出す（hasProjectCheck）。
# 立てないと、インスタンス上の誰でもこのクライアント宛のトークンを取れる。
# 役割をトークンに載せる設定（projectRoleAssertion）も同時に入れる。
api PUT "/management/v1/projects/${PROJECT_ID}" "$(cat <<JSON
{
  "name": "${PROJECT_NAME}",
  "projectRoleAssertion": true,
  "projectRoleCheck": false,
  "hasProjectCheck": true,
  "privateLabelingSetting": "PRIVATE_LABELING_SETTING_UNSPECIFIED"
}
JSON
)" > /dev/null
echo "プロジェクトの設定を更新した（役割をトークンに載せる / このプロジェクトの利用者に限る）"

# ---- ログイン画面の資格 ----
#
# v4 のログイン画面は別コンテナで動き、本体の API を機械ユーザーとして叩く。
# その相手に IAM_LOGIN_CLIENT が無いと、パスワードを入れた後の
# 「認可を完了する」呼び出しだけが permission_denied で落ちる。
# 画面には "Unknown error occurred" としか出ないので、原因はログを見ないと分からない。
MACHINE_USER_ID="$(
  api POST /management/v1/users/_search '{"query":{"limit":100}}' \
    | python3 -c "
import json,sys
d=json.load(sys.stdin)
for u in d.get('result',[]):
    if u.get('userName')=='frestyle-admin':
        print(u['id']); break
" 2>/dev/null
)"
if [ -n "$MACHINE_USER_ID" ]; then
  api PUT "/admin/v1/members/${MACHINE_USER_ID}" '{"roles":["IAM_OWNER","IAM_LOGIN_CLIENT"]}' > /dev/null
  echo "機械ユーザーに IAM_LOGIN_CLIENT を付けた"
else
  echo "機械ユーザー frestyle-admin が見つからない" >&2
  exit 1
fi

# ---- 最初の管理者に admin の役割を与える ----
#
# 役割を作っただけでは誰も持っていない。アプリは「発行者側で admin を持っているか」を
# 見て運営管理者へ昇格させるので、最初の 1 人に付けておかないと、
# 手元で管理者向けの画面を一度も開けない。
ADMIN_USER_ID="$(
  api POST /management/v1/users/_search '{"query":{"limit":100}}' \
    | python3 -c "
import json,sys
d=json.load(sys.stdin)
for u in d.get('result',[]):
    if u.get('userName')=='${ADMIN_LOGIN_NAME}':
        print(u['id']); break
" 2>/dev/null
)"
if [ -n "$ADMIN_USER_ID" ]; then
  GRANT_STATUS="$(
    curl -sS -o /dev/null -w '%{http_code}' -X POST "${ZITADEL_URL}/management/v1/users/${ADMIN_USER_ID}/grants" \
      -H "Authorization: Bearer ${PAT}" -H 'Content-Type: application/json' \
      -d "{\"projectId\":\"${PROJECT_ID}\",\"roleKeys\":[\"admin\"]}"
  )"
  case "$GRANT_STATUS" in
    200|201) echo "${ADMIN_LOGIN_NAME} に admin の役割を与えた" ;;
    409)     echo "${ADMIN_LOGIN_NAME} は既に admin の役割を持っている" ;;
    *)       echo "役割を与えられない（HTTP ${GRANT_STATUS}）" >&2; exit 1 ;;
  esac
else
  echo "初期管理者 ${ADMIN_LOGIN_NAME} が見つからない（docker-compose の初期化設定を確認）" >&2
fi

# ---- ログインのふるまい ----
#
# 目指す形は「メールで入る」が中心で、パスワードは主役ではないログイン。
# ここで設定できるのは次の 4 つ（招待そのものはアプリ側の口）。
#
#   1. メールアドレスだけで入れる（コード / パスキー）
#   2. 外部の IdP でのログイン
#   3. 組織に属する ID として扱う
#   4. 後から強くできる（2 要素）
#
# ignoreUnknownUsernames を立てるのが要点のひとつ。立てないと、存在しない
# メールアドレスに対してだけ違う応答が返り、**そのアドレスが登録済みかどうかを
# 外から確かめられる**（総当たりでメンバーを割り出せる）。
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
# 認証は Zitadel（OIDC）。
#
# issuer とトークンの取得先を分けて指す。issuer はトークンの iss と突き合わせる
# 文字列で、ブラウザが見る URL と同じでなければならない。ここを JWKS の URL から
# 推測すると、発行者を替えた瞬間に「全員 401」になる（以前それで踏んだ）。
OIDC_ISSUER=${ZITADEL_URL}
OIDC_JWKS_URI=${ZITADEL_URL}/oauth/v2/keys
OIDC_TOKEN_URI=${ZITADEL_URL}/oauth/v2/token
OIDC_END_SESSION_URI=${ZITADEL_URL}/oidc/v1/end_session
OIDC_CLIENT_ID=${CLIENT_ID}
OIDC_REDIRECT_URI=${REDIRECT_URI}
# 公開クライアント（PKCE）なので client secret は持たない。
OIDC_CLIENT_SECRET=

# ---- frontend/.env へ貼る ----
VITE_OIDC_AUTHORIZE_URI=${ZITADEL_URL}/oauth/v2/authorize
VITE_OIDC_CLIENT_ID=${CLIENT_ID}
VITE_OIDC_REDIRECT_URI=${REDIRECT_URI}
VITE_OIDC_END_SESSION_URI=${ZITADEL_URL}/oidc/v1/end_session
# offline_access が無いと更新用のトークンが発行されず、
# アクセストークンが切れた瞬間に全員ログイン画面へ飛ぶ。
VITE_OIDC_SCOPE=openid profile email offline_access

# ログイン画面   ${ZITADEL_URL}/ui/v2/login
# コンソール     ${ZITADEL_URL}/ui/console
# 受信メール     http://localhost:8025
ENV
