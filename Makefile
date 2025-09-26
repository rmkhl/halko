MODULES = executor powerunit simulator sensorunit halkoctl storage
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

.PHONY: build
build: clean $(MODULES:%=$(BINDIR)/%)
	@echo "All binaries have been rebuilt."

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

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  help                  Show this help message. (default)"
	@echo "  all                   Build all Go executables to bin/ directory."
	@echo "  prepare               Check for required tools and create/update go.work file to include all modules."
	@echo "  rebuild               Clean and rebuild all executables from scratch."
	@echo "  clean                 Remove the bin/ directory and all built executables."
	@echo "  lint                  Run golangci-lint on all modules."
	@echo "  lint-markdown         Run mdl (markdown linter) on all markdown files."
	@echo "  go-tidy               Run go mod tidy on all modules with go.mod files."
	@echo "  update-modules        Update all go.mod dependencies and tidy them."
	@echo "  install               Install all binaries except simulator to /opt/halko and copy templates/halko.cfg to /etc/opt/halko.cfg if not present."
	@echo "  systemd-units         Create, install, and enable systemd unit files for all binaries except simulator."
	@echo "  fmt-changed           Reformat changed Go files compared to the main branch using golangci-lint."
	@echo "  test                  Run all tests."
	@echo "  test-program-validation  Run program validation tests."
	@echo "  test-shelly-api       Run shelly API tests."
	@echo "  validate              Validate a program.json file using: make validate PROGRAM=path/to/program.json"

.DEFAULT_GOAL := help
