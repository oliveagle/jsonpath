// Copyright 2015, 2021; oliver, DoltHub Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package jsonpath

import (
	"errors"
	"fmt"
	"go/token"
	"go/types"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var ErrGetFromNullObj = errors.New("get attribute from null object")
var ErrKeyError = errors.New("key error: %s not found in object")

func JsonPathLookup(obj interface{}, jpath string) (interface{}, error) {
	c, err := Compile(jpath)
	if err != nil {
		return nil, err
	}
	return c.Lookup(obj)
}

type Compiled struct {
	path  string
	steps []step
}

type step struct {
	op   string
	key  string
	args interface{}
}

func MustCompile(jpath string) *Compiled {
	c, err := Compile(jpath)
	if err != nil {
		panic(err)
	}
	return c
}

func Compile(jpath string) (*Compiled, error) {
	tokens, err := tokenize(jpath)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty path")
	}
	if tokens[0] != "@" && tokens[0] != "$" {
		return nil, fmt.Errorf("$ or @ should in front of path")
	}
	tokens = tokens[1:]
	res := Compiled{
		path:  jpath,
		steps: make([]step, len(tokens)),
	}
	for i, token := range tokens {
		op, key, args, err := parse_token(token)
		if err != nil {
			return nil, err
		}
		res.steps[i] = step{op, key, args}
	}
	return &res, nil
}

func (c *Compiled) String() string {
	return fmt.Sprintf("Compiled lookup: %s", c.path)
}

func (c *Compiled) Lookup(obj interface{}) (interface{}, error) {
	var err error
	for i, s := range c.steps {
		// "key", "idx"
		switch s.op {
		case "key":
			obj, err = get_key(obj, s.key)
			if err != nil {
				return nil, err
			}
		case "idx":
			if len(s.key) > 0 {
				// no key `$[0].test`
				obj, err = get_key(obj, s.key)
				if err != nil {
					return nil, err
				}
			}

			if len(s.args.([]int)) > 1 {
				res := []interface{}{}
				for _, x := range s.args.([]int) {
					//fmt.Println("idx ---- ", x)
					tmp, err := get_idx(obj, x)
					if err != nil {
						return nil, err
					}
					res = append(res, tmp)
				}
				obj = res
			} else if len(s.args.([]int)) == 1 {
				//fmt.Println("idx ----------------3")
				obj, err = get_idx(obj, s.args.([]int)[0])
				if err != nil {
					return nil, err
				}
			} else {
				//fmt.Println("idx ----------------4")
				return nil, fmt.Errorf("cannot index on empty slice")
			}
		case "range":
			if len(s.key) > 0 {
				// no key `$[:1].test`
				obj, err = get_key(obj, s.key)
				if err != nil {
					return nil, err
				}
			}
			if argsv, ok := s.args.([2]interface{}); ok == true {
				obj, err = get_range(obj, argsv[0], argsv[1])
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("range args length should be 2")
			}
		case "filter":
			obj, err = get_key(obj, s.key)
			if err != nil {
				return nil, err
			}
			obj, err = get_filtered(obj, obj, s.args.(string))
			if err != nil {
				return nil, err
			}
		case "recursive":
			obj = getAllDescendants(obj)
			// Heuristic: if next step is key, exclude slices from candidates to avoid double-matching
			// (once as container via implicit map, once as individual elements)
			if i+1 < len(c.steps) && c.steps[i+1].op == "key" {
				if candidates, ok := obj.([]interface{}); ok {
					filtered := []interface{}{}
					for _, cand := range candidates {
						// Filter out Slices (but keep Maps and others)
						// because get_key on Slice iterates children, which are already in candidates
						v := reflect.ValueOf(cand)
						if v.Kind() != reflect.Slice {
							filtered = append(filtered, cand)
						}
					}
					obj = filtered
				}
			}
		case "func":
			// Handle function calls like length()
			// For function calls like $.length(), the key is the function name (e.g., "length")
			// For path-based function calls like $.store.book.length(), the key is empty
			// and we need to evaluate the function on the current object
			if len(s.key) > 0 {
				// This case handles paths like $.store.book.length() where the function
				// is called on the result of the previous path step
				obj, err = eval_func(obj, s.key)
				if err != nil {
					return nil, err
				}
			} else {
				// This case handles direct function calls like $.length() or @.length()
				obj, err = eval_func(obj, s.key)
				if err != nil {
					return nil, err
				}
			}
		default:
			return nil, fmt.Errorf("unsupported jsonpath operation: %s", s.op)
		}
	}
	return obj, nil
}

func tokenize(query string) ([]string, error) {
	tokens := []string{}
	//	token_start := false
	//	token_end := false
	token := ""
	quoteChar := rune(0)

	// fmt.Println("-------------------------------------------------- start")
	for idx, x := range query {
		if quoteChar != 0 {
			if x == quoteChar {
				quoteChar = 0
			} else {
				token += string(x)
			}

			continue
		} else if x == '"' {
			if token == "." {
				token = ""
			}

			quoteChar = x
			continue
		}

		token += string(x)
		// //fmt.Printf("idx: %d, x: %s, token: %s, tokens: %v\n", idx, string(x), token, tokens)
		if idx == 0 {
			if token == "$" || token == "@" {
				tokens = append(tokens, token[:])
				token = ""
				continue
			} else {
				return nil, fmt.Errorf("should start with '$'")
			}
		}
		if token == "." {
			continue
		} else if token == ".." {
			if len(tokens) == 0 || tokens[len(tokens)-1] != ".." {
				tokens = append(tokens, "..")
			}
			token = "."
			continue
		} else {
			// fmt.Println("else: ", string(x), token)
			if strings.Contains(token, "[") {
				// fmt.Println(" contains [ ")
				if x == ']' && !strings.HasSuffix(token, "\\]") {
					if token[0] == '.' {
						tokens = append(tokens, token[1:])
					} else {
						tokens = append(tokens, token[:])
					}
					token = ""
					continue
				}
			} else {
				// fmt.Println(" doesn't contains [ ")
				if x == '.' {
					if token[0] == '.' {
						tokens = append(tokens, token[1:len(token)-1])
					} else {
						tokens = append(tokens, token[:len(token)-1])
					}
					token = "."
					continue
				}
			}
		}
	}

	if quoteChar != 0 {
		token = string(quoteChar) + token
	}

	if len(token) > 0 {
		if token[0] == '.' {
			token = token[1:]
			if token != "*" {
				tokens = append(tokens, token[:])
			} else if len(tokens) > 0 && tokens[len(tokens)-1] == ".." {
				// $..* means recursive descent with scan, * is redundant after ..
				// Don't add * as separate token
			} else if tokens[len(tokens)-1] != "*" {
				tokens = append(tokens, token[:])
			}
		} else {
			if token != "*" {
				tokens = append(tokens, token[:])
			} else if len(tokens) > 0 && tokens[len(tokens)-1] == ".." {
				// $..* means recursive descent with scan, * is redundant after ..
				// Don't add * as separate token
			} else if tokens[len(tokens)-1] != "*" {
				tokens = append(tokens, token[:])
			}
		}
	}
	// fmt.Println("finished tokens: ", tokens)
	// fmt.Println("================================================= done ")
	return tokens, nil
}

/*
op: "root", "key", "idx", "range", "filter", "scan"
*/
func parse_token(token string) (op string, key string, args interface{}, err error) {
	if token == "$" {
		return "root", "$", nil, nil
	}
	if token == "*" {
		return "scan", "*", nil, nil
	}
	if token == ".." {
		return "recursive", "..", nil, nil
	}

	bracket_idx := strings.Index(token, "[")
	if bracket_idx < 0 {
		// Check for function call like length()
		if strings.HasSuffix(token, "()") {
			funcName := strings.TrimSuffix(token, "()")
			return "func", funcName, nil, nil
		}
		return "key", token, nil, nil
	} else {
		key = token[:bracket_idx]
		tail := token[bracket_idx:]
		if len(tail) < 3 {
			err = fmt.Errorf("len(tail) should >=3, %v", tail)
			return
		}
		tail = tail[1 : len(tail)-1]

		//fmt.Println(key, tail)
		if strings.HasPrefix(tail, "?") {
			// filter -------------------------------------------------
			op = "filter"
			// Remove leading ? - the content is everything after ?
			filterContent := tail[1:]
			// Handle filters like [?( @.isbn )] - remove outer parentheses if present
			filterContent = strings.TrimSpace(filterContent)
			if strings.HasPrefix(filterContent, "(") && strings.HasSuffix(filterContent, ")") {
				// Remove outer parentheses
				inner := filterContent[1 : len(filterContent)-1]
				filterContent = strings.TrimSpace(inner)
			}
			args = filterContent
			return
		} else if strings.Contains(tail, ":") {
			// range ----------------------------------------------
			op = "range"
			tails := strings.Split(tail, ":")
			if len(tails) != 2 {
				err = fmt.Errorf("only support one range(from, to): %v", tails)
				return
			}
			var frm interface{}
			var to interface{}
			if frm, err = strconv.Atoi(strings.Trim(tails[0], " ")); err != nil {
				if strings.Trim(tails[0], " ") == "" {
					err = nil
				}
				frm = nil
			}
			if to, err = strconv.Atoi(strings.Trim(tails[1], " ")); err != nil {
				if strings.Trim(tails[1], " ") == "" {
					err = nil
				}
				to = nil
			}
			args = [2]interface{}{frm, to}
			return
		} else if tail == "*" {
			op = "range"
			args = [2]interface{}{nil, nil}
			return
		} else {
			// idx ------------------------------------------------
			op = "idx"
			res := []int{}
			for _, x := range strings.Split(tail, ",") {
				if i, err := strconv.Atoi(strings.Trim(x, " ")); err == nil {
					res = append(res, i)
				} else {
					return "", "", nil, err
				}
			}
			args = res
		}
	}
	return op, key, args, nil
}

func filter_get_from_explicit_path(obj interface{}, path string) (interface{}, error) {
	steps, err := tokenize(path)
	//fmt.Println("f: steps: ", steps, err)
	//fmt.Println(path, steps)
	if err != nil {
		return nil, err
	}
	if steps[0] != "@" && steps[0] != "$" {
		return nil, fmt.Errorf("$ or @ should in front of path")
	}
	steps = steps[1:]
	xobj := obj
	//fmt.Println("f: xobj", xobj)
	for _, s := range steps {
		op, key, args, err := parse_token(s)
		// "key", "idx"
		switch op {
		case "key":
			xobj, err = get_key(xobj, key)
			if err != nil {
				return nil, err
			}
		case "idx":
			if len(args.([]int)) != 1 {
				return nil, fmt.Errorf("don't support multiple index in filter")
			}
			xobj, err = get_key(xobj, key)
			if err != nil {
				return nil, err
			}
			xobj, err = get_idx(xobj, args.([]int)[0])
			if err != nil {
				return nil, err
			}
		case "func":
			// Handle function calls like length()
			xobj, err = get_key(xobj, key)
			if err != nil {
				return nil, err
			}
			xobj, err = eval_func(xobj, key)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported jsonpath operation %s in filter", op)
		}
	}
	return xobj, nil
}

func get_key(obj interface{}, key string) (interface{}, error) {
	if reflect.TypeOf(obj) == nil {
		return nil, ErrGetFromNullObj
	}
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Map:
		// if obj came from stdlib json, its highly likely to be a map[string]interface{}
		// in which case we can save having to iterate the map keys to work out if the
		// key exists
		if jsonMap, ok := obj.(map[string]interface{}); ok {
			val, exists := jsonMap[key]
			if !exists {
				return nil, fmt.Errorf("key error: %s not found in object", key)
			}
			return val, nil
		}
		for _, kv := range reflect.ValueOf(obj).MapKeys() {
			//fmt.Println(kv.String())
			if kv.String() == key {
				return reflect.ValueOf(obj).MapIndex(kv).Interface(), nil
			}
		}
		return nil, fmt.Errorf("key error: %s not found in object", key)
	case reflect.Slice:
		// slice we should get from all objects in it.
		// if key is empty, return the slice itself (for root array filtering)
		if key == "" {
			return obj, nil
		}
		res := []interface{}{}
		for i := 0; i < reflect.ValueOf(obj).Len(); i++ {
			tmp, _ := get_idx(obj, i)
			if v, err := get_key(tmp, key); err == nil {
				res = append(res, v)
			}
		}
		return res, nil
	default:
		return nil, fmt.Errorf("object is not map")
	}
}

func get_idx(obj interface{}, idx int) (interface{}, error) {
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice:
		length := reflect.ValueOf(obj).Len()
		if idx >= 0 {
			if idx >= length {
				return nil, fmt.Errorf("index out of range: len: %v, idx: %v", length, idx)
			}
			return reflect.ValueOf(obj).Index(idx).Interface(), nil
		} else {
			// < 0
			_idx := length + idx
			if _idx < 0 {
				return nil, fmt.Errorf("index out of range: len: %v, idx: %v", length, idx)
			}
			return reflect.ValueOf(obj).Index(_idx).Interface(), nil
		}
	default:
		return nil, fmt.Errorf("object is not Slice")
	}
}

func get_range(obj, frm, to interface{}) (interface{}, error) {
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice:
		length := reflect.ValueOf(obj).Len()
		_frm := 0
		_to := length
		if frm == nil {
			frm = 0
		}
		if to == nil {
			to = length
		}
		if fv, ok := frm.(int); ok == true {
			if fv < 0 {
				_frm = length + fv
			} else {
				_frm = fv
			}
		}
		if tv, ok := to.(int); ok == true {
			if tv < 0 {
				_to = length + tv + 1
			} else {
				_to = tv
			}
		}
		if _frm < 0 || _frm >= length {
			return nil, fmt.Errorf("index [from] out of range: len: %v, from: %v", length, frm)
		}
		// Clamp _to to valid range [0, length] per RFC 9535
		if _to < 0 {
			_to = 0
		}
		if _to > length {
			_to = length
		}
		//fmt.Println("_frm, _to: ", _frm, _to)
		res_v := reflect.ValueOf(obj).Slice(_frm, _to)
		return res_v.Interface(), nil
	case reflect.Map:
		// For wildcard [*] on maps, return all values
		var res []interface{}
		if jsonMap, ok := obj.(map[string]interface{}); ok {
			for _, v := range jsonMap {
				res = append(res, v)
			}
			return res, nil
		}
		keys := reflect.ValueOf(obj).MapKeys()
		for _, k := range keys {
			res = append(res, reflect.ValueOf(obj).MapIndex(k).Interface())
		}
		return res, nil
	default:
		return nil, fmt.Errorf("object is not Slice")
	}
}

func regFilterCompile(rule string) (*regexp.Regexp, error) {
	runes := []rune(rule)
	if len(runes) <= 2 {
		return nil, errors.New("empty rule")
	}

	if runes[0] != '/' || runes[len(runes)-1] != '/' {
		return nil, errors.New("invalid syntax. should be in `/pattern/` form")
	}
	runes = runes[1 : len(runes)-1]
	return regexp.Compile(string(runes))
}

func get_filtered(obj, root interface{}, filter string) ([]interface{}, error) {
	lp, op, rp, err := parse_filter(filter)
	if err != nil {
		return nil, err
	}

	res := []interface{}{}

	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice:
		if op == "=~" {
			// regexp
			pat, err := regFilterCompile(rp)
			if err != nil {
				return nil, err
			}

			for i := 0; i < reflect.ValueOf(obj).Len(); i++ {
				tmp := reflect.ValueOf(obj).Index(i).Interface()
				ok, err := eval_reg_filter(tmp, root, lp, pat)
				if err != nil {
					return nil, err
				}
				if ok == true {
					res = append(res, tmp)
				}
			}
		} else {
			for i := 0; i < reflect.ValueOf(obj).Len(); i++ {
				tmp := reflect.ValueOf(obj).Index(i).Interface()
				ok, err := eval_filter(tmp, root, lp, op, rp)
				if err != nil {
					return nil, err
				}
				if ok == true {
					res = append(res, tmp)
				}
			}
		}
		return res, nil
	case reflect.Map:
		if op == "=~" {
			// regexp
			pat, err := regFilterCompile(rp)
			if err != nil {
				return nil, err
			}

			for _, kv := range reflect.ValueOf(obj).MapKeys() {
				tmp := reflect.ValueOf(obj).MapIndex(kv).Interface()
				ok, err := eval_reg_filter(tmp, root, lp, pat)
				if err != nil {
					return nil, err
				}
				if ok == true {
					res = append(res, tmp)
				}
			}
		} else {
			for _, kv := range reflect.ValueOf(obj).MapKeys() {
				tmp := reflect.ValueOf(obj).MapIndex(kv).Interface()
				ok, err := eval_filter(tmp, root, lp, op, rp)
				if err != nil {
					return nil, err
				}
				if ok == true {
					res = append(res, tmp)
				}
			}
		}
	default:
		return nil, fmt.Errorf("don't support filter on this type: %v", reflect.TypeOf(obj).Kind())
	}

	return res, nil
}

func get_scan(obj interface{}) (interface{}, error) {
	if reflect.TypeOf(obj) == nil {
		return nil, nil
	}
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Map:
		// iterate over keys in sorted by length, then alphabetically
		var res []interface{}
		if jsonMap, ok := obj.(map[string]interface{}); ok {
			var sortedKeys []string
			for k := range jsonMap {
				sortedKeys = append(sortedKeys, k)
			}
			sort.Slice(sortedKeys, func(i, j int) bool {
				if len(sortedKeys[i]) != len(sortedKeys[j]) {
					return len(sortedKeys[i]) < len(sortedKeys[j])
				}
				return sortedKeys[i] < sortedKeys[j]
			})
			for _, k := range sortedKeys {
				res = append(res, jsonMap[k])
			}
			return res, nil
		}
		keys := reflect.ValueOf(obj).MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			ki, kj := keys[i].String(), keys[j].String()
			if len(ki) != len(kj) {
				return len(ki) < len(kj)
			}
			return ki < kj
		})
		for _, k := range keys {
			res = append(res, reflect.ValueOf(obj).MapIndex(k).Interface())
		}
		return res, nil
	case reflect.Slice:
		// slice we should get from all objects in it.
		var res []interface{}
		for i := 0; i < reflect.ValueOf(obj).Len(); i++ {
			tmp := reflect.ValueOf(obj).Index(i).Interface()
			newObj, err := get_scan(tmp)
			if err != nil {
				return nil, err
			}
			res = append(res, newObj.([]interface{})...)
		}
		return res, nil
	default:
		return nil, fmt.Errorf("object is not scannable: %v", reflect.TypeOf(obj).Kind())
	}
}

// @.isbn                 => @.isbn, exists, nil
// @.price < 10           => @.price, <, 10
// @.price <= $.expensive => @.price, <=, $.expensive
// @.author =~ /.*REES/i  => @.author, match, /.*REES/i
// count(@.book) > 0      => count(@.book), >, 0
// match(@.author, 'REES') => match(@.author, 'REES'), exists, nil
func parse_filter(filter string) (lp string, op string, rp string, err error) {
	tmp := ""

	stage := 0
	quoteChar := rune(0)
	parenDepth := 0
	for idx, c := range filter {
		switch c {
		case '\'':
			if quoteChar == 0 {
				quoteChar = c
			} else if c == quoteChar {
				quoteChar = 0
			} else {
				tmp += string(c)
			}
			continue
		case '"':
			if quoteChar == 0 {
				quoteChar = c
			} else if c == quoteChar {
				quoteChar = 0
			} else {
				tmp += string(c)
			}
			continue
		case '(':
			if quoteChar == 0 {
				parenDepth++
			}
			tmp += string(c)
			continue
		case ')':
			if quoteChar == 0 {
				parenDepth--
			}
			tmp += string(c)
			continue
		case ' ':
			if quoteChar != 0 || parenDepth > 0 {
				// Inside quotes or parentheses, keep the space
				tmp += string(c)
				continue
			}
			if tmp != "" {
				switch stage {
				case 0:
					lp = tmp
				case 1:
					op = tmp
				case 2:
					rp = tmp
				}
				tmp = ""
				stage++
				if stage > 2 {
					return "", "", "", errors.New(fmt.Sprintf("invalid char at %d: `%c`", idx, c))
				}
			}
		default:
			tmp += string(c)
		}
	}
	if tmp != "" {
		switch stage {
		case 0:
			lp = tmp
			if strings.HasSuffix(lp, ")") || stage == 0 {
				// Function call without operator, or simple expression without operator
				// set exists operator
				op = "exists"
			}
		case 1:
			op = tmp
		case 2:
			rp = tmp
		}
	}
	return lp, op, rp, err
}

// func parse_filter_v1(filter string) (lp string, op string, rp string, err error) {
// 	tmp := ""
// 	istoken := false
// 	for _, c := range filter {
// 		if istoken == false && c != ' ' {
// 			istoken = true
// 		}
// 		if istoken == true && c == ' ' {
// 			istoken = false
// 		}
// 		if istoken == true {
// 			tmp += string(c)
// 		}
// 		if istoken == false && tmp != "" {
// 			if lp == "" {
// 				lp = tmp[:]
// 				tmp = ""
// 			} else if op == "" {
// 				op = tmp[:]
// 				tmp = ""
// 			} else if rp == "" {
// 				rp = tmp[:]
// 				tmp = ""
// 			}
// 		}
// 	}
// 	if tmp != "" && lp == "" && op == "" && rp == "" {
// 		lp = tmp[:]
// 		op = "exists"
// 		rp = ""
// 		err = nil
// 		return
// 	} else if tmp != "" && rp == "" {
// 		rp = tmp[:]
// 		tmp = ""
// 	}
// 	return lp, op, rp, err
// }

func eval_reg_filter(obj, root interface{}, lp string, pat *regexp.Regexp) (res bool, err error) {
	if pat == nil {
		return false, errors.New("nil pat")
	}
	lp_v, err := get_lp_v(obj, root, lp)
	if err != nil {
		return false, err
	}
	switch v := lp_v.(type) {
	case string:
		return pat.MatchString(v), nil
	default:
		return false, errors.New("only string can match with regular expression")
	}
}

func get_lp_v(obj, root interface{}, lp string) (interface{}, error) {
	// Check if lp is a function call like count(@.xxx) or match(@.xxx, pattern)
	if strings.HasSuffix(lp, ")") {
		return eval_filter_func(obj, root, lp)
	}

	var lp_v interface{}
	if strings.HasPrefix(lp, "@.") {
		return filter_get_from_explicit_path(obj, lp)
	} else if strings.HasPrefix(lp, "$.") {
		return filter_get_from_explicit_path(root, lp)
	} else {
		lp_v = lp
	}
	return lp_v, nil
}

// eval_filter_func evaluates function calls in filter expressions
func eval_filter_func(obj, root interface{}, expr string) (interface{}, error) {
	// Find the first ( that starts the function arguments
	parenIdx := -1
	for i, c := range expr {
		if c == '(' {
			parenIdx = i
			break
		}
	}

	if parenIdx < 0 {
		return nil, fmt.Errorf("invalid function call: %s", expr)
	}

	funcName := strings.TrimSpace(expr[:parenIdx])

	// Find the matching closing parenthesis
	argsStart := parenIdx + 1
	argsEnd := -1
	depth := 1
	for i := argsStart; i < len(expr); i++ {
		if expr[i] == '(' {
			depth++
		} else if expr[i] == ')' {
			depth--
			if depth == 0 {
				argsEnd = i
				break
			}
		}
	}

	if argsEnd < 0 {
		return nil, fmt.Errorf("mismatched parentheses in function call: %s", expr)
	}

	argsStr := expr[argsStart:argsEnd]

	// Split arguments by comma (respecting nested parentheses and quotes)
	var args []string
	current := ""
	argDepth := 0
	quoteChar := rune(0)
	for _, c := range argsStr {
		if quoteChar != 0 {
			if c == quoteChar {
				quoteChar = 0
			}
			current += string(c)
			continue
		}
		if c == '"' || c == '\'' {
			quoteChar = c
			current += string(c)
			continue
		}
		if c == '(' {
			argDepth++
		} else if c == ')' {
			argDepth--
		} else if c == ',' && argDepth == 0 {
			args = append(args, strings.TrimSpace(current))
			current = ""
			continue
		}
		current += string(c)
	}
	if current != "" {
		args = append(args, strings.TrimSpace(current))
	}

	// Evaluate function based on name
	switch funcName {
	case "count":
		return eval_count(obj, root, args)
	case "match":
		return eval_match(obj, root, args)
	case "search":
		return eval_search(obj, root, args)
	case "length":
		return eval_length(obj, root, args)
	default:
		return nil, fmt.Errorf("unsupported function: %s()", funcName)
	}
}

// eval_count evaluates count() function - returns the count of nodes in a nodelist
func eval_count(obj, root interface{}, args []string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("count() requires 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Special case: count(@) or count('') returns the length of the root array
	if arg == "@" || arg == "" {
		// Use root to get the array length
		if root == nil {
			return 0, nil
		}
		rv := reflect.ValueOf(root)
		switch rv.Kind() {
		case reflect.Array, reflect.Slice:
			return rv.Len(), nil
		default:
			// Root is not an array, count as 1
			return 1, nil
		}
	}

	var nodeset interface{}
	if strings.HasPrefix(arg, "@.") {
		nodeset, _ = filter_get_from_explicit_path(obj, arg)
	} else if strings.HasPrefix(arg, "$.") {
		nodeset, _ = filter_get_from_explicit_path(root, arg)
	} else {
		// Literal string - treat as string length
		return len(arg), nil
	}

	// Count nodes in the nodelist
	if nodeset == nil {
		return 0, nil
	}
	switch v := nodeset.(type) {
	case []interface{}:
		return len(v), nil
	default:
		// Single node, count as 1
		return 1, nil
	}
}

// eval_match evaluates match() function - regex with implicit anchoring (^pattern$)
func eval_match(obj, root interface{}, args []string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("match() requires 2 arguments (string, pattern), got %d", len(args))
	}

	// Get the string value
	var strVal string
	if strings.HasPrefix(args[0], "@.") {
		v, err := filter_get_from_explicit_path(obj, args[0])
		if err != nil {
			return nil, err
		}
		if v == nil {
			return false, nil
		}
		strVal = fmt.Sprintf("%v", v)
	} else if strings.HasPrefix(args[0], "$.") {
		v, err := filter_get_from_explicit_path(root, args[0])
		if err != nil {
			return nil, err
		}
		if v == nil {
			return false, nil
		}
		strVal = fmt.Sprintf("%v", v)
	} else {
		strVal = args[0]
	}

	// Get the pattern (remove quotes if present)
	pattern := args[1]
	pattern = strings.Trim(pattern, `"'`)

	// Compile regex with implicit anchoring (^pattern$)
	re, err := regexp.Compile("^" + pattern + "$")
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %v", err)
	}

	return re.MatchString(strVal), nil
}

// eval_search evaluates search() function - regex without anchoring
func eval_search(obj, root interface{}, args []string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("search() requires 2 arguments (string, pattern), got %d", len(args))
	}

	// Get the string value
	var strVal string
	if strings.HasPrefix(args[0], "@.") {
		v, err := filter_get_from_explicit_path(obj, args[0])
		if err != nil {
			return nil, err
		}
		if v == nil {
			return false, nil
		}
		strVal = fmt.Sprintf("%v", v)
	} else if strings.HasPrefix(args[0], "$.") {
		v, err := filter_get_from_explicit_path(root, args[0])
		if err != nil {
			return nil, err
		}
		if v == nil {
			return false, nil
		}
		strVal = fmt.Sprintf("%v", v)
	} else {
		strVal = args[0]
	}

	// Get the pattern (remove quotes if present)
	pattern := args[1]
	pattern = strings.Trim(pattern, `"'`)

	// Compile regex without anchoring
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %v", err)
	}

	return re.MatchString(strVal), nil
}

// eval_length evaluates length() function in filter context
func eval_length(obj, root interface{}, args []string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("length() requires 1 argument, got %d", len(args))
	}

	var val interface{}
	if strings.HasPrefix(args[0], "@.") {
		val, _ = filter_get_from_explicit_path(obj, args[0])
	} else if strings.HasPrefix(args[0], "$.") {
		val, _ = filter_get_from_explicit_path(root, args[0])
	} else {
		val = args[0]
	}

	return get_length(val)
}

func eval_filter(obj, root interface{}, lp, op, rp string) (res bool, err error) {
	lp_v, err := get_lp_v(obj, root, lp)

	// If op is empty, treat it as an exists check (truthy check)
	if op == "" {
		op = "exists"
	}

	if op == "exists" {
		// If lp_v is a function call (contains parentheses), evaluate it
		// and return the boolean result
		if strings.HasSuffix(lp, ")") {
			// It's a function call, get_lp_v should have evaluated it
			// and returned the result (which could be bool, int, etc.)
			switch v := lp_v.(type) {
			case bool:
				return v, nil
			case int, int8, int16, int32, int64, float32, float64:
				// Non-zero values are truthy
				return v != 0, nil
			default:
				// For other types, check if not nil
				return lp_v != nil, nil
			}
		}
		return lp_v != nil, nil
	} else if op == "=~" {
		return false, fmt.Errorf("not implemented yet")
	} else {
		var rp_v interface{}
		if strings.HasPrefix(rp, "@.") {
			rp_v, err = filter_get_from_explicit_path(obj, rp)
		} else if strings.HasPrefix(rp, "$.") {
			rp_v, err = filter_get_from_explicit_path(root, rp)
		} else {
			rp_v = rp
		}
		//fmt.Printf("lp_v: %v, rp_v: %v\n", lp_v, rp_v)
		return cmp_any(lp_v, rp_v, op)
	}
}

// eval_func evaluates function calls like length()
func eval_func(obj interface{}, funcName string) (interface{}, error) {
	switch funcName {
	case "length":
		return get_length(obj)
	default:
		return nil, fmt.Errorf("unsupported function: %s()", funcName)
	}
}

// get_length returns the length of an array, string, or map
func get_length(obj interface{}) (interface{}, error) {
	if obj == nil {
		return nil, nil
	}
	switch v := obj.(type) {
	case []interface{}:
		return len(v), nil
	case string:
		return len(v), nil
	case map[string]interface{}:
		return len(v), nil
	default:
		// Try to use reflection for other types
		rv := reflect.ValueOf(obj)
		switch rv.Kind() {
		case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
			return rv.Len(), nil
		default:
			return nil, fmt.Errorf("length() not supported for type: %T", obj)
		}
	}
}

func isNumber(o interface{}) bool {
	switch v := o.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	case string:
		_, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return true
		} else {
			return false
		}
	}
	return false
}

func cmp_any(obj1, obj2 interface{}, op string) (bool, error) {
	switch op {
	case "<", "<=", "==", ">=", ">":
	default:
		return false, fmt.Errorf("op should only be <, <=, ==, >= and >")
	}

	var exp string
	if isNumber(obj1) && isNumber(obj2) {
		exp = fmt.Sprintf(`%v %s %v`, obj1, op, obj2)
	} else {
		exp = fmt.Sprintf(`"%v" %s "%v"`, obj1, op, obj2)
	}
	//fmt.Println("exp: ", exp)
	fset := token.NewFileSet()
	res, err := types.Eval(fset, nil, 0, exp)
	if err != nil {
		return false, err
	}
	if res.IsValue() == false || (res.Value.String() != "false" && res.Value.String() != "true") {
		return false, fmt.Errorf("result should only be true or false")
	}
	if res.Value.String() == "true" {
		return true, nil
	}

	return false, nil
}

func getAllDescendants(obj interface{}) []interface{} {
	res := []interface{}{}
	var recurse func(curr interface{})
	recurse = func(curr interface{}) {
		res = append(res, curr)
		v := reflect.ValueOf(curr)
		if !v.IsValid() {
			return
		}

		kind := v.Kind()
		if kind == reflect.Ptr {
			v = v.Elem()
			if !v.IsValid() {
				return
			}
			kind = v.Kind()
		}

		switch kind {
		case reflect.Map:
			for _, k := range v.MapKeys() {
				recurse(v.MapIndex(k).Interface())
			}
		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				recurse(v.Index(i).Interface())
			}
		}
	}
	recurse(obj)
	return res
}
