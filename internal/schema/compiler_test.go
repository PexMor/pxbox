package schema

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompiler_Prepare(t *testing.T) {
	compiler := NewCompilerWithCache(64)
	ctx := context.Background()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []string{"name"},
	}

	err := compiler.Prepare(ctx, schema)
	require.NoError(t, err)
}

func TestCompiler_Validate(t *testing.T) {
	compiler := NewCompilerWithCache(64)
	ctx := context.Background()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []string{"name"},
	}

	// Prepare schema
	err := compiler.Prepare(ctx, schema)
	require.NoError(t, err)

	// Valid value
	validValue := map[string]interface{}{
		"name": "test",
	}
	err = compiler.Validate(ctx, "jsonschema", schema, validValue)
	assert.NoError(t, err)

	// Invalid value (missing required field)
	invalidValue := map[string]interface{}{}
	err = compiler.Validate(ctx, "jsonschema", schema, invalidValue)
	assert.Error(t, err)
}

func TestCompiler_ValidateJSONExample(t *testing.T) {
	compiler := NewCompilerWithCache(64)
	ctx := context.Background()

	// JSON example kind should not validate strictly
	schema := map[string]interface{}{
		"example": map[string]interface{}{
			"name": "test",
		},
	}

	value := map[string]interface{}{
		"name": "anything",
	}

	err := compiler.Validate(ctx, "jsonexample", schema, value)
	assert.NoError(t, err) // JSON examples don't validate strictly
}

