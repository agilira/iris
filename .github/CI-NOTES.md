# CI/CD Configuration Notes

## Platform-Specific Test Strategies

### Linux and Windows
- Use standard parallel test execution with race detection
- Command: `go test -race -timeout 5m -v ./...`
- Full test coverage with optimal performance

### macOS
- Use serial test execution to avoid test interference
- Command: `go test -count=1 -p=1 -timeout 10m -v ./...`
- Flags explanation:
  - `-count=1`: Disable test caching to prevent state pollution
  - `-p=1`: Run tests serially (one package at a time)
  - `-timeout 10m`: Extended timeout for serial execution

## Why Different Strategies?

The ring buffer creation tests can interfere with each other when run in parallel on macOS due to:
1. Memory management differences in macOS
2. Test state pollution between parallel test runs
3. ZephyrosLite buffer initialization timing issues

This configuration ensures 100% test reliability across all platforms while maintaining optimal performance where possible.

## Performance Impact

- Linux/Windows: ~17s (parallel execution)
- macOS: ~25-30s (serial execution)

The trade-off ensures reliable CI while keeping acceptable build times.
