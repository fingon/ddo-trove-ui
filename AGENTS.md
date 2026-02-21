# AGENTS.md - Development Guide for DDO Trove UI

This document provides guidelines for agentic coding agents working in this repository.

## Project Overview

DDO Trove UI is a Go-based web application that provides a UI for browsing DDO (Dungeons and Dragons Online) item data from JSON files. It uses the [Templ](https://templ.guide/) package for type-safe HTML templating.

## Build, Lint, and Test Commands

### Running the Application

```bash
# Run with data directories
go run main.go example/local example/local2

# Or use make (builds + runs)
make run
```

### Building

```bash
# Full build (runs lint first, then compiles)
make build

# Build only templates (generates Go from .templ files)
make -C templates
```

### Linting

```bash
# Run golangci-lint (the main linter)
go tool golangci-lint run

# Or use make lint
make lint
```

### Testing

There are currently no unit tests in this project. If tests are added:

```bash
# Run all tests
go test ./...

# Run a single test
go test -run TestName ./...

# Run with verbose output
go test -v -run TestName ./...
```

### Code Generation

```bash
# Generate Go code from .templ files
go tool templ generate -f templates/layout.templ
go tool templ generate -f templates/index.templ
go tool templ generate -f templates/item_list.templ
```

### Development Tools

```bash
# Install required tools (golangci-lint, templ)
make install-tools

# Install pre-commit hooks (requires prek)
prek install

# Run all pre-commit checks
prek run --all-files
```

## Code Style Guidelines

### General

- This is a Go project using Go 1.25+
- Uses [Templ](https://templ.guide/) for HTML generation - **never edit the generated `*_templ.go` files directly**, edit the `.templ` source files instead
- All code must pass `golangci-lint run`

### Formatting

- Uses **gofumpt** (with extra rules enabled) for code formatting
- Uses **goimports** for import organization
- Uses **gci** for import ordering
- Run `golangci-lint run --fix` to auto-fix most formatting issues

### Imports

- Standard library imports first, then third-party imports
- Group imports with blank lines between groups
- Example:
```go
import (
    "context"
    "fmt"
    "log"
    "net/http"

    "github.com/fingon/ddo-trove-ui/db"
    "github.com/fingon/ddo-trove-ui/templates"
)
```

### Naming Conventions

- **Variables/Functions**: camelCase
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE for grouped constants
- **Types/Structs**: PascalCase
- **JSON struct tags**: snake_case (enforced by tagliatelle linter)
- **Package name**: lowercase, short (e.g., `db`, `templates`)
- **Receiver names**: short (1-2 chars), consistent (e.g., `i` for Item, `s` for Struct)

### Error Handling

- Use `fmt.Errorf` with `%w` for wrapped errors
- Return errors from functions when appropriate
- Use named return values for clearer error documentation
- Example:
```go
func LoadItemsFromDir(dirPath string) (*AllItems, error) {
    // ...
    if err != nil {
        return nil, fmt.Errorf("failed to read directory: %w", err)
    }
    // ...
}
```

### Linter Configuration

The project uses golangci-lint with these specific rules:

**Enabled linters:**
- bidichk, bodyclose, durationcheck, errchkjson, errname
- exptostd, goconst, gocritic, intrange, misspell
- nestif, nilerr, perfsprint, reassign, revive
- sloglint, sqlclosecheck, tagalign, tparallel
- unconvert, usetesting, wastedassign, whitespace

**Key settings:**
- `nestif` min-complexity: 7
- JSON struct tags must use snake_case (tagliatelle)
- revive rules enabled for: context-as-argument, error-naming, error-return, error-strings, exported, package-comments, receiver-naming, unused-parameter, var-naming, etc.

**Exceptions:**
- `db/item.go` has exceptions for tagliatelle and var-naming (due to external JSON schema compatibility)

### Pre-commit Hooks

The project uses pre-commit hooks (configured in `.pre-commit-config.yaml`):

1. Built-in hooks: check-case-conflict, check-merge-conflict, check-yaml, detect-private-key, end-of-file-fixer, mixed-line-ending, trailing-whitespace
2. Local hooks:
   - `fmt-golangci-lint`: Runs `golangci-lint run --fix` on Go files
   - `fmt-templ-guide`: Runs `templ fmt` on .templ files

### File Organization

```
/Users/mstenber/projects/ddo-trove-ui
├── main.go              # Application entry point and HTTP handlers
├── db/
│   └── item.go          # Data models and filtering logic
├── templates/
│   ├── *.templ          # Templ source files (edit these!)
│   └── *_templ.go       # Generated Go files (do not edit)
├── static/              # CSS and other static assets
└── Makefile             # Build commands
```

### Working with Templ Files

When editing UI components:

1. Edit the `.templ` source files in `templates/`
2. Run `templ generate -f <file>` or `make -C templates` to regenerate Go code
3. The generated files are committed to the repository

Example templ file structure:
```templ
package templates

func Index(...) templ.Component {
    return templHTML()
}
```

### Context Usage

- Use `context.Background()` for templ Render calls (as seen in main.go)
- Pass context through HTTP handlers when needed

### Logging

- Use the standard `log` package for logging
- Log important events: startup, reloads, errors
- Use descriptive log messages with context
