# githubapi package

This package provides a single `GHClient` struct for all GitHub API operations (labels, issues, discussions, PRs).

- All input types are defined in `types.go`.
- The main interface is `GitHubClient` (see `interfaces.go`).
- All methods are implemented on `GHClient` (see `client.go`).

This design allows for easy mocking and extension.
