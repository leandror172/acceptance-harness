package verify

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/leandror172/acceptance-harness/harness"
)

func contextWithStdout(t *testing.T, stdout string) *harness.Context {
	return &harness.Context{T: t, Stdout: stdout}
}

func TestValueAtPath(t *testing.T) {
	jsonStr := `{"app":{"status":"active","count":3},"items":[{"name":"a"},{"name":"b"}],"empty":[]}`
	var root interface{}
	err := json.Unmarshal([]byte(jsonStr), &root)
	assert.NoError(t, err)

	tests := []struct {
		path  string
		want  interface{}
		found bool
	}{
		{"$", root, true},
		{"", root, true},
		{"app.status", "active", true},
		{"app.count", 3.0, true}, // JSON numbers are float64
		{"items.0.name", "a", true},
		{"items.1.name", "b", true},
		{"$.app.status", "active", true},
		{"app.missing", nil, false},
		{"items.9", nil, false},
		{"items.x", nil, false},
		{"app.status.deeper", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, found := valueAtPath(root, tt.path)
			assert.Equal(t, tt.found, found)
			if tt.found {
				assert.EqualValues(t, tt.want, got)
			}
		})
	}
}

func TestOutputIsValidJSON_AcceptsValidJSON(t *testing.T) {
	ctx := contextWithStdout(t, `{"key":"value"}`)
	OutputIsValidJSON()(ctx)
}

func TestOutputJSONHasKey_FindsTopLevelKey(t *testing.T) {
	ctx := contextWithStdout(t, `{"key":"value"}`)
	OutputJSONHasKey("key")(ctx)
}

func TestOutputJSONHasValue_NumericEqualValues(t *testing.T) {
	ctx := contextWithStdout(t, `{"count":3}`)
	OutputJSONHasValue("count", 3)(ctx)
}

func TestJSONFieldEquals_NestedAndIndexed(t *testing.T) {
	jsonStr := `{"app":{"status":"active"},"items":[{"name":"a"}]}`
	ctx := contextWithStdout(t, jsonStr)

	JSONFieldEquals("app.status", "active")(ctx)
	JSONFieldEquals("items.0.name", "a")(ctx)
}

func TestJSONArrayLen_CountsElements(t *testing.T) {
	jsonStr := `{"items":[{"name":"a"},{"name":"b"}]}`
	ctx := contextWithStdout(t, jsonStr)

	JSONArrayLen("items", 2)(ctx)
}

func TestJSONArrayEvery_ChecksFieldOnAllElements(t *testing.T) {
	jsonStr := `{"items":[{"state":"done"},{"state":"done"}]}`
	ctx := contextWithStdout(t, jsonStr)

	JSONArrayEvery("items", "state", "done")(ctx)
}
