# syntax=docker/dockerfile:1
#
# ステージング(Lightsail)用に frontend を nginx 静的配信コンテナ化する Dockerfile。
# 本番は S3 + CloudFront 配信のためこの Dockerfile は使わない（ステージング検証専用）。
#
# VITE_* はビルド時にバンドルへ焼き込まれるため build-arg で受け取る。
# サブドメインごとに redirect_uri（Cognito callback）が変わるため、deploy 時に
# VITE_REDIRECT_URI をそのサブドメインの値で渡してビルドする。
# API は同一ホストの /api/* を Caddy が backend に振るので VITE_API_BASE_URL は
# 空（相対パス）でよい。

# --- Build stage ---
FROM node:20-bookworm-slim AS builder
WORKDIR /app

# 依存だけ先に入れてレイヤキャッシュを効かせる。
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Vite に焼き込む公開設定（すべて build-arg 経由）。
ARG VITE_API_BASE_URL=""
ARG VITE_COGNITO_DOMAIN=""
ARG VITE_CLIENT_ID=""
ARG VITE_REDIRECT_URI=""
ARG VITE_RESPONSE_TYPE="code"
ARG VITE_SCOPE="openid"
ARG VITE_WEB_SOCKET_URL_AI_CHAT=""
ARG VITE_WEB_SOCKET_URL_CHAT=""

ENV VITE_API_BASE_URL=$VITE_API_BASE_URL \
    VITE_COGNITO_DOMAIN=$VITE_COGNITO_DOMAIN \
    VITE_CLIENT_ID=$VITE_CLIENT_ID \
    VITE_REDIRECT_URI=$VITE_REDIRECT_URI \
    VITE_RESPONSE_TYPE=$VITE_RESPONSE_TYPE \
    VITE_SCOPE=$VITE_SCOPE \
    VITE_WEB_SOCKET_URL_AI_CHAT=$VITE_WEB_SOCKET_URL_AI_CHAT \
    VITE_WEB_SOCKET_URL_CHAT=$VITE_WEB_SOCKET_URL_CHAT

RUN npm run build

# --- Runtime stage ---
FROM nginx:1.27-alpine
COPY nginx.staging.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
