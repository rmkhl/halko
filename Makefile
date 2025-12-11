MODULES = controlunit powerunit simulator sensorunit halkoctl
BINDIR = bin

.PHONY: all
all: prepare clean $(MODULES:%=$(BINDIR)/%)

$(BINDIR)/%: %/main.go | $(BINDIR)
	go build -o $@ ./$*/

$(BINDIR):
	mkdir -p $(BINDIR)

.PHONY: clean
clean:
	rm -rf $(BINDIR)
	@rm -rf webapp/dist webapp/.parcel-cache webapp/node_modules
	@echo "✓ Cleaned Go binaries and webapp artifacts"

.PHONY: distclean
distclean: clean
	@rm -rf .nodejs
	@echo "✓ Removed local Node.js installation"

.PHONY: prepare
prepare:
	@echo "Checking for required tools..."
	@if ! command -v go > /dev/null; then \
		echo "Error: 'go' command is not available. Please install Go before continuing."; \
		exit 1; \
	else \
		echo "✓ Go is installed"; \
	fi
	@if ! command -v golangci-lint > /dev/null; then \
		echo "Warning: 'golangci-lint' is not available. Some make targets will not work."; \
		echo "Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
	else \
		echo "✓ golangci-lint is installed"; \
	fi
	@if ! command -v mdl > /dev/null; then \
		echo "Warning: 'mdl' is not available. Markdown linting will not work."; \
	else \
		echo "✓ mdl is installed"; \
	fi
	@echo "Creating or updating go.work file with all modules..."
	@if [ ! -f go.work ]; then \
		echo "Initializing new go.work file..."; \
		go work init; \
	else \
		echo "Clearing existing go.work file..."; \
		rm -f go.work && go work init; \
	fi
	@# Add all modules that have go.mod files
	@for mod in $(MODULES) types tests; do \
		if [ -f $$mod/go.mod ]; then \
			echo "Adding $$mod to go.work..."; \
			go work use ./$$mod; \
		fi; \
	done
	@echo "Updated go.work file to include all modules."
	@echo "Checking for Node.js..."
	@if [ -f .nodejs/bin/node ]; then \
		export PATH="$$(pwd)/.nodejs/bin:$$PATH"; \
	fi; \
	if ! command -v node > /dev/null; then \
		echo "Node.js not found. Installing Node.js 18 locally in .nodejs/..."; \
		mkdir -p .nodejs; \
		ARCH=$$(uname -m); \
		OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
		if [ "$$ARCH" = "x86_64" ]; then \
			NODE_ARCH="x64"; \
		elif [ "$$ARCH" = "aarch64" ] || [ "$$ARCH" = "arm64" ]; then \
			NODE_ARCH="arm64"; \
		else \
			echo "Error: Unsupported architecture: $$ARCH"; \
			exit 1; \
		fi; \
		NODE_VERSION="18.20.5"; \
		NODE_DIST="node-v$${NODE_VERSION}-$${OS}-$${NODE_ARCH}"; \
		echo "Downloading Node.js $${NODE_VERSION} for $${OS}-$${NODE_ARCH}..."; \
		curl -fsSL "https://nodejs.org/dist/v$${NODE_VERSION}/$${NODE_DIST}.tar.gz" -o .nodejs/node.tar.gz; \
		echo "Extracting Node.js..."; \
		tar -xzf .nodejs/node.tar.gz -C .nodejs --strip-components=1; \
		rm .nodejs/node.tar.gz; \
		echo "✓ Node.js installed to .nodejs/"; \
	else \
		NODE_VERSION=$$(node -v | sed 's/v//'); \
		NODE_MAJOR=$$(echo $$NODE_VERSION | cut -d. -f1); \
		if [ $$NODE_MAJOR -lt 18 ]; then \
			echo "Warning: Node.js $$NODE_VERSION detected. Node.js 18+ is recommended."; \
		else \
			echo "✓ Node.js $$(node -v) is installed"; \
		fi; \
	fi

.PHONY: build
build: prepare clean $(MODULES:%=$(BINDIR)/%)
	@echo "All Go binaries have been rebuilt."
	@echo "Installing webapp dependencies..."
	@if [ -f .nodejs/bin/node ]; then \
		export PATH="$$(pwd)/.nodejs/bin:$$PATH"; \
	fi; \
	cd webapp && npm install
	@echo "✓ Webapp dependencies installed"
	@echo "Building webapp for production..."
	@if [ -f .nodejs/bin/node ]; then \
		export PATH="$$(pwd)/.nodejs/bin:$$PATH"; \
	fi; \
	cd webapp && npm run build
	@echo "✓ Webapp built to webapp/dist/"
	@echo "All binaries and webapp have been rebuilt."

.PHONY: lint
lint:
	@for mod in $(MODULES) types tests; do \
		if [ -f $$mod/go.mod ]; then \
			echo "Linting $$mod..."; \
			(cd $$mod && golangci-lint run ./... || true); \
		fi; \
	done

.PHONY: lint-markdown
lint-markdown:
	@if command -v mdl > /dev/null; then \
		echo "Linting markdown files..."; \
		mdl . || true; \
	else \
		echo "Warning: mdl is not installed. Skipping markdown linting."; \
	fi

.PHONY: go-tidy
go-tidy:
	@echo "Running go mod tidy on all modules..."
	@find . -name "go.mod" -type f | while read modfile; do \
		moddir=$$(dirname "$$modfile"); \
		echo "Tidying $$moddir..."; \
		(cd "$$moddir" && go mod tidy); \
	done
	@echo "All modules have been tidied."

.PHONY: update-modules
update-modules:
	@for mod in $(MODULES) types tests; do \
		if [ -f $$mod/go.mod ]; then \
			echo "Updating $$mod..."; \
			(cd $$mod && go get -u ./... && go mod tidy); \
		fi; \
	done

.PHONY: install
install: clean all
	@echo "Installing binaries to /opt/halko (excluding simulator)..."
	sudo install -d /opt/halko
	for bin in $(MODULES); do \
		if [ "$$bin" != "simulator" ]; then \
			sudo install -m 755 $(BINDIR)/$$bin /opt/halko/; \
		fi; \
	done
	@echo "Creating data directory at /var/opt/halko..."
	sudo install -d -m 755 /var/opt/halko
	@echo "Installing config to /etc/opt/halko.cfg if not present..."
	sudo install -d /etc/opt
	@if [ ! -f /etc/opt/halko.cfg ]; then \
		sudo install -m 644 templates/halko.cfg /etc/opt/halko.cfg; \
		 echo "Installed default config to /etc/opt/halko.cfg"; \
	else \
		 echo "/etc/opt/halko.cfg already exists, not overwriting."; \
	fi

.PHONY: systemd-units
systemd-units: install
	@echo "Creating and installing systemd unit files for all binaries except simulator..."
	for bin in $(MODULES); do \
		if [ "$$bin" != "simulator" ]; then \
			sudo cp templates/halko-daemon.service /etc/systemd/system/halko@$$bin.service; \
			sudo sed -i "s/%i/$$bin/g" /etc/systemd/system/halko@$$bin.service; \
			sudo systemctl daemon-reload; \
			sudo systemctl enable halko@$$bin.service; \
			if systemctl is-active --quiet halko@$$bin.service; then \
				sudo systemctl restart halko@$$bin.service; \
			else \
				sudo systemctl start halko@$$bin.service; \
			fi; \
		fi; \
	done
	@echo "Systemd unit files installed and services enabled."

.PHONY: fmt-changed
fmt-changed:
	@for mod in $(MODULES) types; do \
		if [ -f $$mod/go.mod ]; then \
			echo "Formatting changed files in $$mod..."; \
			(cd $$mod && git diff --name-only master...HEAD | grep '\.go$$' | xargs -r golangci-lint run --fix -v || true); \
		fi; \
	done
	@echo "Reformatted changed Go files compared to main branch using golangci-lint."

.PHONY: test
test:
	@echo "Running all tests..."
	@$(MAKE) test-config || true
	@$(MAKE) test-program-validation || true
	@$(MAKE) test-shelly-api || true
	@echo "All tests completed."

.PHONY: test-config
test-config:
	@echo "Running configuration tests..."
	@cd tests && go test -v -run TestConfig

.PHONY: test-program-validation
test-program-validation:
	@echo "Running program validation tests..."
	@cd tests && go test -v -run TestProgramValidation

.PHONY: test-shelly-api
test-shelly-api:
	@echo "Running shelly API tests..."
	@cd tests && go test -v -run TestShellyAPI

.PHONY: validate
validate:
	@if [ -z "$(PROGRAM)" ]; then \
		echo "Usage: make validate PROGRAM=path/to/program.json"; \
		echo "Example: make validate PROGRAM=example/example-program-delta.json"; \
		exit 1; \
	fi
	@if [ ! -f $(BINDIR)/halkoctl ]; then \
		echo "Building halkoctl..."; \
		$(MAKE) $(BINDIR)/halkoctl; \
	fi
	@echo "Validating program: $(PROGRAM)"
	@$(BINDIR)/halkoctl validate -program $(PROGRAM) -verbose

.PHONY: images
images: build
	@echo "Ensuring fsdb directory exists..."
	@mkdir -p fsdb
	@echo "Generating nginx configuration for webapp Docker container..."
	@$(BINDIR)/halkoctl -c halko-docker.cfg nginx -port 80 -output webapp/nginx-docker.conf
	@echo "Removing existing Docker images..."
	@docker-compose down --remove-orphans || true
	@docker-compose rm -f || true
	@docker images --filter "reference=halko_*" -q | xargs -r docker rmi -f || true
	@echo "Building new Docker images..."
	@BUILDKIT_PROGRESS=plain docker-compose build --no-cache
	@echo "Docker images have been rebuilt."

# ============================================================================
# WebApp Targets
# ============================================================================

.PHONY: webapp-dev
webapp-dev: prepare
	@echo "Installing webapp dependencies..."
	@if [ -f .nodejs/bin/node ]; then \
		export PATH="$$(pwd)/.nodejs/bin:$$PATH"; \
	fi; \
	cd webapp && npm install
	@echo "Starting webapp development server..."
	@if [ -f .nodejs/bin/node ]; then \
		export PATH="$$(pwd)/.nodejs/bin:$$PATH"; \
	fi; \
	cd webapp && npm start

.PHONY: build-webapp
build-webapp: prepare $(BINDIR)/halkoctl
	@echo "Building webapp for production (host installation)..."
	@if [ -f .nodejs/bin/node ]; then \
		export PATH="$$(pwd)/.nodejs/bin:$$PATH"; \
	fi; \
	cd webapp && npm install
	@echo "Building production bundle..."
	@if [ -f .nodejs/bin/node ]; then \
		export PATH="$$(pwd)/.nodejs/bin:$$PATH"; \
	fi; \
	cd webapp && npm run build
	@echo "✓ Webapp built successfully to webapp/dist/"
	@echo "Generating nginx configuration for host installation..."
	@if [ -f /etc/opt/halko.cfg ]; then \
		CONFIG_FILE="/etc/opt/halko.cfg"; \
		echo "Using global config: $$CONFIG_FILE"; \
	else \
		CONFIG_FILE="halko.cfg"; \
		echo "Using local config: $$CONFIG_FILE"; \
	fi; \
	$(BINDIR)/halkoctl -c $$CONFIG_FILE nginx -port 80 -output webapp/nginx-host.conf
	@echo "✓ Generated webapp/nginx-host.conf for host installation"
	@echo "  To serve: Copy webapp/dist/* to your web server root and nginx-host.conf to nginx sites"

.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Main Targets:"
	@echo "  help                       Show this help message. (default)"
	@echo "  all                        Build all Go executables to bin/ directory."
	@echo "  prepare                    Check for required tools (Go, Node.js), install Node.js if needed, and setup workspace."
	@echo "  build                      Clean and rebuild all Go executables and webapp from scratch."
	@echo "  clean                      Remove bin/ directory and webapp artifacts (keeps Node.js installation)."
	@echo "  distclean                  Like clean, but also removes local Node.js installation."
	@echo "  images                     Rebuild everything and recreate all Docker images (including webapp)."
	@echo ""
	@echo "Go Backend Targets:"
	@echo "  lint                       Run golangci-lint on all modules."
	@echo "  lint-markdown              Run mdl (markdown linter) on all markdown files."
	@echo "  go-tidy                    Run go mod tidy on all modules with go.mod files."
	@echo "  update-modules             Update all go.mod dependencies and tidy them."
	@echo "  install                    Install all binaries except simulator to /opt/halko."
	@echo "  systemd-units              Create, install, and enable systemd unit files."
	@echo "  fmt-changed                Reformat changed Go files compared to the main branch."
	@echo ""
	@echo "Test Targets:"
	@echo "  test                       Run all tests."
	@echo "  test-program-validation    Run program validation tests."
	@echo "  test-shelly-api            Run shelly API tests."
	@echo "  validate                   Validate a program.json file: make validate PROGRAM=path/to/program.json"
	@echo ""
	@echo "WebApp Targets:"
	@echo "  webapp-dev                 Start webapp development server with hot reload."
	@echo "  build-webapp               Build webapp for production (host installation) to webapp/dist/."

.DEFAULT_GOAL := help
