# --- Build stage ---
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 静的バイナリ・最小サイズ
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# --- Runtime stage ---
# PHP コード実行環境のため debian-slim + php-cli を使用。
# distroless から変更済み。コード実行時は disable_functions で危険関数を無効化する。
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends php-cli ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /
COPY --from=builder /out/server /server

# nonroot ユーザー (UID=65532) は distroless と同じ値を踏襲
RUN groupadd -g 65532 nonroot && useradd -u 65532 -g nonroot -s /sbin/nologin nonroot

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/server"]
