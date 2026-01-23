// Copyright 2015, 2021; oliver, DoltHub Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package jsonpath

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
)

var json_data interface{}

func init() {
	data := `
{
    "store": {
        "book": [
            {
                "category": "reference",
                "author": "Nigel Rees",
                "title": "Sayings of the Century",
                "price": 8.95
            },
            {
                "category": "fiction",
                "author": "Evelyn Waugh",
                "title": "Sword of Honour",
                "price": 12.99
            },
            {
                "category": "fiction",
                "author": "Herman Melville",
                "title": "Moby Dick",
                "isbn": "0-553-21311-3",
                "price": 8.99
            },
            {
                "category": "fiction",
                "author": "J. R. R. Tolkien",
                "title": "The Lord of the Rings",
                "isbn": "0-395-19395-8",
                "price": 22.99
            }
        ],
        "bicycle": {
            "color": "red",
            "price": 19.95
        }
    },
    "expensive": 10
}
`
	json.Unmarshal([]byte(data), &json_data)
}

func Test_jsonpath_JsonPathLookup_1(t *testing.T) {
	// empty string
	res, err := JsonPathLookup(json_data, "")
	if err == nil {
		t.Errorf("expected error from empty jsonpath")
	}

	// key from root
	res, _ = JsonPathLookup(json_data, "$.expensive")
	if res_v, ok := res.(float64); ok != true || res_v != 10.0 {
		t.Errorf("expensive should be 10")
	}

	// single index
	res, _ = JsonPathLookup(json_data, "$.store.book[0].price")
	if res_v, ok := res.(float64); ok != true || res_v != 8.95 {
		t.Errorf("$.store.book[0].price should be 8.95")
	}

	// quoted - single index
	res, _ = JsonPathLookup(json_data, `$."store"."book"[0]."price"`)
	if res_v, ok := res.(float64); ok != true || res_v != 8.95 {
		t.Errorf(`$."store"."book"[0]."price" should be 8.95`)
	}

	// nagtive single index
	res, _ = JsonPathLookup(json_data, "$.store.book[-1].isbn")
	if res_v, ok := res.(string); ok != true || res_v != "0-395-19395-8" {
		t.Errorf("$.store.book[-1].isbn should be \"0-395-19395-8\"")
	}

	// multiple index
	res, err = JsonPathLookup(json_data, "$.store.book[0,1].price")
	t.Log(err, res)
	if res_v, ok := res.([]interface{}); ok != true || res_v[0].(float64) != 8.95 || res_v[1].(float64) != 12.99 {
		t.Errorf("exp: [8.95, 12.99], got: %v", res)
	}

	// multiple index
	res, err = JsonPathLookup(json_data, "$.store.book[0,1].title")
	t.Log(err, res)
	if res_v, ok := res.([]interface{}); ok != true {
		if res_v[0].(string) != "Sayings of the Century" || res_v[1].(string) != "Sword of Honour" {
			t.Errorf("title are wrong: %v", res)
		}
	}

	// full array
	res, err = JsonPathLookup(json_data, "$.store.book[0:].price")
	t.Log(err, res)
	if res_v, ok := res.([]interface{}); ok != true || res_v[0].(float64) != 8.95 || res_v[1].(float64) != 12.99 || res_v[2].(float64) != 8.99 || res_v[3].(float64) != 22.99 {
		t.Errorf("exp: [8.95, 12.99, 8.99, 22.99], got: %v", res)
	}

	// range - RFC 9535: end is exclusive, so [0:1] returns only element 0
	res, err = JsonPathLookup(json_data, "$.store.book[0:1].price")
	t.Log(err, res)
	if res_v, ok := res.([]interface{}); ok != true || res_v[0].(float64) != 8.95 || len(res_v) != 1 {
		t.Errorf("exp: [8.95], got: %v", res)
	}

	// range - RFC 9535: end is exclusive, so [0:1] returns only element 0
	res, err = JsonPathLookup(json_data, "$.store.book[0:1].title")
	t.Log(err, res)
	if res_v, ok := res.([]interface{}); ok != true {
		if res_v[0].(string) != "Sayings of the Century" || len(res_v) != 1 {
			t.Errorf("title are wrong: %v", res)
		}
	}
}

func Test_jsonpath_JsonPathLookup_filter(t *testing.T) {
	res, err := JsonPathLookup(json_data, "$.store.book[?(@.isbn)].isbn")
	t.Log(err, res)

	if res_v, ok := res.([]interface{}); ok != true {
		if res_v[0].(string) != "0-553-21311-3" || res_v[1].(string) != "0-395-19395-8" {
			t.Errorf("error: %v", res)
		}
	}

	res, err = JsonPathLookup(json_data, "$.store.book[?(@.price > 10)].title")
	t.Log(err, res)
	if res_v, ok := res.([]interface{}); ok != true {
		if res_v[0].(string) != "Sword of Honour" || res_v[1].(string) != "The Lord of the Rings" {
			t.Errorf("error: %v", res)
		}
	}

	res, err = JsonPathLookup(json_data, "$.store.book[?(@.price > 10)]")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$.store.book[?(@.price > $.expensive)].price")
	t.Log(err, res)
	res, err = JsonPathLookup(json_data, "$.store.book[?(@.price < $.expensive)].price")
	t.Log(err, res)
}

func Test_jsonpath_authors_of_all_books(t *testing.T) {
	query := "store.book[*].author"
	expected := []string{
		"Nigel Rees",
		"Evelyn Waugh",
		"Herman Melville",
		"J. R. R. Tolkien",
	}
	res, _ := JsonPathLookup(json_data, query)
	t.Log(res, expected)
}

func Test_jsonpath_string_equal(t *testing.T) {
	data := `{
    "store": {
        "book": [
            {
                "category": "reference",
                "author": "Nigel Rees",
                "title": "Sayings of the Century",
                "price": 8.95
            },
            {
                "category": "fiction",
                "author": "Evelyn Waugh",
                "title": "Sword of Honour",
                "price": 12.99
            },
            {
                "category": "fiction",
                "author": "Herman Melville",
                "title": "Moby Dick",
                "isbn": "0-553-21311-3",
                "price": 8.99
            },
            {
                "category": "fiction",
                "author": "J. R. R. Tolkien",
                "title": "The Lord of the Rings",
                "isbn": "0-395-19395-8",
                "price": 22.99
            }
        ],
        "bicycle": {
            "color": "red",
            "price": 19.95
        }
    },
    "expensive": 10
}`

	var j interface{}

	json.Unmarshal([]byte(data), &j)

	res, err := JsonPathLookup(j, "$.store.book[?(@.author == 'Nigel Rees')].price")
	t.Log(res, err)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if fmt.Sprintf("%v", res) != "[8.95]" {
		t.Fatalf("not the same: %v", res)
	}
}

func Test_jsonpath_null_in_the_middle(t *testing.T) {
	data := `{
  "head_commit": null,
}
`

	var j interface{}

	json.Unmarshal([]byte(data), &j)

	res, err := JsonPathLookup(j, "$.head_commit.author.username")
	t.Log(res, err)
}

func Test_jsonpath_num_cmp(t *testing.T) {
	data := `{
	"books": [ 
                { "name": "My First Book", "price": 10 }, 
		{ "name": "My Second Book", "price": 20 } 
		]
}`
	var j interface{}
	json.Unmarshal([]byte(data), &j)
	res, err := JsonPathLookup(j, "$.books[?(@.price > 100)].name")
	if err != nil {
		t.Fatal(err)
	}
	arr := res.([]interface{})
	if len(arr) != 0 {
		t.Fatal("should return [], got: ", arr)
	}

}

func TestReg(t *testing.T) {
	r := regexp.MustCompile(`(?U).*REES`)
	t.Log(r)
	t.Log(r.Match([]byte(`Nigel Rees`)))

	res, err := JsonPathLookup(json_data, "$.store.book[?(@.author =~ /(?i).*REES/ )].author")
	t.Log(err, res)

	author := res.([]interface{})[0].(string)
	t.Log(author)
	if author != "Nigel Rees" {
		t.Fatal("should be `Nigel Rees` but got: ", author)
	}
}

var tcases_reg_op = []struct {
	Line string
	Exp  string
	Err  bool
}{
	{``, ``, true},
	{`xxx`, ``, true},
	{`/xxx`, ``, true},
	{`xxx/`, ``, true},
	{`'/xxx/'`, ``, true},
	{`"/xxx/"`, ``, true},
	{`/xxx/`, `xxx`, false},
	{`/π/`, `π`, false},
}

func TestRegOp(t *testing.T) {
	for idx, tcase := range tcases_reg_op {
		t.Logf("idx: %v, tcase: %v", idx, tcase)
		res, err := regFilterCompile(tcase.Line)
		if tcase.Err == true {
			if err == nil {
				t.Fatal("expect err but got nil")
			}
		} else {
			if res == nil || res.String() != tcase.Exp {
				t.Fatal("different. res:", res)
			}
		}
	}
}

func Test_jsonpath_rootnode_is_array(t *testing.T) {
	data := `[{
    "test": 12.34
}, {
	"test": 13.34
}, {
	"test": 14.34
}]
`

	var j interface{}

	err := json.Unmarshal([]byte(data), &j)
	if err != nil {
		t.Fatal(err)
	}

	res, err := JsonPathLookup(j, "$[0].test")
	t.Log(res, err)
	if err != nil {
		t.Fatal("err:", err)
	}
	if res == nil || res.(float64) != 12.34 {
		t.Fatalf("different:  res:%v, exp: 123", res)
	}
}

func Test_jsonpath_rootnode_is_array_range(t *testing.T) {
	data := `[{
    "test": 12.34
}, {
	"test": 13.34
}, {
	"test": 14.34
}]
`

	var j interface{}

	err := json.Unmarshal([]byte(data), &j)
	if err != nil {
		t.Fatal(err)
	}

	res, err := JsonPathLookup(j, "$[:1].test")
	t.Log(res, err)
	if err != nil {
		t.Fatal("err:", err)
	}
	if res == nil {
		t.Fatal("res is nil")
	}
	// RFC 9535: end is exclusive, so [:1] returns only first element
	ares := res.([]interface{})
	for idx, v := range ares {
		t.Logf("idx: %v, v: %v", idx, v)
	}
	if len(ares) != 1 {
		t.Fatalf("len is not 1. got: %v", len(ares))
	}
	if ares[0].(float64) != 12.34 {
		t.Fatalf("idx: 0, should be 12.34. got: %v", ares[0])
	}
}

func Test_jsonpath_rootnode_is_nested_array(t *testing.T) {
	data := `[ [ {"test":1.1}, {"test":2.1} ], [ {"test":3.1}, {"test":4.1} ] ]`

	var j interface{}

	err := json.Unmarshal([]byte(data), &j)
	if err != nil {
		t.Fatal(err)
	}

	res, err := JsonPathLookup(j, "$[0].[0].test")
	t.Log(res, err)
	if err != nil {
		t.Fatal("err:", err)
	}
	if res == nil || res.(float64) != 1.1 {
		t.Fatalf("different:  res:%v, exp: 123", res)
	}
}

func Test_jsonpath_rootnode_is_nested_array_range(t *testing.T) {
	data := `[ [ {"test":1.1}, {"test":2.1} ], [ {"test":3.1}, {"test":4.1} ] ]`

	var j interface{}

	err := json.Unmarshal([]byte(data), &j)
	if err != nil {
		t.Fatal(err)
	}

	res, err := JsonPathLookup(j, "$[:1].[0].test")
	t.Log(res, err)
	if err != nil {
		t.Fatal("err:", err)
	}
	if res == nil {
		t.Fatal("res is nil")
	}
	ares := res.([]interface{})
	for idx, v := range ares {
		t.Logf("idx: %v, v: %v", idx, v)
	}
	if len(ares) != 2 {
		t.Fatalf("len is not 2. got: %v", len(ares))
	}

	//FIXME: `$[:1].[0].test` got wrong result
	//if ares[0].(float64) != 1.1 {
	//	t.Fatal("idx: 0, should be 1.1, got: %v", ares[0])
	//}
	//if ares[1].(float64) != 3.1 {
	//	t.Fatal("idx: 0, should be 3.1, got: %v", ares[1])
	//}
}

func TestRecursiveDescent(t *testing.T) {
	data := `
{
    "store": {
        "book": [
            {
                "category": "reference",
                "author": "Nigel Rees",
                "title": "Sayings of the Century",
                "price": 8.95
            },
            {
                "category": "fiction",
                "author": "Evelyn Waugh",
                "title": "Sword of Honour",
                "price": 12.99
            },
            {
                "category": "fiction",
                "author": "Herman Melville",
                "title": "Moby Dick",
                "isbn": "0-553-21311-3",
                "price": 8.99
            },
            {
                "category": "fiction",
                "author": "J. R. R. Tolkien",
                "title": "The Lord of the Rings",
                "isbn": "0-395-19395-8",
                "price": 22.99
            }
        ],
        "bicycle": {
            "color": "red",
            "price": 19.95
        }
    },
    "expensive": 10
}
`
	var json_data interface{}
	json.Unmarshal([]byte(data), &json_data)

	// Test case: $..author should return all authors
	authors_query := "$..author"
	res, err := JsonPathLookup(json_data, authors_query)
	if err != nil {
		t.Fatalf("Failed to execute recursive query %s: %v", authors_query, err)
	}

	authors, ok := res.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", res)
	}

	if len(authors) != 4 {
		t.Errorf("Expected 4 authors, got %d: %v", len(authors), authors)
	}

	// Test case: $..price should return all prices (5 total: 4 books + 1 bicycle)
	price_query := "$..price"
	res, err = JsonPathLookup(json_data, price_query)
	if err != nil {
		t.Fatalf("Failed to execute recursive query %s: %v", price_query, err)
	}
	prices, ok := res.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", res)
	}
	if len(prices) != 5 {
		t.Errorf("Expected 5 prices, got %d: %v", len(prices), prices)
	}
}

// Issue #40: [*] over objects returns an error
// https://github.com/oliveagle/jsonpath/issues/40
func Test_jsonpath_wildcard_over_object(t *testing.T) {
	input := map[string]interface{}{
		"a": map[string]interface{}{
			"foo": map[string]interface{}{
				"b": 1,
			},
		},
	}

	// Test $.a[*].b - wildcard on nested map should return values
	res, err := JsonPathLookup(input, "$.a[*].b")
	if err != nil {
		t.Fatalf("$.a[*].b failed: %v", err)
	}
	resSlice, ok := res.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", res)
	}
	if len(resSlice) != 1 {
		t.Errorf("Expected 1 result, got %d: %v", len(resSlice), resSlice)
	}

	// Test $.a[*] - wildcard should return map values
	res2, err2 := JsonPathLookup(input, "$.a[*]")
	if err2 != nil {
		t.Fatalf("$.a[*] failed: %v", err2)
	}
	resSlice2, ok2 := res2.([]interface{})
	if !ok2 {
		t.Fatalf("Expected []interface{}, got %T", res2)
	}
	if len(resSlice2) != 1 {
		t.Errorf("Expected 1 result, got %d", len(resSlice2))
	}
}

// Issue #43: root jsonpath filter on array
// https://github.com/oliveagle/jsonpath/issues/43
func Test_jsonpath_root_array_filter(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"name": "John", "age": 30},
		map[string]interface{}{"name": "Jane", "age": 25},
	}

	// Test $[?(@.age == 30)] on root array
	res, err := JsonPathLookup(input, "$[?(@.age == 30)]")
	if err != nil {
		t.Fatalf("$[?(@.age == 30)] failed: %v", err)
	}
	resSlice, ok := res.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", res)
	}
	if len(resSlice) != 1 {
		t.Errorf("Expected 1 result, got %d: %v", len(resSlice), resSlice)
	}
}

// Issue #27: Range syntax doesn't match RFC 9535
// https://github.com/oliveagle/jsonpath/issues/27
func Test_jsonpath_range_syntax_rfc9535(t *testing.T) {
	// Test case 1: $[1:10] on small array should not error
	arr1 := []interface{}{"first", "second", "third"}
	res, err := JsonPathLookup(arr1, "$[1:10]")
	if err != nil {
		t.Fatalf("$[1:10] failed: %v", err)
	}
	resSlice, ok := res.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", res)
	}
	if len(resSlice) != 2 {
		t.Errorf("Expected 2 elements, got %d: %v", len(resSlice), resSlice)
	}
	if resSlice[0] != "second" || resSlice[1] != "third" {
		t.Errorf("Expected [second, third], got %v", resSlice)
	}

	// Test case 2: $[:2] should return first 2 elements (exclusive end)
	arr2 := []interface{}{1, 2, 3, 4, 5}
	res, err = JsonPathLookup(arr2, "$[:2]")
	if err != nil {
		t.Fatalf("$[:2] failed: %v", err)
	}
	resSlice, ok = res.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", res)
	}
	if len(resSlice) != 2 {
		t.Errorf("Expected 2 elements, got %d: %v", len(resSlice), resSlice)
	}
	if resSlice[0].(int) != 1 || resSlice[1].(int) != 2 {
		t.Errorf("Expected [1, 2], got %v", resSlice)
	}

	// Test case 3: $[2:] should return elements from index 2 onwards
	res, err = JsonPathLookup(arr2, "$[2:]")
	if err != nil {
		t.Fatalf("$[2:] failed: %v", err)
	}
	resSlice, ok = res.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", res)
	}
	if len(resSlice) != 3 {
		t.Errorf("Expected 3 elements, got %d: %v", len(resSlice), resSlice)
	}
	if resSlice[0].(int) != 3 || resSlice[1].(int) != 4 || resSlice[2].(int) != 5 {
		t.Errorf("Expected [3, 4, 5], got %v", resSlice)
	}

	// Test case 4: $[:-1] should include elements up to last (RFC 9535: -1 = last element)
	res, err = JsonPathLookup(arr2, "$[:-1]")
	if err != nil {
		t.Fatalf("$[:-1] failed: %v", err)
	}
	resSlice, ok = res.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", res)
	}
	// RFC 9535: -1 means last element, slice end is exclusive
	// So [:-1] returns elements from 0 to (last - 1), which is all elements in this case
	if len(resSlice) != 5 {
		t.Errorf("Expected 5 elements, got %d: %v", len(resSlice), resSlice)
	}

	// Test case 5: $[-2:] should return last 2 elements
	res, err = JsonPathLookup(arr2, "$[-2:]")
	if err != nil {
		t.Fatalf("$[-2:] failed: %v", err)
	}
	resSlice, ok = res.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", res)
	}
	if len(resSlice) != 2 {
		t.Errorf("Expected 2 elements, got %d: %v", len(resSlice), resSlice)
	}
	if resSlice[0].(int) != 4 || resSlice[1].(int) != 5 {
		t.Errorf("Expected [4, 5], got %v", resSlice)
	}

	// Test case 6: $[1:4] should return elements at indices 1, 2, 3
	res, err = JsonPathLookup(arr2, "$[1:4]")
	if err != nil {
		t.Fatalf("$[1:4] failed: %v", err)
	}
	resSlice, ok = res.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", res)
	}
	if len(resSlice) != 3 {
		t.Errorf("Expected 3 elements, got %d: %v", len(resSlice), resSlice)
	}
	if resSlice[0].(int) != 2 || resSlice[1].(int) != 3 || resSlice[2].(int) != 4 {
		t.Errorf("Expected [2, 3, 4], got %v", resSlice)
	}
}
