# --- Build stage ---
# 本サービス本体をビルドする stage。Go ツールチェインで純静的バイナリ（CGO 無効）を作る。
FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 静的バイナリ・最小サイズ（distroless/static で動かすため CGO 無効の純静的にする）。
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# --- Runtime stage ---
# distroless/static: 静的 Go バイナリのみを載せる（イメージ ~60MB）。go/php ランタイムは含めない。
# 学習者コードのサンドボックス実行は code-runner サイドカーへ HTTP 委譲する想定で、
# 本番 ECS は CODE_RUNNER_URL=http://127.0.0.1:9000 を注入する（**distroless backend は
# CODE_RUNNER_URL 必須**＝in-process 実行はできない。ローカルは `go run ./cmd/server` を使う）。
# distroless は ca-certificates / tzdata / nonroot(65532) を同梱し、HTTPS 呼び出しが可能。
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
