# GitHub Demo CLI Extension

A GitHub CLI extension for automating repository hydration tasks, such as creating issues, discussions, pull requests, and labels.

## Installation

```bash
gh extension install chrisreddington/gh-demo
```

## Usage

```bash
# Hydrate a repository with demo content
gh demo hydrate --owner myuser --repo myrepo

# Hydrate only with issues
gh demo hydrate --issues --no-discussions --no-prs

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

## File Structure

The hydration tool expects JSON files in the `.github/demos/` directory:

- `.github/demos/issues.json`: Array of issue objects
- `.github/demos/discussions.json`: Array of discussion objects
- `.github/demos/prs.json`: Array of pull request objects

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.