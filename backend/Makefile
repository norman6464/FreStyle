# FreStyle backend - 開発 / 検証 用 Makefile
#
# 主要 target:
#   make openapi      … swag init で OpenAPI spec を 再 生成 (docs/swagger.{json,yaml})
#   make verify       … gofmt / vet / build / test / archlint を 一括 実行
#   make archlint     … クリーンアーキテクチャ依存方向ルール (CLAUDE.md §2) を 静的 検証
#   make run          … go run ./cmd/server で ローカル サーバ 起動
#
# swag CLI が PATH に なければ `go install github.com/swaggo/swag/cmd/swag@v1.16.4` を 自動 実行 する。
# バージョン を 固定 する のは spec 差分 が CLI バージョン で 変わる の を 防ぐ ため。

SWAG_VERSION := v1.16.4

# go install は GOBIN > GOPATH/bin の 優先 順位 で 配置 する。
# どちら で も 拾える よう に GOBIN が セット されて いれば そちら を 優先 する。
GO_INSTALL_DIR := $(shell go env GOBIN)
ifeq ($(GO_INSTALL_DIR),)
GO_INSTALL_DIR := $(shell go env GOPATH)/bin
endif
SWAG := $(GO_INSTALL_DIR)/swag

.PHONY: openapi
openapi:
	@if [ ! -x "$(SWAG)" ]; then \
	  echo ">>> installing swag CLI ($(SWAG_VERSION))..."; \
	  go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION); \
	fi
	@echo ">>> regenerating OpenAPI spec at docs/"
	@$(SWAG) init --generalInfo cmd/server/main.go --output docs --parseDependency --parseInternal

.PHONY: verify
verify:
	@GOFMT_TMP=$$(mktemp) && \
	  gofmt -l . | tee "$$GOFMT_TMP" && \
	  if [ -s "$$GOFMT_TMP" ]; then echo "ERROR: gofmt issues above"; rm -f "$$GOFMT_TMP"; exit 1; fi && \
	  rm -f "$$GOFMT_TMP"
	go vet ./...
	go build ./...
	go test ./...
	$(MAKE) archlint

# クリーンアーキテクチャ依存方向ルール（CLAUDE.md §2）を静的に検証する。
.PHONY: archlint
archlint:
	@echo ">>> archlint: クリーンアーキテクチャ依存方向チェック"
	@go run ./cmd/archlint .

.PHONY: run
run:
	go run ./cmd/server
