#!/usr/bin/env bash
#
# 認可の設定と、宣言した意図（AUTH_MODE）が噛み合っているかを確かめる。
#
# 「設定が空ならビルドを止める」だけにしていた頃は、発行者がまだ無いあいだ
# フロントを一切出せなかった。かといって素通しすると、**押しても何も起きない
# ログインボタン**が本番に出る。そこで、止めるのをやめる代わりに
# 「止まった状態で出す」ことを名前の付いた選択肢にして、記録を残す。
#
# 押せて何も起きない状態を作らせないのは、ここではなく型の役目
# （frontend/src/features/auth の useOidcLogin が返す合併に、押せる枝が無い）。
# このスクリプトが見るのは **宣言と実態が食い違っていないか** だけ。
#
# ワークフローに直接書かず切り出してあるのは、4 通りの組み合わせを CI で
# 毎回試せるようにするため（.github/workflows/ci-frontend.yml）。
# シェルに書いた分岐は、書いた時に 1 度手で試して終わりになりやすい。
#
# 入力（環境変数）:
#   AUTH_MODE                 configured | unconfigured
#   SITE_ORIGIN               配信先のオリジン（例 https://frestyle.jp）
#   VITE_OIDC_AUTHORIZE_URI   認可要求の宛先
#   VITE_OIDC_CLIENT_ID       クライアント ID
#   VITE_OIDC_REDIRECT_URI    ログイン後の戻り先
#
# 終了コード: 0 = 進んでよい / 1 = 止める
set -euo pipefail

AUTH_MODE="${AUTH_MODE:-configured}"
SITE_ORIGIN="${SITE_ORIGIN:-https://frestyle.jp}"
SUMMARY="${GITHUB_STEP_SUMMARY:-/dev/null}"

# 変数名は必ず波括弧で囲む。直後に全角文字が来ると、古い bash が
# それを変数名の一部として読み「unbound variable」で落ちる（実際に踏んだ）。
missing=""
for k in VITE_OIDC_AUTHORIZE_URI VITE_OIDC_CLIENT_ID VITE_OIDC_REDIRECT_URI; do
  eval "v=\${$k:-}"
  [ -n "$v" ] || missing="${missing} $k"
done

case "$AUTH_MODE" in
  configured | unconfigured) ;;
  *)
    echo "::error::AUTH_MODE は configured か unconfigured のどちらかです（受け取った値: $AUTH_MODE）"
    exit 1
    ;;
esac

if [ "$AUTH_MODE" = "configured" ] && [ -n "${missing}" ]; then
  echo "::error::認可の設定が空です:${missing} — 発行者を決めて Secrets を入れるか、ログインを止めたまま出すなら auth_mode に unconfigured を指定してください"
  exit 1
fi

# 逆向きの取り違えも落とす。「止める」と言ったのに値が揃っているのは、
# 指定を間違えたか、Secrets を入れたのに宣言を戻し忘れたかのどちらか。
if [ "$AUTH_MODE" = "unconfigured" ] && [ -z "${missing}" ]; then
  echo "::error::auth_mode に unconfigured を指定していますが、認可の設定は揃っています。止める理由がありません"
  exit 1
fi

if [ "$AUTH_MODE" = "unconfigured" ]; then
  echo "::warning::ログインを止めた状態で出します（欠けている設定:${missing}）"
  {
    echo "### ログイン: 停止"
    echo "認可の設定が無いため、ログインボタンは押せない状態で配信されます。"
    echo "欠けている設定:${missing}"
  } >> "$SUMMARY"
  exit 0
fi

# ここから先は configured かつ設定が揃っている場合だけ。
#
# **値が「有る」ことと「使える」ことは別。** 戻り先のオリジンが実際に配信する
# ホストと違うと、ここは通り、発行者の画面へ飛んで redirect_uri_mismatch で
# 初めて分かる（押して初めて分かる種類の壊れ方）。
redirect_origin=$(printf '%s' "$VITE_OIDC_REDIRECT_URI" | sed -E 's#^(https?://[^/]+).*#\1#')
if [ "$redirect_origin" != "$SITE_ORIGIN" ]; then
  echo "::error::VITE_OIDC_REDIRECT_URI のオリジンが配信先と違います（戻り先: $redirect_origin / 配信先: $SITE_ORIGIN）"
  exit 1
fi

echo "### ログイン: 有効" >> "$SUMMARY"
