package specprose_test

import (
	"testing"

	"github.com/psyb0t/go-btse/internal/specprose"
	"github.com/stretchr/testify/assert"
)

func TestNormalizePreservesWireValues(t *testing.T) {
	t.Parallel()

	const (
		wireValue      = "a \u2014 b"
		descriptionKey = "description"
		summaryKey     = "summary"
	)

	document := map[string]any{
		descriptionKey: wireValue,
		"items":        []any{map[string]any{summaryKey: wireValue}},
		"const":        wireValue,
		"enum":         []any{wireValue},
		"default":      map[string]any{"title": wireValue},
		"example":      map[string]any{descriptionKey: wireValue},
		"examples":     []any{map[string]any{summaryKey: wireValue}},
	}

	specprose.Normalize(document)

	assert.Equal(t, "a. b", document[descriptionKey])
	assert.Equal(t, []any{map[string]any{summaryKey: "a. b"}},
		document["items"])
	assert.Equal(t, wireValue, document["const"])
	assert.Equal(t, []any{wireValue}, document["enum"])
	assert.Equal(t, map[string]any{"title": wireValue}, document["default"])
	assert.Equal(t, map[string]any{descriptionKey: wireValue},
		document["example"])
	assert.Equal(t, []any{map[string]any{summaryKey: wireValue}},
		document["examples"])
}
