// Copyright 2015, 2021; oliver, DoltHub Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package jsonpath

import "testing"

// Tokenizer tests - verify that JSONPath queries are correctly tokenized

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
