# GitHub Demo CLI Extension

A GitHub CLI extension for automating repository hydration tasks, such as creating issues, discussions, pull requests, and labels.

## Installation

```bash
gh extension install chrisreddington/gh-demo
```

## Usage

```bash
# Hydrate a repository with demo content (uses default .github/demos directory)
gh demo hydrate --owner myuser --repo myrepo

# Hydrate with custom configuration directory
gh demo hydrate --owner myuser --repo myrepo --config-path custom/config/path

# Hydrate only with issues
gh demo hydrate --issues --no-discussions --no-prs

# Enable debug mode for detailed logging
gh demo hydrate --owner myuser --repo myrepo --debug

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

## File Structure

The hydration tool expects JSON files in the configuration directory. By default, this is the `.github/demos/` directory, but it can be customized using the `--config-path` flag:

- `<config-path>/issues.json`: Array of issue objects
- `<config-path>/discussions.json`: Array of discussion objects
- `<config-path>/prs.json`: Array of pull request objects
- `<config-path>/labels.json`: Array of label objects (optional - labels referenced in other files will be auto-created with defaults)

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

- **GitHub API Client**: Uses dependency injection with `NewGHClientWithClients()` for unit tests and `NewGHClient()` for integration tests
- **Authentication**: Unit tests use mocks and don't require GitHub credentials
- **CI/CD**: Runs tests in short mode to skip integration tests that require authentication
- **Coverage**: All business logic is covered by unit tests using table-driven test patterns

### Test Environment

Tests run against mock implementations by default. Integration tests that require GitHub authentication will:
- Skip automatically if no credentials are available
- Run in short mode during CI/CD to avoid authentication issues
- Can be enabled locally when GitHub CLI authentication is configured
````