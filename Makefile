GIT_REV?=$$(git rev-parse --short HEAD)
DATE?=$$(date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION?=$$(git describe --tags --always)
LDFLAGS="-s -w -X main.version=$(VERSION)-$(GIT_REV) -X main.date=$(DATE)"
goos?=${INPUT_GOOS}
goarch?=${INPUT_GOARCH}
NAME:=mandel

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: default help

default: help

## Build:
prepare: ## Download depencies and prepare dev env
	@pre-commit install
	@go mod download
	@go mod vendor

build:  ## Builds the game
	@CGO_ENABLED=1 go build -ldflags=$(LDFLAGS) -o ./bin/$(NAME) .

buildwin: ## Builds game for Win64
	@CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -ldflags=$(LDFLAGS) -o ./bin/$(NAME).exe .

build-ci: ## Optimized build for CI
	@echo $(goos)/$(goarch)
	@GOOS=$(goos) GOARCH=$(goarch) CGO_ENABLED=1 go build -ldflags=$(LDFLAGS) -o ./bin/$(NAME)_$(goos)_$(goarch) .

## Test:
coverage:  ## Run test coverage suite
	@go test ./... --coverprofile=cov.out
	@go tool cover --html=cov.out

## Help:
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
