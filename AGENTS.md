# AGENTS.md

Instructions for agentic coding assistants working in this repository.

## Project Overview

Go web server (macaron-based) for issuing offline licenses for AppsCode products.
Module path: `go.bytebuilders.dev/offline-license-server`

## Build, Lint, and Test Commands

### Build

```bash
# Local binary build (via Docker)
make build

# Cross-compile
make build-linux-amd64
make build-darwin-arm64
```

### Lint

```bash
# Lint via Docker (uses golangci-lint)
make lint

# Format code (runs goimports, gofmt, reimport3.py)
make fmt
```

### Test

```bash
# Run all unit tests
make test

# Run a single test
go test -mod=vendor -race -run TestFormatRFC822Email ./pkg/server/

# Run tests for a specific package
go test -mod=vendor -race ./pkg/server/...

# Run with verbose output
go test -mod=vendor -race -v ./pkg/server/
```

### Verify

```bash
# Check generated files and go modules are up to date
make verify
```

## Code Style

### Imports

Group imports into three blocks separated by blank lines:
1. Standard library
2. Local project imports (`go.bytebuilders.dev/offline-license-server/...`)
3. Third-party imports

Use import aliases when the default package name is ambiguous or verbose:
- `ep "gomodules.xyz/email-providers"`
- `v "gomodules.xyz/x/version"`
- `gdrive "gomodules.xyz/gdrive-utils"`
- `apierrors "k8s.io/apimachinery/pkg/api/errors"`
- `metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"`

### Formatting

- Use `gofmt` with the `interface{} -> any` rewrite rule (configured in `.golangci.yml`).
- Use `goimports` for import ordering.
- Use `interface{}` only if required by Go version; prefer `any`.
- Shell scripts use 4-space indentation (`shfmt -ci -i 4`).

### Types and Naming

- Exported types use PascalCase: `Server`, `Options`, `LicenseForm`, `ProductLicense`.
- Unexported fields use camelCase: `listmonkHost`, `blockedDomains`.
- Constructor pattern: `NewXxx()` functions (`New()`, `NewOptions()`, `NewRootCmd()`).
- Handler methods: `HandleXxx()` (`HandleRegisterEmail`, `HandleIssueLicense`).
- Command builders: `NewCmdXxx()` (`NewCmdRun`, `NewCmdCreate`).
- Configuration structs use `AddFlags(fs *pflag.FlagSet)` method for CLI flags.

### Error Handling

- Return errors up the call stack; do not log-and-continue in library code.
- Use `fmt.Errorf("context: %v", err)` or `github.com/pkg/errors.Wrap(err, "msg")` for wrapping.
- In HTTP handlers, write error status then respond with `err.Error()` as the body.
- Use `klog.ErrorS(err, "message", "key", value)` for structured error logging.
- Silently discard errors with explicit intent: `_ = s.geodb.Close() // nolint:errcheck`

### Testing

- Test files use `_test.go` suffix.
- Both internal (`package server`) and external (`package server_test`) test packages are used.
- Table-driven tests with `t.Run()` are preferred.
- Use `t.SkipNow()` for tests requiring external credentials/services.
- Tests that need Google API clients or SMTP should be skipped in CI.

### License Header

All `.go` files must include the Apache 2.0 license header. Enforcement:
```bash
make check-license   # verify
make add-license     # add missing headers
```

### Struct Tags

- JSON tags: `json:"field_name,omitempty"`
- Form binding tags: `form:"field_name" binding:"Required;Email"`
- Use `json:",inline"` for embedded structs in JSON serialization.

### Concurrency

- Use `go func()` goroutines for fire-and-forget background work.
- Catch panics in goroutines via `defer` + `recover` or let them propagate to `panic(err)`.
- Use `context.TODO()` when a real context is not yet available.

## CI Pipeline

`make ci` runs: `verify` -> `check-license` -> `lint` -> `build` -> `unit-tests`
