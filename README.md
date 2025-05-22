# GitHub Demo - Repository Hydration Tool

This is a GitHub CLI extension that automates repository hydration tasks, such as creating issues, discussions, pull requests, and labels using the GitHub API. It's designed for quickly setting up demo environments, onboarding repositories, or testing purposes.

## Installation

```bash
# Clone the repository
git clone https://github.com/chrisreddington/gh-demo.git

# Navigate to the directory
cd gh-demo

# Build the extension
go build -o gh-demo

# Add to your PATH or move to a directory in your PATH
```

## Usage

```bash
# Basic usage
gh-demo hydrate <owner> <repo> [config-path]

# Example: Hydrate a repository using the default hydrate.json in the current directory
gh-demo hydrate octocat my-repo

# Example: Specify a custom configuration file
gh-demo hydrate octocat my-repo ./path/to/hydrate.json
```

## Configuration Schema Documentation

The hydration configuration uses a JSON file (default: `hydrate.json`) to define the resources that should be created in a GitHub repository. Below are the schemas for each object type:

### Root Schema

The root schema defines the structure of the hydration configuration file:

```json
{
  "issues": [],      // Array of Issue objects
  "discussions": [], // Array of Discussion objects
  "prs": [],         // Array of PR objects
  "labels": []       // Array of Label objects
}
```

### Issue Schema

Each issue object supports the following properties:

```json
{
  "title": "Issue title",          // Required: Title of the issue
  "body": "Issue description",     // Required: Body/content of the issue
  "labels": ["bug", "help wanted"], // Optional: Array of label names
  "assignees": ["username1"]        // Optional: Array of GitHub usernames
}
```

Example:

```json
{
  "title": "Welcome to the repository!",
  "body": "This is a welcome issue for new contributors.",
  "labels": ["documentation", "good-first-issue"],
  "assignees": ["octocat"]
}
```

### Discussion Schema

Each discussion object supports the following properties:

```json
{
  "title": "Discussion title",    // Required: Title of the discussion
  "body": "Discussion content",   // Required: Body/content of the discussion
  "category": "General"           // Required: Name of the discussion category
}
```

Example:

```json
{
  "title": "Introduce yourself",
  "body": "Use this thread to introduce yourself to the team!",
  "category": "General"
}
```

### Pull Request Schema

Each PR object supports the following properties:

```json
{
  "title": "PR title",             // Required: Title of the pull request
  "body": "PR description",        // Required: Body/content of the pull request
  "head": "feature-branch",        // Required: Head branch name
  "base": "main",                  // Required: Base branch name
  "draft": false                   // Optional: Whether the PR is a draft (default: false)
}
```

Example:

```json
{
  "title": "Add new feature",
  "body": "This PR implements the new feature X.",
  "head": "feature-x",
  "base": "main",
  "draft": true
}
```

### Label Schema

Each label object supports the following properties:

```json
{
  "name": "bug",                        // Required: Name of the label
  "color": "ff0000",                    // Required: Color of the label in hex format (without #)
  "description": "Something is wrong"   // Optional: Description of the label
}
```

Example:

```json
{
  "name": "enhancement",
  "color": "a2eeef",
  "description": "New feature or request"
}
```

## Complete Example Configuration

Here's a complete example configuration file:

```json
{
  "labels": [
    {
      "name": "bug",
      "color": "d73a4a",
      "description": "Something isn't working"
    },
    {
      "name": "enhancement",
      "color": "a2eeef",
      "description": "New feature or request"
    }
  ],
  "issues": [
    {
      "title": "Welcome to the repository",
      "body": "This repository is set up for demonstration purposes.",
      "labels": ["documentation"]
    },
    {
      "title": "Getting Started Guide",
      "body": "We need a guide to help users getting started.",
      "labels": ["enhancement", "documentation"],
      "assignees": ["octocat"]
    }
  ],
  "discussions": [
    {
      "title": "Introduce yourself",
      "body": "Let's get to know each other! Share your name and interests.",
      "category": "General"
    }
  ],
  "prs": [
    {
      "title": "Update documentation",
      "body": "This PR updates the README with installation instructions.",
      "head": "docs-update",
      "base": "main"
    }
  ]
}
```

## Development

### Prerequisites

- Go 1.16 or higher
- GitHub CLI installed and authenticated
- Git

### Building and Testing

```bash
# Build the extension
go build

# Run tests
go test ./...

# Format code
go fmt ./...
```

### Project Structure

- `/cmd`: Main command implementations
- `/internal/githubapi`: GitHub API client interfaces and implementations
- `/internal/hydrate`: Hydration logic and schema definitions
- `/prd`: Product documentation and requirements

## License

[MIT License](LICENSE)