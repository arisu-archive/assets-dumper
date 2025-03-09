# Change these variables as necessary.
MAIN_PACKAGE_PATH := ./main.go
TMP_DIR := ./tmp
BINARY_NAME := asset-dumper

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

.PHONY: prepare
prepare:
	mkdir -p ${TMP_DIR}/bin
	go env -w CGO_ENABLED=1
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
	go install github.com/onsi/ginkgo/v2/ginkgo@v2.22.2
	go install github.com/vektra/mockery/v2@v2.53.0

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: no-dirty
no-dirty:
	git diff --exit-code

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## audit: run quality control checks
.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1 run ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

.PHONY: mocks
mocks:
	@mockery --all

.PHONY: generate-testdata
generate-testdata:
	./scripts/generate_xxhash_testdata.sh

## test: run unit tests
.PHONY: test
test:
	go test -race -cover -coverprofile="coverage.out" ./...

## build: build the application
.PHONY: build
build:
	go build -o=${TMP_DIR}/bin/${BINARY_NAME} ${MAIN_PACKAGE_PATH}

## run: run the application
.PHONY: run
run: build
	${TMP_DIR}/bin/${BINARY_NAME}

# ==================================================================================== #
# OPERATIONS
# ==================================================================================== #

## push: push changes to the remote Git repository
.PHONY: push
push: tidy audit no-dirty
	git push

## production/deploy: deploy the application to production
.PHONY: production/build
production/build:
	GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o=${TMP_DIR}/bin/linux_amd64/${BINARY_NAME} ${MAIN_PACKAGE_PATH}
