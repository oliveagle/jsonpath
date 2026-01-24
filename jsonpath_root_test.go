// Copyright 2015, 2021; oliver, DoltHub Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package jsonpath

import (
	"encoding/json"
	"testing"
)

// Root node and recursive descent tests

func Test_jsonpath_rootnode_is_array(t *testing.T) {
	data := `[{
     "test": 12.34
 }, {
	"test": 13.34
 }, {
	"test": 14.34}]
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
	"test": 14.34}]
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
