// Copyright 2015, 2021; oliver, DoltHub Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package jsonpath

import (
	"encoding/json"
	"fmt"
	"go/token"
	"go/types"
	"reflect"
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

var token_cases = []map[string]interface{}{
	map[string]interface{}{
		"query":  "$.store.*",
		"tokens": []string{"$", "store", "*"},
	},
	map[string]interface{}{
		"query":  "$.store..price",
		"tokens": []string{"$", "store", "..", "price"},
	},
	map[string]interface{}{
		"query":  "$.store.book[*].author",
		"tokens": []string{"$", "store", "book[*]", "author"},
	},
	map[string]interface{}{
		"query":  "$..book[2]",
		"tokens": []string{"$", "..", "book[2]"},
	},
	map[string]interface{}{
		"query":  "$..book[(@.length-1)]",
		"tokens": []string{"$", "..", "book[(@.length-1)]"},
	},
	map[string]interface{}{
		"query":  "$..book[0,1]",
		"tokens": []string{"$", "..", "book[0,1]"},
	},
	map[string]interface{}{
		"query":  "$..book[:2]",
		"tokens": []string{"$", "..", "book[:2]"},
	},
	map[string]interface{}{
		"query":  "$..book[?(@.isbn)]",
		"tokens": []string{"$", "..", "book[?(@.isbn)]"},
	},
	map[string]interface{}{
		"query":  "$..book[?(@.price <= $.expensive)]",
		"tokens": []string{"$", "..", "book[?(@.price <= $.expensive)]"},
	},
	map[string]interface{}{
		"query":  "$..book[?(@.author =~ /.*REES/i)]",
		"tokens": []string{"$", "..", "book[?(@.author =~ /.*REES/i)]"},
	},
	map[string]interface{}{
		"query":  "$..book[?(@.author =~ /.*REES\\]/i)]",
		"tokens": []string{"$", "..", "book[?(@.author =~ /.*REES\\]/i)]"},
	},
	map[string]interface{}{
		"query":  "$..*",
		"tokens": []string{"$", ".."},
	},
	// New test cases for recursive descent
	map[string]interface{}{
		"query":  "$..author",
		"tokens": []string{"$", "..", "author"},
	},
	map[string]interface{}{
		"query":  "$....author",
		"tokens": []string{"$", "..", "author"},
	},
}

func Test_jsonpath_tokenize(t *testing.T) {
	for _, tcase := range token_cases {
		query := tcase["query"].(string)
		expected := tcase["tokens"].([]string)
		t.Run(query, func(t *testing.T) {
			tokens, err := tokenize(query)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(tokens) != len(expected) {
				t.Errorf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
			}
			for i, token := range tokens {
				if i < len(expected) && token != expected[i] {
					t.Errorf("token[%d]: expected %q, got %q", i, expected[i], token)
				}
			}
		})
	}
}

var parse_token_cases = []map[string]interface{}{
	map[string]interface{}{
		"token": "$",
		"op":    "root",
		"key":   "$",
		"args":  nil,
	},
	map[string]interface{}{
		"token": "store",
		"op":    "key",
		"key":   "store",
		"args":  nil,
	},

	// idx --------------------------------------
	map[string]interface{}{
		"token": "book[2]",
		"op":    "idx",
		"key":   "book",
		"args":  []int{2},
	},
	map[string]interface{}{
		"token": "book[-1]",
		"op":    "idx",
		"key":   "book",
		"args":  []int{-1},
	},
	map[string]interface{}{
		"token": "book[0,1]",
		"op":    "idx",
		"key":   "book",
		"args":  []int{0, 1},
	},
	map[string]interface{}{
		"token": "[0]",
		"op":    "idx",
		"key":   "",
		"args":  []int{0},
	},

	// range ------------------------------------
	map[string]interface{}{
		"token": "book[1:-1]",
		"op":    "range",
		"key":   "book",
		"args":  [2]interface{}{1, -1},
	},
	map[string]interface{}{
		"token": "book[*]",
		"op":    "range",
		"key":   "book",
		"args":  [2]interface{}{nil, nil},
	},
	map[string]interface{}{
		"token": "book[:2]",
		"op":    "range",
		"key":   "book",
		"args":  [2]interface{}{nil, 2},
	},
	map[string]interface{}{
		"token": "book[-2:]",
		"op":    "range",
		"key":   "book",
		"args":  [2]interface{}{-2, nil},
	},

	// filter --------------------------------
	map[string]interface{}{
		"token": "book[?( @.isbn      )]",
		"op":    "filter",
		"key":   "book",
		"args":  "@.isbn",
	},
	map[string]interface{}{
		"token": "book[?(@.price < 10)]",
		"op":    "filter",
		"key":   "book",
		"args":  "@.price < 10",
	},
	map[string]interface{}{
		"token": "book[?(@.price <= $.expensive)]",
		"op":    "filter",
		"key":   "book",
		"args":  "@.price <= $.expensive",
	},
	map[string]interface{}{
		"token": "book[?(@.author =~ /.*REES/i)]",
		"op":    "filter",
		"key":   "book",
		"args":  "@.author =~ /.*REES/i",
	},
	map[string]interface{}{
		"token": "*",
		"op":    "scan",
		"key":   "*",
		"args":  nil,
	},
}

func Test_jsonpath_parse_token(t *testing.T) {
	for idx, tcase := range parse_token_cases {
		t.Logf("[%d] - tcase: %v", idx, tcase)
		token := tcase["token"].(string)
		exp_op := tcase["op"].(string)
		exp_key := tcase["key"].(string)
		exp_args := tcase["args"]

		op, key, args, err := parse_token(token)
		t.Logf("[%d] - expected: op: %v, key: %v, args: %v\n", idx, exp_op, exp_key, exp_args)
		t.Logf("[%d] - got: err: %v, op: %v, key: %v, args: %v\n", idx, err, op, key, args)
		if op != exp_op {
			t.Errorf("ERROR: op(%v) != exp_op(%v)", op, exp_op)
			return
		}
		if key != exp_key {
			t.Errorf("ERROR: key(%v) != exp_key(%v)", key, exp_key)
			return
		}

		if op == "idx" {
			if args_v, ok := args.([]int); ok == true {
				for i, v := range args_v {
					if v != exp_args.([]int)[i] {
						t.Errorf("ERROR: different args: [%d], (got)%v != (exp)%v", i, v, exp_args.([]int)[i])
						return
					}
				}
			} else {
				t.Errorf("ERROR: idx op should expect args:[]int{} in return, (got)%v", reflect.TypeOf(args))
				return
			}
		}

		if op == "range" {
			if args_v, ok := args.([2]interface{}); ok == true {
				t.Logf("%v", args_v)
				exp_from := exp_args.([2]interface{})[0]
				exp_to := exp_args.([2]interface{})[1]
				if args_v[0] != exp_from {
					t.Errorf("(from)%v != (exp_from)%v", args_v[0], exp_from)
					return
				}
				if args_v[1] != exp_to {
					t.Errorf("(to)%v != (exp_to)%v", args_v[1], exp_to)
					return
				}
			} else {
				t.Errorf("ERROR: range op should expect args:[2]interface{}, (got)%v", reflect.TypeOf(args))
				return
			}
		}

		if op == "filter" {
			if args_v, ok := args.(string); ok == true {
				t.Logf("%s", args_v)
				if exp_args.(string) != args_v {
					t.Errorf("len(args) not expected: (got)%v != (exp)%v", len(args_v), len(exp_args.([]string)))
					return
				}

			} else {
				t.Errorf("ERROR: filter op should expect args:[]string{}, (got)%v", reflect.TypeOf(args))
			}
		}
	}
}

func Test_jsonpath_get_key(t *testing.T) {
	obj := map[string]interface{}{
		"key": 1,
	}
	res, err := get_key(obj, "key")
	t.Logf("err: %v, res: %v", err, res)
	if err != nil {
		t.Errorf("failed to get key: %v", err)
		return
	}
	if res.(int) != 1 {
		t.Errorf("key value is not 1: %v", res)
		return
	}

	res, err = get_key(obj, "hah")
	t.Logf("err: %v, res: %v", err, res)
	if err == nil {
		t.Errorf("key error not raised")
		return
	}
	if res != nil {
		t.Errorf("key error should return nil res: %v", res)
		return
	}

	obj2 := 1
	res, err = get_key(obj2, "key")
	t.Logf("err: %v, res: %v", err, res)
	if err == nil {

		t.Errorf("object is not map error not raised")
		return
	}
	obj3 := map[string]string{"key": "hah"}
	res, err = get_key(obj3, "key")
	if res_v, ok := res.(string); ok != true || res_v != "hah" {
		t.Logf("err: %v, res: %v", err, res)
		t.Errorf("map[string]string support failed")
	}

	obj4 := []map[string]interface{}{
		map[string]interface{}{
			"a": 1,
		},
		map[string]interface{}{
			"a": 2,
		},
	}
	res, err = get_key(obj4, "a")
	t.Logf("err: %v, res: %v", err, res)
}

func Test_jsonpath_get_idx(t *testing.T) {
	obj := []interface{}{1, 2, 3, 4}
	res, err := get_idx(obj, 0)
	t.Logf("err: %v, res: %v", err, res)
	if err != nil {
		t.Errorf("failed to get_idx(obj,0): %v", err)
		return
	}
	if v, ok := res.(int); ok != true || v != 1 {
		t.Errorf("failed to get int 1")
	}

	res, err = get_idx(obj, 2)
	t.Logf("err: %v, res: %v", err, res)
	if v, ok := res.(int); ok != true || v != 3 {
		t.Errorf("failed to get int 3")
	}
	res, err = get_idx(obj, 4)
	t.Logf("err: %v, res: %v", err, res)
	if err == nil {
		t.Errorf("index out of range  error not raised")
		return
	}

	res, err = get_idx(obj, -1)
	t.Logf("err: %v, res: %v", err, res)
	if err != nil {
		t.Errorf("failed to get_idx(obj, -1): %v", err)
		return
	}
	if v, ok := res.(int); ok != true || v != 4 {
		t.Errorf("failed to get int 4")
	}

	res, err = get_idx(obj, -4)
	t.Logf("err: %v, res: %v", err, res)
	if v, ok := res.(int); ok != true || v != 1 {
		t.Errorf("failed to get int 1")
	}

	res, err = get_idx(obj, -5)
	t.Logf("err: %v, res: %v", err, res)
	if err == nil {
		t.Errorf("index out of range  error not raised")
		return
	}

	obj1 := 1
	res, err = get_idx(obj1, 1)
	if err == nil {
		t.Errorf("object is not Slice error not raised")
		return
	}

	obj2 := []int{1, 2, 3, 4}
	res, err = get_idx(obj2, 0)
	t.Logf("err: %v, res: %v", err, res)
	if err != nil {
		t.Errorf("failed to get_idx(obj2,0): %v", err)
		return
	}
	if v, ok := res.(int); ok != true || v != 1 {
		t.Errorf("failed to get int 1")
	}
}

func Test_jsonpath_get_range(t *testing.T) {
	obj := []int{1, 2, 3, 4, 5}

	res, err := get_range(obj, 0, 2)
	t.Logf("err: %v, res: %v", err, res)
	if err != nil {
		t.Errorf("failed to get_range: %v", err)
	}
	if res.([]int)[0] != 1 || res.([]int)[1] != 2 {
		t.Errorf("failed get_range: %v, expect: [1,2]", res)
	}

	obj1 := []interface{}{1, 2, 3, 4, 5}
	res, err = get_range(obj1, 3, -1)
	t.Logf("err: %v, res: %v", err, res)
	if err != nil {
		t.Errorf("failed to get_range: %v", err)
	}
	t.Logf("%v", res.([]interface{}))
	if res.([]interface{})[0] != 4 || res.([]interface{})[1] != 5 {
		t.Errorf("failed get_range: %v, expect: [4,5]", res)
	}

	res, err = get_range(obj1, nil, 2)
	t.Logf("err: %v, res:%v", err, res)
	if res.([]interface{})[0] != 1 || res.([]interface{})[1] != 2 {
		t.Errorf("from support nil failed: %v", res)
	}

	res, err = get_range(obj1, nil, nil)
	t.Logf("err: %v, res:%v", err, res)
	if len(res.([]interface{})) != 5 {
		t.Errorf("from, to both nil failed")
	}

	res, err = get_range(obj1, -2, nil)
	t.Logf("err: %v, res:%v", err, res)
	if res.([]interface{})[0] != 4 || res.([]interface{})[1] != 5 {
		t.Errorf("from support nil failed: %v", res)
	}

	obj2 := 2
	res, err = get_range(obj2, 0, 1)
	t.Logf("err: %v, res: %v", err, res)
	if err == nil {
		t.Errorf("object is Slice error not raised")
	}
}

func Test_jsonpath_get_scan(t *testing.T) {
	obj := map[string]interface{}{
		"key": 1,
	}
	res, err := get_scan(obj)
	if err != nil {
		t.Errorf("failed to scan: %v", err)
		return
	}
	if res.([]interface{})[0] != 1 {
		t.Errorf("scanned value is not 1: %v", res)
		return
	}

	obj2 := 1
	res, err = get_scan(obj2)
	if err == nil || err.Error() != "object is not scannable: int" {
		t.Errorf("object is not scannable error not raised")
		return
	}

	obj3 := map[string]string{"key1": "hah1", "key2": "hah2", "key3": "hah3"}
	res, err = get_scan(obj3)
	if err != nil {
		t.Errorf("failed to scan: %v", err)
		return
	}
	res_v, ok := res.([]interface{})
	if !ok {
		t.Errorf("scanned result is not a slice")
	}
	if len(res_v) != 3 {
		t.Errorf("scanned result is of wrong length")
	}
	if v, ok := res_v[0].(string); !ok || v != "hah1" {
		t.Errorf("scanned result contains unexpected value: %v", v)
	}
	if v, ok := res_v[1].(string); !ok || v != "hah2" {
		t.Errorf("scanned result contains unexpected value: %v", v)
	}
	if v, ok := res_v[2].(string); !ok || v != "hah3" {
		t.Errorf("scanned result contains unexpected value: %v", v)
	}

	obj4 := map[string]interface{}{
		"key1": "abc",
		"key2": 123,
		"key3": map[string]interface{}{
			"a": 1,
			"b": 2,
			"c": 3,
		},
		"key4": []interface{}{1, 2, 3},
		"key5": nil,
	}
	res, err = get_scan(obj4)
	res_v, ok = res.([]interface{})
	if !ok {
		t.Errorf("scanned result is not a slice")
	}
	if len(res_v) != 5 {
		t.Errorf("scanned result is of wrong length")
	}
	if v, ok := res_v[0].(string); !ok || v != "abc" {
		t.Errorf("scanned result contains unexpected value: %v", v)
	}
	if v, ok := res_v[1].(int); !ok || v != 123 {
		t.Errorf("scanned result contains unexpected value: %v", v)
	}
	if v, ok := res_v[2].(map[string]interface{}); !ok || v["a"].(int) != 1 || v["b"].(int) != 2 || v["c"].(int) != 3 {
		t.Errorf("scanned result contains unexpected value: %v", v)
	}
	if v, ok := res_v[3].([]interface{}); !ok || v[0].(int) != 1 || v[1].(int) != 2 || v[2].(int) != 3 {
		t.Errorf("scanned result contains unexpected value: %v", v)
	}
	if res_v[4] != nil {
		t.Errorf("scanned result contains unexpected value: %v", res_v[4])
	}
}

func Test_jsonpath_types_eval(t *testing.T) {
	fset := token.NewFileSet()
	res, err := types.Eval(fset, nil, 0, "1 < 2")
	t.Logf("err: %v, res: %v, res.Type: %v, res.Value: %v, res.IsValue: %v", err, res, res.Type, res.Value, res.IsValue())
}

var tcase_parse_filter = []map[string]interface{}{
	// 0
	map[string]interface{}{
		"filter":  "@.isbn",
		"exp_lp":  "@.isbn",
		"exp_op":  "exists",
		"exp_rp":  "",
		"exp_err": nil,
	},
	// 1
	map[string]interface{}{
		"filter":  "@.price < 10",
		"exp_lp":  "@.price",
		"exp_op":  "<",
		"exp_rp":  "10",
		"exp_err": nil,
	},
	// 2
	map[string]interface{}{
		"filter":  "@.price <= $.expensive",
		"exp_lp":  "@.price",
		"exp_op":  "<=",
		"exp_rp":  "$.expensive",
		"exp_err": nil,
	},
	// 3
	map[string]interface{}{
		"filter":  "@.author =~ /.*REES/i",
		"exp_lp":  "@.author",
		"exp_op":  "=~",
		"exp_rp":  "/.*REES/i",
		"exp_err": nil,
	},

	// 4
	{
		"filter": "@.author == 'Nigel Rees'",
		"exp_lp": "@.author",
		"exp_op": "==",
		"exp_rp": "Nigel Rees",
	},
}

func Test_jsonpath_parse_filter(t *testing.T) {

	//for _, tcase := range tcase_parse_filter[4:] {
	for _, tcase := range tcase_parse_filter {
		lp, op, rp, _ := parse_filter(tcase["filter"].(string))
		t.Log(tcase)
		t.Logf("lp: %v, op: %v, rp: %v", lp, op, rp)
		if lp != tcase["exp_lp"].(string) {
			t.Errorf("%s(got) != %v(exp_lp)", lp, tcase["exp_lp"])
			return
		}
		if op != tcase["exp_op"].(string) {
			t.Errorf("%s(got) != %v(exp_op)", op, tcase["exp_op"])
			return
		}
		if rp != tcase["exp_rp"].(string) {
			t.Errorf("%s(got) != %v(exp_rp)", rp, tcase["exp_rp"])
			return
		}
	}
}

var tcase_filter_get_from_explicit_path = []map[string]interface{}{
	// 0
	map[string]interface{}{
		// 0 {"a": 1}
		"obj":      map[string]interface{}{"a": 1},
		"query":    "$.a",
		"expected": 1,
	},
	map[string]interface{}{
		// 1 {"a":{"b":1}}
		"obj":      map[string]interface{}{"a": map[string]interface{}{"b": 1}},
		"query":    "$.a.b",
		"expected": 1,
	},
	map[string]interface{}{
		// 2 {"a": {"b":1, "c":2}}
		"obj":      map[string]interface{}{"a": map[string]interface{}{"b": 1, "c": 2}},
		"query":    "$.a.c",
		"expected": 2,
	},
	map[string]interface{}{
		// 3 {"a": {"b":1}, "b": 2}
		"obj":      map[string]interface{}{"a": map[string]interface{}{"b": 1}, "b": 2},
		"query":    "$.a.b",
		"expected": 1,
	},
	map[string]interface{}{
		// 4 {"a": {"b":1}, "b": 2}
		"obj":      map[string]interface{}{"a": map[string]interface{}{"b": 1}, "b": 2},
		"query":    "$.b",
		"expected": 2,
	},
	map[string]interface{}{
		// 5 {'a': ['b',1]}
		"obj":      map[string]interface{}{"a": []interface{}{"b", 1}},
		"query":    "$.a[0]",
		"expected": "b",
	},
}

func Test_jsonpath_filter_get_from_explicit_path(t *testing.T) {

	for idx, tcase := range tcase_filter_get_from_explicit_path {
		obj := tcase["obj"]
		query := tcase["query"].(string)
		expected := tcase["expected"]

		res, err := filter_get_from_explicit_path(obj, query)
		t.Log(idx, err, res)
		if err != nil {
			t.Errorf("flatten_cases: failed: [%d] %v", idx, err)
		}
		// t.Logf("typeof(res): %v, typeof(expected): %v", reflect.TypeOf(res), reflect.TypeOf(expected))
		if reflect.TypeOf(res) != reflect.TypeOf(expected) {
			t.Errorf("different type: (res)%v != (expected)%v", reflect.TypeOf(res), reflect.TypeOf(expected))
			continue
		}
		switch expected.(type) {
		case map[string]interface{}:
			if len(res.(map[string]interface{})) != len(expected.(map[string]interface{})) {
				t.Errorf("two map with differnt lenght: (res)%v, (expected)%v", res, expected)
			}
		default:
			if res != expected {
				t.Errorf("res(%v) != expected(%v)", res, expected)
			}
		}
	}
}

var tcase_eval_filter = []map[string]interface{}{
	// 0
	map[string]interface{}{
		"obj":  map[string]interface{}{"a": 1},
		"root": map[string]interface{}{},
		"lp":   "@.a",
		"op":   "exists",
		"rp":   "",
		"exp":  true,
	},
	// 1
	map[string]interface{}{
		"obj":  map[string]interface{}{"a": 1},
		"root": map[string]interface{}{},
		"lp":   "@.b",
		"op":   "exists",
		"rp":   "",
		"exp":  false,
	},
	// 2
	map[string]interface{}{
		"obj":  map[string]interface{}{"a": 1},
		"root": map[string]interface{}{"a": 1},
		"lp":   "$.a",
		"op":   "exists",
		"rp":   "",
		"exp":  true,
	},
	// 3
	map[string]interface{}{
		"obj":  map[string]interface{}{"a": 1},
		"root": map[string]interface{}{"a": 1},
		"lp":   "$.b",
		"op":   "exists",
		"rp":   "",
		"exp":  false,
	},
	// 4
	map[string]interface{}{
		"obj":  map[string]interface{}{"a": 1, "b": map[string]interface{}{"c": 2}},
		"root": map[string]interface{}{"a": 1, "b": map[string]interface{}{"c": 2}},
		"lp":   "$.b.c",
		"op":   "exists",
		"rp":   "",
		"exp":  true,
	},
	// 5
	map[string]interface{}{
		"obj":  map[string]interface{}{"a": 1, "b": map[string]interface{}{"c": 2}},
		"root": map[string]interface{}{},
		"lp":   "$.b.a",
		"op":   "exists",
		"rp":   "",
		"exp":  false,
	},

	// 6
	map[string]interface{}{
		"obj":  map[string]interface{}{"a": 3},
		"root": map[string]interface{}{"a": 3},
		"lp":   "$.a",
		"op":   ">",
		"rp":   "1",
		"exp":  true,
	},
}

func Test_jsonpath_eval_filter(t *testing.T) {
	for idx, tcase := range tcase_eval_filter[1:] {
		t.Logf("------------------------------")
		obj := tcase["obj"].(map[string]interface{})
		root := tcase["root"].(map[string]interface{})
		lp := tcase["lp"].(string)
		op := tcase["op"].(string)
		rp := tcase["rp"].(string)
		exp := tcase["exp"].(bool)
		t.Logf("idx: %v, lp: %v, op: %v, rp: %v, exp: %v", idx, lp, op, rp, exp)
		got, err := eval_filter(obj, root, lp, op, rp)

		if err != nil {
			t.Errorf("idx: %v, failed to eval: %v", idx, err)
			return
		}
		if got != exp {
			t.Errorf("idx: %v, %v(got) != %v(exp)", idx, got, exp)
		}

	}
}

var (
	ifc1 interface{} = "haha"
	ifc2 interface{} = "ha ha"
)
var tcase_cmp_any = []map[string]interface{}{

	map[string]interface{}{
		"obj1": 1,
		"obj2": 1,
		"op":   "==",
		"exp":  true,
		"err":  nil,
	},
	map[string]interface{}{
		"obj1": 1,
		"obj2": 2,
		"op":   "==",
		"exp":  false,
		"err":  nil,
	},
	map[string]interface{}{
		"obj1": 1.1,
		"obj2": 2.0,
		"op":   "<",
		"exp":  true,
		"err":  nil,
	},
	map[string]interface{}{
		"obj1": "1",
		"obj2": "2.0",
		"op":   "<",
		"exp":  true,
		"err":  nil,
	},
	map[string]interface{}{
		"obj1": "1",
		"obj2": "2.0",
		"op":   ">",
		"exp":  false,
		"err":  nil,
	},
	map[string]interface{}{
		"obj1": 1,
		"obj2": 2,
		"op":   "=~",
		"exp":  false,
		"err":  "op should only be <, <=, ==, >= and >",
	}, {
		"obj1": ifc1,
		"obj2": ifc1,
		"op":   "==",
		"exp":  true,
		"err":  nil,
	}, {
		"obj1": ifc2,
		"obj2": ifc2,
		"op":   "==",
		"exp":  true,
		"err":  nil,
	}, {
		"obj1": 20,
		"obj2": "100",
		"op":   ">",
		"exp":  false,
		"err":  nil,
	},
}

func Test_jsonpath_cmp_any(t *testing.T) {
	for idx, tcase := range tcase_cmp_any {
		//for idx, tcase := range tcase_cmp_any[8:] {
		t.Logf("idx: %v, %v %v %v, exp: %v", idx, tcase["obj1"], tcase["op"], tcase["obj2"], tcase["exp"])
		res, err := cmp_any(tcase["obj1"], tcase["obj2"], tcase["op"].(string))
		exp := tcase["exp"].(bool)
		exp_err := tcase["err"]
		if exp_err != nil {
			if err == nil {
				t.Errorf("idx: %d error not raised: %v(exp)", idx, exp_err)
				break
			}
		} else {
			if err != nil {
				t.Errorf("idx: %v, error: %v", idx, err)
				break
			}
		}
		if res != exp {
			t.Errorf("idx: %v, %v(got) != %v(exp)", idx, res, exp)
			break
		}
	}
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

func BenchmarkJsonPathLookupCompiled(b *testing.B) {
	c, err := Compile("$.store.book[0].price")
	if err != nil {
		b.Fatalf("%v", err)
	}
	for n := 0; n < b.N; n++ {
		res, err := c.Lookup(json_data)
		if res_v, ok := res.(float64); ok != true || res_v != 8.95 {
			b.Errorf("$.store.book[0].price should be 8.95")
		}
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkJsonPathLookup(b *testing.B) {
	for n := 0; n < b.N; n++ {
		res, err := JsonPathLookup(json_data, "$.store.book[0].price")
		if res_v, ok := res.(float64); ok != true || res_v != 8.95 {
			b.Errorf("$.store.book[0].price should be 8.95")
		}
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkJsonPathLookup_0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.expensive")
	}
}

func BenchmarkJsonPathLookup_1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[0].price")
	}
}

func BenchmarkJsonPathLookup_2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[-1].price")
	}
}

func BenchmarkJsonPathLookup_3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[0,1].price")
	}
}

func BenchmarkJsonPathLookup_4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[0:2].price")
	}
}

func BenchmarkJsonPathLookup_5(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[?(@.isbn)].price")
	}
}

func BenchmarkJsonPathLookup_6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[?(@.price > 10)].title")
	}
}

func BenchmarkJsonPathLookup_7(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[?(@.price < $.expensive)].price")
	}
}

func BenchmarkJsonPathLookup_8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[:].price")
	}
}

func BenchmarkJsonPathLookup_9(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[?(@.author == 'Nigel Rees')].price")
	}
}

func BenchmarkJsonPathLookup_10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[?(@.author =~ /(?i).*REES/)].price")
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

// Benchmarks

func BenchmarkJsonPathLookup_Simple(b *testing.B) {
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{"author": "A", "price": 10.0},
				map[string]interface{}{"author": "B", "price": 20.0},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		JsonPathLookup(data, "$.store.book[0].author")
	}
}

func BenchmarkJsonPathLookup_Filter(b *testing.B) {
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{"author": "A", "price": 10.0},
				map[string]interface{}{"author": "B", "price": 20.0},
				map[string]interface{}{"author": "C", "price": 30.0},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		JsonPathLookup(data, "$.store.book[?(@.price > 15)].author")
	}
}

func BenchmarkJsonPathLookup_Range(b *testing.B) {
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{"author": "A", "price": 10.0},
				map[string]interface{}{"author": "B", "price": 20.0},
				map[string]interface{}{"author": "C", "price": 30.0},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		JsonPathLookup(data, "$.store.book[0:2].price")
	}
}

func BenchmarkJsonPathLookup_Recursive(b *testing.B) {
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{"author": "A", "price": 10.0},
				map[string]interface{}{"author": "B", "price": 20.0},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		JsonPathLookup(data, "$..author")
	}
}

func BenchmarkJsonPathLookup_RootArrayFilter(b *testing.B) {
	data := []interface{}{
		map[string]interface{}{"name": "John", "age": 30},
		map[string]interface{}{"name": "Jane", "age": 25},
		map[string]interface{}{"name": "Bob", "age": 35},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		JsonPathLookup(data, "$[?(@.age > 25)]")
	}
}

func BenchmarkCompileAndLookup(b *testing.B) {
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{"author": "A", "price": 10.0},
				map[string]interface{}{"author": "B", "price": 20.0},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := Compile("$.store.book[?(@.price > 10)].author")
		c.Lookup(data)
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
