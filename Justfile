# tail-claude-hud

# Default: run tests
default: test

# -----------------------------------------------------------
# Build
# -----------------------------------------------------------

# Build the binary
build:
    go build -o bin/tail-claude-hud ./cmd/tail-claude-hud

# Build with version info
build-release version:
    go build -ldflags "-X main.version={{version}}" -o bin/tail-claude-hud ./cmd/tail-claude-hud

# Install to GOPATH/bin
install:
    go install ./cmd/tail-claude-hud

# -----------------------------------------------------------
# Test
# -----------------------------------------------------------

# Run all tests
test:
    go test ./...

# Run tests with verbose output
test-verbose:
    go test -v ./...

# Run tests with race detector
test-race:
    go test -race ./...

# Run benchmarks
bench:
    go test -bench=. -benchmem ./internal/...

# -----------------------------------------------------------
# Lint
# -----------------------------------------------------------

# Format code
fmt:
    go fmt ./...

# Vet code
vet:
    go vet ./...

# Run all checks
check: fmt vet test

# -----------------------------------------------------------
# Dev
# -----------------------------------------------------------

# Run with sample stdin (pipe a JSON file)
run-sample:
    cat testdata/sample-stdin.json | go run ./cmd/tail-claude-hud

# Benchmark a single render tick
bench-tick:
    @echo "Build first, then use hyperfine:"
    @echo "  just build"
    @echo "  hyperfine --warmup 3 'cat testdata/sample-stdin.json | ./bin/tail-claude-hud'"

# Clean build artifacts
clean:
    rm -rf bin/ cpu.prof mem.prof
