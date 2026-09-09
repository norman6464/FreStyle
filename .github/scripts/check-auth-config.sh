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
# （frontend/src/shared/lib/auth の readFirebaseAuthConfig が返す合併に、
# 設定が欠けたまま SDK を初期化できる枝が無い）。
# このスクリプトが見るのは **宣言と実態が食い違っていないか** だけ。
#
# ワークフローに直接書かず切り出してあるのは、正例・負例を CI で毎回試せるように
# するため（.github/workflows/ci-frontend.yml）。シェルに書いた分岐は、
# 書いた時に 1 度手で試して終わりになりやすい。
#
# 入力（環境変数）:
#   AUTH_MODE                  configured | unconfigured
#   VITE_FIREBASE_API_KEY      GCIP のクライアント API キー
#   VITE_FIREBASE_AUTH_DOMAIN  サインインの窓口になるホスト（例 <project>.firebaseapp.com）
#   VITE_FIREBASE_PROJECT_ID   GCP プロジェクト ID
#
# 終了コード: 0 = 進んでよい / 1 = 止める
set -euo pipefail

AUTH_MODE="${AUTH_MODE:-configured}"
SUMMARY="${GITHUB_STEP_SUMMARY:-/dev/null}"

# 変数名は必ず波括弧で囲む。直後に全角文字が来ると、古い bash が
# それを変数名の一部として読み「unbound variable」で落ちる（実際に踏んだ）。
missing=""
for k in VITE_FIREBASE_API_KEY VITE_FIREBASE_AUTH_DOMAIN VITE_FIREBASE_PROJECT_ID; do
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
  echo "::error::認可の設定が空です:${missing} — 設定を入れるか、ログインを止めたまま出すなら auth_mode に unconfigured を指定してください"
  exit 1
fi

# 逆向きの取り違えも落とす。「止める」と言ったのに値が揃っているのは、
# 指定を間違えたか、設定を入れたのに宣言を戻し忘れたかのどちらか。
if [ "$AUTH_MODE" = "unconfigured" ] && [ -z "${missing}" ]; then
  echo "::error::auth_mode に unconfigured を指定していますが、認可の設定は揃っています。止める理由がありません"
  exit 1
fi

if [ "$AUTH_MODE" = "unconfigured" ]; then
  echo "::warning::ログインを止めた状態で出します（欠けている設定:${missing}）"
  {
    echo "### ログイン: 停止"
    echo "認可の設定が無いため、ログインは止まった状態で配信されます。"
    echo "欠けている設定:${missing}"
  } >> "$SUMMARY"
  exit 0
fi

# ここから先は configured かつ設定が揃っている場合だけ。
#
# **値が「有る」ことと「使える」ことは別。**
#
# (1) authDomain は素のホスト名。ここに https:// やパスを付けて渡すと、SDK は
#     組み立てた URL でサインインの窓口を探しに行き、押して初めて壊れる。
case "$VITE_FIREBASE_AUTH_DOMAIN" in
  *://* | */*)
    echo "::error::VITE_FIREBASE_AUTH_DOMAIN はホスト名だけで指定します（受け取った値: $VITE_FIREBASE_AUTH_DOMAIN）"
    exit 1
    ;;
esac

# (2) authDomain と projectId が別プロジェクトを指していないか。
#     この構成には表示名がどちらも FreStyle の GCP プロジェクトが 2 つあり
#     （frontend 用と backend 用）、片方だけ書き換える取り違えが起こりうる。
#     既定の <projectId>.firebaseapp.com 形のときだけ照合する（カスタムの
#     authDomain を使う場合はこの形にならないので、その時は照合しない）。
case "$VITE_FIREBASE_AUTH_DOMAIN" in
  *.firebaseapp.com)
    domain_project="${VITE_FIREBASE_AUTH_DOMAIN%.firebaseapp.com}"
    if [ "$domain_project" != "$VITE_FIREBASE_PROJECT_ID" ]; then
      echo "::error::VITE_FIREBASE_AUTH_DOMAIN と VITE_FIREBASE_PROJECT_ID が別のプロジェクトを指しています（authDomain: $domain_project / projectId: $VITE_FIREBASE_PROJECT_ID）"
      exit 1
    fi
    ;;
esac

echo "### ログイン: 有効" >> "$SUMMARY"
