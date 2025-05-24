# GitHub Demo - Copilot Instructions

## Project Overview
This project (gh-demo) is a GitHub CLI Extension written in Go to automate repository hydration tasks. It creates issues, discussions, pull requests, and labels using the GitHub API via the `go-gh` library.

## Code Quality Standards

### Context Management
- **ALL** functions that perform I/O operations MUST accept `context.Context` as the first parameter
- Pass context through the entire call chain: CLI → business logic → API clients
- Use `context.WithTimeout()` for all external API calls to prevent hanging

### Error Handling
- Use typed errors from `/internal/errors` package, never string-based error matching
- Wrap ALL errors with meaningful context using `fmt.Errorf("operation failed: %w", err)`
- Every error must be handled. When fixing errcheck linting errors, you may not ignore them.
  ```

### Structured Logging
- Replace ALL `fmt.Printf` statements with structured logger calls
- Use the logger from `/internal/common/logger.go`
- Include request IDs for tracing API operations
- Log levels: Debug (internal state), Info (user actions), Error (failures)
- Example:
  ```go
  logger.Info("creating issue", "title", issue.Title, "labels", len(issue.Labels))
  ```

### Testing Requirements
- Achieve minimum 80% test coverage for all packages
- Use table-driven tests for validation logic and multiple scenarios
- Mock ALL external dependencies (GitHub API, file system operations)
- Test error paths, not just happy paths
- Write tests before implementation (TDD approach)

### Go Language Standards
- Follow standard Go idioms and best practices
- Use comprehensive GoDoc comments for all exported symbols
- Keep functions small and focused on a single responsibility
- Use meaningful, consistent naming (no arbitrary abbreviations)
- Prefer strongly typed interfaces over `interface{}`
- Separate interface definitions from implementations

### Pre-Commit Requirements
- The code must build: `go build .`
- All tests must pass: `go test ./...`
- Code must be formatted: `go fmt ./...`
- Linting must pass: `golangci-lint run`
- No hardcoded paths - use relative paths from project root
- Documentation updated for any API changes

## Development Workflow

### Build and Test Commands
```bash
go build                    # Build the application
go test ./...               # Run all tests
go fmt ./...                # Format code
golangci-lint run           # Run linter (if available)
go run main.go <subcommand> # Run with subcommand (e.g., hydrate)
```

### Repository Structure
```
/cmd/           - CLI command implementations (hydrate.go, root.go)
/internal/      - Internal packages and business logic
  /common/      - Shared utilities (logger, interfaces)
  /config/      - Configuration constants
  /errors/      - Typed error definitions
  /githubapi/   - GitHub API client and interfaces
  /hydrate/     - Core hydration logic and file parsing
  /types/       - Shared type definitions
/prd/           - Product documentation
main.go         - Application entry point
```

## Implementation Guidelines

### API Client Design
- Use the `go-gh` library for all GitHub API interactions
- Define interfaces for all API clients to enable mocking
- Never hardcode repository or owner names - accept as parameters
- All API operations must include timeout handling

### Command Structure
- Add new features as subcommands following existing patterns
- Provide clear, actionable error messages for users
- Ensure output works for both human reading and script parsing
- Maintain consistency in command naming and option structure

### File and Path Handling
- Always use paths relative to project root (see `findProjectRoot` utility)
- Never hardcode absolute paths
- Handle file operations gracefully with proper error context

### Documentation Standards
- Update `README.md` for new features
- As new coding quality standards emerge, update `.github/copilot-instructions.md`
- Include usage examples for complex operations
- Document design decisions and architectural choices
- Maintain accurate API documentation for internal packages