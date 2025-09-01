# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Testing
- Run all tests: `go test ./...`
- Run tests with race detection and coverage: `go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/pq.out -covermode=atomic ./...`
- Run single test: `go test -run TestName`
- Generate test coverage report: `make test_coverage`

### Linting and Formatting
- Run linter: `golangci-lint run`
- Format code: `go fmt ./...`

### Building
- Build: `go build ./...`
- Check dependencies: `go mod tidy`

## Architecture Overview

This is a Go library implementing a thread-safe binary heap-based priority queue with the following key components:

### Core Types
- **`Item` interface**: Defines the contract for items that can be stored in the priority queue. Items must provide:
  - `ID()` - unique identifier
  - `Priority()` - int64 priority for sorting (min-heap, lower values have higher priority)
  - `GroupID()` - group identifier for bulk operations

- **`BinHeap[T Item]`**: Generic binary heap implementation with:
  - Thread-safe operations using sync.Cond
  - Configurable maximum capacity with overflow protection
  - Existence tracking via shadow map
  - Monotonic stack for efficient bulk removals

### Key Files
- `binary_heap.go`: Main priority queue implementation with Insert, ExtractMin, Remove operations
- `monotonic_stack.go`: Utility for tracking removal intervals efficiently
- `binary_heap_test.go`: Comprehensive test suite with concurrency tests
- `monotonic_stack_test.go`: Tests for the monotonic stack utility

### Design Patterns
- Uses generics (`BinHeap[T Item]`) for type safety
- Implements standard binary heap algorithms (fixUp/fixDown for heap property maintenance)
- Thread-safe with conditional variables for blocking ExtractMin when empty
- Bulk removal optimization using monotonic stack to track contiguous removal ranges

### Testing Approach
Tests are located in the root directory and use:
- `github.com/stretchr/testify` for assertions
- Concurrent testing with goroutines
- Coverage tracking with atomic mode
- Test items implement the Item interface with configurable priority/groupID