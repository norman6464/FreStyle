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
# サンドボックスでユーザコードを実行するため、 PHP CLI / Go ツールチェイン / JDK を同居させる。
# golang:1.26-bookworm をそのまま runtime に使い、 必要なパッケージのみ apt-install。
# Java は単一ファイルソースランチャ (`java Main.java`) を使うため javac 入りの JDK が要る。
FROM golang:1.26-bookworm

RUN apt-get update && \
    apt-get install -y --no-install-recommends php-cli default-jdk-headless ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /
COPY --from=builder /out/server /server

# nonroot ユーザー (UID=65532) は distroless 互換値を踏襲。
RUN groupadd -g 65532 nonroot && useradd -u 65532 -g nonroot -s /sbin/nologin nonroot

# Go の build cache を nonroot が書ける場所に置く（共有してコンパイル高速化）。
RUN mkdir -p /tmp/go-build-cache && chown -R nonroot:nonroot /tmp/go-build-cache
ENV GOCACHE=/tmp/go-build-cache GOMODCACHE=/tmp/go-mod-cache

# GOCACHE warmup:
# 学習者の最初の `go run` で cold compile を 走らせると ECS Fargate の I/O 遅延で
# 6-8 秒かかり 初回の体験を著しく損なう。 image build 時に hello-world を 1 回
# `go run` して fmt / runtime / os 等の標準パッケージをプリコンパイルしておくと、
# 以降の cold start でも 実行は ~1 秒台に収まる。
#
# nonroot 切替前に warmup する (cache 所有者を nonroot に揃えるため最後に chown)。
RUN printf 'package main\nimport "fmt"\nfunc main() { fmt.Println("warmup") }\n' > /tmp/warmup.go && \
    GOCACHE=/tmp/go-build-cache GO111MODULE=off HOME=/tmp \
        go run /tmp/warmup.go > /dev/null 2>&1 && \
    rm /tmp/warmup.go && \
    chown -R nonroot:nonroot /tmp/go-build-cache

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/server"]
