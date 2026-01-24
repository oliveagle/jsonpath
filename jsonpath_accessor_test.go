// Copyright 2015, 2021; oliver, DoltHub Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package jsonpath

import "testing"

// Accessor tests - test get_key, get_idx, get_range, get_scan functions

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
