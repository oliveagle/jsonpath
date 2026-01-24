// Copyright 2015, 2021; oliver, DoltHub Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package jsonpath

import (
	"reflect"
	"testing"
)

// Token parser tests - verify that tokens are correctly parsed into operations

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
