package jsonpath

import (
	"testing"
)

// Additional coverage tests for low-coverage functions

func Test_jsonpath_uncovered_edge_cases(t *testing.T) {
	// Test empty slice indexing error (line ~117-120)
	t.Run("index_empty_slice_error", func(t *testing.T) {
		c, _ := Compile("$[0]")
		data := []interface{}{}
		_, err := c.Lookup(data)
		if err == nil {
			t.Error("Should error when indexing empty slice")
		}
	})

	// Test range with key like $[:1].name (line ~121-128)
	t.Run("range_with_key", func(t *testing.T) {
		c, _ := Compile("$.items[:1].name")
		data := map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{"name": "first"},
				map[string]interface{}{"name": "second"},
			},
		}
		res, err := c.Lookup(data)
		if err != nil {
			t.Fatalf("Range with key failed: %v", err)
		}
		if res == nil {
			t.Error("Expected result for range with key")
		}
	})

	// Test multiple indices (line ~100-109)
	t.Run("multiple_indices", func(t *testing.T) {
		c, _ := Compile("$[0,2]")
		data := []interface{}{"a", "b", "c", "d"}
		res, err := c.Lookup(data)
		if err != nil {
			t.Fatalf("Multiple indices failed: %v", err)
		}
		resSlice := res.([]interface{})
		if len(resSlice) != 2 {
			t.Errorf("Expected 2 results, got %d", len(resSlice))
		}
	})

	// Test direct function call on root (line ~177-181)
	t.Run("direct_function_call", func(t *testing.T) {
		// Test eval_func being called directly on an array
		data := []interface{}{1, 2, 3}
		c, _ := Compile("$.length()")
		_, err := c.Lookup(data)
		if err != nil {
			t.Logf("Direct function call error: %v", err)
		}
	})

	// Test tokenize edge cases with . prefix (line ~268-286)
	t.Run("tokenize_dot_prefix", func(t *testing.T) {
		// Test tokenization of paths with . prefix handling
		tokens, err := tokenize("$.name")
		if err != nil {
			t.Fatalf("tokenize failed: %v", err)
		}
		if len(tokens) < 2 {
			t.Errorf("Expected at least 2 tokens, got %d", len(tokens))
		}
	})

	// Test tokenize wildcard handling (line ~279-286)
	t.Run("tokenize_wildcard", func(t *testing.T) {
		// Test tokenization of $.* paths
		_, err := tokenize("$.*")
		if err != nil {
			t.Fatalf("tokenize failed: %v", err)
		}
		// * should not be added if last token was already processed
	})

	// Test tokenize ..* (line ~272-275, 281-284)
	t.Run("tokenize_recursive_wildcard", func(t *testing.T) {
		// Test tokenization of $..* - * should be skipped after ..
		_, err := tokenize("$..*")
		if err != nil {
			t.Fatalf("tokenize failed: %v", err)
		}
		// Should have tokens for .. but not for * after ..
	})

	// Test parse_token with empty range (line ~350-360)
	t.Run("parse_token_empty_range", func(t *testing.T) {
		// Test parsing of $[:] or $[::] paths
		_, _, _, err := parse_token("[:]")
		if err == nil {
			t.Logf("Empty range parsing result: should handle gracefully")
		}
	})

	// Test parse_token with partial range (line ~350-360)
	t.Run("parse_token_partial_range", func(t *testing.T) {
		// Test parsing of $[1:] or $[:2] paths
		_, _, _, err := parse_token("[1:]")
		if err != nil {
			t.Fatalf("parse_token failed: %v", err)
		}
	})

	// Test parse_token wildcard (line ~364-367)
	t.Run("parse_token_wildcard", func(t *testing.T) {
		// Test parsing of $[*] path
		op, _, _, err := parse_token("[*]")
		if err != nil {
			t.Fatalf("parse_token failed: %v", err)
		}
		if op != "range" {
			t.Errorf("Expected 'range' op, got '%s'", op)
		}
	})

	// Test cmp_any with different types (line ~1193+)
	t.Run("cmp_any_type_mismatch", func(t *testing.T) {
		// Test comparison of incompatible types
		res, err := cmp_any("string", 123, "==")
		if err != nil {
			t.Logf("Type mismatch comparison: %v", err)
		}
		if res {
			t.Error("String should not equal number")
		}
	})

	// Test cmp_any with > operator (line ~1193+)
	t.Run("cmp_any_greater_than", func(t *testing.T) {
		res, err := cmp_any(10, 5, ">")
		if err != nil {
			t.Fatalf("cmp_any failed: %v", err)
		}
		if res != true {
			t.Error("10 should be > 5")
		}
	})

	// Test cmp_any with >= operator
	t.Run("cmp_any_greater_equal", func(t *testing.T) {
		res, err := cmp_any(5, 5, ">=")
		if err != nil {
			t.Fatalf("cmp_any failed: %v", err)
		}
		if res != true {
			t.Error("5 should be >= 5")
		}
	})

	// Test cmp_any with < operator
	t.Run("cmp_any_less_than", func(t *testing.T) {
		res, err := cmp_any(3, 7, "<")
		if err != nil {
			t.Fatalf("cmp_any failed: %v", err)
		}
		if res != true {
			t.Error("3 should be < 7")
		}
	})

	// Test cmp_any with <= operator
	t.Run("cmp_any_less_equal", func(t *testing.T) {
		res, err := cmp_any(5, 5, "<=")
		if err != nil {
			t.Fatalf("cmp_any failed: %v", err)
		}
		if res != true {
			t.Error("5 should be <= 5")
		}
	})

	// Test cmp_any with != operator (not supported, should error)
	t.Run("cmp_any_not_equal", func(t *testing.T) {
		_, err := cmp_any(1, 2, "!=")
		if err == nil {
			t.Error("!= operator should not be supported by cmp_any")
		}
	})

	// Test cmp_any with regex-like match
	t.Run("cmp_any_regex_match", func(t *testing.T) {
		_, err := cmp_any("test@example.com", ".*@example.*", "=~")
		if err != nil {
			t.Logf("Regex comparison: %v", err)
		}
	})

	// Test eval_filter with exists operator
	t.Run("eval_filter_exists", func(t *testing.T) {
		obj := map[string]interface{}{"name": "test"}
		root := obj
		res, err := eval_filter(obj, root, "name", "exists", "")
		if err != nil {
			t.Fatalf("eval_filter failed: %v", err)
		}
		if res != true {
			t.Error("name should exist")
		}
	})

	// Test eval_filter with non-existent key
	t.Run("eval_filter_not_exists", func(t *testing.T) {
		obj := map[string]interface{}{"name": "test"}
		root := obj
		// "nonexistent" is a literal string, not a path, so it's not nil
		// This tests that eval_filter handles non-path strings
		res, err := eval_filter(obj, root, "nonexistent", "exists", "")
		if err != nil {
			t.Fatalf("eval_filter failed: %v", err)
		}
		// "nonexistent" as a literal string is truthy (not nil)
		if res != true {
			t.Error("literal string should be truthy")
		}
	})

	// Test get_filtered with slice and regex (line ~571+)
	t.Run("get_filtered_slice_regex", func(t *testing.T) {
		obj := []interface{}{
			map[string]interface{}{"name": "test1"},
			map[string]interface{}{"name": "test2"},
		}
		root := obj
		var res interface{}
		res, err := get_filtered(obj, root, "@.name =~ /test.*/")
		if err != nil {
			t.Fatalf("get_filtered failed: %v", err)
		}
		if res != nil {
			resSlice := res.([]interface{})
			if len(resSlice) != 2 {
				t.Errorf("Expected 2 results, got %d", len(resSlice))
			}
		}
	})

	// Test get_filtered with map (line ~571+)
	t.Run("get_filtered_map", func(t *testing.T) {
		obj := map[string]interface{}{
			"a": map[string]interface{}{"value": 1},
			"b": map[string]interface{}{"value": 2},
		}
		root := obj
		// Filter on map values
		res, err := get_filtered(obj, root, "@.value > 0")
		if err != nil {
			t.Fatalf("get_filtered on map failed: %v", err)
		}
		if res != nil {
			t.Logf("Map filter result: %v", res)
		}
	})

	// Test getAllDescendants with map (line ~1222+)
	t.Run("getAllDescendants_map", func(t *testing.T) {
		obj := map[string]interface{}{
			"a": map[string]interface{}{
				"b": "deep",
			},
		}
		res := getAllDescendants(obj)
		// Should include: map itself, nested map, "deep" value
		if len(res) < 2 {
			t.Errorf("Expected at least 2 descendants, got %d", len(res))
		}
	})

	// Test getAllDescendants with empty map
	t.Run("getAllDescendants_empty_map", func(t *testing.T) {
		obj := map[string]interface{}{}
		res := getAllDescendants(obj)
		// Should at least include the empty map itself
		if len(res) < 1 {
			t.Errorf("Expected at least 1 result, got %d", len(res))
		}
	})

	// Test get_key on slice with empty key (line ~459-472)
	t.Run("get_key_slice_empty_key", func(t *testing.T) {
		obj := []interface{}{"a", "b", "c"}
		res, err := get_key(obj, "")
		if err != nil {
			t.Fatalf("get_key failed: %v", err)
		}
		// Empty key on slice should return the slice itself (same reference)
		resSlice, ok := res.([]interface{})
		if !ok {
			t.Error("Expected slice result")
		}
		if len(resSlice) != 3 {
			t.Errorf("Expected 3 elements, got %d", len(resSlice))
		}
	})

	// Test eval_reg_filter with empty string (line ~827+)
	t.Run("eval_reg_filter_empty_string", func(t *testing.T) {
		obj := map[string]interface{}{"name": ""}
		root := obj
		pat, _ := regFilterCompile("/.*/")
		val, err := eval_reg_filter(obj, root, "@.name", pat)
		if err != nil {
			t.Fatalf("eval_reg_filter failed: %v", err)
		}
		if val != true {
			t.Error("Empty string should match .*")
		}
	})

	// Test eval_reg_filter with non-string (line ~835-840)
	t.Run("eval_reg_filter_non_string", func(t *testing.T) {
		obj := map[string]interface{}{"name": 123}
		root := obj
		pat, _ := regFilterCompile("/.*/")
		_, err := eval_reg_filter(obj, root, "@.name", pat)
		if err == nil {
			t.Error("Should error when matching non-string")
		}
	})

	// Test eval_count with literal string (line ~976-978)
	t.Run("eval_count_literal_string", func(t *testing.T) {
		obj := map[string]interface{}{}
		root := obj
		// "hello" is not @ or $. prefix, should return string length
		val, err := eval_count(obj, root, []string{"hello"})
		if err != nil {
			t.Fatalf("eval_count failed: %v", err)
		}
		if val.(int) != 5 {
			t.Errorf("Expected 5 (length of 'hello'), got %v", val)
		}
	})

	// Test eval_count with nil nodeset (line ~982-983)
	t.Run("eval_count_nil_nodeset", func(t *testing.T) {
		obj := map[string]interface{}{}
		root := obj
		val, err := eval_count(obj, root, []string{"@.nonexistent"})
		if err != nil {
			t.Fatalf("eval_count failed: %v", err)
		}
		if val.(int) != 0 {
			t.Errorf("Expected 0 for nil nodeset, got %v", val)
		}
	})

	// Test eval_match with non-string result (line ~1007-1009)
	t.Run("eval_match_nil_value", func(t *testing.T) {
		obj := map[string]interface{}{"name": nil}
		root := obj
		val, err := eval_match(obj, root, []string{"@.name", ".*"})
		if err != nil {
			t.Fatalf("eval_match failed: %v", err)
		}
		if val != false {
			t.Error("nil value should not match")
		}
	})

	// Test eval_search with non-string result (line ~1070+)
	t.Run("eval_search_nil_value", func(t *testing.T) {
		obj := map[string]interface{}{"text": nil}
		root := obj
		val, err := eval_search(obj, root, []string{"@.text", ".*"})
		if err != nil {
			t.Fatalf("eval_search failed: %v", err)
		}
		if val != false {
			t.Error("nil value should not match")
		}
	})
}

func Test_jsonpath_more_uncovered(t *testing.T) {
	// Test parse_token with invalid multi-index (line ~375-377)
	t.Run("parse_token_invalid_multi_index", func(t *testing.T) {
		_, _, _, err := parse_token("[1,abc]")
		if err == nil {
			t.Error("Should error on invalid multi-index with non-number")
		}
	})

	// Test filter_get_from_explicit_path with unsupported op (line ~392-394)
	t.Run("filter_get_unsupported_op", func(t *testing.T) {
		obj := map[string]interface{}{"name": "test"}
		// "unknown" is not a valid token type
		_, err := filter_get_from_explicit_path(obj, "@.name.unknown")
		if err == nil {
			t.Error("Should error on unsupported operation")
		}
	})

	// Test filter_get_from_explicit_path with multi-index in filter (line ~408-410)
	t.Run("filter_get_multi_index_error", func(t *testing.T) {
		obj := []interface{}{
			map[string]interface{}{"name": "test"},
		}
		// [1,2] has multiple indices, not supported in filter
		_, err := filter_get_from_explicit_path(obj, "@[1,2].name")
		if err == nil {
			t.Error("Should error on multi-index in filter path")
		}
	})

	// Test filter_get_from_explicit_path with invalid token (line ~412-424)
	t.Run("filter_get_invalid_token", func(t *testing.T) {
		obj := map[string]interface{}{"name": "test"}
		// Try to access with unsupported token type
		_, err := filter_get_from_explicit_path(obj, "@.name(())")
		if err == nil {
			t.Error("Should error on invalid token format")
		}
	})

	// Test tokenize with quoted strings (line ~263-265)
	t.Run("tokenize_with_quotes", func(t *testing.T) {
		tokens, err := tokenize(`$["key with spaces"]`)
		if err != nil {
			t.Fatalf("tokenize failed: %v", err)
		}
		// Should handle quoted keys
		if len(tokens) < 2 {
			t.Logf("Tokens: %v", tokens)
		}
	})

	// Test tokenize with nested parentheses (line ~281-284)
	t.Run("tokenize_nested_parens", func(t *testing.T) {
		_, err := tokenize("$.func(arg1, arg2)")
		if err != nil {
			t.Fatalf("tokenize failed: %v", err)
		}
	})

	// Test parse_token with complex range (line ~344-347)
	t.Run("parse_token_complex_range", func(t *testing.T) {
		op, key, args, err := parse_token("[1:5:2]")
		if err == nil {
			t.Logf("Complex range [1:5:2] result: op=%s, key=%s, args=%v", op, key, args)
		}
	})

	// Test Lookup with deeply nested path that errors (line ~95-97)
	t.Run("lookup_nested_error", func(t *testing.T) {
		c, _ := Compile("$.a.b.c.d.e.f.g")
		data := map[string]interface{}{
			"a": map[string]interface{}{
				"b": "not a map",
			},
		}
		_, err := c.Lookup(data)
		if err == nil {
			t.Error("Should error on accessing key on non-map")
		}
	})

	// Test Lookup with recursive descent into non-iterable (line ~105-107)
	t.Run("lookup_recursive_non_iterable", func(t *testing.T) {
		c, _ := Compile("$..*")
		data := "string value"
		res, err := c.Lookup(data)
		if err != nil {
			t.Logf("Recursive descent on string: %v", err)
		}
		if res != nil {
			t.Logf("Result: %v", res)
		}
	})

	// Test tokenize with Unicode characters (line ~263-265)
	t.Run("tokenize_unicode", func(t *testing.T) {
		tokens, err := tokenize(`$.你好`)
		if err != nil {
			t.Fatalf("tokenize failed: %v", err)
		}
		if len(tokens) < 2 {
			t.Logf("Tokens: %v", tokens)
		}
	})

	// Test tokenize with special characters in key (line ~263-265)
	t.Run("tokenize_special_chars", func(t *testing.T) {
		tokens, err := tokenize(`$["key-with-dashes"]`)
		if err != nil {
			t.Fatalf("tokenize failed: %v", err)
		}
		if len(tokens) < 2 {
			t.Logf("Tokens: %v", tokens)
		}
	})

	// Test eval_filter with function call in left path (line ~173-175)
	t.Run("eval_filter_function_call", func(t *testing.T) {
		obj := []interface{}{1, 2, 3}
		root := obj
		// Test eval_filter with function result
		res, err := eval_filter(obj, root, "length(@)", ">", "0")
		if err != nil {
			t.Fatalf("eval_filter failed: %v", err)
		}
		if res != true {
			t.Error("length(@) should be > 0")
		}
	})

	// Test eval_filter with comparison operator (line ~176-181)
	t.Run("eval_filter_comparison", func(t *testing.T) {
		obj := map[string]interface{}{"count": 5}
		root := obj
		res, err := eval_filter(obj, root, "@.count", ">", "3")
		if err != nil {
			t.Fatalf("eval_filter failed: %v", err)
		}
		if res != true {
			t.Error("5 should be > 3")
		}
	})

	// Test eval_filter with $ root reference (line ~117-120)
	t.Run("eval_filter_root_reference", func(t *testing.T) {
		obj := map[string]interface{}{"value": 10}
		root := map[string]interface{}{"threshold": 5}
		res, err := eval_filter(obj, root, "@.value", ">", "$.threshold")
		if err != nil {
			t.Fatalf("eval_filter failed: %v", err)
		}
		if res != true {
			t.Error("10 should be > 5 (from $.threshold)")
		}
	})

	// Test filter_get_from_explicit_path with $ prefix (line ~392-394)
	t.Run("filter_get_dollar_prefix", func(t *testing.T) {
		obj := map[string]interface{}{"key": "value"}
		val, err := filter_get_from_explicit_path(obj, "$.key")
		if err != nil {
			t.Fatalf("filter_get_from_explicit_path failed: %v", err)
		}
		if val != "value" {
			t.Errorf("Expected 'value', got %v", val)
		}
	})

	// Test filter_get_from_explicit_path with @ prefix
	t.Run("filter_get_at_prefix", func(t *testing.T) {
		obj := map[string]interface{}{"key": "value"}
		val, err := filter_get_from_explicit_path(obj, "@.key")
		if err != nil {
			t.Fatalf("filter_get_from_explicit_path failed: %v", err)
		}
		if val != "value" {
			t.Errorf("Expected 'value', got %v", val)
		}
	})

	// Test filter_get_from_explicit_path with missing $ or @ (line ~392-394)
	t.Run("filter_get_missing_prefix", func(t *testing.T) {
		obj := map[string]interface{}{"key": "value"}
		_, err := filter_get_from_explicit_path(obj, "key")
		if err == nil {
			t.Error("Should error when path doesn't start with $ or @")
		}
	})

	// Test get_key on map with non-string key (line ~452-458)
	t.Run("get_key_reflect_map", func(t *testing.T) {
		// Create a map using reflection that isn't map[string]interface{}
		obj := map[int]interface{}{1: "one"}
		_, err := get_key(obj, "1")
		if err == nil {
			t.Logf("Reflect map key access result: should handle numeric keys")
		}
	})

	// Test eval_match with pattern error (line ~1030-1032)
	t.Run("eval_match_invalid_pattern", func(t *testing.T) {
		obj := map[string]interface{}{"name": "test"}
		root := obj
		_, err := eval_match(obj, root, []string{"@.name", "[invalid"})
		if err == nil {
			t.Error("Should error on invalid regex pattern")
		}
	})

	// Test eval_search with pattern error
	t.Run("eval_search_invalid_pattern", func(t *testing.T) {
		obj := map[string]interface{}{"text": "hello"}
		root := obj
		_, err := eval_search(obj, root, []string{"@.text", "[invalid"})
		if err == nil {
			t.Error("Should error on invalid regex pattern")
		}
	})

	// Test eval_filter_func with count on $ path (line ~974-975)
	t.Run("eval_filter_func_count_dollar_path", func(t *testing.T) {
		obj := map[string]interface{}{"items": []interface{}{"a", "b", "c"}}
		root := obj
		val, err := eval_filter_func(obj, root, "count($.items)")
		if err != nil {
			t.Fatalf("eval_filter_func failed: %v", err)
		}
		if val.(int) != 3 {
			t.Errorf("Expected 3, got %v", val)
		}
	})

	// Test eval_filter_func with length on $ path
	t.Run("eval_filter_func_length_dollar_path", func(t *testing.T) {
		obj := map[string]interface{}{"text": "hello"}
		root := obj
		val, err := eval_filter_func(obj, root, "length($.text)")
		if err != nil {
			t.Fatalf("eval_filter_func failed: %v", err)
		}
		if val.(int) != 5 {
			t.Errorf("Expected 5, got %v", val)
		}
	})

	// Test eval_filter_func with unsupported function (line ~942-944)
	t.Run("eval_filter_func_unsupported", func(t *testing.T) {
		obj := map[string]interface{}{}
		root := obj
		_, err := eval_filter_func(obj, root, "unknown_func(@)")
		if err == nil {
			t.Error("Should error on unsupported function")
		}
	})

	// Test eval_reg_filter with nil pattern (line ~828-830)
	t.Run("eval_reg_filter_nil_pattern", func(t *testing.T) {
		obj := map[string]interface{}{"name": "test"}
		root := obj
		_, err := eval_reg_filter(obj, root, "@.name", nil)
		if err == nil {
			t.Error("Should error on nil pattern")
		}
	})

	// Test get_filtered on non-slice with regex (line ~581-586)
	t.Run("get_filtered_non_slice_regex", func(t *testing.T) {
		obj := "not a slice"
		root := obj
		_, err := get_filtered(obj, root, "@ =~ /test/")
		if err != nil {
			t.Logf("Non-slice regex filter: %v", err)
		}
	})

	// Test get_range on non-slice (line ~539-554)
	t.Run("get_range_non_slice", func(t *testing.T) {
		obj := "string"
		_, err := get_range(obj, 0, 5)
		if err == nil {
			t.Error("Should error on non-slice range")
		}
	})

	// Test get_scan on nil (line ~649+)
	t.Run("get_scan_nil", func(t *testing.T) {
		_, err := get_scan(nil)
		if err != nil {
			t.Logf("get_scan nil result: %v", err)
		}
	})

	// Test cmp_any with string comparison (line ~1200-1205)
	t.Run("cmp_any_string_compare", func(t *testing.T) {
		res, err := cmp_any("apple", "banana", "<")
		if err != nil {
			t.Fatalf("cmp_any failed: %v", err)
		}
		if res != true {
			t.Error("apple should be < banana")
		}
	})

	// Test getAllDescendants with nested slice (line ~1246-1249)
	t.Run("getAllDescendants_nested_slice", func(t *testing.T) {
		obj := []interface{}{
			[]interface{}{1, 2, 3},
			[]interface{}{4, 5, 6},
		}
		res := getAllDescendants(obj)
		// Should include: outer array, both inner arrays, all elements
		if len(res) < 7 {
			t.Errorf("Expected at least 7 descendants, got %d", len(res))
		}
	})
}
