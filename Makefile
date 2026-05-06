SHELL := /bin/bash
.DEFAULT_GOAL := help

REPO_ROOT := $(shell pwd)

.PHONY: help
help: ## List available targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: test
test: ## Run tests, excluding the intentionally-failing exercise tests
	@pkgs=$$(go list ./... | grep -v '/exercises$$'); \
	if [ -z "$$pkgs" ]; then echo "no testable packages"; exit 0; fi; \
	go test $$pkgs

.PHONY: test-exercises
test-exercises: ## Run exercise tests (these fail by design until students complete them)
	-@pkgs=$$(go list ./... | grep '/exercises$$'); \
	if [ -z "$$pkgs" ]; then echo "no exercise packages"; exit 0; fi; \
	go test $$pkgs

.PHONY: test-lesson
test-lesson: ## Run tests for one lesson, both exercises and solutions (LESSON=NN-name)
	@test -n "$(LESSON)" || (echo "usage: make test-lesson LESSON=NN-name" && exit 1)
	-go test ./lessons/$(LESSON)/exercises/...
	go test ./lessons/$(LESSON)/solutions/...

.PHONY: lint
lint: ## Run golangci-lint (must be installed locally)
	golangci-lint run

.PHONY: fmt
fmt: ## Format Go code with gofmt and goimports
	gofmt -w .
	goimports -w .

.PHONY: new-lesson
new-lesson: ## Scaffold a new lesson (NAME=NN-name)
	@test -n "$(NAME)" || (echo "usage: make new-lesson NAME=NN-name" && exit 1)
	go run ./tools/new-lesson -name $(NAME)

.PHONY: slides-dev
slides-dev: ## Serve one lesson's deck locally on http://localhost:8000 (LESSON=NN-name)
	@test -n "$(LESSON)" || (echo "usage: make slides-dev LESSON=NN-name" && exit 1)
	go run ./tools/slides-dev -lesson $(LESSON) -repo $(REPO_ROOT)

.PHONY: slides-build
slides-build: ## Build the static slides site into dist/
	go run ./tools/build-index -lessons lessons -shared shared/reveal -out dist

.PHONY: slides-docker
slides-docker: ## Build the deploy image and run it locally on http://localhost:8080
	docker build -t go-training-slides:local -f deploy/Dockerfile .
	@echo "starting container on http://localhost:8080  (Ctrl-C to stop)"
	docker run --rm -p 8080:8080 -e PORT=8080 go-training-slides:local
