# Version from git tag (e.g. v1.2.3), falls back to commit hash if no tag
GIT_TAG    := $(shell git describe --tags --always 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -X github.com/songquanpeng/one-api/common.Version=$(GIT_TAG) \
              -X github.com/songquanpeng/one-api/common.BuildCommit=$(GIT_COMMIT) \
              -X github.com/songquanpeng/one-api/common.BuildTime=$(BUILD_TIME)

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o uniapi .

.PHONY: install
install:
	# https://golangci-lint.run/docs/welcome/install/local/
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.10.1

	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	# go install go.uber.org/nilaway/cmd/nilaway@latest
	# go install github.com/mitranim/gow@latest
	# go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

.PHONY: lint
lint:
	goimports -local module,github.com/songquanpeng/one-api -w .
	go mod tidy
	gofmt -s -w .
	go vet
	# nilaway ./...
	golangci-lint run -c .golangci.yml
	govulncheck ./...

# Development targets
.PHONY: dev-modern
dev-modern:
	@cd web/modern && yarn dev

# Default dev target
.PHONY: dev
dev: dev-modern

# Build targets
.PHONY: build-frontend-modern
build-frontend-modern:
	@cd web/modern && yarn && VITE_APP_VERSION=$(GIT_TAG) yarn build

# Default build target
.PHONY: build-frontend
build-frontend: build-frontend-modern

# Build development version
.PHONY: build-frontend-dev-modern
build-frontend-dev-modern:
	@cd web/modern && yarn build

# Default dev build target
.PHONY: build-frontend-dev
build-frontend-dev: build-frontend-dev-modern

# Help target
.PHONY: help-dev
help-dev:
	@echo "Development targets:"
	@echo "  dev-modern        Start modern template development server"
	@echo "  dev               Start modern template development server (default)"
	@echo ""
	@echo "Build targets:"
	@echo "  build-frontend-modern      Build modern template for production"
	@echo ""
	@echo "Development build targets:"
	@echo "  build-frontend-dev-modern  Build modern template for development"
