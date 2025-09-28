.PHONY: help doc deps build test test.int test.all install test.tf tart.image.list tart.ils install-and-plan clean api executor deps.tmux deps.tart console dietpi.create dietpi.run dietpi.delete help.dietpi
.DEFAULT_GOAL := help

## Show available make targets
help: ## show help and available targets
	@echo ""
	@echo "terraform-provider-tart - Make targets"
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_.-]+:.*## / { gsub(/^[[:space:]]+|[[:space:]]+$$/, "", $$1); printf "  %-24s %s\n", $$1, $$2 }' $(MAKEFILE_LIST) | sort -k1,1

## Sync Go module dependencies
deps: ## tidy and download go modules
	go mod tidy && go mod download

## Run Go tests for the provider
test: ## run go tests
	go test ./...

## Run integration tests (tests/)
test.int: ## run integration tests (tests/)
	go test -v ./tests
test.all: ## run all tests with verbose output
	go test -v ./...

## Start the Tart API server (loads .env if present)
.PHONY: api
api: deps.tart ## start the Tart API server
	@set -e; \
	if [ -f .env ]; then echo "Loading .env"; set -a; . ./.env; set +a; fi; \
	GO111MODULE=on go run ./cmd/tart-api

## Start the Tart executor daemon (loads .env if present)
.PHONY: executor
executor: deps.tart ## start the Tart executor daemon
	@set -e; \
	if [ -f .env ]; then echo "Loading .env"; set -a; . ./.env; set +a; fi; \
	GO111MODULE=on go run ./tart-executor

## Ensure tmux is installed (attempt Homebrew install on macOS)
.PHONY: deps.tmux
deps.tmux: ## ensure tmux is available
	@set -e; \
	if command -v tmux >/dev/null 2>&1; then \
	  echo "tmux found: $$(command -v tmux)"; \
	else \
	  echo "tmux not found. Attempting to install..."; \
	  if command -v brew >/dev/null 2>&1; then \
	    echo "Using Homebrew to install tmux"; \
	    brew list tmux >/dev/null 2>&1 || brew install tmux; \
	  else \
	    echo "Homebrew not found. Please install tmux manually (e.g., brew install tmux)"; \
	    exit 1; \
	  fi; \
	fi

## Ensure Tart CLI is installed (attempt Homebrew install on macOS)
.PHONY: deps.tart
deps.tart: ## ensure Tart CLI is available
	@set -e; \
	if command -v tart >/dev/null 2>&1; then \
	  echo "tart found: $$(command -v tart)"; \
	else \
	  echo "tart not found. Attempting to install..."; \
	  if command -v brew >/dev/null 2>&1; then \
	    echo "Using Homebrew to install Tart"; \
	    brew list tart >/dev/null 2>&1 || brew install cirruslabs/cli/tart; \
	  else \
	    echo "Homebrew not found. Please install Tart manually: brew install cirruslabs/cli/tart"; \
	    exit 1; \
	  fi; \
	fi

## Open a tmux console with API, executor, and Terraform plan
.PHONY: console

console: ## open tmux console (3 panes: api, executor, shell)
	@bash ./addons/tmux-console.sh
DIETPI_NAME ?= hello-from-dietpi
DIETPI_URL ?= https://dietpi.com/downloads/images/DietPi_VM-ARM64.img.xz

dietpi.create: ## build a DietPi-based ARM64 VM locally via debootstrap script (addons/)
	@set -e; \
	if [ -f .env ]; then set -a; . ./.env; set +a; fi; \
	if [ ! -x ./addons/tart-debootstrap-dietpi-vm.sh ]; then \
	  echo "addons/tart-debootstrap-dietpi-vm.sh not found or not executable"; exit 1; \
	fi; \
	echo "Running DietPi VM builder (this may take a while)..."; \
	bash ./addons/tart-debootstrap-dietpi-vm.sh

dietpi.run: ## run VM and wait until user stops it (POST /vms/{id}/run)
	@set -e; \
	if [ -f .env ]; then set -a; . ./.env; set +a; fi; \
	API_URL="$${TART_API_URL:-http://localhost:8085/api}"; \
	NAME="$${DIETPI_NAME:-$(DIETPI_NAME)}"; \
	echo "Running $$NAME via $$API_URL (this will block until the VM stops)"; \
	curl -fsS -X POST "$$API_URL/vms/$$NAME/run" || (echo "Run failed" && exit 1)

dietpi.delete: ## delete DietPi VM via API
	@set -e; \
	if [ -f .env ]; then set -a; . ./.env; set +a; fi; \
	API_URL="$${TART_API_URL:-http://localhost:8085/api}"; \
	NAME="$${DIETPI_NAME:-$(DIETPI_NAME)}"; \
	echo "Deleting $$NAME via $$API_URL"; \
	curl -fsS -X DELETE "$$API_URL/vms/$$NAME" || (echo "Delete failed" && exit 1)

help.dietpi: ## show DietPi workflow usage and variables
	@echo ""; \
	echo "DietPi workflow targets"; \
	echo "  make dietpi.create        # build local DietPi-based ARM64 VM via debootstrap script"; \
	echo "  make dietpi.run           # run VM and wait until user stops it"; \
	echo "  make dietpi.delete        # delete VM"; \
	echo ""; \
	echo "Variables (override via CLI or .env):"; \
	echo "  DIETPI_NAME  (default: hello-from-dietpi)"; \
	echo "  DIETPI_URL   (default: https://dietpi.com/downloads/images/DietPi_VM-ARM64.img.xz)"; \
	echo "  TART_API_URL (default: http://localhost:8085/api)"; \
	echo ""; \
	echo "Examples:"; \
	echo "  DIETPI_NAME=myvm make dietpi.create"; \
	echo "  DIETPI_URL=ghcr.io/vendor/image:tag DIETPI_NAME=myvm make dietpi.create"; \
	echo "  make dietpi.delete DIETPI_NAME=myvm"

## Build the provider binary
build: ## build the provider binary
	@set -e; \
		if [ -z "$$ver" ] && [ -f VERSION ]; then ver="$$(tr -d ' \n\r' < VERSION)"; fi; \
		os="$${os:-$$(go env GOOS)}"; \
		arch="$${arch:-$$(go env GOARCH)}"; \
		if [ -z "$$ver" ]; then echo "Set ver=X.Y.Z or create VERSION file"; exit 1; fi; \
		dir="$$HOME/.terraform.d/plugins/local/tart/tart/$$ver/$${os}_$${arch}"; \
		mkdir -p "$$dir"; \
		echo "Building provider for $$os/$$arch -> $$dir"; \
		GOOS="$$os" GOARCH="$$arch" go build -o "$$dir/terraform-provider-tart" ./cmd/terraform-provider-tart; \
		echo "Installed to: $$dir"

## Run Terraform fmt/init/validate/plan (non-destructive). Override TF_DIR to target another folder.
TF_DIR ?= .
test.tf: ## run terraform fmt/validate/plan
	echo "Running Terraform checks in $(TF_DIR)"
	terraform -chdir=$(TF_DIR) init -upgrade -input=false
	terraform -chdir=$(TF_DIR) validate
	terraform -chdir=$(TF_DIR) plan -no-color -out=$(TF_DIR)/.tf-plan.out

## Clean Terraform state and plan files in TF_DIR
clean: ## remove Terraform states, lockfile, plan files and local provider binary
	@echo "Cleaning Terraform state in $(TF_DIR)..."; \
	rm -rf $(TF_DIR)/.terraform $(TF_DIR)/.terraform.lock.hcl $(TF_DIR)/*.tfstate $(TF_DIR)/*.tfstate.backup $(TF_DIR)/*.tfplan $(TF_DIR)/.tf-plan.out
	@echo "Cleaning local build artifacts..."
	rm -f terraform-provider-tart

## List available local Tart images/VMs
tart.image.list: ## list local Tart images
	@tart list

## IPSW/ISO list for running VMs (best-effort)
tart.ils: ## list .ipsw or iso files in use (note: none are used at runtime); show cache and hints
	@set -e; \
	echo "Note: Tart does not mount .ipsw or ISO at runtime; VMs boot from their own disks."; \
	running="$$(tart list | awk 'NR>1 && $$NF=="running" {print $$2}')"; \
	if [ -z "$$running" ]; then echo "No running VMs."; exit 0; fi; \
	echo ""; echo "Running VMs:"; echo "$$running" | sed 's/^/  - /'; \
	echo ""; echo "macOS IPSW cache:"; \
	ipsw_dir="$$HOME/Library/Containers/com.cirruslabs.tart/Data/Library/Caches/com.cirruslabs.tart/ipsw"; \
	if [ -d "$$ipsw_dir" ]; then ls -1 "$$ipsw_dir"/*.ipsw 2>/dev/null || echo "  (none)"; else echo "  (no IPSW cache dir)"; fi; \
	echo ""; \
	for vm in $$running; do \
	  if tart run "$$vm" -- sw_vers >/dev/null 2>&1; then \
	    echo "VM $$vm: macOS (likely created from an IPSW; exact file not tracked)"; \
	  else \
	    echo "VM $$vm: non-macOS (likely installed from an ISO; original ISO not tracked)"; \
	  fi; \
	done

## Build, install provider (no bump), and run terraform init/plan in TF_DIR
.PHONY: install-and-plan
install-and-plan: build install ## build, install provider and terraform init/plan
	@echo "Re-initializing Terraform in $(TF_DIR)" && \
	rm -rf $(TF_DIR)/.terraform $(TF_DIR)/.terraform.lock.hcl && \
	terraform -chdir=$(TF_DIR) init -upgrade -input=false && \
	terraform -chdir=$(TF_DIR) plan -input=false -lock=false -no-color -out=$(TF_DIR)/.tf-plan.out || true
