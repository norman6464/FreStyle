# FreStyle backend - 開発 / 検証 用 Makefile
#
# 主要 target:
#   make openapi      … swag init で OpenAPI spec を 再 生成 (docs/swagger.{json,yaml})
#   make verify       … gofmt / vet / build / test / archlint / apispec-lint / naminglint を 一括 実行
#   make archlint     … クリーンアーキテクチャ依存方向ルール (CLAUDE.md §2) を 静的 検証
#   make apispec-lint … Gin ルート ↔ swaggo @Router 注釈 の 整合 (CLAUDE.md §2.7) を 検証
#   make naminglint   … usecase 命名・構造規約 (CLAUDE.md §2.3) を 検証
#   make test-integration … docker-compose で PostgreSQL を起動し -tags=integration の結合テストを実行
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

SQLC_VERSION := v1.31.1
SQLC := $(GO_INSTALL_DIR)/sqlc

# sqlc 生成: queries/*.sql + schema.sql → 型安全な Go (internal/adapter/persistence/sqlcgen)。
# 生成物はコミット対象。クエリ / schema を変更したら必ず `make sqlc` で再生成する。
.PHONY: sqlc
sqlc:
	@if [ ! -x "$(SQLC)" ]; then \
	  echo ">>> installing sqlc CLI ($(SQLC_VERSION))..."; \
	  go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION); \
	fi
	@echo ">>> sqlc generate"
	@$(SQLC) generate

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
	$(MAKE) apispec-lint
	$(MAKE) naminglint

# クリーンアーキテクチャ依存方向ルール（CLAUDE.md §2）を静的に検証する。
.PHONY: archlint
archlint:
	@echo ">>> archlint: クリーンアーキテクチャ依存方向チェック"
	@go run ./cmd/archlint .

# Gin ルート ↔ swaggo @Router 注釈の整合（CLAUDE.md §2.7）を検証する。
.PHONY: apispec-lint
apispec-lint:
	@echo ">>> apispec-lint: ルート ↔ swaggo 注釈チェック"
	@go run ./cmd/apispec-lint .

# usecase 命名・構造規約（CLAUDE.md §2.3: struct + コンストラクタ + Execute）を検証する。
.PHONY: naminglint
naminglint:
	@echo ">>> naminglint: usecase 命名規約チェック"
	@go run ./cmd/naminglint .

# 結合テスト: docker-compose で本物の PostgreSQL を起動し、-tags=integration のテストを回す。
# 終了時は必ずコンテナを破棄する（テスト失敗でも teardown する）。
INTEGRATION_DSN := postgres://frestyle:frestyle@localhost:5433/frestyle_integration?sslmode=disable
.PHONY: test-integration
test-integration:
	docker compose -f docker-compose.integration.yml up -d --wait
	@TEST_DATABASE_URL='$(INTEGRATION_DSN)' go test -tags=integration -run Integration ./... ; \
	  status=$$? ; \
	  docker compose -f docker-compose.integration.yml down -v ; \
	  exit $$status

.PHONY: run
run:
	go run ./cmd/server
