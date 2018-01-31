package jsonpath

import (
	"fmt"
	"github.com/mohae/utilitybelt/deepcopy"
	//"golang.org/x/tools/go/types"
	"go/token"
	"go/types"
	"reflect"
	"strconv"
	"strings"
	"errors"
)

var ErrGetFromNullObj = errors.New("get attribute from null object")

func JsonPathLookup(obj interface{}, jpath string) (interface{}, error) {
	steps, err := tokenize(jpath)
	//fmt.Println("f: steps: ", steps, err)
	//fmt.Println(jpath, steps)
	if err != nil {
		return nil, err
	}
	if steps[0] != "@" && steps[0] != "$" {
		return nil, fmt.Errorf("$ or @ should in front of path")
	}
	steps = steps[1:]
	xobj := deepcopy.Iface(obj)
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
			//fmt.Println("idx ----------------1")
			xobj, err = get_key(xobj, key)
			if err != nil {
				return nil, err
			}

			if len(args.([]int)) > 1 {
				//fmt.Println("idx ----------------2")
				res := []interface{}{}
				for _, x := range args.([]int) {
					//fmt.Println("idx ---- ", x)
					tmp, err := get_idx(xobj, x)
					if err != nil {
						return nil, err
					}
					res = append(res, tmp)
				}
				xobj = res
			} else if len(args.([]int)) == 1 {
				//fmt.Println("idx ----------------3")
				xobj, err = get_idx(xobj, args.([]int)[0])
				if err != nil {
					return nil, err
				}
			} else {
				//fmt.Println("idx ----------------4")
				return nil, fmt.Errorf("cannot index on empty slice")
			}
		case "range":
			xobj, err = get_key(xobj, key)
			if err != nil {
				return nil, err
			}
			if argsv, ok := args.([2]interface{}); ok == true {
				xobj, err = get_range(xobj, argsv[0], argsv[1])
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("range args length should be 2")
			}
		case "filter":
			xobj, err = get_key(xobj, key)
			if err != nil {
				return nil, err
			}
			xobj, err = get_filtered(xobj, obj, args.(string))
		default:
			return nil, fmt.Errorf("expression don't support in filter")
		}
	}
	return xobj, nil
}

func tokenize(query string) ([]string, error) {
	tokens := []string{}
	//	token_start := false
	//	token_end := false
	token := ""

	// fmt.Println("-------------------------------------------------- start")
	for idx, x := range query {
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
			if tokens[len(tokens)-1] != "*" {
				tokens = append(tokens, "*")
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
	if len(token) > 0 {
		if token[0] == '.' {
			token = token[1:]
			if token != "*" {
				tokens = append(tokens, token[:])
			} else if tokens[len(tokens)-1] != "*" {
				tokens = append(tokens, token[:])
			}
		} else {
			if token != "*" {
				tokens = append(tokens, token[:])
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

	bracket_idx := strings.Index(token, "[")
	if bracket_idx < 0 {
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
		if strings.Contains(tail, "?") {
			// filter -------------------------------------------------
			op = "filter"
			if strings.HasPrefix(tail, "?(") && strings.HasSuffix(tail, ")") {
				args = strings.Trim(tail[2:len(tail)-1], " ")
			}
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
				frm = nil
			}
			if to, err = strconv.Atoi(strings.Trim(tails[1], " ")); err != nil {
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
		default:
			return nil, fmt.Errorf("expression don't support in filter")
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
		for _, kv := range reflect.ValueOf(obj).MapKeys() {
			//fmt.Println(kv.String())
			if kv.String() == key {
				return reflect.ValueOf(obj).MapIndex(kv).Interface(), nil
			}
		}
		return nil, fmt.Errorf("key error: %s not found in object", key)
	case reflect.Slice:
		// slice we should get from all objects in it.
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
				_to = tv + 1
			}
		}
		if _frm < 0 || _frm >= length {
			return nil, fmt.Errorf("index [from] out of range: len: %v, from: %v", length, frm)
		}
		if _to < 0 || _to > length {
			return nil, fmt.Errorf("index [to] out of range: len: %v, to: %v", length, to)
		}
		//fmt.Println("_frm, _to: ", _frm, _to)
		res_v := reflect.ValueOf(obj).Slice(_frm, _to)
		return res_v.Interface(), nil
	default:
		return nil, fmt.Errorf("object is not Slice")
	}
}

func get_filtered(obj, root interface{}, filter string) ([]interface{}, error) {
	lp, op, rp, err := parse_filter(filter)
	if err != nil {
		return nil, err
	}

	res := []interface{}{}

	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice:
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
		return res, nil
	case reflect.Map:
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
	default:
		return nil, fmt.Errorf("don't support filter on this type: %v", reflect.TypeOf(obj).Kind())
	}

	return res, nil
}

// @.isbn                 => @.isbn, exists, nil
// @.price < 10           => @.price, <, 10
// @.price <= $.expensive => @.price, <=, $.expensive
// @.author =~ /.*REES/i  => @.author, match, /.*REES/i

func parse_filter(filter string) (lp string, op string, rp string, err error) {
	tmp := ""

	stage := 0
	str_embrace := false
	for idx, c := range filter {
		switch c {
		case '\'':
			if str_embrace == false {
				str_embrace = true
			} else {
				switch stage {
				case 0: lp = tmp
				case 1: op = tmp
				case 2: rp = tmp
				}
				tmp = ""
			}
		case ' ':
			if str_embrace == true {
				tmp += string(c)
				continue
			}
			switch stage {
			case 0: lp = tmp
			case 1: op = tmp
			case 2: rp = tmp
			}
			tmp = ""

			stage += 1
			if stage > 2 {
				return "", "", "", errors.New(fmt.Sprintf("invalid char at %d: `%s`", idx, c))
			}
		default:
			tmp += string(c)
		}
	}
	if tmp != "" {
		switch stage {
		case 0:
			lp = tmp
			op = "exists"
		case 1: op = tmp
		case 2: rp = tmp
		}
		tmp = ""
	}
	return lp, op, rp, err
}

func parse_filter_v1(filter string) (lp string, op string, rp string, err error) {
	tmp := ""
	istoken := false
	for _, c := range filter {
		if istoken == false && c != ' ' {
			istoken = true
		}
		if istoken == true && c == ' ' {
			istoken = false
		}
		if istoken == true {
			tmp += string(c)
		}
		if istoken == false && tmp != "" {
			if lp == "" {
				lp = tmp[:]
				tmp = ""
			} else if op == "" {
				op = tmp[:]
				tmp = ""
			} else if rp == "" {
				rp = tmp[:]
				tmp = ""
			}
		}
	}
	if tmp != "" && lp == "" && op == "" && rp == "" {
		lp = tmp[:]
		op = "exists"
		rp = ""
		err = nil
		return
	} else if tmp != "" && rp == "" {
		rp = tmp[:]
		tmp = ""
	}
	return lp, op, rp, err
}

func eval_filter(obj, root interface{}, lp, op, rp string) (res bool, err error) {
	var lp_v interface{}
	//fmt.Println(obj, root)
	//fmt.Printf("lp: %v, op: %v, rp: %v\n", lp, op, rp)
	if strings.HasPrefix(lp, "@.") {
		//fmt.Println("@. ----------------")
		lp_v, err = filter_get_from_explicit_path(obj, lp)
	} else if strings.HasPrefix(lp, "$.") {
		lp_v, err = filter_get_from_explicit_path(root, lp)
	} else {
		lp_v = lp
	}

	if op == "exists" {
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

func isNumber(s string) bool {
	dot_cnt := 0
	for _, c := range s {
		if c == '.' {
			dot_cnt += 1
			if dot_cnt > 1 {
				return false
			}
		} else if ( c >= '0' && c <= '9') {
			continue
		} else {
			return false
		}
	}
	return true
}

func cmp_any(obj1, obj2 interface{}, op string) (bool, error) {
	switch op {
	case "<", "<=", "==", ">=", ">":
	default:
		return false, fmt.Errorf("op should only be <, <=, ==, >= and >")
	}


	var exp string
	if isNumber(fmt.Sprintf("%s", obj1)) && isNumber(fmt.Sprintf("%s", obj2)) {
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
