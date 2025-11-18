# JSON Schema $ref Resolution with Allowlist

## Overview

PxBox supports JSON Schema `$ref` resolution with an optional allowlist to restrict which URLs can be referenced. This provides security by preventing arbitrary external schema references.

## Usage

### Creating a Compiler with Allowlist

```go
import "pxbox/internal/schema"

// Create compiler with allowlist
allowlist := []string{
    "https://example.com/schemas/*",
    "https://trusted-domain.com/api/*",
    "file:///local/schemas/*",
}
compiler := schema.NewCompilerWithCacheAndAllowlist(64, allowlist)
```

### Allowlist Patterns

The allowlist supports several pattern types:

1. **Exact Match**: `"https://example.com/schema.json"`

   - Matches exactly this URL

2. **Prefix Match**: `"https://example.com/schemas/*"`

   - Matches any URL starting with this prefix

3. **Domain Match**: `"https://example.com"`

   - Matches any URL from this domain

4. **Local Files**: `"file://*"`
   - Matches any file:// URL

### Example

```go
// Allowlist configuration
allowlist := []string{
    "https://json.schemastore.org/*",  // JSON Schema Store
    "https://example.com/api/schemas/*", // Your API schemas
}

// Create compiler
compiler := schema.NewCompilerWithCacheAndAllowlist(64, allowlist)

// Prepare schema with $ref
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "email": map[string]interface{}{
            "$ref": "https://json.schemastore.org/email.json", // Allowed
        },
    },
}

// This will succeed if URL is in allowlist
err := compiler.Prepare(ctx, schema)

// This will fail if URL is not in allowlist
badSchema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "data": map[string]interface{}{
            "$ref": "https://malicious.com/schema.json", // Not allowed
        },
    },
}
err := compiler.Prepare(ctx, badSchema) // Returns error
```

## Security Considerations

- **Empty Allowlist**: If allowlist is `nil` or empty, all `$ref` URLs are allowed (backward compatible)
- **Production**: Always configure an allowlist in production to prevent SSRF attacks
- **Pattern Matching**: Use specific patterns rather than broad wildcards when possible

## Environment Configuration

You can configure the allowlist via environment variables or application configuration:

```go
// Example: Read from environment
allowlistStr := os.Getenv("SCHEMA_REF_ALLOWLIST")
var allowlist []string
if allowlistStr != "" {
    allowlist = strings.Split(allowlistStr, ",")
}
compiler := schema.NewCompilerWithCacheAndAllowlist(64, allowlist)
```
