# GitHub Demo - Copilot Instructions

## Project Overview
This project (gh-demo) is a GitHub CLI Extension written in Go to automate repository hydration tasks. It creates issues, discussions, pull requests, and labels using the GitHub API via the `go-gh` library.

## Code Quality Standards

### Go Language & Function Design
- Follow standard Go idioms with comprehensive GoDoc comments for all exported symbols
- **Maximum function size**: 80 lines including comments and blank lines
- **Complex functions (>50 lines)** must be broken down into smaller, focused helper functions
- Use Single Responsibility Principle - each function does exactly one thing well
- Avoid deeply nested logic - use early returns and guard clauses
- If a function takes >5 parameters, use struct or options pattern
- Prefer strongly typed interfaces over `interface{}`

### Naming Standards
- **Use full descriptive names**: `pullRequest` not `pr`, `repository` not `repo`, `discussion` not `disc`
- **Avoid generic names**: `createIssueResponse` not `resp`, `validateUserError` not `err`
- **Collections use plural**: `pullRequests` not `prs`, `repositories` not `repos`
- **Boolean variables**: `shouldCreateLabels`, `hasValidationErrors`
- **Function names**: verb-noun pattern - `validatePullRequestData`, `extractLabelNames`

### Code Organization
- Extract common patterns into helper functions (timeout context creation, error wrapping, response validation)
- Group related functionality: validation, API calls, result processing in separate functions
- Use composition over large functions - pipeline of smaller functions
- Consistent error handling patterns across similar operations

### Testing Requirements
- Write tests before implementation (TDD approach) with minimum 80% coverage
- Use table-driven tests with individual test cases ≤30 lines of setup
- Break test functions >100 lines into multiple focused test functions
- Mock ALL external dependencies (GitHub API, file system operations)
- Test error paths, not just happy paths
- Use descriptive test names: `TestCreatePullRequest_WithInvalidHead_ReturnsValidationError`

### Error Handling & Context
- Use typed errors from `/internal/errors` package, never string-based matching
- Wrap ALL errors with meaningful context: `fmt.Errorf("operation failed: %w", err)`
- **ALL** I/O functions MUST accept `context.Context` as first parameter
- Use `context.WithTimeout()` for all external API calls
- Create specific error variables: `repositoryFetchError`, `pullRequestValidationError`
- Always include relevant identifiers (IDs, names, paths) in error context

## Development Workflow

### Pre-Commit Requirements
- Code must build: `go build .`
- All tests pass: `go test ./...`
- Code formatted: `go fmt ./...`
- Linting passes: `golangci-lint run`
- No hardcoded paths - use relative paths from project root

### Repository Structure
```
/cmd/           - CLI command implementations
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

### API Client & Commands
- Use `go-gh` library for all GitHub API interactions with timeout handling
- Define interfaces for all API clients to enable mocking
- Never hardcode repository/owner names - accept as parameters
- Add features as subcommands with clear, actionable error messages
- Ensure output works for both human reading and script parsing

### Structured Logging & File Handling
- Prefer structured logger from `/internal/common/logger.go` over `fmt.Printf` 
- Use log levels: Debug (internal state), Info (user actions), Error (failures)
- Always use paths relative to project root (see `findProjectRoot` utility)
- Handle file operations gracefully with proper error context

### Refactoring Triggers
- **When to refactor**: If function >50 lines, immediately break it down
- **Code duplication**: If copy-pasting >3 lines, extract into helper function
- **Magic numbers**: Replace with named constants in `/internal/config`
- **Boy Scout Rule**: When touching code, improve at least one nearby issue

### Performance & Documentation
- Focus on readability first, optimize hot paths with benchmarks
- Use `defer` for cleanup, check `ctx.Done()` in long-running operations
- Document "why" not "what" - code should be self-documenting
- Use TODO comments with GitHub issue references for technical debt
- Update `README.md` and `.github/copilot-instructions.md` as new features and patterns emerge

### Common Patterns
- **Timeout helper**: Create standardized timeout creation functions
- **Response processing**: Use descriptive variables like `repositoryResponse`, extract validation helpers
- **Error handling helpers**: Create reusable patterns for common error scenarios
- **Context passing**: Pass through entire call chain: CLI → business logic → API clients