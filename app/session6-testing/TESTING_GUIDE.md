# Session 6: Testing & Quality Assurance - COMPLETE GUIDE

## Overview

This session focuses on **comprehensive testing** for our EASM application. Session 6 is a copy of Session 5 code with **complete test coverage** added.

## Quick Start

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./internal/model/...
```

## Testing Patterns Used

Always test:

- ✅ Empty input
- ✅ Nil values
- ✅ Maximum values
- ✅ Minimum values
- ✅ Boundary conditions
- ✅ Invalid input
- ✅ Special characters (SQL injection attempts)

### 4. Benchmarking

Performance testing for critical functions:

```go
func BenchmarkValidateDomain(b *testing.B) {
    validator := NewAssetValidator()
    for i := 0; i < b.N; i++ {
        validator.ValidateDomain("api.example.com")
    }
}
```

Run with:

```bash
go test -bench=BenchmarkValidateDomain -benchmem
```

## Coverage Analysis

### Check Coverage

```bash
# Quick check
go test -cover ./...

# Detailed by package
go test -cover ./internal/model/...
go test -cover ./internal/validator/...
```

### HTML Coverage Report

```bash
# Generate profile
go test -coverprofile="coverage.out" ./...

# View in browser (shows line-by-line coverage)
go tool cover -html="coverage.out"
```

**Color coding:**

- 🟢 Green = Covered
- 🔴 Red = Not covered
- ⚪ Gray = Not executable

## Running Specific Tests

```bash
# Run all tests
go test ./...

# Run specific package
go test ./internal/model/...

# Run specific test function
go test -run TestIsValidType

# Run specific sub-test
go test -run TestIsValidType/valid_domain

# Run with race detection
go test -race ./...

# Run with verbose output
go test -v ./internal/validator/...

# Failed tests only
go test -v ./... | grep -E "FAIL|RUN"
```

### Pre-Commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running tests..."
go test ./...
if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

echo "Running linter..."
golangci-lint run
if [ $? -ne 0 ]; then
    echo "Linter failed. Commit aborted."
    exit 1
fi
```

### Next Steps

1. Run all tests: `go test -v -cover ./...`
2. Generate coverage report: `go test -coverprofile=coverage.out ./...`
3. View HTML report: `go tool cover -html=coverage.out`
4. Complete exercises (service, handler, integration tests)
5. Set up CI/CD pipeline
6. Add more edge case tests

### Resources

- [Go Testing Documentation](https://go.dev/doc/tutorial/add-a-test)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Advanced Testing Patterns](https://go.dev/blog/subtests)
- [Testify Library](https://github.com/stretchr/testify)
- [Go Test Best Practices](https://go.dev/wiki/TestComments)
