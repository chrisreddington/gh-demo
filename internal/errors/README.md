# Error Handling Conventions

This document outlines the standardized error handling patterns used throughout the gh-demo codebase.

## Core Principles

1. **Structured Errors**: Use `LayeredError` for consistent error categorization and context
2. **Type Safety**: Always use safe type assertions to prevent panics
3. **Error Context**: Include sufficient context for debugging (file paths, operation names, etc.)
4. **Partial Failures**: Handle partial failures gracefully with `ErrorCollector` or `PartialFailureError`

## Error Types

### LayeredError
Provides structured error information with layers, operations, and context:
- **Layer**: Category of error (api, validation, file, config, context)
- **Operation**: Specific operation being performed
- **Message**: Human-readable description
- **Cause**: Underlying error that caused this error
- **Context**: Additional key-value pairs for debugging

### PartialFailureError
Used when some operations succeed and others fail in batch operations.

### ErrorCollector
Utility for collecting multiple errors and converting them to appropriate return types.

## Safe Type Assertion Helpers

### AsLayeredError(err error) *LayeredError
Safely converts an error to LayeredError, returning nil if conversion fails.

```go
if layeredErr := errors.AsLayeredError(err); layeredErr != nil {
    // Safe to use layeredErr
}
```

### WithContextSafe(err error, key, value string) error
Safely adds context to an error if it's a LayeredError, otherwise returns original error.

```go
err = errors.WithContextSafe(err, "file_path", "/path/to/file")
```

### WrapWithOperation(err error, layer, operation, message string) error
Wraps any error with operation context, creating a LayeredError if needed.

```go
err = errors.WrapWithOperation(originalErr, "api", "create_issue", "failed to create issue")
```

## Usage Patterns

### API Operations
```go
if err := client.CreateIssue(ctx, issue); err != nil {
    if errors.IsContextError(err) {
        return errors.ContextError("create_issue", err)
    }
    err = errors.WrapWithOperation(err, "api", "create_issue", "failed to create issue")
    err = errors.WithContextSafe(err, "title", issue.Title)
    return err
}
```

### File Operations
```go
content, err := os.ReadFile(path)
if err != nil {
    err = errors.WrapWithOperation(err, "file", "read_config", "failed to read configuration file")
    return nil, errors.WithContextSafe(err, "path", path)
}
```

### Batch Operations with ErrorCollector
```go
collector := errors.NewErrorCollector("cleanup_issues")

for _, issue := range issues {
    if err := client.DeleteIssue(ctx, issue.NodeID); err != nil {
        wrappedErr := errors.WrapWithOperation(err, "cleanup", "delete_issue", "failed to delete issue")
        wrappedErr = errors.WithContextSafe(wrappedErr, "title", issue.Title)
        collector.Add(wrappedErr)
        continue
    }
}

return collector.Result()
```

### Type Assertions
**❌ Never do this (unsafe):**
```go
layeredErr := err.(*errors.LayeredError).WithContext("key", "value")
```

**✅ Always do this (safe):**
```go
err = errors.WithContextSafe(err, "key", "value")
```

## Layer Categories

- **api**: GitHub API operations
- **validation**: Input validation and parameter checking
- **file**: File system operations
- **config**: Configuration loading and parsing
- **context**: Context cancellation and timeout handling
- **cleanup**: Cleanup and deletion operations

## Operation Names

Use descriptive operation names that indicate the specific action being performed:
- `create_issue`, `delete_issue`, `list_issues`
- `read_config`, `parse_config`, `validate_config`
- `find_project_root`, `read_labels`, `ensure_labels`

## Error Context

Always include relevant context for debugging:
- File paths for file operations
- Object titles/names for API operations
- Node IDs for deletion operations
- Request/response information for network operations

## Backward Compatibility

When updating cleanup functions that previously returned `[]string`, use this pattern:

```go
// Convert collected errors to string slice for backward compatibility
if result := collector.Result(); result != nil {
    if partialErr, ok := result.(*errors.PartialFailureError); ok {
        return partialErr.Errors
    }
    return []string{result.Error()}
}
return nil
```

This ensures existing callers continue to work while providing structured error handling internally.