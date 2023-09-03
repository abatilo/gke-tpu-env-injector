SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.ONESHELL:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

.PHONY: help
help: ## View help information
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: asdf-bootstrap
asdf-bootstrap: .tool-versions
	cat .tool-versions | cut -f 1 -d ' ' | xargs -n 1 asdf plugin-add || true

.PHONY: helm-bootstrap
helm-bootstrap: ## Add helm repos
	helm repo add jetstack https://charts.jetstack.io
	helm repo update

.PHONY: up
up: asdf-bootstrap ## Run dev environment
	ctlptl apply -f ctlptl.yaml
	skaffold dev

.PHONY: ci
ci: helm-bootstrap ## Setup CI environment
	ctlptl apply -f ctlptl.yaml
	skaffold run

.PHONY: clean
clean: ## Delete all dev environment resources
	ctlptl delete -f ctlptl.yaml
