# GitHub Demo - Copilot Instructions

## Project Overview
This is a GitHub CLI Extension written in Go to automate repository hydration tasks, such as creating issues, discussions, pull requests, and labels using the GitHub API. It leverages the `go-gh` library and is structured for extensibility and testability.

## Code Standards

- Please use a test-driven development approach. Start writing tests based on expectation, then write the implementation.
- Use table-driven tests for unit tests to cover multiple scenarios in a single test function.

### Required Before Commit
- All tests must pass: `go test ./...`
- Code must be properly formatted: `go fmt ./...`
- Linting must pass: `golangci-lint run` (if configured)
- Ensure documentation is up-to-date for any new commands, features, or API changes
- Do not hard code any paths. Make them relative to the project root
- Verify new features follow established patterns for CLI, API, and test structure

### Go Patterns
- Follow standard Go idioms and best practices
- Use GoDoc comments for all exported functions, types, and packages:
  ```go
  // FunctionName does something specific
  func FunctionName() {}
  ```
- Handle errors explicitly and return them up the call stack
- Use meaningful variable and function names
- Keep functions small and focused on a single responsibility
- Separate interface definitions from implementations where appropriate
- Having multiple files in a package is acceptable (e.g. `client.go`, `interfaces.go`, `types.go`), but keep related code together
- Use `context.Context` for passing request-scoped values and cancellation signals

### Code Quality Standards
- **Type Consolidation**: Avoid duplicating type definitions across packages. Shared types should be defined in a common location or use consistent naming patterns
- **Naming Consistency**: Use consistent variable naming throughout the codebase (e.g., prefer `pullRequests` over `prs` for clarity, or establish a consistent abbreviation pattern)
- **Documentation Coverage**: All exported functions, types, methods, and packages must have comprehensive GoDoc comments that explain purpose, parameters, return values, and usage examples where appropriate
- **Type Safety**: Prefer strongly typed interfaces over generic `interface{}` where possible
- **Error Handling**: Provide meaningful error messages that include context about what operation failed

## Development Flow

- Build: `go build`
- Test: `go test ./...`
- Lint: `golangci-lint run` (if available)
- Format: `go fmt ./...`
- Run: `go run main.go <subcommand>` (e.g., `go run main.go hydrate`)

## Repository Structure
- `/cmd`: Main command implementations and CLI structure
  - Subcommand files (e.g., `hydrate.go`, `root.go`)
- `/internal`: Internal packages for business logic and API clients
  - `/githubapi`: GitHub API client interfaces and implementations (issues, PRs, labels, discussions)
  - `/hydrate`: Hydration logic, file parsing, and test utilities
- `/prd`: Product documentation (e.g., `hydrate.md`)
- `main.go`: Entry point for the application
- `go.mod` & `go.sum`: Go module declarations and dependency tracking

## Key Guidelines

1. **User Experience Focus**:
   - CLI commands should be intuitive and provide clear feedback
   - Handle errors gracefully with helpful messages
   - Output should be suitable for both human and script consumption

2. **Command Structure**:
   - New features should be added as subcommands or options to the CLI
   - Maintain consistency in command structure and naming

3. **API Usage**:
   - Use the `go-gh` library for all GitHub API interactions
   - Prefer interfaces for API clients to enable mocking and testing
   - Avoid hardcoding repository or owner names; pass them as parameters or config

4. **Testing**:
   - Write unit tests for all business logic and API interactions
   - Use mocks for external dependencies (see `/internal/hydrate/hydrate_test.go`)
   - Ensure test coverage for file parsing and label management

5. **Documentation**:
   - Update `prd/hydrate.md` and `README.md` when adding new features or commands
   - Document usage examples and any complex logic or design decisions

6. **Paths and File Handling**:
   - Always use paths relative to the project root (see `findProjectRoot` in `/internal/hydrate/hydrate.go`)
   - Do not hard code absolute paths