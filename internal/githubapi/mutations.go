// Package githubapi contains GraphQL mutation definitions for GitHub API operations.
// This file centralizes all GraphQL mutations used by the GitHub client.
package githubapi

// createLabelMutation creates a new label in a repository
const createLabelMutation = `
	mutation CreateLabel($repositoryId: ID!, $name: String!, $color: String!, $description: String) {
		createLabel(input: {
			repositoryId: $repositoryId
			name: $name
			color: $color
			description: $description
		}) {
			label {
				id
				name
				color
				description
			}
		}
	}
`

// createIssueMutation creates a new issue in a repository
const createIssueMutation = `
	mutation CreateIssue($repositoryId: ID!, $title: String!, $body: String, $labelIds: [ID!], $assigneeIds: [ID!]) {
		createIssue(input: {
			repositoryId: $repositoryId
			title: $title
			body: $body
			labelIds: $labelIds
			assigneeIds: $assigneeIds
		}) {
			issue {
				id
				number
				title
				url
			}
		}
	}
`

// createPullRequestMutation creates a new pull request in a repository
const createPullRequestMutation = `
	mutation CreatePullRequest($repositoryId: ID!, $title: String!, $body: String, $headRefName: String!, $baseRefName: String!) {
		createPullRequest(input: {
			repositoryId: $repositoryId
			title: $title
			body: $body
			headRefName: $headRefName
			baseRefName: $baseRefName
		}) {
			pullRequest {
				id
				number
				title
				url
			}
		}
	}
`

// getRepositoryIdQuery gets the repository ID needed for mutations
const getRepositoryIdQuery = `
	query GetRepositoryId($owner: String!, $name: String!) {
		repository(owner: $owner, name: $name) {
			id
		}
	}
`

// getLabelIdQuery gets label ID by name for issue/PR creation
const getLabelIdQuery = `
	query GetLabelId($owner: String!, $name: String!, $labelName: String!) {
		repository(owner: $owner, name: $name) {
			label(name: $labelName) {
				id
			}
		}
	}
`

// getUserIdQuery gets user ID by login for assignee operations
const getUserIdQuery = `
	query GetUserId($login: String!) {
		user(login: $login) {
			id
		}
	}
`

// listLabelsQuery lists all labels in a repository with pagination support
const listLabelsQuery = `
	query($owner: String!, $name: String!) {
		repository(owner: $owner, name: $name) {
			labels(first: 100) {
				nodes {
					name
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	}
`

// repositoryWithDiscussionCategoriesQuery gets repository ID and discussion categories
const repositoryWithDiscussionCategoriesQuery = `
	query($owner: String!, $name: String!) {
		repository(owner: $owner, name: $name) {
			id
			discussionCategories(first: 50) {
				nodes {
					id
					name
				}
			}
		}
	}
`

// createDiscussionMutation creates a new discussion in a repository
const createDiscussionMutation = `
	mutation($input: CreateDiscussionInput!) {
		createDiscussion(input: $input) {
			discussion {
				id
				number
				title
				url
			}
		}
	}
`

// deleteDiscussionMutation deletes a discussion by its node ID
const deleteDiscussionMutation = `
	mutation DeleteDiscussion($discussionId: ID!) {
		deleteDiscussion(input: {id: $discussionId}) {
			clientMutationId
			discussion {
				id
				title
			}
		}
	}
`

// labelByNameQuery gets label ID by name for discussion/labelable operations
const labelByNameQuery = `
	query($owner: String!, $repo: String!, $name: String!) {
		repository(owner: $owner, name: $repo) {
			label(name: $name) {
				id
			}
		}
	}
`

// addLabelsToLabelableMutation adds labels to any labelable object (issues, PRs, discussions)
const addLabelsToLabelableMutation = `
	mutation($input: AddLabelsToLabelableInput!) {
		addLabelsToLabelable(input: $input) {
			clientMutationId
		}
	}
`

// addLabelsToLabelableMutationWithParams adds labels to labelable with explicit parameters
const addLabelsToLabelableMutationWithParams = `
	mutation AddLabelsToPR($labelableId: ID!, $labelIds: [ID!]!) {
		addLabelsToLabelable(input: {
			labelableId: $labelableId
			labelIds: $labelIds
		}) {
			clientMutationId
		}
	}
`

// addAssigneesToAssignableMutation adds assignees to any assignable object (issues, PRs)
const addAssigneesToAssignableMutation = `
	mutation AddAssigneesToPR($assignableId: ID!, $assigneeIds: [ID!]!) {
		addAssigneesToAssignable(input: {
			assignableId: $assignableId
			assigneeIds: $assigneeIds
		}) {
			clientMutationId
		}
	}
`

// listIssuesQuery lists all issues in a repository with pagination support
const listIssuesQuery = `
	query($owner: String!, $name: String!, $first: Int!, $after: String) {
		repository(owner: $owner, name: $name) {
			issues(first: $first, after: $after, states: [OPEN]) {
				nodes {
					id
					number
					title
					body
					labels(first: 20) {
						nodes {
							name
						}
					}
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	}
`

// listDiscussionsQuery lists all discussions in a repository with pagination support
const listDiscussionsQuery = `
	query($owner: String!, $name: String!, $first: Int!, $after: String) {
		repository(owner: $owner, name: $name) {
			discussions(first: $first, after: $after) {
				nodes {
					id
					number
					title
					body
					category {
						name
					}
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	}
`

// listPullRequestsQuery lists all pull requests in a repository with pagination support
const listPullRequestsQuery = `
	query($owner: String!, $name: String!, $first: Int!, $after: String) {
		repository(owner: $owner, name: $name) {
			pullRequests(first: $first, after: $after, states: [OPEN]) {
				nodes {
					id
					number
					title
					body
					headRefName
					baseRefName
					labels(first: 20) {
						nodes {
							name
						}
					}
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	}
`

// deleteIssueMutation deletes an issue by closing it (GitHub doesn't allow permanent deletion)
const deleteIssueMutation = `
	mutation DeleteIssue($issueId: ID!) {
		closeIssue(input: {
			issueId: $issueId
		}) {
			issue {
				id
				state
			}
		}
	}
`

// deletePullRequestMutation deletes a pull request by closing it
const deletePullRequestMutation = `
	mutation DeletePullRequest($pullRequestId: ID!) {
		closePullRequest(input: {
			pullRequestId: $pullRequestId
		}) {
			pullRequest {
				id
				state
			}
		}
	}
`

// deleteLabelMutation deletes a label by name
const deleteLabelMutation = `
	mutation DeleteLabel($labelId: ID!) {
		deleteLabel(input: {
			id: $labelId
		}) {
			clientMutationId
		}
	}
`

// getLabelByNameQuery gets a label ID by name for deletion
const getLabelByNameQuery = `
	query($owner: String!, $name: String!, $labelName: String!) {
		repository(owner: $owner, name: $name) {
			label(name: $labelName) {
				id
			}
		}
	}
`

// ProjectV2 mutations and queries

// createProjectV2Mutation creates a new ProjectV2 for a repository owner
const createProjectV2Mutation = `
	mutation CreateProjectV2($ownerId: ID!, $title: String!) {
		createProjectV2(input: {
			ownerId: $ownerId
			title: $title
		}) {
			projectV2 {
				id
				number
				title
				url
			}
		}
	}
`

// addProjectV2ItemByIdMutation adds an item to a ProjectV2 by content ID
const addProjectV2ItemByIdMutation = `
	mutation AddProjectV2ItemById($projectId: ID!, $contentId: ID!) {
		addProjectV2ItemById(input: {
			projectId: $projectId
			contentId: $contentId
		}) {
			item {
				id
			}
		}
	}
`

// getProjectV2Query retrieves a ProjectV2 by ID
const getProjectV2Query = `
	query GetProjectV2($projectId: ID!) {
		node(id: $projectId) {
			... on ProjectV2 {
				id
				number
				title
				description
				url
			}
		}
	}
`

// getRepositoryOwnerIdQuery gets the owner ID for creating projects
const getRepositoryOwnerIdQuery = `
	query GetRepositoryOwnerId($owner: String!) {
		repositoryOwner(login: $owner) {
			id
		}
	}
`
