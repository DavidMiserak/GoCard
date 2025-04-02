# Filename: Makefile

.PHONY: clean
clean:
	rm -f GoCard gocard-linux-amd64 gocard-macos-amd64 gocard-windows-amd64.exe
	go clean

.PHONY: pre-commit-setup
pre-commit-setup:
	@echo "Setting up pre-commit hooks..."
	@echo "consider running <pre-commit autoupdate> to get the latest versions"
	pre-commit install
	pre-commit install --install-hooks
	pre-commit run --all-files

GoCard:
	go fmt ./...
	CGO_ENABLED=0 go build -o GoCard ./cmd/gocard

.PHONY: format
format:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go fmt ./...
	go test -timeout 3m -race ./... | tee test.log

.PHONY: test-cover
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tee coverage.log

.PHONY: test-cover-html
test-cover-html: test-cover
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Open with: xdg-open coverage.html"

.PHONY: test-cover-verbose
test-cover-verbose:
	go test -v -covermode=count -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: test-cover-report
test-cover-report:
	@./scripts/generate_coverage_report.sh

# Target for checking if coverage meets threshold (e.g., 70%)
.PHONY: test-cover-check
test-cover-check:
	@go test -coverprofile=coverage.out ./...
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | tr -d '%'); \
	echo "Total coverage: $$coverage%"; \
	if [ $$(echo "$$coverage < 70" | bc -l) -eq 1 ]; then \
		echo "Coverage is below threshold of 70%"; \
		exit 1; \
	else \
		echo "Coverage meets threshold of 70%"; \
	fi

.PHONY: build-all
build-all: gocard-linux-amd64 gocard-macos-amd64 gocard-windows-amd64.exe

gocard-linux-amd64:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -o $@ ./cmd/gocard

gocard-macos-amd64:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -v -o $@ ./cmd/gocard

gocard-windows-amd64.exe:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -v -o $@ ./cmd/gocard
