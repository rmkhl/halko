MODULES = configurator executor powerunit simulator
BINDIR = bin

.PHONY: all
all: $(MODULES:%=$(BINDIR)/%)

$(BINDIR)/%: %/main.go | $(BINDIR)
	go build -o $@ ./$*/

$(BINDIR):
	mkdir -p $(BINDIR)

.PHONY: clean
clean:
	rm -rf $(BINDIR)

.PHONY: lint
lint:
	@for mod in $(MODULES) types; do \
		if [ -f $$mod/go.mod ]; then \
			echo "Linting $$mod..."; \
			(cd $$mod && golangci-lint run ./... || true); \
		fi; \
	done

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
	@echo "Installing config to /etc/opt/halko.cfg if not present..."
	sudo install -d /etc/opt
	@if [ ! -f /etc/opt/halko.cfg ]; then \
		sudo install -m 644 halko.cfg.sample /etc/opt/halko.cfg; \
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
			sudo systemctl enable --now halko@$$bin.service; \
		fi; \
	done
	sudo systemctl daemon-reload
	@echo "Systemd unit files installed and services enabled."

.PHONY: fmt-changed
fmt-changed:
	@git diff --name-only master...HEAD | grep '\.go$$' | xargs -r golangci-lint run --fix
	@echo "Reformatted changed Go files compared to main branch using golangci-lint."

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  help             Show this help message. (default)"
	@echo "  all              Build all Go executables to bin/ directory."
	@echo "  clean            Remove the bin/ directory and all built executables."
	@echo "  lint             Run golangci-lint on all modules."
	@echo "  update-modules   Update all go.mod dependencies and tidy them."
	@echo "  install           Install all binaries except simulator to /opt/halko and copy halko.cfg.sample to /etc/opt/halko.cfg if not present."
	@echo "  systemd-units    Create, install, and enable systemd unit files for all binaries except simulator."
	@echo "  fmt-changed      Reformat changed Go files compared to the main branch using golangci-lint."

.DEFAULT_GOAL := help
