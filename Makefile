MODULES = configurator executor powerunit simulator
BINDIR = bin

.PHONY: all
all: $(MODULES:%=$(BINDIR)/%)

$(BINDIR)/%: %/main.go | $(BINDIR)
	GO111MODULE=on go build -o $@ ./$*/

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

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  help             Show this help message. (default)"
	@echo "  all              Build all Go executables to bin/ directory."
	@echo "  clean            Remove the bin/ directory and all built executables."
	@echo "  lint             Run golangci-lint on all modules."
	@echo "  update-modules   Update all go.mod dependencies and tidy them."

.DEFAULT_GOAL := help
