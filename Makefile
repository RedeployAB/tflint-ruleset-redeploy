PACKAGES := $$(go list ./... | grep -v integration)

.DEFAULT_GOAL := build

# Build targets
build:
ifeq ($(OS),Windows_NT)
	go build -o tflint-ruleset-redeploy.exe
else
	go build -o tflint-ruleset-redeploy
endif

install: build
ifeq ($(OS),Windows_NT)
	mkdir -p $(USERPROFILE)/.tflint.d/plugins
	mv tflint-ruleset-redeploy.exe $(USERPROFILE)/.tflint.d/plugins
else
	mkdir -p ~/.tflint.d/plugins
	mv ./tflint-ruleset-redeploy ~/.tflint.d/plugins
endif

# Test targets
test:
	go test --count=1 $(PACKAGES)

coverage:
	go test -race --count=1 -coverprofile=coverage.out $(PACKAGES)
	go tool cover -html=coverage.out -o coverage.html

benchmarks:
	go test -bench=. -benchmem ./rules/

e2e: install
	go test -v ./integration/

# Code quality targets
lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Utility targets
clean:
	rm -f tflint-ruleset-redeploy tflint-ruleset-redeploy.exe
	rm -f coverage.out coverage.html

help:
	@echo "Available targets:"
	@echo "  build      - Build the plugin"
	@echo "  install    - Install plugin locally"
	@echo "  test       - Run unit tests"
	@echo "  coverage   - Generate coverage report"
	@echo "  benchmarks - Run benchmarks"
	@echo "  e2e        - Run end-to-end tests"
	@echo "  lint       - Run golangci-lint"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  vulncheck  - Scan for known vulnerabilities (govulncheck)"
	@echo "  clean      - Remove build artifacts"

.PHONY: build install test coverage benchmarks e2e lint fmt vet vulncheck clean help
