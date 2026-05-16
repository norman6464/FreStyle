# FreStyle backend - 開発 / 検証 用 Makefile
#
# 主要 target:
#   make openapi      … swag init で OpenAPI spec を 再 生成 (docs/swagger.{json,yaml})
#   make verify       … go build / vet / test を 一括 実行
#   make run          … go run ./cmd/server で ローカル サーバ 起動
#
# swag CLI が PATH に なければ `go install github.com/swaggo/swag/cmd/swag@latest` を 自動 実行 する。

SWAG := $(shell go env GOPATH)/bin/swag

.PHONY: openapi
openapi:
	@if [ ! -x "$(SWAG)" ]; then \
	  echo ">>> installing swag CLI..."; \
	  go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@echo ">>> regenerating OpenAPI spec at docs/"
	@$(SWAG) init --generalInfo cmd/server/main.go --output docs --parseDependency --parseInternal

.PHONY: verify
verify:
	@gofmt -l . | tee /tmp/gofmt-issues.txt
	@[ ! -s /tmp/gofmt-issues.txt ] || (echo "ERROR: gofmt issues above"; exit 1)
	go vet ./...
	go build ./...
	go test ./...

.PHONY: run
run:
	go run ./cmd/server
