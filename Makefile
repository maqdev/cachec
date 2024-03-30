TEST_OPTS=--race
LINT_OPTS=

LINTER_VERSION=v1.57.2
SQLC_VERSION=v1.26.0
PROTOBUF_DOCKER=jaegertracing/protobuf:v0.5.0

USER_ID = $(shell id -u)
GROUP_ID = $(shell id -g)

.PHONY: install-tools
install-tools:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(LINTER_VERSION)
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)

.PHONY: sqlc
sqlc:
	@sqlc generate

.PHONY: proto
proto:
	@mkdir -p ./gen/proto
	@docker run --rm -v "${PWD}":"/data/" -w "/data/" --user "$(USER_ID):$(GROUP_ID)" "${PROTOBUF_DOCKER}" --go_out="./" --proto_path "./" "./gen/proto/*.proto"

.PHONY: test
test:
	@go test $(TEST_OPTS) ./...

.PHONY: lint
lint:
	@golangci-lint run -v --timeout 30m --exclude-use-default $(LINT_OPTS)

.PHONY: _enable_lint_fix
_enable_lint_fix:
	@$(eval LINT_OPTS=--fix)

.PHONY: lint-n-fix
lint-n-fix: _enable_lint_fix lint

.PHONY: clean
clean:
	rm -rf ./gen/*
