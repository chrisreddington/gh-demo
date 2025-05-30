# GitHub Demo CLI Extension

A GitHub CLI extension for automating repository hydration tasks, such as creating issues, discussions, pull requests, and labels. Built with Go using the `go-gh` library for GitHub API interactions.

## Installation

```bash
gh extension install chrisreddington/gh-demo
```

## Features

### 🎯 **Repository Hydration**
Create issues, discussions, and pull requests from JSON configuration files with automatic label creation and support for assignees, labels, and other GitHub metadata.

### 🧹 **Cleanup Operations**  
Clean existing repository content with intelligent preservation rules, granular control, and regex pattern support to preserve important content.

### 🔍 **Development Tools**
Dry run mode, flexible configuration paths, structured logging with debug mode, and comprehensive error handling for production use.

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

### ProjectV2 Integration

Create a GitHub ProjectV2 and automatically organize all hydrated content:

```bash
# Create a project with default configuration
gh demo hydrate --owner myuser --repo myrepo --create-project

# Create a project with custom configuration
gh demo hydrate --owner myuser --repo myrepo --create-project --project-config .github/demos/my-project.json

# Fail if project creation fails (default: continue without project)
gh demo hydrate --owner myuser --repo myrepo --create-project --fail-on-project-error

# Preview project creation
gh demo hydrate --owner myuser --repo myrepo --create-project --dry-run
```

**Important**: Project creation requires your GitHub token to have `write:org` (for organization projects) or `write:user` (for user projects) scope. If project creation fails due to insufficient permissions, the command will continue with standard hydration unless `--fail-on-project-error` is specified.

### Help

```bash
# Get help
gh demo hydrate --help
```

## Project Configuration

The `--create-project` flag creates a GitHub ProjectV2 and associates all created issues, discussions, and pull requests with it. Project configuration is defined in JSON format.

**Custom Fields Support:** The tool now supports creating custom fields including single select options with proper color validation. All field types supported by GitHub ProjectV2 are available: `single_select`, `text`, `number`, and `date`.

### Project Configuration Schema

| Field       | Type                    | Description                                    | Required |
|-------------|-------------------------|------------------------------------------------|----------|
| title       | string                  | Project title                                  | No*      |
| description | string                  | Project description                            | No       |
| visibility  | string                  | Project visibility ("private" or "public")    | No*      |
| fields      | []ProjectV2Field        | Custom project fields                          | No       |
| views       | []ProjectV2View         | Project views and layouts                      | No       |
| templates   | []ProjectV2Template     | Default field values for content types        | No       |

*Default values are provided if not specified.

### Project Field Schema

| Field       | Type                         | Description                                    | Required |
|-------------|------------------------------|------------------------------------------------|----------|
| name        | string                       | Field name                                     | Yes      |
| type        | string                       | Field type ("single_select", "text", "number", "date") | Yes      |
| description | string                       | Field description                              | No       |
| options     | []ProjectV2FieldOption       | Options for single_select fields              | No       |

#### ProjectV2FieldOption Schema

| Field       | Type   | Description                                    | Required |
|-------------|--------|------------------------------------------------|----------|
| name        | string | Option display name                            | Yes      |
| description | string | Option description                             | Yes      |
| color       | string | Option color (see allowed values below)       | Yes      |

#### Allowed Color Values for Single Select Options

The `color` field must use one of the following GitHub enum values:

- `GRAY` - Light gray
- `BLUE` - Blue  
- `GREEN` - Green
- `YELLOW` - Yellow
- `ORANGE` - Orange
- `RED` - Red
- `PINK` - Pink
- `PURPLE` - Purple

**Note:** Color values are case-insensitive. Invalid colors will default to `GRAY`.

#### Reserved Field Names

Avoid using these reserved field names which conflict with GitHub's built-in project fields:
- `Status` - Use alternative names like "Workflow Status", "Progress", or "State"
- `Assignees` - Built-in field for issue/PR assignees
- `Labels` - Built-in field for issue/PR labels  
- `Milestone` - Built-in field for milestones
- `Repository` - Built-in field for the source repository

### Example Project Configuration

```json
{
  "title": "Repository Demo Project",
  "description": "Demonstration project for repository hydration",
  "visibility": "private",
  "fields": [
    {
      "name": "Priority",
      "type": "single_select",
      "description": "Priority level for the content item",
      "options": [
        {
          "name": "🔥 Critical",
          "description": "Urgent items requiring immediate attention",
          "color": "RED"
        },
        {
          "name": "⚠️ High",
          "description": "Important items that should be addressed soon", 
          "color": "ORANGE"
        },
        {
          "name": "📋 Medium",
          "description": "Standard priority items",
          "color": "YELLOW"
        },
        {
          "name": "📝 Low",
          "description": "Nice to have items",
          "color": "GREEN"
        }
      ]
    },
    {
      "name": "Status",
      "type": "single_select", 
      "description": "Current status of the item",
      "options": [
        {
          "name": "📋 To Do",
          "description": "Items that haven't been started",
          "color": "GRAY"
        },
        {
          "name": "🔄 In Progress", 
          "description": "Items currently being worked on",
          "color": "YELLOW"
        },
        {
          "name": "✅ Done",
          "description": "Completed items", 
          "color": "GREEN"
        }
      ]
    },
    {
      "name": "Effort Estimate",
      "type": "text",
      "description": "Estimated effort required"
    },
    {
      "name": "Due Date",
      "type": "date",
      "description": "Target completion date"
    }
  ],
  "views": [
    {
      "name": "All Items",
      "description": "Complete view of all project items",
      "layout": "table", 
      "fields": ["title", "assignees", "status", "priority"]
    }
  ]
}
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

## Configuration

The hydration tool uses JSON configuration files to define the content to create. By default, it looks in the `.github/demos/` directory, but you can specify a custom path using the `--config-path` flag:

- `<config-path>/issues.json`: Array of issue objects
- `<config-path>/discussions.json`: Array of discussion objects  
- `<config-path>/prs.json`: Array of pull request objects
- `<config-path>/labels.json`: Array of label objects (optional - labels referenced in other files will be auto-created with defaults)
- `<config-path>/preserve.json`: Configuration for objects to preserve during cleanup operations (optional)
- `<config-path>/project-config.json`: ProjectV2 configuration for project creation (optional)

### Example Configuration Files

Example configuration files are included in the `.github/demos/` directory:

- [`issues.json`](.github/demos/issues.json) - Sample issue definitions
- [`discussions.json`](.github/demos/discussions.json) - Sample discussion definitions  
- [`prs.json`](.github/demos/prs.json) - Sample pull request definitions
- [`labels.json`](.github/demos/labels.json) - Sample label definitions with colors
- [`preserve.json`](.github/demos/preserve.json) - Preservation rules for cleanup operations
- [`project-config.json`](.github/demos/project-config.json) - Sample project configuration with fields and views

## Development

### Building and Running

```bash
# Build the extension
go build -o gh-demo .

# Install locally for testing (from the project root)
gh extension install .
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests in short mode (skips integration tests)
go test -short ./...

# Run tests with verbose output
go test -v ./...
```

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

### Contributing

We welcome contributions! For detailed coding standards, architecture guidelines, and development practices, please read [`.github/copilot-instructions.md`](.github/copilot-instructions.md).

Please follow these steps:
1. **Fork the repository** and create a feature branch
2. **Follow the coding standards** outlined in `.github/copilot-instructions.md`
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

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.