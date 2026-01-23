package jsonpath

import "testing"

// Issue #36: Add .length() function support
// https://github.com/oliveagle/jsonpath/issues/36
func Test_jsonpath_length_function(t *testing.T) {
	// Test case 1: Get length of an array
	arr := []interface{}{1, 2, 3, 4, 5}
	res, err := JsonPathLookup(arr, "$.length()")
	if err != nil {
		t.Fatalf("$.length() failed: %v", err)
	}
	if res.(int) != 5 {
		t.Errorf("Expected 5, got %v", res)
	}

	// Test case 2: Get length of a string
	str := "hello"
	res, err = JsonPathLookup(str, "$.length()")
	if err != nil {
		t.Fatalf("$.length() on string failed: %v", err)
	}
	if res.(int) != 5 {
		t.Errorf("Expected 5, got %v", res)
	}

	// Test case 3: Use length() in filter
	books := []interface{}{
		map[string]interface{}{"title": "Book1", "pages": 100},
		map[string]interface{}{"title": "Book2", "pages": 250},
		map[string]interface{}{"title": "Book3", "pages": 50},
	}
	// $[?(@.pages > length($.books))] - would select books with pages > length of books (3)
	res, err = JsonPathLookup(books, "$[?(@.pages > 3)]")
	if err != nil {
		t.Fatalf("$[?(@.pages > 3)] failed: %v", err)
	}
	resSlice, ok := res.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", res)
	}
	// Should return all books since pages (100, 250, 50) are all > 3
	if len(resSlice) != 3 {
		t.Errorf("Expected 3 books, got %d: %v", len(resSlice), resSlice)
	}

	// Test case 4: Get length of a map
	obj := map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": 3,
	}
	res, err = JsonPathLookup(obj, "$.length()")
	if err != nil {
		t.Fatalf("$.length() on map failed: %v", err)
	}
	if res.(int) != 3 {
		t.Errorf("Expected 3, got %v", res)
	}

	// Test case 5: length() with absolute path
	store := map[string]interface{}{
		"book": []interface{}{
			map[string]interface{}{"title": "Book1"},
			map[string]interface{}{"title": "Book2"},
		},
	}
	res, err = JsonPathLookup(store, "$.book.length()")
	if err != nil {
		t.Fatalf("$.book.length() failed: %v", err)
	}
	if res.(int) != 2 {
		t.Errorf("Expected 2, got %v", res)
	}

	// Test case 6: Use length() in filter with root path
	res, err = JsonPathLookup(books, "$[?(@.pages > $.length())]")
	if err != nil {
		t.Fatalf("$[?(@.pages > $.length())] failed: %v", err)
	}
	resSlice, ok = res.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", res)
	}
	// $.length() on root books returns 3, so pages > 3 returns all
	if len(resSlice) != 3 {
		t.Errorf("Expected 3 books, got %d: %v", len(resSlice), resSlice)
	}
}

// Issue #41: RFC 9535 function support - count(), match(), search()
// https://github.com/oliveagle/jsonpath/issues/41
func Test_jsonpath_rfc9535_functions(t *testing.T) {
	// === count() function tests ===
	t.Run("count", func(t *testing.T) {
		books := []interface{}{
			map[string]interface{}{"title": "Book1", "author": "AuthorA"},
			map[string]interface{}{"title": "Book2", "author": "AuthorB"},
			map[string]interface{}{"title": "Book3", "author": "AuthorC"},
		}

		// Test $[?count(@) > 1] - count current array (3 books)
		// count(@) returns the length of the current iteration array
		res, err := JsonPathLookup(books, "$[?count(@) > 1]")
		if err != nil {
			t.Fatalf("$[?count(@) > 1] failed: %v", err)
		}
		resSlice, ok := res.([]interface{})
		if !ok {
			t.Fatalf("Expected []interface{}, got %T", res)
		}
		// count(@) returns 3, so 3 > 1 is true, should return all books
		if len(resSlice) != 3 {
			t.Errorf("Expected 3 books, got %d", len(resSlice))
		}

		// Test $[?count(@) > 2] - count is 3, 3 > 2 is true
		res, err = JsonPathLookup(books, "$[?count(@) > 2]")
		if err != nil {
			t.Fatalf("$[?count(@) > 2] failed: %v", err)
		}
		resSlice, ok = res.([]interface{})
		if !ok {
			t.Fatalf("Expected []interface{}, got %T", res)
		}
		if len(resSlice) != 3 {
			t.Errorf("Expected 3 books, got %d", len(resSlice))
		}

		// Test $[?count(@) == 3] - exact count match
		res, err = JsonPathLookup(books, "$[?count(@) == 3]")
		if err != nil {
			t.Fatalf("$[?count(@) == 3] failed: %v", err)
		}
		resSlice, ok = res.([]interface{})
		if !ok {
			t.Fatalf("Expected []interface{}, got %T", res)
		}
		if len(resSlice) != 3 {
			t.Errorf("Expected 3 books, got %d", len(resSlice))
		}

		// Test count with absolute path
		store := map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{"title": "Book1"},
				map[string]interface{}{"title": "Book2"},
			},
		}
		// count($.book) returns 2
		res, err = JsonPathLookup(store, "$.book[?count($.book) > 1]")
		if err != nil {
			t.Fatalf("$.book[?count($.book) > 1] failed: %v", err)
		}
	})

	// === match() function tests (implicit anchoring ^pattern$) ===
	t.Run("match", func(t *testing.T) {
		books := []interface{}{
			map[string]interface{}{"title": "Book1", "author": "Nigel Rees"},
			map[string]interface{}{"title": "Book2", "author": "Evelyn Waugh"},
			map[string]interface{}{"title": "Book3", "author": "Herman Melville"},
		}

		// match() with implicit anchoring - pattern must match entire string
		res, err := JsonPathLookup(books, "$[?match(@.author, 'Nigel Rees')]")
		if err != nil {
			t.Fatalf("$[?match(@.author, 'Nigel Rees')] failed: %v", err)
		}
		resSlice, ok := res.([]interface{})
		if !ok {
			t.Fatalf("Expected []interface{}, got %T", res)
		}
		if len(resSlice) != 1 {
			t.Errorf("Expected 1 book (Nigel Rees), got %d: %v", len(resSlice), resSlice)
		}

		// match with regex pattern (implicit anchoring)
		res, err = JsonPathLookup(books, "$[?match(@.author, '.*Rees')]")
		if err != nil {
			t.Fatalf("$[?match(@.author, '.*Rees')] failed: %v", err)
		}
		resSlice, ok = res.([]interface{})
		if !ok {
			t.Fatalf("Expected []interface{}, got %T", res)
		}
		if len(resSlice) != 1 {
			t.Errorf("Expected 1 book matching .*Rees, got %d", len(resSlice))
		}

		// match should fail if pattern doesn't match entire string
		res, err = JsonPathLookup(books, "$[?match(@.author, 'Rees')]")
		if err != nil {
			t.Fatalf("$[?match(@.author, 'Rees')] failed: %v", err)
		}
		resSlice, ok = res.([]interface{})
		if !ok {
			t.Fatalf("Expected []interface{}, got %T", res)
		}
		// 'Rees' alone won't match 'Nigel Rees' due to implicit anchoring (^Rees$ != Nigel Rees)
		if len(resSlice) != 0 {
			t.Errorf("Expected 0 books (Rees alone doesn't match 'Nigel Rees'), got %d", len(resSlice))
		}
	})

	// === search() function tests (no anchoring) ===
	t.Run("search", func(t *testing.T) {
		books := []interface{}{
			map[string]interface{}{"title": "Book1", "author": "Nigel Rees"},
			map[string]interface{}{"title": "Book2", "author": "Evelyn Waugh"},
			map[string]interface{}{"title": "Book3", "author": "Herman Melville"},
		}

		// search() without anchoring - pattern can match anywhere
		res, err := JsonPathLookup(books, "$[?search(@.author, 'Rees')]")
		if err != nil {
			t.Fatalf("$[?search(@.author, 'Rees')] failed: %v", err)
		}
		resSlice, ok := res.([]interface{})
		if !ok {
			t.Fatalf("Expected []interface{}, got %T", res)
		}
		// search finds 'Rees' anywhere in the string
		if len(resSlice) != 1 {
			t.Errorf("Expected 1 book containing 'Rees', got %d: %v", len(resSlice), resSlice)
		}

		// search with regex pattern
		res, err = JsonPathLookup(books, "$[?search(@.author, '.*Rees')]")
		if err != nil {
			t.Fatalf("$[?search(@.author, '.*Rees')] failed: %v", err)
		}
		resSlice, ok = res.([]interface{})
		if !ok {
			t.Fatalf("Expected []interface{}, got %T", res)
		}
		if len(resSlice) != 1 {
			t.Errorf("Expected 1 book matching .*Rees, got %d", len(resSlice))
		}

		// search should find partial matches
		res, err = JsonPathLookup(books, "$[?search(@.author, 'Waugh')]")
		if err != nil {
			t.Fatalf("$[?search(@.author, 'Waugh')] failed: %v", err)
		}
		resSlice, ok = res.([]interface{})
		if !ok {
			t.Fatalf("Expected []interface{}, got %T", res)
		}
		if len(resSlice) != 1 {
			t.Errorf("Expected 1 book containing 'Waugh', got %d", len(resSlice))
		}
	})

	// === match vs search comparison ===
	t.Run("match_vs_search", func(t *testing.T) {
		data := []interface{}{
			map[string]interface{}{"text": "hello world"},
			map[string]interface{}{"text": "hello"},
			map[string]interface{}{"text": "world"},
		}

		// match requires full string match
		res, err := JsonPathLookup(data, "$[?match(@.text, 'hello')]")
		if err != nil {
			t.Fatalf("$[?match(@.text, 'hello')] failed: %v", err)
		}
		resSlice, ok := res.([]interface{})
		if !ok {
			t.Fatalf("Expected []interface{}, got %T", res)
		}
		// match: ^hello$ doesn't match "hello world"
		if len(resSlice) != 1 {
			t.Errorf("Expected 1 book (exact match 'hello'), got %d", len(resSlice))
		}

		// search finds substring
		res, err = JsonPathLookup(data, "$[?search(@.text, 'hello')]")
		if err != nil {
			t.Fatalf("$[?search(@.text, 'hello')] failed: %v", err)
		}
		resSlice, ok = res.([]interface{})
		if !ok {
			t.Fatalf("Expected []interface{}, got %T", res)
		}
		// search finds "hello" in "hello world" and "hello"
		if len(resSlice) != 2 {
			t.Errorf("Expected 2 books containing 'hello', got %d", len(resSlice))
		}
	})
}
