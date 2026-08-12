# FreStyle backend - 開発 / 検証 用 Makefile
#
# 主要 target:
#   make openapi      … swag init で OpenAPI spec を 再 生成 (docs/swagger.{json,yaml})
#   make fmt          … gofumpt -w でコードを自動整形（commit 前に実行）
#   make verify       … gofumpt / vet / build / test / archlint / apispec-lint / naminglint を 一括 実行
#   make archlint     … クリーンアーキテクチャ依存方向ルール (CLAUDE.md §2) を 静的 検証
#   make apispec-lint … Gin ルート ↔ swaggo @Router 注釈 の 整合 (CLAUDE.md §2.7) を 検証
#   make naminglint   … usecase 命名・構造規約 (CLAUDE.md §2.3) を 検証
#   make test-integration … docker-compose で PostgreSQL を起動し -tags=integration の結合テストを実行
#   make run          … go run ./cmd/server で ローカル サーバ 起動
#   make local-up     … ローカル開発用の PostgreSQL + メールキャッチャーを起動
#   make local-down   … 上を停止(データは残る)。make local-reset で volume ごと破棄
#   make local-seed   … ダミーデータ投入 (SIZE=small|medium|large)
#   make local-psql   … ローカル DB に psql で接続
#   make local-slow   … pg_stat_statements から遅いクエリ上位 20 件を表示
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

GOFUMPT_VERSION := v0.10.0
GOFUMPT := $(GO_INSTALL_DIR)/gofumpt

# gofumpt を用意する。
#
# 存在チェックだけでは不十分で、バージョンまで確認して違えば入れ直す。別バージョンが既に
# 入っている環境だと整形結果が CI とずれるため（CI は v0.10.0 固定。v0.11.0 は多行呼び出しの
# 閉じ括弧の扱いが変わっており、ローカルで整形済みでも CI の gofumpt check が fail する）。
.PHONY: ensure-gofumpt
ensure-gofumpt:
	@if ! "$(GOFUMPT)" --version 2>/dev/null | grep -q '^$(GOFUMPT_VERSION) '; then \
	  echo ">>> installing gofumpt ($(GOFUMPT_VERSION))..."; \
	  go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION); \
	fi

# gofumpt(gofmt 上位互換)でコードを自動整形する。CI(gofumpt check)が未整形を fail にするので、
# commit 前に `make fmt` で整える運用。生成コードは gofumpt clean のため対象外。
.PHONY: fmt
fmt: ensure-gofumpt
	@echo ">>> gofumpt -w ."
	@$(GOFUMPT) -w .

.PHONY: verify
verify:
	@if [ ! -x "$(GOFUMPT)" ]; then go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION); fi
	@GOFMT_TMP=$$(mktemp) && \
	  $(GOFUMPT) -l . | tee "$$GOFMT_TMP" && \
	  if [ -s "$$GOFMT_TMP" ]; then echo "ERROR: gofumpt issues above (run 'make fmt')"; rm -f "$$GOFMT_TMP"; exit 1; fi && \
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

# ---------------------------------------------------------------------------
# ローカル開発環境 (docker-compose.local.yml)
#
# アプリ(Go / Vite)はホストで動かし、DB とメールキャッチャーだけコンテナに置く。
# Go の再ビルドは数秒・Vite は HMR が効くため、アプリまでコンテナに入れると遅くなる。
# 結合テスト用(docker-compose.integration.yml / 5433 / tmpfs)とは別物。
# ---------------------------------------------------------------------------
LOCAL_COMPOSE := docker-compose.local.yml
LOCAL_DSN ?= postgres://frestyle:frestyle@localhost:5432/frestyle?sslmode=disable
SIZE ?= small

.PHONY: local-up
local-up:
	docker compose -f $(LOCAL_COMPOSE) up -d --wait
	@echo ""
	@echo "DB   : $(LOCAL_DSN)"
	@echo "Mail : http://localhost:8025"
	@echo ""
	@echo "次に .env へ以下を設定して 'make run' でサーバを起動する:"
	@echo "  DATABASE_URL=$(LOCAL_DSN)"
	@echo "  MAIL_SMTP_HOST=localhost"

.PHONY: local-down
local-down:
	docker compose -f $(LOCAL_COMPOSE) down

# volume ごと破棄してまっさらにする。スキーマは次回起動時の AutoMigrate で再構築される。
.PHONY: local-reset
local-reset:
	docker compose -f $(LOCAL_COMPOSE) down -v
	@echo "破棄しました。make local-up → make run(AutoMigrate) → make local-seed の順で作り直せます。"

# ダミーデータ投入。スキーマは AutoMigrate 済みが前提(先に make run を一度実行する)。
# 規模: SIZE=small(既定) | medium | large
.PHONY: local-seed
local-seed:
	@docker compose -f $(LOCAL_COMPOSE) exec -T db \
	  psql -U frestyle -d frestyle -v ON_ERROR_STOP=1 -v size=$(SIZE) \
	  < scripts/seed-local.sql

.PHONY: local-psql
local-psql:
	@docker compose -f $(LOCAL_COMPOSE) exec db psql -U frestyle -d frestyle

# 実測でボトルネックを特定する。体感ではなくここを根拠にする。
.PHONY: local-slow
local-slow:
	@docker compose -f $(LOCAL_COMPOSE) exec -T db psql -U frestyle -d frestyle -c "\
	  SELECT calls, \
	         round(mean_exec_time::numeric, 2) AS mean_ms, \
	         round(total_exec_time::numeric, 2) AS total_ms, \
	         left(query, 90) AS query \
	    FROM pg_stat_statements \
	   WHERE query NOT LIKE '%pg_stat_statements%' \
	   ORDER BY total_exec_time DESC LIMIT 20;"

.PHONY: local-slow-reset
local-slow-reset:
	@docker compose -f $(LOCAL_COMPOSE) exec -T db psql -U frestyle -d frestyle \
	  -c "SELECT pg_stat_statements_reset();" >/dev/null && echo "統計をリセットしました。"
