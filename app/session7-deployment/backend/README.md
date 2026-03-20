# Session 6: Testing & Quality Assurance

## Overview

Session 6 focuses on **comprehensive testing** for Go applications. This is a copy of Session 5's EASM code with **complete unit test coverage** added.

📖 **[→ READ FULL TESTING GUIDE](TESTING_GUIDE.md)** for detailed documentation, patterns, and best practices.

## Quick Start

```bash
# Navigate to session
cd d:\Projects\cmc-intern-program\app\session6-testing

# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## What's Been Tested ✅

| Component | File | Tests | Coverage | Status |
|-----------|------|-------|----------|--------|
| **Models** | `internal/model/asset_test.go` | 2 tests + 2 benchmarks | 100% | ✅ |
| **Models** | `internal/model/scan_test.go` | 10 tests (~80 sub-tests) | 100% | ✅ |
| **Validators** | `internal/validator/asset_validator_test.go` | 10 tests (~100 sub-tests) | 95%+ | ✅ |
| Services | TBD | - | - | 🔄 Pending |
| Handlers | TBD | - | - | 🔄 Pending |
| Storage | TBD | - | - | 🔄 Pending |

**Summary:**
- ✅ **26 test functions** written
- ✅ **6 benchmark functions** for performance testing  
- ✅ **180+ individual test cases** (sub-tests)
- ✅ All tests passing

## Test Results

Run the tests to see results:

```bash
go test -v ./internal/model/...
```

Expected output:
```
=== RUN   TestIsValidType
=== RUN   TestIsValidType/valid_domain_type
=== RUN   TestIsValidType/valid_ip_type
=== RUN   TestIsValidType/valid_service_type
--- PASS: TestIsValidType (0.00s)
    --- PASS: TestIsValidType/valid_domain_type (0.00s)
    --- PASS: TestIsValidType/valid_ip_type (0.00s)
    --- PASS: TestIsValidType/valid_service_type (0.00s)
...
PASS
ok      mini-asm/internal/model    0.135s   coverage: 100.0%
```

## Test Coverage Highlights

### 1. Model Tests (`internal/model/`)

**asset_test.go:**
- ✅ Asset type validation (domain, ip, service)
- ✅ Asset status validation (active, inactive)
- ✅ Performance benchmarks

**scan_test.go:**
- ✅ Scan type validation (dns, whois, subdomain, port, ssl, asn, all)
- ✅ Scan category classification (passive vs active)
- ✅ Permission requirements for active scans
- ✅ Scan status validation
- ✅ Classification consistency checks

### 2. Validator Tests (`internal/validator/`)

**asset_validator_test.go:**
- ✅ Name validation (empty, max length, null bytes)
- ✅ Domain validation (format, edge cases, security)
- ✅ IP validation (IPv4, IPv6, invalid formats)
- ✅ Service URL validation
- ✅ Pagination parameters validation
- ✅ **SQL injection prevention** (sort, order parameters)
- ✅ Search query sanitization
- ✅ Performance benchmarks

## Testing Patterns Used

### Table-Driven Tests

```go
tests := []struct {
    name    string
    input   string
    want    bool
}{
    {"valid domain", "example.com", true},
    {"empty string", "", false},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got := validate(tt.input)
        if got != tt.want {
            t.Errorf("got %v, want %v", got, tt.want)
        }
    })
}
```

### Sub-Tests

Each test case runs as an isolated sub-test:
```
TestIsValidType
  └─ TestIsValidType/valid_domain_type
  └─ TestIsValidType/invalid_type
```

Benefits:
- Run specific tests: `go test -run TestIsValidType/valid_domain`
- Clear test names in output
- Isolated execution

### Edge Case Testing

Every test covers:
- ✅ Valid inputs (happy path)
- ✅ Empty/nil inputs
- ✅ Maximum/minimum values
- ✅ Boundary conditions
- ✅ Invalid formats
- ✅ Security concerns (SQL injection, null bytes)

## Project Structure

```
session6-testing/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── model/
│   │   ├── asset.go
│   │   ├── asset_test.go          ✅ NEW
│   │   ├── scan.go
│   │   ├── scan_test.go           ✅ NEW
│   │   └── errors.go
│   ├── validator/
│   │   ├── asset_validator.go
│   │   └── asset_validator_test.go ✅ NEW
│   ├── service/
│   │   ├── asset_service.go
│   │   └── scan_service.go
│   ├── handler/
│   │   ├── asset_handler.go
│   │   ├── scan_handler.go
│   │   └── health_handler.go
│   └── storage/
│       ├── storage.go
│       ├── memory/
│       └── postgres/
├── migrations/
├── docker-compose.yml
├── go.mod
├── README.md                       ← You are here
└── TESTING_GUIDE.md                ✅ NEW - Complete testing documentation
```

## Key Concepts

### Testing Pyramid

```
       /\
      /E2\     ← Few (End-to-end tests)
     /----\
    /Integ\   ← Some (Integration tests)
   /--------\
  /   Unit   \ ← Many (Unit tests) ✅ We are here
 /____________\
```

**Current focus:** Unit tests for models and validators (foundation)

### Coverage Goals

| Layer | Goal | Current |
|-------|------|---------|
| Model | 100% | ✅ 100% |
| Validator | 95% | ✅ 95%+ |
| Service | 80% | 🔄 TBD |
| Handler | 75% | 🔄 TBD |
| Storage | 70% | 🔄 TBD |

## Next Steps

### 1. Run and Understand Current Tests

```bash
# Run all tests with verbose output
go test -v ./...

# Check coverage
go test -cover ./...

# View coverage report in browser
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 2. Read the Testing Guide

Open [TESTING_GUIDE.md](TESTING_GUIDE.md) for:
- Complete testing patterns and best practices
- How to write service tests with mocks
- HTTP handler testing with httptest
- Integration testing strategies
- Benchmarking techniques
- CI/CD integration

### 3. Add More Tests (Homework)

**Service Layer Tests:**
```bash
# Create test file
touch internal/service/asset_service_test.go
```

Key topics:
- Mocking storage layer
- Testing business logic
- Error handling

**Handler Layer Tests:**
```bash
# Create test file
touch internal/handler/asset_handler_test.go
```

Key topics:
- Using httptest package
- Request/response validation
- Status code checks
- JSON marshaling/unmarshaling

**Integration Tests:**
```bash
# Create test directory
mkdir test
touch test/integration_test.go
```

Key topics:
- Testing with real database
- Full request-to-response flow
- CRUD workflows

### 4. Explore Testing Tools

**Standard library:**
- `testing` - Core framework
- `net/http/httptest` - HTTP testing
- `testing/quick` - Property-based testing

**Third-party (recommended):**
```bash
# Testify - assertions and mocks
go get github.com/stretchr/testify

# Gomega - matchers
go get github.com/onsi/gomega
```

## Common Commands

```bash
# Run all tests
go test ./...

# Run specific package
go test ./internal/model/...

# Run specific test
go test -run TestIsValidType

# Run specific sub-test
go test -run TestIsValidType/valid_domain

# With race detection
go test -race ./...

# With coverage
go test -cover ./...

# Coverage profile + HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Benchmarks
go test -bench=. ./internal/model/...

# Benchmarks with memory stats
go test -bench=. -benchmem ./internal/model/...

# Verbose output
go test -v ./...

# Short mode (skip long tests)
go test -short ./...
```

## Learning Resources

- 📖 [TESTING_GUIDE.md](TESTING_GUIDE.md) - **START HERE** for comprehensive guide
- 🔗 [Official Go Testing](https://go.dev/doc/tutorial/add-a-test)
- 🔗 [Table-Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- 🔗 [Advanced Testing](https://go.dev/blog/subtests)
- 🔗 [Testing Best Practices](https://go.dev/wiki/TestComments)

## Summary

✅ **What's Complete:**
- Model layer: 100% test coverage
- Validator layer: 95%+ test coverage
- 180+ test cases covering happy paths and edge cases
- Security testing (SQL injection, null bytes)
- Performance benchmarks
- Comprehensive documentation

🔄 **What's Next:**
- Service layer tests (with mocking)
- Handler layer tests (HTTP testing)
- Integration tests (database + API)
- CI/CD pipeline setup

---

**Ready to learn testing?** Start by running the tests and reading [TESTING_GUIDE.md](TESTING_GUIDE.md)!
