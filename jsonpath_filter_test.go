package jsonpath

import (
	"reflect"
	"testing"
)

// Filter parsing and evaluation tests

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
