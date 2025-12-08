.PHONY: all build test lint generate clean download-specs help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
BINARY_NAME=gofhir

# Directories
CMD_DIR=./cmd/gofhir
PKG_DIR=./pkg/...
INTERNAL_DIR=./internal/...
SPECS_DIR=./specs

# FHIR versions
FHIR_R4_VERSION=4.0.1
FHIR_R4B_VERSION=4.3.0
FHIR_R5_VERSION=5.0.0

all: lint test build

## build: Build the CLI binary
build:
	$(GOBUILD) -o bin/$(BINARY_NAME) $(CMD_DIR)

## test: Run all tests
test:
	$(GOTEST) -v -race -coverprofile=coverage.out $(PKG_DIR) $(INTERNAL_DIR)

## test-short: Run tests without race detector (faster)
test-short:
	$(GOTEST) -v -coverprofile=coverage.out $(PKG_DIR) $(INTERNAL_DIR)

## coverage: Show test coverage in browser
coverage: test
	$(GOCMD) tool cover -html=coverage.out

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format code
fmt:
	$(GOFMT) -s -w .

## generate: Generate FHIR types from StructureDefinitions
generate:
	$(GOCMD) run $(CMD_DIR) generate --specs $(SPECS_DIR) --output ./pkg/fhir

## generate-r4: Generate only R4 types
generate-r4:
	$(GOCMD) run $(CMD_DIR) generate --specs $(SPECS_DIR) --output ./pkg/fhir --version r4

## generate-r4b: Generate only R4B types
generate-r4b:
	$(GOCMD) run $(CMD_DIR) generate --specs $(SPECS_DIR) --output ./pkg/fhir --version r4b

## generate-r5: Generate only R5 types
generate-r5:
	$(GOCMD) run $(CMD_DIR) generate --specs $(SPECS_DIR) --output ./pkg/fhir --version r5

## generate-all: Generate types for all FHIR versions
generate-all: generate-r4 generate-r4b generate-r5

## download-specs: Download FHIR specifications
download-specs:
	./scripts/download-specs.sh

## download-specs-r4: Download only R4 specifications
download-specs-r4:
	./scripts/download-specs.sh r4

## clean: Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

## deps: Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## verify: Verify dependencies
verify:
	$(GOMOD) verify

## update: Update dependencies
update:
	$(GOGET) -u ./...
	$(GOMOD) tidy

## help: Show this help
help:
	@echo "GoFHIR - FHIR Toolkit for Go"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'
