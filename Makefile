SH ?= ash

IMAGE_NAME := idempotency-manager

DOCKER_WORKDIR ?= `pwd`
DOCKER_RUN := docker run --rm -v $(DOCKER_WORKDIR):/app -it
DOCKER_BIN := $(shell which docker)

define docker_run
	if [ -n "$(DOCKER_BIN)" ]; then echo "Running on $1 container..." && $(DOCKER_RUN) --name $1-$2 $4 $1 $3; else $3; fi;
endef

default: help ## Defaults to help

help: ## Get help
	@echo "Make tasks:\n"
	@grep -hE '^[%a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m  %-10s\033[0m %s\n", $$1, $$2}'
	@echo ""
.PHONY: help

build :## Builds docker image
	docker build -t $(IMAGE_NAME) .
.PHONY: build

shell: ## Runs shell inside container
	@$(call docker_run,$(IMAGE_NAME),$@,$(SH))
.PHONY: shell

test: ## Runs tests
	@$(call docker_run,$(IMAGE_NAME),$@,go test)
.PHONY: test

test-%: ## Runs specific test
	@$(call docker_run,$(IMAGE_NAME),$@,go test -run $*)
.PHONY: test-%

example-%: ## Runs an example from folder examples by number
	@$(eval EXAMPLE := $(shell ls examples/$*-* | head -n 1))
	@$(call docker_run,$(IMAGE_NAME),$@,go run $(EXAMPLE))
.PHONY: example-%

make-%: ## Runs make tasks
	@$(call docker_run,$(IMAGE_NAME),$@,make $*)
.PHONY: make-%

cover: make-cover ## Runs tests with cover and output its result
.PHONY: cover

stop-%: ## Stops container
	docker rm -f $(IMAGE_NAME)-$*

docs: ## Runs godoc server
	@$(call docker_run,$(IMAGE_NAME),$@,godoc -http=:6060,-p 6060:6060)
.PHONY: docs

make-cover:
	@mkdir -p tmp
	go test -coverprofile=tmp/cover.out
	go tool cover -html=tmp/cover.out -o tmp/cover.html
	@rm -rf tmp/cover.out
.PHONY: make-cover
