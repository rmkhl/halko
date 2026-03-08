MODULES = controlunit powerunit simulator sensorunit halkoctl dbusunit
BINDIR = bin

# Build flags - use OPTIMIZED=yes for memory-constrained environments (Raspberry Pi)
# Standard flags (default): Full debugging symbols, faster execution
GOFLAGS_STANDARD =

# Optimized flags: Smaller binaries (~30% reduction), lower memory footprint
# -ldflags="-s -w": Strip debug info and symbol table
# -trimpath: Remove absolute file paths (reproducibility)
GOFLAGS_OPTIMIZED = -ldflags="-s -w" -trimpath

# Select flags based on OPTIMIZED variable
ifeq ($(OPTIMIZED),yes)
GOFLAGS = $(GOFLAGS_OPTIMIZED)
BUILD_TYPE = optimized (stripped, -s -w, -trimpath)
else
GOFLAGS = $(GOFLAGS_STANDARD)
BUILD_TYPE = standard (with debug symbols)
endif

.PHONY: all
all: clean $(MODULES:%=$(BINDIR)/%)
	@echo "✓ Build completed: $(BUILD_TYPE)"

$(BINDIR)/%: %/main.go | $(BINDIR)
	@go build $(GOFLAGS) -o $@ ./$*/

$(BINDIR):
	@mkdir -p $(BINDIR)

.PHONY: clean
clean:
	@rm -rf $(BINDIR)
	@echo "✓ Cleaned Go binaries"

.PHONY: distclean
distclean: clean clean-webapp clean-arduino
	@rm -rf .nodejs node_modules .arduino-cli .arduino-data .arduino-user
	@echo "✓ Removed local Node.js installation and node modules"
	@echo "✓ Removed local Arduino CLI installation"
	@echo "✓ Removed Arduino firmware build artifacts"

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

	@if ! command -v tmux > /dev/null; then \
		echo "Warning: 'tmux' is not available. The 'make tmux-debug-run' target will not work."; \
		echo "Install with: sudo apt install tmux (Debian/Ubuntu) or brew install tmux (macOS)"; \
	else \
		echo "✓ tmux is installed"; \
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
	@echo "Installing root dependencies (markdownlint-cli2)..."
	@if [ -f .nodejs/bin/node ]; then \
		export PATH="$$(pwd)/.nodejs/bin:$$PATH"; \
	fi; \
	npm install
	@echo "✓ Root dependencies installed"
	@echo "Installing webapp dependencies (including ESLint)..."
	@if [ -f .nodejs/bin/node ]; then \
		export PATH="$$(pwd)/.nodejs/bin:$$PATH"; \
	fi; \
	cd webapp && npm install
	@echo "✓ Webapp dependencies installed"

# Arduino workspace-local paths
ARDUINO_DATA_DIR := $(shell pwd)/.arduino-data
ARDUINO_USER_DIR := $(shell pwd)/.arduino-user
ARDUINO_CLI_CONFIG := $(shell pwd)/arduino-cli.yaml

# File target: creates arduino-cli.yaml config file (host-specific)
arduino-cli.yaml:
	@echo "Creating Arduino CLI configuration..."
	@mkdir -p $(ARDUINO_DATA_DIR) $(ARDUINO_USER_DIR)
	@echo "board_manager:" > $@
	@echo "    additional_urls: []" >> $@
	@echo "directories:" >> $@
	@echo "    data: $(ARDUINO_DATA_DIR)" >> $@
	@echo "    user: $(ARDUINO_USER_DIR)" >> $@
	@echo "✓ Config file created at $@"
	@if [ -d .vscode ]; then \
		echo "Updating VS Code IntelliSense configuration..."; \
		$(MAKE) .vscode/c_cpp_properties.json; \
	fi

# File target: creates VS Code C/C++ IntelliSense configuration
.vscode/c_cpp_properties.json:
	@mkdir -p .vscode
	@echo "{" > $@
	@echo "    \"configurations\": [" >> $@
	@echo "        {" >> $@
	@echo "            \"name\": \"Arduino\"," >> $@
	@echo "            \"includePath\": [" >> $@
	@echo "                \"\$${workspaceFolder}/**\"," >> $@
	@echo "                \"\$${workspaceFolder}/.arduino-data/packages/arduino/hardware/avr/1.8.7/cores/arduino\"," >> $@
	@echo "                \"\$${workspaceFolder}/.arduino-data/packages/arduino/hardware/avr/1.8.7/variants/eightanaloginputs\"," >> $@
	@echo "                \"\$${workspaceFolder}/.arduino-data/packages/arduino/tools/avr-gcc/7.3.0-atmel3.6.1-arduino7/avr/include\"," >> $@
	@echo "                \"\$${workspaceFolder}/.arduino-data/packages/arduino/tools/avr-gcc/7.3.0-atmel3.6.1-arduino7/lib/gcc/avr/7.3.0/include\"," >> $@
	@echo "                \"\$${workspaceFolder}/.arduino-user/libraries/MAX6675_library\"," >> $@
	@echo "                \"\$${workspaceFolder}/.arduino-user/libraries/LiquidCrystal/src\"" >> $@
	@echo "            ]," >> $@
	@echo "            \"defines\": [" >> $@
	@echo "                \"ARDUINO=10607\"," >> $@
	@echo "                \"ARDUINO_AVR_NANO\"," >> $@
	@echo "                \"ARDUINO_ARCH_AVR\"," >> $@
	@echo "                \"F_CPU=16000000L\"," >> $@
	@echo "                \"__AVR_ATmega328P__\"" >> $@
	@echo "            ]," >> $@
	@echo "            \"compilerPath\": \"\$${workspaceFolder}/.arduino-data/packages/arduino/tools/avr-gcc/7.3.0-atmel3.6.1-arduino7/bin/avr-gcc\"," >> $@
	@echo "            \"cStandard\": \"c11\"," >> $@
	@echo "            \"cppStandard\": \"c++11\"," >> $@
	@echo "            \"intelliSenseMode\": \"gcc-x64\"" >> $@
	@echo "        }" >> $@
	@echo "    ]," >> $@
	@echo "    \"version\": 4" >> $@
	@echo "}" >> $@
	@echo "✓ VS Code IntelliSense configuration created"

.PHONY: arduino-help
arduino-help:
	@echo ""
	@echo "Arduino Firmware Development:"
	@echo ""
	@echo "First-time setup:"
	@echo "  make prepare-arduino            # Install Arduino CLI and AVR core (workspace-local)"
	@echo ""
	@echo "Build and upload targets:"
	@echo "  make build-arduino              # Compile firmware (requires prepare-arduino)"
	@echo "  make upload-arduino             # Upload to /dev/ttyUSB0 (auto-compiles first)"
	@echo "  make upload-arduino PORT=/dev/ttyUSB1  # Upload to specific port"
	@echo "  make backup-arduino             # Backup existing firmware from device"
	@echo "  make restore-arduino BACKUP=path.hex  # Restore backed-up firmware"
	@echo "  make clean-arduino              # Remove firmware/ build directory"
	@echo ""
	@echo "Direct arduino-cli usage (after prepare-arduino):"
	@echo "  .arduino-cli/bin/arduino-cli --config-file \$$(pwd)/arduino-cli.yaml compile --fqbn arduino:avr:nano:cpu=atmega328 --output-dir firmware sensorunit/arduino/sensorunit"
	@echo "  .arduino-cli/bin/arduino-cli --config-file \$$(pwd)/arduino-cli.yaml upload -p /dev/ttyUSB0 --fqbn arduino:avr:nano:cpu=atmega328 --input-dir firmware"
	@echo ""
	@echo "Note: All targets use workspace-local Arduino installation (.arduino-cli/, .arduino-data/, .arduino-user/)"
	@echo ""

.PHONY: prepare-arduino
prepare-arduino: arduino-cli.yaml
	@echo "Checking for Arduino CLI..."
	@if [ -f .arduino-cli/bin/arduino-cli ]; then \
		export PATH="$$(pwd)/.arduino-cli/bin:$$PATH"; \
	fi; \
	if command -v arduino-cli > /dev/null; then \
		ARDUINO_VERSION=$$(arduino-cli version | head -n1 | awk '{print $$3}'); \
		echo "✓ Arduino CLI $$ARDUINO_VERSION is already installed"; \
		ARDUINO_CLI_CMD="arduino-cli"; \
	else \
		echo "Arduino CLI not found. Installing locally to .arduino-cli/..."; \
		mkdir -p .arduino-cli/bin; \
		ARCH=$$(uname -m); \
		OS=$$(uname -s); \
		if [ "$$ARCH" = "x86_64" ]; then \
			ARDUINO_ARCH="64bit"; \
		elif [ "$$ARCH" = "aarch64" ] || [ "$$ARCH" = "arm64" ]; then \
			ARDUINO_ARCH="ARM64"; \
		elif [ "$$ARCH" = "armv7l" ]; then \
			ARDUINO_ARCH="ARMv7"; \
		elif [ "$$ARCH" = "armv6l" ]; then \
			ARDUINO_ARCH="ARMv6"; \
		else \
			echo "Error: Unsupported architecture: $$ARCH"; \
			exit 1; \
		fi; \
		ARDUINO_VERSION="1.0.4"; \
		if [ "$$OS" = "Linux" ]; then \
			ARDUINO_DIST="arduino-cli_$${ARDUINO_VERSION}_Linux_$${ARDUINO_ARCH}"; \
			ARCHIVE_EXT="tar.gz"; \
		elif [ "$$OS" = "Darwin" ]; then \
			ARDUINO_DIST="arduino-cli_$${ARDUINO_VERSION}_macOS_$${ARDUINO_ARCH}"; \
			ARCHIVE_EXT="tar.gz"; \
		else \
			echo "Error: Unsupported OS: $$OS"; \
			exit 1; \
		fi; \
		echo "Downloading Arduino CLI $${ARDUINO_VERSION} for $${OS}-$${ARDUINO_ARCH}..."; \
		curl -fsSL "https://github.com/arduino/arduino-cli/releases/download/v$${ARDUINO_VERSION}/$${ARDUINO_DIST}.$${ARCHIVE_EXT}" -o .arduino-cli/arduino-cli.$${ARCHIVE_EXT}; \
		echo "Extracting Arduino CLI..."; \
		tar -xzf .arduino-cli/arduino-cli.$${ARCHIVE_EXT} -C .arduino-cli/bin; \
		rm .arduino-cli/arduino-cli.$${ARCHIVE_EXT}; \
		chmod +x .arduino-cli/bin/arduino-cli; \
		echo "✓ Arduino CLI installed to .arduino-cli/bin/"; \
		ARDUINO_CLI_CMD=".arduino-cli/bin/arduino-cli"; \
	fi; \
	echo "Installing Arduino AVR core (for Arduino Nano)..."; \
	$$ARDUINO_CLI_CMD --config-file $(ARDUINO_CLI_CONFIG) core update-index; \
	$$ARDUINO_CLI_CMD --config-file $(ARDUINO_CLI_CONFIG) core install arduino:avr; \
	echo "✓ Arduino AVR core installed to $(ARDUINO_DATA_DIR)"; \
	echo "Installing required libraries..."; \
	$$ARDUINO_CLI_CMD --config-file $(ARDUINO_CLI_CONFIG) lib install "MAX6675 library@1.1.2"; \
	$$ARDUINO_CLI_CMD --config-file $(ARDUINO_CLI_CONFIG) lib install "LiquidCrystal@1.0.7"; \
	echo "✓ Required libraries installed to $(ARDUINO_USER_DIR)"; \
	echo ""; \
	echo "✓ Arduino CLI setup complete (workspace-local installation)"; \
	echo "  Data directory: $(ARDUINO_DATA_DIR)"; \
	echo "  User libraries: $(ARDUINO_USER_DIR)"; \
	echo "  Config file: $(ARDUINO_CLI_CONFIG)"; \
	@$(MAKE) arduino-help
	echo "  Config file: $(ARDUINO_CLI_CONFIG)"; \
	@$(MAKE) arduino-help

.PHONY: build
build: clean $(MODULES:%=$(BINDIR)/%)
	@echo "All Go binaries have been rebuilt: $(BUILD_TYPE)"

.PHONY: build-arduino
build-arduino:
	@echo "Compiling Arduino firmware for Nano (ATmega328P)..."
	@if [ ! -f .arduino-cli/bin/arduino-cli ]; then \
		echo "Error: Arduino CLI not found. Run 'make prepare-arduino' first."; \
		exit 1; \
	fi
	@if [ ! -f $(ARDUINO_CLI_CONFIG) ]; then \
		echo "Error: Arduino config not found. Run 'make prepare-arduino' first."; \
		exit 1; \
	fi
	@mkdir -p firmware
	@.arduino-cli/bin/arduino-cli --config-file $(ARDUINO_CLI_CONFIG) compile \
		--fqbn arduino:avr:nano:cpu=atmega328 \
		--output-dir firmware \
		sensorunit/arduino/sensorunit
	@echo "✓ Firmware compiled to firmware/sensorunit.ino.hex"
	@echo "  Board: Arduino Nano (ATmega328P)"
	@ls -lh firmware/sensorunit.ino.* 2>/dev/null | awk '{print "  " $$9 " (" $$5 ")"}'

.PHONY: clean-arduino
clean-arduino:
	@echo "Cleaning Arduino firmware build artifacts..."
	@rm -rf firmware
	@echo "✓ Firmware directory removed"

.PHONY: upload-arduino
upload-arduino: build-arduino
	@echo "Uploading firmware to Arduino Nano..."
	@if [ ! -f .arduino-cli/bin/arduino-cli ]; then \
		echo "Error: Arduino CLI not found. Run 'make prepare-arduino' first."; \
		exit 1; \
	fi
	@if [ -z "$(PORT)" ]; then \
		if [ ! -e /dev/ttyUSB0 ]; then \
			echo "Error: Arduino not found at /dev/ttyUSB0."; \
			echo "Specify port with: make upload-arduino PORT=/dev/ttyUSB1"; \
			exit 1; \
		fi; \
		UPLOAD_PORT="/dev/ttyUSB0"; \
	else \
		UPLOAD_PORT="$(PORT)"; \
	fi; \
	echo "Uploading to $$UPLOAD_PORT..."; \
	.arduino-cli/bin/arduino-cli --config-file $(ARDUINO_CLI_CONFIG) upload \
		-p $$UPLOAD_PORT \
		--fqbn arduino:avr:nano:cpu=atmega328 \
		--input-dir firmware
	@echo "✓ Firmware uploaded successfully"
	@echo "  Note: Arduino will reset automatically after upload"

.PHONY: backup-arduino
backup-arduino:
	@echo "Backing up Arduino firmware..."
	@if [ ! -f .arduino-cli/bin/arduino-cli ]; then \
		echo "Error: Arduino CLI not found. Run 'make prepare-arduino' first."; \
		exit 1; \
	fi
	@AVRDUDE=$$(find $(ARDUINO_DATA_DIR)/packages/arduino/tools/avrdude -name "avrdude" -type f | head -1); \
	AVRDUDE_CONF=$$(find $(ARDUINO_DATA_DIR)/packages/arduino/tools/avrdude -name "avrdude.conf" -type f | head -1); \
	if [ -z "$$AVRDUDE" ] || [ -z "$$AVRDUDE_CONF" ]; then \
		echo "Error: avrdude not found. Arduino AVR core may not be installed."; \
		echo "Run 'make prepare-arduino' to install it."; \
		exit 1; \
	fi; \
	if [ -z "$(PORT)" ]; then \
		if [ ! -e /dev/ttyUSB0 ]; then \
			echo "Error: Arduino not found at /dev/ttyUSB0."; \
			echo "Specify port with: make backup-arduino PORT=/dev/ttyUSB1"; \
			exit 1; \
		fi; \
		BACKUP_PORT="/dev/ttyUSB0"; \
	else \
		BACKUP_PORT="$(PORT)"; \
	fi; \
	mkdir -p firmware/backup; \
	TIMESTAMP=$$(date +%Y%m%d_%H%M%S); \
	BACKUP_PREFIX="firmware/backup/arduino_backup_$$TIMESTAMP"; \
	echo "Reading flash memory from $$BACKUP_PORT..."; \
	$$AVRDUDE -C $$AVRDUDE_CONF -v -p atmega328p -c arduino -P $$BACKUP_PORT -b 115200 \
		-U flash:r:$$BACKUP_PREFIX.hex:i || exit 1; \
	echo "Reading EEPROM from $$BACKUP_PORT..."; \
	$$AVRDUDE -C $$AVRDUDE_CONF -v -p atmega328p -c arduino -P $$BACKUP_PORT -b 115200 \
		-U eeprom:r:$$BACKUP_PREFIX.eep:i || exit 1; \
	echo ""; \
	echo "✓ Firmware backed up successfully:"; \
	ls -lh $$BACKUP_PREFIX.* | awk '{print "  " $$9 " (" $$5 ")"}'

.PHONY: restore-arduino
restore-arduino:
	@echo "Restoring Arduino firmware from backup..."
	@if [ ! -f .arduino-cli/bin/arduino-cli ]; then \
		echo "Error: Arduino CLI not found. Run 'make prepare-arduino' first."; \
		exit 1; \
	fi
	@if [ -z "$(BACKUP)" ]; then \
		echo "Error: BACKUP parameter required."; \
		echo ""; \
		echo "Available backups:"; \
		if [ -d firmware/backup ]; then \
			ls -1t firmware/backup/*.hex 2>/dev/null | head -10 | sed 's/^/  /' || echo "  (no backups found)"; \
		else \
			echo "  (no backup directory found)"; \
		fi; \
		echo ""; \
		echo "Usage:"; \
		echo "  make restore-arduino BACKUP=firmware/backup/arduino_backup_20260308_143022.hex"; \
		echo "  make restore-arduino BACKUP=firmware/backup/arduino_backup_20260308_143022.hex PORT=/dev/ttyUSB1"; \
		exit 1; \
	fi
	@if [ ! -f "$(BACKUP)" ]; then \
		echo "Error: Backup file not found: $(BACKUP)"; \
		exit 1; \
	fi
	@if [ -z "$(PORT)" ]; then \
		if [ ! -e /dev/ttyUSB0 ]; then \
			echo "Error: Arduino not found at /dev/ttyUSB0."; \
			echo "Specify port with: make restore-arduino BACKUP=... PORT=/dev/ttyUSB1"; \
			exit 1; \
		fi; \
		RESTORE_PORT="/dev/ttyUSB0"; \
	else \
		RESTORE_PORT="$(PORT)"; \
	fi; \
	echo "Restoring $(BACKUP) to $$RESTORE_PORT..."; \
	.arduino-cli/bin/arduino-cli upload \
		-p $$RESTORE_PORT \
		--fqbn arduino:avr:nano:cpu=atmega328 \
		--input-file $(BACKUP)
	@echo "✓ Firmware restored successfully"
	@echo "  Note: Arduino has been reset"

.PHONY: lint
lint: lint-golang lint-markdown lint-webapp
	@echo "✓ All linting completed"

.PHONY: lint-golang
lint-golang:
	@for mod in $(MODULES) types tests; do \
		if [ -f $$mod/go.mod ]; then \
			echo "Linting $$mod..."; \
			(cd $$mod && golangci-lint run ./... || true); \
		fi; \
	done

.PHONY: lint-markdown
lint-markdown:
	@echo "Linting markdown files..."
	@if [ -f .nodejs/bin/node ]; then \
		export PATH="$$(pwd)/.nodejs/bin:$$PATH"; \
	fi; \
	if [ -f node_modules/.bin/markdownlint-cli2 ]; then \
		npm run lint:markdown || true; \
	else \
		echo "Warning: markdownlint-cli2 is not installed. Run 'make prepare' first."; \
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
		 echo ""; \
		 echo "⚠️  IMPORTANT: Edit /etc/opt/halko.cfg before starting services:"; \
		 echo "   - Set network_interface to match your system (run 'ip addr show')"; \
		 echo "   - Set serial_device to your Arduino path (e.g., /dev/ttyUSB0)"; \
		 echo "   - Set shelly_address to your Shelly device IP"; \
		 echo "   See templates/README.md for details"; \
		 echo ""; \
	else \
		 echo "/etc/opt/halko.cfg already exists, not overwriting."; \
	fi

.PHONY: systemd-units
systemd-units: install
	@echo "Creating and installing systemd unit files for all binaries except simulator..."
	for bin in $(MODULES); do \
		if [ "$$bin" != "simulator" ]; then \
			if [ "$$bin" = "dbusunit" ]; then \
				sudo cp templates/halko-dbusunit.service /etc/systemd/system/halko-dbusunit.service; \
				sudo systemctl daemon-reload; \
				sudo systemctl enable halko-dbusunit.service; \
				if systemctl is-active --quiet halko-dbusunit.service; then \
					sudo systemctl restart halko-dbusunit.service; \
				else \
					sudo systemctl start halko-dbusunit.service; \
				fi; \
			else \
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
		fi; \
	done
	@echo "Systemd unit files installed and services enabled."

.PHONY: install-webapp
install-webapp: build-webapp
	@echo "Installing webapp to /var/www/halko..."
	@sudo install -d /var/www/halko
	@sudo cp -r webapp/dist/* /var/www/halko/
	@echo "Installing nginx configuration..."
	@sudo install -m 644 webapp/nginx-host.conf /etc/nginx/sites-available/halko
	@echo ""
	@echo "✓ Webapp installed to /var/www/halko"
	@echo "✓ Nginx config installed to /etc/nginx/sites-available/halko"
	@echo ""
	@echo "To enable and start:"
	@echo "  sudo ln -s /etc/nginx/sites-available/halko /etc/nginx/sites-enabled/"
	@echo "  sudo nginx -t"
	@echo "  sudo systemctl reload nginx"
	@echo ""
	@echo "Access the webapp at http://localhost/ or http://your-server-ip/"

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

.PHONY: clean-webapp
clean-webapp:
	@echo "Cleaning webapp build artifacts..."
	@rm -rf webapp/dist webapp/node_modules webapp/.parcel-cache
	@echo "✓ Webapp cleaned"

.PHONY: run-webapp
run-webapp: $(BINDIR)/halkoctl
	@echo "Starting webapp development server..."
	@if [ -f .nodejs/bin/node ]; then \
		export PATH="$$(pwd)/.nodejs/bin:$$PATH"; \
	fi; \
	cd webapp && npm start

.PHONY: build-webapp
build-webapp: clean-webapp $(BINDIR)/halkoctl
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

.PHONY: lint-webapp
lint-webapp:
	@echo "Linting webapp..."
	@if [ -f .nodejs/bin/node ]; then \
		export PATH="$$(pwd)/.nodejs/bin:$$PATH"; \
	fi; \
	cd webapp && npm run lint

.PHONY: tmux-debug-run tmux-debug-stop
tmux-debug-run: all
	@LOGLEVEL=$(LOGLEVEL) SIMULATOR=$(SIMULATOR) ./scripts/tmux-debug-start.sh

tmux-debug-stop:
	@./scripts/tmux-debug-stop.sh

.PHONY: monitor-memory
monitor-memory:
	@echo "Starting memory monitor for Halko processes..."
	@./scripts/monitor-memory.py $(MONITOR_ARGS)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  help                       Show this help message (default)."
	@echo "  prepare                    Check for required tools, install Node.js if needed, setup workspace."
	@echo ""
	@echo "Arduino Firmware:"
	@echo "  prepare-arduino            Install Arduino CLI locally and setup AVR core + libraries."
	@echo "                               Creates arduino-cli.yaml config file automatically."
	@echo "  build-arduino              Compile Arduino firmware for Nano (ATmega328P) to firmware/."
	@echo "  upload-arduino             Upload compiled firmware to Arduino (default: /dev/ttyUSB0)."
	@echo "                               Override port: make upload-arduino PORT=/dev/ttyUSB1"
	@echo "  backup-arduino             Backup existing firmware from Arduino to firmware/backup/."
	@echo "                               Creates timestamped .hex and .eep files. Override port as above."
	@echo "  restore-arduino            Restore a backed-up firmware to Arduino."
	@echo "                               Usage: make restore-arduino BACKUP=firmware/backup/file.hex"
	@echo "  clean-arduino              Remove Arduino firmware build artifacts."
	@echo "  arduino-help               Show Arduino CLI usage information."
	@echo ""
	@echo "Build Targets:"
	@echo "  all                        Build all Go executables to bin/ directory."
	@echo "  build                      Clean and rebuild all Go executables."
	@echo "  clean                      Remove bin/ directory (Go binaries only)."
	@echo "  clean-webapp               Remove webapp build artifacts (dist/, node_modules, cache)."
	@echo "  distclean                  Like clean + clean-webapp, plus removes local Node.js installation."
	@echo ""
	@echo "Production Installation (Raspberry Pi / Host):"
	@echo "  install                    Install all binaries (except simulator) to /opt/halko."
	@echo "  systemd-units              Create, install, and enable systemd service units."
	@echo "  install-webapp             Install webapp to /var/www/halko with nginx config."
	@echo ""
	@echo "  Note: For memory-constrained systems, use: OPTIMIZED=yes make build"
	@echo "        This reduces binary size by ~30%."
	@echo ""
	@echo "Development & Testing:"
	@echo "  run-webapp                 Start webapp development server with hot reload."
	@echo "  build-webapp               Build webapp for production to webapp/dist/."
	@echo "  test                       Run all tests (config, program validation, Shelly API)."
	@echo "  test-config                Run configuration loading tests."
	@echo "  test-program-validation    Run program JSON validation tests."
	@echo "  test-shelly-api            Run Shelly device API compatibility tests."
	@echo "  monitor-memory             Monitor process memory usage (requires running processes)."
	@echo "                               Examples: make monitor-memory MONITOR_ARGS='-p controlunit -i 5'"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint                       Run all linters (golang, markdown, webapp)."
	@echo "  lint-golang                Run golangci-lint on all Go modules."
	@echo "  lint-markdown              Run markdownlint-cli2 on all markdown files."
	@echo "  lint-webapp                Run ESLint on webapp TypeScript/React code."
	@echo "  fmt-changed                Reformat changed Go files compared to main branch."
	@echo "  go-tidy                    Run go mod tidy on all modules."
	@echo "  update-modules             Update all go.mod dependencies and tidy them."
	@echo ""
	@echo "Tmux Debug Environment:"
	@echo "  tmux-debug-run             Start services in tmux session for native debugging."
	@echo "                               Starts: simulator, powerunit, controlunit, webapp"
	@echo "                               Default loglevel: 3 (DEBUG)"
	@echo "                               Usage: LOGLEVEL=4 make tmux-debug-run"
	@echo "                               Usage: SIMULATOR=thermodynamic make tmux-debug-run"
	@echo "                               Usage: LOGLEVEL=4 SIMULATOR=differential make tmux-debug-run"
	@echo "  tmux-debug-stop            Stop and terminate the tmux debug session."

.DEFAULT_GOAL := help
