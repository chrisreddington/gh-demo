# Product Requirement Document: Hydrate Functionality

## Why
Many developers and teams want to quickly set up demo environments in GitHub repositories for onboarding, workshops, or testing. Manually creating issues, discussions, and pull requests is time-consuming and error-prone. Automating this process helps users get started faster, ensures consistency, and reduces friction in demo or onboarding scenarios.

## Problem Statement
- Setting up a repository for demos or onboarding requires repetitive manual steps.
- Users often forget to create all the necessary issues, discussions, or pull requests, leading to incomplete demos.
- Manual setup is slow and can introduce inconsistencies between demo environments.

## Goals
- Enable users to quickly and reliably hydrate a repository with demo content (issues, discussions, pull requests).
- Reduce manual effort and errors in setting up demo repositories.
- Ensure a consistent experience for all users running demos or onboarding sessions.

## User Stories

### User Story 1
As a developer running a workshop, I want to automatically populate my repository with demo issues, discussions, and pull requests, so that participants can immediately interact with realistic content.

### User Story 2
As a new team member, I want to hydrate a repository with onboarding tasks and discussions, so I know exactly what steps to follow and where to ask questions.

### User Story 3
As a maintainer, I want to ensure that every demo repository is set up the same way, so that I can avoid confusion and support requests from users.

## User Story 4
As the product developer of the hydration feature, I want to provide a consistent schema and format to represent the issues, discussions and pull requests that will be created. This schema should be documented, so that users can easily understand how to customize the content being created.

## Acceptance Criteria
- [ ] Users can run a single command to hydrate a repository with demo issues, discussions, and pull requests.
- [ ] The command provides clear feedback on what was created and if any errors occurred.
- [ ] The process is idempotent: running it multiple times does not create duplicate content.
- [ ] Users can specify which types of content (issues, discussions, pull requests) to include.
- [ ] The feature works for both public and private repositories.
- [ ] Documentation is available to guide users through the hydration process.
- [ ] The schema for issues, discussions, and pull requests is well-defined and documented. It should not just include titles and bodies, but also labels, assignees, and other relevant metadata.
- [ ] The command line should use the `go-gh` package to interact with the GitHub API, ensuring compatibility and ease of use. Implement this using the GraphQL API.
- [ ] The hydration process should handle rate limits gracefully, retrying as necessary without failing completely.
- [ ] Users can configure the path to the configuration files using a `--config-path` flag, with `.github/demos` as the default location.
