package jsonpath

import (
	"testing"
)

// Additional coverage tests for low-coverage functions

func Test_MustCompile(t *testing.T) {
	// Test valid path
	c := MustCompile("$.store.book[0].price")
	if c == nil {
		t.Fatal("MustCompile returned nil for valid path")
	}
	if c.path != "$.store.book[0].price" {
		t.Errorf("Expected path '$.store.book[0].price', got '%s'", c.path)
	}

	// Test String() method
	str := c.String()
	expected := "Compiled lookup: $.store.book[0].price"
	if str != expected {
		t.Errorf("String() expected '%s', got '%s'", expected, str)
	}

	// Test MustCompile with valid Lookup
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{"price": 8.95},
			},
		},
	}
	res, err := c.Lookup(data)
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if res.(float64) != 8.95 {
		t.Errorf("Expected 8.95, got %v", res)
	}

	// Test MustCompile panic on invalid path
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustCompile did not panic on invalid path")
		}
	}()
	MustCompile("invalid[path")
}

func Test_parse_filter_v1_skipped(t *testing.T) {
	// parse_filter_v1 is abandoned code (v1 parser), not used in current implementation
	// Skipping coverage for this function as it's not part of the active codebase
	t.Skip("parse_filter_v1 is abandoned v1 code, not used in current implementation")
}

func Test_jsonpath_eval_match_coverage(t *testing.T) {
	// Test eval_match with nil value from path
	t.Run("match_with_nil_from_path", func(t *testing.T) {
		obj := map[string]interface{}{"name": nil}
		root := obj
		res, err := eval_match(obj, root, []string{"@.name", ".*"})
		if err != nil {
			t.Fatalf("eval_match failed: %v", err)
		}
		if res != false {
			t.Errorf("Expected false for nil value, got %v", res)
		}
	})

	// Test eval_match with $.path and nil
	t.Run("match_with_dollar_path_nil", func(t *testing.T) {
		obj := map[string]interface{}{"name": nil}
		root := obj
		res, err := eval_match(obj, root, []string{"$.name", ".*"})
		if err != nil {
			t.Fatalf("eval_match failed: %v", err)
		}
		if res != false {
			t.Errorf("Expected false for nil value, got %v", res)
		}
	})

	// Test eval_match with non-string value (should still work via fmt.Sprintf)
	t.Run("match_with_non_string", func(t *testing.T) {
		obj := map[string]interface{}{"num": 123}
		root := obj
		res, err := eval_match(obj, root, []string{"@.num", "123"})
		if err != nil {
			t.Fatalf("eval_match failed: %v", err)
		}
		if res != true {
			t.Errorf("Expected true for num=123 matching '123', got %v", res)
		}
	})

	// Test eval_match with wrong argument count
	t.Run("match_wrong_args", func(t *testing.T) {
		_, err := eval_match(nil, nil, []string{"only_one"})
		if err == nil {
			t.Error("eval_match should error with 1 arg")
		}

		_, err = eval_match(nil, nil, []string{"one", "two", "three"})
		if err == nil {
			t.Error("eval_match should error with 3 args")
		}
	})
}

func Test_jsonpath_eval_search_coverage(t *testing.T) {
	// Test eval_search with nil value from path
	t.Run("search_with_nil_from_path", func(t *testing.T) {
		obj := map[string]interface{}{"name": nil}
		root := obj
		res, err := eval_search(obj, root, []string{"@.name", ".*"})
		if err != nil {
			t.Fatalf("eval_search failed: %v", err)
		}
		if res != false {
			t.Errorf("Expected false for nil value, got %v", res)
		}
	})

	// Test eval_search with $.path and nil
	t.Run("search_with_dollar_path_nil", func(t *testing.T) {
		obj := map[string]interface{}{"name": nil}
		root := obj
		res, err := eval_search(obj, root, []string{"$.name", ".*"})
		if err != nil {
			t.Fatalf("eval_search failed: %v", err)
		}
		if res != false {
			t.Errorf("Expected false for nil value, got %v", res)
		}
	})

	// Test eval_search with wrong argument count
	t.Run("search_wrong_args", func(t *testing.T) {
		_, err := eval_search(nil, nil, []string{"only_one"})
		if err == nil {
			t.Error("eval_search should error with 1 arg")
		}

		_, err = eval_search(nil, nil, []string{"one", "two", "three"})
		if err == nil {
			t.Error("eval_search should error with 3 args")
		}
	})
}

func Test_jsonpath_eval_count_coverage(t *testing.T) {
	// Test eval_count with wrong argument count
	t.Run("count_wrong_args", func(t *testing.T) {
		_, err := eval_count(nil, nil, []string{})
		if err == nil {
			t.Error("eval_count should error with 0 args")
		}

		_, err = eval_count(nil, nil, []string{"one", "two"})
		if err == nil {
			t.Error("eval_count should error with 2 args")
		}
	})

	// Test eval_count with non-array type
	t.Run("count_non_array", func(t *testing.T) {
		obj := map[string]interface{}{"items": "not an array"}
		root := obj
		res, err := eval_count(obj, root, []string{"@.items"})
		if err != nil {
			t.Fatalf("eval_count failed: %v", err)
		}
		// Should return 0 or handle gracefully
		t.Logf("eval_count on string returned: %v", res)
	})
}

func Test_jsonpath_eval_reg_filter_coverage(t *testing.T) {
	// Test eval_reg_filter with various types
	t.Run("reg_filter_on_map", func(t *testing.T) {
		obj := map[string]interface{}{"name": "test"}
		root := obj
		pat, err := regFilterCompile("/test/")
		if err != nil {
			t.Fatalf("regFilterCompile failed: %v", err)
		}

		ok, err := eval_reg_filter(obj, root, "@.name", pat)
		if err != nil {
			t.Fatalf("eval_reg_filter failed: %v", err)
		}
		if ok != true {
			t.Errorf("Expected true for 'test' matching /test/, got %v", ok)
		}
	})

	// Test eval_reg_filter with non-matching pattern
	t.Run("reg_filter_no_match", func(t *testing.T) {
		obj := map[string]interface{}{"name": "other"}
		root := obj
		pat, err := regFilterCompile("/test/")
		if err != nil {
			t.Fatalf("regFilterCompile failed: %v", err)
		}

		ok, err := eval_reg_filter(obj, root, "@.name", pat)
		if err != nil {
			t.Fatalf("eval_reg_filter failed: %v", err)
		}
		if ok != false {
			t.Errorf("Expected false for 'other' not matching /test/, got %v", ok)
		}
	})
}

func Test_jsonpath_get_scan_coverage(t *testing.T) {
	// Test get_scan with nested map containing various types
	t.Run("scan_nested_map", func(t *testing.T) {
		obj := map[string]interface{}{
			"a": []interface{}{1, 2, 3},
			"b": map[string]interface{}{"nested": true},
			"c": "string",
			"d": nil,
		}
		res, err := get_scan(obj)
		if err != nil {
			t.Fatalf("get_scan failed: %v", err)
		}
		resSlice, ok := res.([]interface{})
		if !ok {
			t.Fatalf("Expected []interface{}, got %T", res)
		}
		if len(resSlice) != 4 {
			t.Errorf("Expected 4 results, got %d", len(resSlice))
		}
	})

	// Test get_scan with empty map
	t.Run("scan_empty_map", func(t *testing.T) {
		obj := map[string]interface{}{}
		res, err := get_scan(obj)
		if err != nil {
			t.Fatalf("get_scan on empty map failed: %v", err)
		}
		resSlice, ok := res.([]interface{})
		if !ok {
			t.Fatalf("Expected []interface{}, got %T", res)
		}
		if len(resSlice) != 0 {
			t.Errorf("Expected 0 results, got %d", len(resSlice))
		}
	})
}
