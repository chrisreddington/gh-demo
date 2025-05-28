# GitHub Demo CLI Extension

A GitHub CLI extension for automating repository hydration tasks, such as creating issues, discussions, pull requests, and labels. Built with Go using the `go-gh` library for GitHub API interactions.

## Installation

```bash
gh extension install chrisreddington/gh-demo
```

## Usage

### Basic Hydration

```bash
# Hydrate a repository with demo content (uses default .github/demos directory)
gh demo hydrate --owner myuser --repo myrepo

# Hydrate with custom configuration directory
gh demo hydrate --owner myuser --repo myrepo --config-path custom/config/path

# Hydrate only specific content types (flags default to true)
gh demo hydrate --owner myuser --repo myrepo --discussions=false --prs=false

# Enable debug mode for detailed logging
gh demo hydrate --owner myuser --repo myrepo --debug

# Preview what would be created without actually doing it
gh demo hydrate --owner myuser --repo myrepo --dry-run
```

### Cleanup Operations

```bash
# Clean all existing content before hydrating
gh demo hydrate --owner myuser --repo myrepo --clean

# Clean specific content types
gh demo hydrate --owner myuser --repo myrepo --clean-issues --clean-labels

# Clean with preservation rules
gh demo hydrate --owner myuser --repo myrepo --clean --preserve-config .github/demos/preserve.json

# Preview cleanup operations
gh demo hydrate --owner myuser --repo myrepo --clean --dry-run
```

### Help

```bash
# Get help
gh demo hydrate --help
```

## Schema Documentation

This section documents the schema for each type of object used in the application.

### Issue Schema

Issues are defined with the following properties:

| Field     | Type     | Description                                   | Required |
|-----------|----------|-----------------------------------------------|----------|
| title     | string   | Title of the issue                            | Yes      |
| body      | string   | Content of the issue                          | Yes      |
| labels    | []string | List of labels to apply to the issue          | No       |
| assignees | []string | List of GitHub usernames to assign the issue to | No     |

Example:
```json
{
  "title": "Add dark mode support",
  "body": "The application should support dark mode for better user experience at night.",
  "labels": ["enhancement", "ui"],
  "assignees": ["octocat"]
}
```

### Discussion Schema

Discussions are defined with the following properties:

| Field    | Type     | Description                           | Required |
|----------|----------|---------------------------------------|----------|
| title    | string   | Title of the discussion               | Yes      |
| body     | string   | Content of the discussion             | Yes      |
| category | string   | Category of the discussion (must be an existing discussion category in the repo) | Yes |
| labels   | []string | List of labels to apply to the discussion | No    |

Example:
```json
{
  "title": "How should we implement dark mode?",
  "body": "Let's discuss the best approach for implementing dark mode in our application.",
  "category": "Ideas",
  "labels": ["ui", "discussion"]
}
```

### Pull Request Schema

Pull requests are defined with the following properties:

| Field     | Type     | Description                                   | Required |
|-----------|----------|-----------------------------------------------|----------|
| title     | string   | Title of the pull request                     | Yes      |
| body      | string   | Description of the changes                    | Yes      |
| head      | string   | Name of the branch containing the changes     | Yes      |
| base      | string   | Name of the base branch to merge into         | Yes      |
| labels    | []string | List of labels to apply to the pull request   | No       |
| assignees | []string | List of GitHub usernames to assign the PR to  | No       |

Example:
```json
{
  "title": "Implement dark mode",
  "body": "This PR adds dark mode support for the application.",
  "head": "feature/dark-mode",
  "base": "main",
  "labels": ["enhancement", "ui"],
  "assignees": ["octocat"]
}
```

### Label Schema

Labels can be explicitly defined with custom colors and descriptions. Labels referenced in issues, discussions, or pull requests that aren't explicitly defined will be auto-created with default styling.

| Field       | Type   | Description                                    | Required |
|-------------|--------|------------------------------------------------|----------|
| name        | string | Name of the label                              | Yes      |
| description | string | Description of what the label represents       | No       |
| color       | string | Hexadecimal color code (without # prefix)     | Yes      |

Example:
```json
{
  "name": "bug",
  "description": "Something isn't working",
  "color": "d73a4a"
}
```

### Preserve Configuration Schema

The preserve configuration file allows you to specify which objects should be preserved during cleanup operations. This is useful when you want to clean demo content but keep certain important issues, discussions, pull requests, or labels.

| Field                          | Type     | Description                                                      |
|--------------------------------|----------|------------------------------------------------------------------|
| issues.preserve_by_title       | []string | Preserve issues with these titles (supports regex patterns)     |
| issues.preserve_by_label       | []string | Preserve issues that have any of these labels                   |
| issues.preserve_by_id          | []string | Preserve issues with these GitHub node IDs                      |
| discussions.preserve_by_title  | []string | Preserve discussions with these titles (supports regex patterns)|
| discussions.preserve_by_category| []string| Preserve discussions in these categories                        |
| discussions.preserve_by_id     | []string | Preserve discussions with these GitHub node IDs                 |
| pull_requests.preserve_by_title| []string | Preserve PRs with these titles (supports regex patterns)        |
| pull_requests.preserve_by_label| []string | Preserve PRs that have any of these labels                      |
| pull_requests.preserve_by_id   | []string | Preserve PRs with these GitHub node IDs                         |
| labels.preserve_by_name        | []string | Preserve labels with these exact names                          |

Example:
```json
{
  "issues": {
    "preserve_by_title": ["^Release.*", "Important.*"],
    "preserve_by_label": ["critical", "security"],
    "preserve_by_id": ["I_kwDOA"]
  },
  "discussions": {
    "preserve_by_title": ["Monthly.*"],
    "preserve_by_category": ["Announcements"],
    "preserve_by_id": []
  },
  "pull_requests": {
    "preserve_by_title": ["^feat:.*"],
    "preserve_by_label": ["release"],
    "preserve_by_id": []
  },
  "labels": {
    "preserve_by_name": ["bug", "feature", "documentation"]
  }
}
```

## Features

### 🎯 **Repository Hydration**
- Create issues, discussions, and pull requests from JSON configuration files
- Automatic label creation with customizable colors and descriptions
- Support for assignees, labels, and other GitHub metadata
- Configurable content types (include/exclude issues, discussions, PRs)

### 🧹 **Cleanup Operations**
- Clean existing repository content before hydration
- Granular cleanup control (issues, discussions, PRs, labels separately)
- Intelligent preservation rules with regex pattern support
- Preserve by title, labels, categories, or GitHub node IDs

### 🔍 **Dry Run Mode**
- Preview all operations without making actual changes
- See what would be created, updated, or deleted
- Perfect for testing configurations and preservation rules

### ⚙️ **Flexible Configuration**
- Customizable configuration directory path
- Auto-discovery from current git repository context
- Comprehensive error handling with detailed feedback
- Structured logging with debug mode support

### 🛡️ **Production Ready**
- Context-aware timeout handling for all operations
- Comprehensive test coverage with mocked dependencies
- GraphQL API integration via `go-gh` library
- Rate limit handling and retry logic

## File Structure

The hydration tool expects JSON files in the configuration directory. By default, this is the `.github/demos/` directory, but it can be customized using the `--config-path` flag:

- `<config-path>/issues.json`: Array of issue objects
- `<config-path>/discussions.json`: Array of discussion objects
- `<config-path>/prs.json`: Array of pull request objects
- `<config-path>/labels.json`: Array of label objects (optional - labels referenced in other files will be auto-created with defaults)
- `<config-path>/preserve.json`: Configuration for objects to preserve during cleanup operations (optional)

### Default Configuration Directory

```bash
gh demo hydrate --owner myorg --repo myrepo
# Uses .github/demos/ directory by default
```

### Custom Configuration Directory

```bash
gh demo hydrate --owner myorg --repo myrepo --config-path custom/config/path
# Uses custom/config/path/ directory relative to project root
```

### Example Files

Example configuration files are included in the `.github/demos/` directory:

- [`issues.json`](.github/demos/issues.json) - Sample issue definitions
- [`discussions.json`](.github/demos/discussions.json) - Sample discussion definitions  
- [`prs.json`](.github/demos/prs.json) - Sample pull request definitions
- [`labels.json`](.github/demos/labels.json) - Sample label definitions with colors
- [`preserve.json`](.github/demos/preserve.json) - Preservation rules for cleanup operations

## Project Architecture

This project follows clean architecture principles with clear separation of concerns:

```
├── cmd/                    # CLI command implementations
│   ├── hydrate.go         # Main hydrate command with flags and validation
│   └── root.go            # Root command setup
├── internal/              # Internal packages (not for external use)
│   ├── common/            # Shared utilities and interfaces
│   │   ├── interfaces.go  # Common interfaces (Logger, etc.)
│   │   └── logger.go      # Structured logging implementation
│   ├── config/            # Configuration management
│   │   └── constants.go   # Application constants and Configuration struct
│   ├── errors/            # Typed error definitions and handling
│   │   └── types.go       # Custom error types with context
│   ├── githubapi/         # GitHub API client and interfaces
│   │   ├── client.go      # GitHub client implementation
│   │   ├── interfaces.go  # API client interfaces for mocking
│   │   └── mutations.go   # GraphQL mutations for create/delete operations
│   ├── hydrate/           # Core business logic
│   │   ├── hydrate.go     # Main hydration and cleanup logic
│   │   └── preserve.go    # Preservation logic for cleanup operations
│   ├── testutil/          # Test utilities and helpers
│   └── types/             # Shared data structures
│       └── models.go      # Issue, Discussion, PR, Label models
├── .github/demos/         # Example configuration files
└── main.go               # Application entry point
```

### Design Principles

- **Dependency Injection**: All external dependencies (GitHub API, file system) are injected via interfaces
- **Context Propagation**: All operations accept `context.Context` for timeout and cancellation handling
- **Comprehensive Error Handling**: Typed errors with operation context for better debugging
- **Configuration Over Convention**: Flexible configuration with sensible defaults
- **Test-Driven Development**: High test coverage with mocked dependencies

## Testing

This project follows a test-driven development approach with comprehensive test coverage.

### Test Types

1. **Unit Tests**: Use mocked dependencies for fast, isolated testing
2. **Integration Tests**: Test real GitHub API interactions (skipped in CI without auth)

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests in short mode (skips integration tests)
go test -short ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/githubapi/...
```

### Testing Strategy

- **Table-Driven Tests**: Comprehensive test cases with individual test functions, aiming for ≤30 lines of setup
- **Test Coverage**: Minimum 80% coverage with focus on error paths and edge cases
- **GitHub API Client**: Uses dependency injection with `NewGHClientWithClients()` for unit tests and `NewGHClient()` for integration tests
- **Authentication**: Unit tests use mocks and don't require GitHub credentials
- **CI/CD**: Runs tests in short mode to skip integration tests that require authentication
- **Mocking**: ALL external dependencies (GitHub API, file system) are mocked for unit tests

### Test Environment

Tests run against mock implementations by default. Integration tests that require GitHub authentication will:
- Skip automatically if no credentials are available
- Run in short mode during CI/CD to avoid authentication issues
- Can be enabled locally when GitHub CLI authentication is configured

## Development

### Code Quality Standards

This project follows strict code quality standards:

- **Maximum function size**: 80 lines including comments and blank lines
- **Complex functions (>50 lines)** must be broken down into smaller, focused helper functions
- **Descriptive naming**: Use full names (`pullRequest` not `pr`, `repository` not `repo`)
- **Context handling**: ALL I/O functions MUST accept `context.Context` as first parameter
- **Error wrapping**: Wrap ALL errors with meaningful context using typed errors

### Pre-Commit Requirements

Before committing, ensure:

```bash
# Code builds without errors
go build .

# All tests pass
go test ./...

# Code is properly formatted
go fmt ./...

# Linting passes (if golangci-lint is available)
golangci-lint run
```

### Building and Running

```bash
# Build the extension
go build -o gh-demo .

# Install locally for testing (from the project root)
gh extension install .
```

## Contributing

We welcome contributions! Please follow these guidelines:

1. **Fork the repository** and create a feature branch
2. **Follow the coding standards** outlined in the Development section
3. **Write tests** for all new functionality (minimum 80% coverage)
4. **Update documentation** including this README if needed
5. **Test your changes** with real GitHub repositories
6. **Submit a pull request** with a clear description of changes

### Reporting Issues

When reporting issues, please include:
- Command that failed and full error output
- Repository permissions and authentication status
- Configuration files used (sanitized of sensitive information)
- Expected vs actual behavior

### Feature Requests

Feature requests are welcome! Please:
- Check existing issues to avoid duplicates
- Describe the use case and problem being solved
- Provide examples of how the feature would be used
- Consider implementation complexity and maintenance burden

## License

This project is licensed under the MIT License - see the LICENSE file for details.