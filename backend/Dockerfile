# --- Build stage ---
# 本サービス本体をビルドする stage。 ランタイム image でも Go ツールチェインを使うため
# bookworm ベースに揃える（alpine の musl と debian の glibc は互換性がない）。
FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 静的バイナリ・最小サイズ
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# --- Runtime stage ---
# サンドボックスでユーザコードを実行するため、 PHP CLI と Go ツールチェインを同居させる。
# golang:1.26-bookworm をそのまま runtime に使い、 必要なパッケージのみ apt-install。
FROM golang:1.26-bookworm

RUN apt-get update && \
    apt-get install -y --no-install-recommends php-cli ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /
COPY --from=builder /out/server /server

# nonroot ユーザー (UID=65532) は distroless 互換値を踏襲。
RUN groupadd -g 65532 nonroot && useradd -u 65532 -g nonroot -s /sbin/nologin nonroot

# Go の build cache を nonroot が書ける場所に置く（共有してコンパイル高速化）。
RUN mkdir -p /tmp/go-build-cache && chown -R nonroot:nonroot /tmp/go-build-cache
ENV GOCACHE=/tmp/go-build-cache GOMODCACHE=/tmp/go-mod-cache

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/server"]
