package verify

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/stretchr/testify/assert"

	"github.com/leandror172/acceptance-harness/harness"
)

// OutputIsValidJSON asserts that stdout contains valid JSON.
func OutputIsValidJSON() func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		if ctx.Stdout == "" {
			ctx.T.Errorf("expected non-empty stdout, got empty string")
			return
		}
		var result interface{}
		if err := json.Unmarshal([]byte(ctx.Stdout), &result); err != nil {
			ctx.T.Errorf("stdout is not valid JSON: %v\nstdout: %s", err, ctx.Stdout)
		}
	}
}

// OutputJSONHasKey asserts stdout is valid JSON with the given top-level key.
func OutputJSONHasKey(key string) func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(ctx.Stdout), &result); err != nil {
			ctx.T.Errorf("stdout is not valid JSON: %v\nstdout: %s", err, ctx.Stdout)
			return
		}
		if _, exists := result[key]; !exists {
			keys := make([]string, 0, len(result))
			for k := range result {
				keys = append(keys, k)
			}
			ctx.T.Errorf("JSON missing key %q. Available: %s\nstdout: %s",
				key, strings.Join(keys, ", "), ctx.Stdout)
		}
	}
}

// OutputJSONHasValue asserts stdout JSON has expected at the given top-level key.
// Numeric comparison is value-based (an int expected matches a JSON float).
func OutputJSONHasValue(key string, expected interface{}) func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(ctx.Stdout), &result); err != nil {
			ctx.T.Errorf("stdout is not valid JSON: %v\nstdout: %s", err, ctx.Stdout)
			return
		}
		actual, exists := result[key]
		if !exists {
			ctx.T.Errorf("JSON missing key %q\nstdout: %s", key, ctx.Stdout)
			return
		}
		assert.EqualValues(ctx.T, expected, actual, "JSON key %q\nstdout: %s", key, ctx.Stdout)
	}
}

// JSONFieldEquals asserts the value at a dotted path in stdout JSON equals
// expected. Path "$" is the root; "app.status" walks nested objects. Numeric
// comparison is value-based, so an int expected matches a JSON float.
func JSONFieldEquals(path string, expected interface{}) func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		root, ok := jsonRoot(ctx)
		if !ok {
			return
		}
		actual, found := valueAtPath(root, path)
		if !assert.True(ctx.T, found, "JSON path %q not found\nstdout: %s", path, ctx.Stdout) {
			return
		}
		assert.EqualValues(ctx.T, expected, actual, "JSON path %q\nstdout: %s", path, ctx.Stdout)
	}
}

// JSONArrayLen asserts the array at path has exactly n elements.
func JSONArrayLen(path string, n int) func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		arr, ok := jsonArrayAt(ctx, path)
		if !ok {
			return
		}
		assert.Len(ctx.T, arr, n, "JSON path %q length\nstdout: %s", path, ctx.Stdout)
	}
}

// JSONArrayEvery asserts every element (an object) at path has element[field] == value.
func JSONArrayEvery(path, field, value string) func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		arr, ok := jsonArrayAt(ctx, path)
		if !ok {
			return
		}
		for i, el := range arr {
			obj, isObj := el.(map[string]interface{})
			if !assert.True(ctx.T, isObj, "JSON path %q[%d] is not an object", path, i) {
				continue
			}
			assert.EqualValues(ctx.T, value, obj[field],
				"JSON path %q[%d].%s\nstdout: %s", path, i, field, ctx.Stdout)
		}
	}
}

func jsonRoot(ctx *harness.Context) (interface{}, bool) {
	var root interface{}
	if err := json.Unmarshal([]byte(ctx.Stdout), &root); err != nil {
		ctx.T.Errorf("stdout is not valid JSON: %v\nstdout: %s", err, ctx.Stdout)
		return nil, false
	}
	return root, true
}

func jsonArrayAt(ctx *harness.Context, path string) ([]interface{}, bool) {
	root, ok := jsonRoot(ctx)
	if !ok {
		return nil, false
	}
	v, found := valueAtPath(root, path)
	if !assert.True(ctx.T, found, "JSON path %q not found\nstdout: %s", path, ctx.Stdout) {
		return nil, false
	}
	arr, isArr := v.([]interface{})
	if !assert.True(ctx.T, isArr, "JSON path %q is not an array\nstdout: %s", path, ctx.Stdout) {
		return nil, false
	}
	return arr, true
}

// valueAtPath walks a dotted path into nested JSON. "$" (or "") returns root.
// A segment indexes a map by key, or — when the current node is an array — by
// integer position (e.g. "actionable.0.company", "eval.paths.0").
func valueAtPath(root interface{}, path string) (interface{}, bool) {
	if path == "" || path == "$" {
		return root, true
	}
	path = strings.TrimPrefix(path, "$.")
	cur := root
	for _, seg := range strings.Split(path, ".") {
		switch node := cur.(type) {
		case map[string]interface{}:
			v, ok := node[seg]
			if !ok {
				return nil, false
			}
			cur = v
		case []interface{}:
			idx, err := strconv.Atoi(seg)
			if err != nil || idx < 0 || idx >= len(node) {
				return nil, false
			}
			cur = node[idx]
		default:
			return nil, false
		}
	}
	return cur, true
}
