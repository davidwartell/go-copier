GO_BUILD=go build
GO_TEST=go clean -testcache && go test -v -race
GO_CLEAN=go mod tidy && go mod clean
GO_MOD_TIDY=go mod tidy
IS_RELEASE=false
DEBUG_GO_BUILD_OPTIONS=-race
RELEASE_GO_BUILD_OPTIONS=
GO_BUILD_OPTIONS=$(DEBUG_GO_BUILD_OPTIONS)
DEBUG_GLOBAL_LDFLAGS=
RELEASE_GLOBAL_LDFLAGS=-ldflags "-s -w"
LDFLAGS_GLOBAL=${DEBUG_GLOBAL_LDFLAGS}

all: go-build

.PHONY: release
release: ## Make the release target first to build with release flags and without debugging symbols e.g.: make release all
	$(info RELEASE FLAGS)
	$(eval IS_RELEASE = true)
	$(eval LDFLAGS_GLOBAL = $(RELEASE_GLOBAL_LDFLAGS))
	$(eval GO_BUILD_OPTIONS = $(RELEASE_GO_BUILD_OPTIONS))
	@: # this suppresses message nothing to do for target

.PHONY: go-build
go-build: ## Compile golang lib with debug symbols.
	$(GO_BUILD) $(GO_BUILD_OPTIONS) $(LDFLAGS_GLOBAL) ./... ; \

.PHONY: test
test: ## Run unit tests.
	$(GO_TEST) ${LDFLAGS_GLOBAL} -v ./...

# G304 - machineid/helper.go:31 false positive on file name in variable
# G404 - weak random generator for exponential backoff times is fine (backoff/backoff.go:71)
# G402 - TLS InsecureSkipVerify set true on purpose used by httpcommon for http probing through a proxy
.PHONY: gosec
gosec: ## Run gosec to static analyze for security vulnerabilities..
	gosec -exclude=G304,G404,G402 ./...

.PHONY: clean
clean: ## Clean.
	$(GO_CLEAN)

# FIXME remove exclude for "-D staticcheck" when deprecated logging API usage is fixed
.PHONY: golint
golint: ## Run golangci-lint linter.
	golangci-lint run -D govet -D staticcheck

GO_MODULES += "github.com/pkg/errors"

GO_MODULES_TOOLS += "github.com/securego/gosec/v2/cmd/gosec"
GO_MODULES_TOOLS += "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
GO_MODULES_TOOLS += "google.golang.org/protobuf/cmd/protoc-gen-go"

.PHONY: setup
setup: ## Update / download required go modules.
	for m in $(GO_MODULES); do \
			echo "go get -u $$m"; \
			go get -u $$m; \
	done

	for m in $(GO_MODULES_TOOLS); do \
			echo "go get -d $$m"; \
			go get -d $$m; \
	done
	$(GO_MOD_TIDY)

help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)
