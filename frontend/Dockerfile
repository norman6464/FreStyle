# syntax=docker/dockerfile:1
#
# ステージング(Lightsail)用に frontend を nginx 静的配信コンテナ化する Dockerfile。
# 本番は S3 + CloudFront 配信のためこの Dockerfile は使わない（ステージング検証専用）。
#
# VITE_* はビルド時にバンドルへ焼き込まれるため build-arg で受け取る。
# サブドメインごとに戻り先が変わるため、deploy 時に
# VITE_OIDC_REDIRECT_URI をそのサブドメインの値で渡してビルドする。
# API は同一ホストの /api/* を Caddy が backend に振るので VITE_API_BASE_URL は
# 空（相対パス）でよい。

# --- Build stage ---
FROM node:22-bookworm-slim AS builder
WORKDIR /app

# pnpm は Node に同梱の corepack から、package.json の packageManager が指す版で用意する
# （端末・CI・ここで同じ版に揃える）。
RUN corepack enable

# 依存だけ先に入れてレイヤキャッシュを効かせる。
# pnpm-workspace.yaml も要る（install スクリプトを許すパッケージの一覧がここにある）。
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .

# Vite に焼き込む公開設定（すべて build-arg 経由）。
ARG VITE_API_BASE_URL=""
ARG VITE_OIDC_AUTHORIZE_URI=""
ARG VITE_OIDC_CLIENT_ID=""
ARG VITE_OIDC_REDIRECT_URI=""
ARG VITE_OIDC_END_SESSION_URI=""
# 既定に profile / email / offline_access を含める。
# openid だけだと、発行者は id_token に名前もメールも載せず、
# offline_access が無ければ更新用のトークンをそもそも発行しない
# （アクセストークンが切れた瞬間に全員ログイン画面へ飛ぶ）。
ARG VITE_OIDC_SCOPE="openid profile email offline_access"

ENV VITE_API_BASE_URL=$VITE_API_BASE_URL \
    VITE_OIDC_AUTHORIZE_URI=$VITE_OIDC_AUTHORIZE_URI \
    VITE_OIDC_CLIENT_ID=$VITE_OIDC_CLIENT_ID \
    VITE_OIDC_REDIRECT_URI=$VITE_OIDC_REDIRECT_URI \
    VITE_OIDC_END_SESSION_URI=$VITE_OIDC_END_SESSION_URI \
    VITE_OIDC_SCOPE=$VITE_OIDC_SCOPE

RUN pnpm run build

# --- Runtime stage ---
FROM nginx:1.27-alpine
COPY nginx.staging.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
