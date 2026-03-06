// Copyright 2015, 2021; oliver, DoltHub Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package jsonpath

import (
	"testing"
)

// TestParseFilter tests the parse_filter function
// Covers lines 758-837 in jsonpath.go
func TestParseFilter(t *testing.T) {
	tests := []struct {
		name      string
		filter    string
		wantLp    string
		wantOp    string
		wantRp    string
		wantErr   bool
		coverCase string
	}{
		{
			name:      "Simple exists check",
			filter:    "@.isbn",
			wantLp:    "@.isbn",
			wantOp:    "exists",
			wantRp:    "",
			wantErr:   false,
			coverCase: "Basic exists check without operator",
		},
		{
			name:      "String comparison with single quotes",
			filter:    "@.author == 'Nigel Rees'",
			wantLp:    "@.author",
			wantOp:    "==",
			wantRp:    "Nigel Rees",
			wantErr:   false,
			coverCase: "Single quote handling (lines 766-773)",
		},
		{
			name:      "String comparison with single quotes - space in value",
			filter:    "@.author == 'John Doe'",
			wantLp:    "@.author",
			wantOp:    "==",
			wantRp:    "John Doe",
			wantErr:   false,
			coverCase: "Single quote with space inside (lines 766-773)",
		},
		{
			name:      "String comparison with double quotes",
			filter:    `@.author == "Nigel Rees"`,
			wantLp:    "@.author",
			wantOp:    "==",
			wantRp:    "Nigel Rees",
			wantErr:   false,
			coverCase: "Double quote handling (lines 775-782)",
		},
		{
			name:      "String comparison with double quotes - space in value",
			filter:    `@.title == "Book Title"`,
			wantLp:    "@.title",
			wantOp:    "==",
			wantRp:    "Book Title",
			wantErr:   false,
			coverCase: "Double quote with space inside (lines 775-782)",
		},
		{
			name:      "Function call count with comparison",
			filter:    "count(@.book) > 0",
			wantLp:    "count(@.book)",
			wantOp:    ">",
			wantRp:    "0",
			wantErr:   false,
			coverCase: "Nested parentheses handling (lines 784-795)",
		},
		{
			name:      "Function call with nested parentheses",
			filter:    "count(@.items) >= 5",
			wantLp:    "count(@.items)",
			wantOp:    ">=",
			wantRp:    "5",
			wantErr:   false,
			coverCase: "Nested parentheses with >= operator (lines 784-795)",
		},
		{
			name:      "Expression with spaces around operator",
			filter:    "@.price > 10",
			wantLp:    "@.price",
			wantOp:    ">",
			wantRp:    "10",
			wantErr:   false,
			coverCase: "Space handling as delimiter (lines 796-800)",
		},
		{
			name:      "Expression with multiple spaces",
			filter:    "@.price  <  20",
			wantLp:    "@.price",
			wantOp:    "<",
			wantRp:    "20",
			wantErr:   false,
			coverCase: "Multiple spaces handling (lines 796-800)",
		},
		{
			name:      "Function call without operator - exists check",
			filter:    "count(@.book)",
			wantLp:    "count(@.book)",
			wantOp:    "exists",
			wantRp:    "",
			wantErr:   false,
			coverCase: "Function call without operator (lines 825-829)",
		},
		{
			name:      "Simple path without operator",
			filter:    "@.name",
			wantLp:    "@.name",
			wantOp:    "exists",
			wantRp:    "",
			wantErr:   false,
			coverCase: "Simple path without operator (lines 825-829)",
		},
		{
			name:      "Invalid filter - too many parts",
			filter:    "@.a + @.b + @.c",
			wantLp:    "",
			wantOp:    "",
			wantRp:    "",
			wantErr:   true,
			coverCase: "Error case - invalid extra parts (lines 813-815)",
		},
		{
			name:      "Numeric comparison less than",
			filter:    "@.price < 10",
			wantLp:    "@.price",
			wantOp:    "<",
			wantRp:    "10",
			wantErr:   false,
			coverCase: "Less than operator",
		},
		{
			name:      "Numeric comparison less than or equal",
			filter:    "@.price <= 10",
			wantLp:    "@.price",
			wantOp:    "<=",
			wantRp:    "10",
			wantErr:   false,
			coverCase: "Less than or equal operator",
		},
		{
			name:      "Numeric comparison greater than or equal",
			filter:    "@.price >= 10",
			wantLp:    "@.price",
			wantOp:    ">=",
			wantRp:    "10",
			wantErr:   false,
			coverCase: "Greater than or equal operator",
		},
		{
			name:      "Numeric comparison greater than",
			filter:    "@.price > 10",
			wantLp:    "@.price",
			wantOp:    ">",
			wantRp:    "10",
			wantErr:   false,
			coverCase: "Greater than operator",
		},
		{
			name:      "Comparison with root path on right side",
			filter:    "@.price < $.expensive",
			wantLp:    "@.price",
			wantOp:    "<",
			wantRp:    "$.expensive",
			wantErr:   false,
			coverCase: "Root path reference on right side",
		},
		{
			name:      "Single quote with special characters",
			filter:    "@.code == 'A-B'",
			wantLp:    "@.code",
			wantOp:    "==",
			wantRp:    "A-B",
			wantErr:   false,
			coverCase: "Single quote with hyphen (lines 766-773)",
		},
		{
			name:      "Double quote with special characters",
			filter:    `@.path == "a/b/c"`,
			wantLp:    "@.path",
			wantOp:    "==",
			wantRp:    "a/b/c",
			wantErr:   false,
			coverCase: "Double quote with slashes (lines 775-782)",
		},
		{
			name:      "Parentheses in function with spaces",
			filter:    "count( @.book ) > 0",
			wantLp:    "count( @.book )",
			wantOp:    ">",
			wantRp:    "0",
			wantErr:   false,
			coverCase: "Spaces inside parentheses (lines 796-800)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lp, op, rp, err := parse_filter(tt.filter)

			if (err != nil) != tt.wantErr {
				t.Errorf("parse_filter(%q) error = %v, wantErr %v", tt.filter, err, tt.wantErr)
				return
			}

			if lp != tt.wantLp {
				t.Errorf("parse_filter(%q) lp = %q, want %q", tt.filter, lp, tt.wantLp)
			}
			if op != tt.wantOp {
				t.Errorf("parse_filter(%q) op = %q, want %q", tt.filter, op, tt.wantOp)
			}
			if rp != tt.wantRp {
				t.Errorf("parse_filter(%q) rp = %q, want %q", tt.filter, rp, tt.wantRp)
			}

			if tt.coverCase != "" {
				t.Logf("Covered: %s", tt.coverCase)
			}
		})
	}
}

// TestParseFilterEdgeCases tests edge cases for parse_filter
func TestParseFilterEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		filter    string
		wantLp    string
		wantOp    string
		wantRp    string
		wantErr   bool
		coverCase string
	}{
		{
			name:      "Empty string after quotes",
			filter:    "@.name == ''",
			wantLp:    "@.name",
			wantOp:    "==",
			wantRp:    "",
			wantErr:   false,
			coverCase: "Empty single quoted string",
		},
		{
			name:      "Empty string after double quotes",
			filter:    `@.name == ""`,
			wantLp:    "@.name",
			wantOp:    "==",
			wantRp:    "",
			wantErr:   false,
			coverCase: "Empty double quoted string",
		},
		{
			name:      "Quote at end of filter",
			filter:    "@.x == 'test'",
			wantLp:    "@.x",
			wantOp:    "==",
			wantRp:    "test",
			wantErr:   false,
			coverCase: "Quote at end of filter",
		},
		{
			name:      "Multiple words in quotes",
			filter:    "@.title == 'The Lord of the Rings'",
			wantLp:    "@.title",
			wantOp:    "==",
			wantRp:    "The Lord of the Rings",
			wantErr:   false,
			coverCase: "Multiple words in quotes",
		},
		{
			name:      "Nested function call",
			filter:    "count(@.a.b) > 0",
			wantLp:    "count(@.a.b)",
			wantOp:    ">",
			wantRp:    "0",
			wantErr:   false,
			coverCase: "Nested function call",
		},
		{
			name:      "Quote containing parenthesis",
			filter:    "@.formula == 'f(x)'",
			wantLp:    "@.formula",
			wantOp:    "==",
			wantRp:    "f(x)",
			wantErr:   false,
			coverCase: "Quote containing parenthesis",
		},
		{
			name:      "Single quote with double quote inside",
			filter:    "@.text == 'say \"hello\"'",
			wantLp:    "@.text",
			wantOp:    "==",
			wantRp:    "say \"hello\"",
			wantErr:   false,
			coverCase: "Single quote with double quote inside (lines 771-773)",
		},
		{
			name:      "Double quote with single quote inside",
			filter:    `@.text == "it's"`,
			wantLp:    "@.text",
			wantOp:    "==",
			wantRp:    "it's",
			wantErr:   false,
			coverCase: "Double quote with single quote inside (lines 780-782)",
		},
		{
			name:      "Only left side no operator",
			filter:    "@.name",
			wantLp:    "@.name",
			wantOp:    "exists",
			wantRp:    "",
			wantErr:   false,
			coverCase: "Only left side, no operator",
		},
		{
			name:      "Left and operator only",
			filter:    "@.x >",
			wantLp:    "@.x",
			wantOp:    ">",
			wantRp:    "",
			wantErr:   false,
			coverCase: "Left and operator only (lines 830-831)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lp, op, rp, err := parse_filter(tt.filter)

			if (err != nil) != tt.wantErr {
				t.Errorf("parse_filter(%q) error = %v, wantErr %v", tt.filter, err, tt.wantErr)
				return
			}

			if lp != tt.wantLp {
				t.Errorf("parse_filter(%q) lp = %q, want %q", tt.filter, lp, tt.wantLp)
			}
			if op != tt.wantOp {
				t.Errorf("parse_filter(%q) op = %q, want %q", tt.filter, op, tt.wantOp)
			}
			if rp != tt.wantRp {
				t.Errorf("parse_filter(%q) rp = %q, want %q", tt.filter, rp, tt.wantRp)
			}

			if tt.coverCase != "" {
				t.Logf("Covered: %s", tt.coverCase)
			}
		})
	}
}
