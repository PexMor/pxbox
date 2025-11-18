package schema

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	js "github.com/santhosh-tekuri/jsonschema/v5"
)

type Compiler struct {
	compiler      *js.Compiler
	cache         *expirable.LRU[string, *js.Schema]
	refAllowlist  []string // Allowed URL patterns for $ref resolution
}

// NewCompilerWithCache creates a new compiler with cache
func NewCompilerWithCache(maxSize int) *Compiler {
	return NewCompilerWithCacheAndAllowlist(maxSize, nil)
}

// NewCompilerWithCacheAndAllowlist creates a new compiler with cache and $ref allowlist
func NewCompilerWithCacheAndAllowlist(maxSize int, allowlist []string) *Compiler {
	c := js.NewCompiler()
	c.ExtractAnnotations = true
	
	return &Compiler{
		compiler:     c,
		cache:        expirable.NewLRU[string, *js.Schema](maxSize, nil, time.Hour),
		refAllowlist: allowlist,
	}
}

// matchesPattern checks if a URL matches an allowlist pattern
// Supports:
// - Exact match: "https://example.com/schema.json"
// - Domain match: "https://example.com/*"
// - Path prefix: "https://example.com/schemas/*"
// - Local file: "file://*"
func matchesPattern(urlStr, pattern string) bool {
	// Exact match
	if urlStr == pattern {
		return true
	}

	// Wildcard pattern
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(urlStr, prefix)
	}

	// Try parsing as URLs for domain matching
	u1, err1 := url.Parse(urlStr)
	u2, err2 := url.Parse(pattern)
	if err1 == nil && err2 == nil {
		// Domain match
		if u1.Host == u2.Host {
			return true
		}
	}

	return false
}

func (c *Compiler) key(schema map[string]interface{}) string {
	b, _ := json.Marshal(schema)
	return string(b)
}

// Prepare compiles and caches a schema
func (c *Compiler) Prepare(ctx context.Context, schema map[string]interface{}) error {
	key := c.key(schema)
	if _, ok := c.cache.Get(key); ok {
		return nil // Already cached
	}

	// Validate $ref URLs against allowlist if configured
	if len(c.refAllowlist) > 0 {
		if err := c.validateRefs(schema); err != nil {
			return fmt.Errorf("$ref validation failed: %w", err)
		}
	}

	// Convert schema to JSON bytes
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Add resource to compiler
	// Use a hash-based URL to avoid URL parsing issues with JSON content
	hash := fmt.Sprintf("%x", schemaBytes)
	resourceURL := fmt.Sprintf("mem://schema/%s.json", hash[:16])
	if err := c.compiler.AddResource(resourceURL, bytes.NewReader(schemaBytes)); err != nil {
		return fmt.Errorf("failed to add resource: %w", err)
	}

	// Compile schema
	compiled, err := c.compiler.Compile(resourceURL)
	if err != nil {
		return fmt.Errorf("failed to compile schema: %w", err)
	}

	c.cache.Add(key, compiled)
	return nil
}

// validateRefs recursively validates all $ref URLs in a schema against the allowlist
func (c *Compiler) validateRefs(schema interface{}) error {
	switch v := schema.(type) {
	case map[string]interface{}:
		// Check for $ref
		if ref, ok := v["$ref"].(string); ok {
			if !c.isRefAllowed(ref) {
				return fmt.Errorf("$ref URL not allowed: %s (not in allowlist)", ref)
			}
		}
		// Recursively check nested objects and arrays
		for _, val := range v {
			if err := c.validateRefs(val); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, item := range v {
			if err := c.validateRefs(item); err != nil {
				return err
			}
		}
	}
	return nil
}

// isRefAllowed checks if a $ref URL is allowed by the allowlist
func (c *Compiler) isRefAllowed(refURL string) bool {
	// Empty allowlist means allow all (backward compatible)
	if len(c.refAllowlist) == 0 {
		return true
	}

	// Check against allowlist patterns
	for _, pattern := range c.refAllowlist {
		if matchesPattern(refURL, pattern) {
			return true
		}
	}
	return false
}

// Validate validates a value against a schema
func (c *Compiler) Validate(ctx context.Context, kind string, schema map[string]interface{}, value map[string]interface{}) error {
	if kind == "jsonexample" {
		// For JSON examples, we don't validate strictly
		return nil
	}

	key := c.key(schema)
	compiled, ok := c.cache.Get(key)
	if !ok {
		// Try to prepare it
		if err := c.Prepare(ctx, schema); err != nil {
			return err
		}
		compiled, _ = c.cache.Get(key)
		if compiled == nil {
			return fmt.Errorf("schema not found in cache after preparation")
		}
	}

	// Convert value to JSON for validation
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	var valueRaw interface{}
	if err := json.Unmarshal(valueBytes, &valueRaw); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	if err := compiled.Validate(valueRaw); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

