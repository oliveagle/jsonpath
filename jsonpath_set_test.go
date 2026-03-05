// Copyright 2025 oliveagle
// Use of this source code is governed by an MIT-style
// license that can be found in LICENSE file.

package jsonpath

import "testing"

func TestJsonPathSet_SimpleKey(t *testing.T) {
	obj := map[string]interface{}{
		"numbers": []interface{}{1, 2, 3},
		"other": "value",
	}

	c := MustCompile("$.numbers[2]")
	result, err := c.Set(obj, 99)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	numbers := resultMap["numbers"].([]interface{})

	if len(numbers) != 3 || numbers[2] != 99 {
		t.Errorf("numbers should be [1, 99, 3], got %v", numbers)
	}

	if resultMap["other"] != "value" {
		t.Error("other key should be preserved")
	}
}

func TestJsonPathSet_CompiledSet(t *testing.T) {
	obj := map[string]interface{}{"name": "John"}

	compiled := MustCompile("$.name")
	result, err := compiled.Set(obj, "Jane")
	if err != nil {
		t.Fatalf("Compiled.Set failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if resultMap["name"] != "Jane" {
		t.Errorf("Expected name to be 'Jane', got '%v'", resultMap["name"])
	}

	if obj["name"] != "John" {
		t.Error("Original object was modified")
	}
}

func TestJsonPathSet_NestedKey(t *testing.T) {
	obj := map[string]interface{}{
		"store": map[string]interface{}{
			"book": map[string]interface{}{"title": "Old Title"},
		},
	}

	result, err := JsonPathSet(obj, "$.store.book.title", "New Title")
	if err != nil {
		t.Fatalf("JsonPathSet failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	store := resultMap["store"].(map[string]interface{})
	book := store["book"].(map[string]interface{})
	if book["title"] != "New Title" {
		t.Errorf("Expected title to be 'New Title', got '%v'", book["title"])
	}
}

func TestJsonPathSet_DeepCopy(t *testing.T) {
	original := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{"value": "original"},
		},
	}

	result, err := JsonPathSet(original, "$.level1.level2.value", "modified")
	if err != nil {
		t.Fatalf("JsonPathSet failed: %v", err)
	}

	origLevel1 := original["level1"].(map[string]interface{})
	origLevel2 := origLevel1["level2"].(map[string]interface{})
	if origLevel2["value"] != "original" {
		t.Error("Original nested map was modified")
	}

	resultMap := result.(map[string]interface{})
	resultLevel1 := resultMap["level1"].(map[string]interface{})
	resultLevel2 := resultLevel1["level2"].(map[string]interface{})
	if resultLevel2["value"] != "modified" {
		t.Error("Result nested map was not modified")
	}
}

func TestJsonPathSet_EmptyPath(t *testing.T) {
	obj := map[string]interface{}{"name": "John"}
	compiled := &Compiled{path: "", steps: []step{}}
	_, err := compiled.Set(obj, "value")
	if err == nil {
		t.Error("Expected error for empty path")
	}
}

func TestJsonPathSet_NullObject(t *testing.T) {
	_, err := JsonPathSet(nil, "$.name", "value")
	if err == nil {
		t.Error("Expected error for nil object")
	}
}

func TestJsonPathSet_IndexOutOfRange(t *testing.T) {
	obj := map[string]interface{}{"numbers": []interface{}{1, 2, 3}}
	_, err := JsonPathSet(obj, "$.numbers[10]", 99)
	if err == nil {
		t.Error("Expected error for index out of range")
	}
}

func TestJsonPathSet_NegativeIndexOutOfRange(t *testing.T) {
	obj := map[string]interface{}{"numbers": []interface{}{1, 2, 3}}
	_, err := JsonPathSet(obj, "$.numbers[-10]", 99)
	if err == nil {
		t.Error("Expected error for negative index out of range")
	}
}

func TestJsonPathSet_KeyNotFound(t *testing.T) {
	obj := map[string]interface{}{"name": "John"}
	// Note: JsonPathSet creates non-existent keys (loose mode)
	result, err := JsonPathSet(obj, "$.age", 30)
	if err != nil {
		t.Errorf("Expected no error for creating new key, got %v", err)
	}
	resultMap := result.(map[string]interface{})
	if resultMap["age"] != 30 {
		t.Errorf("Expected age to be 30, got %v", resultMap["age"])
	}
}

func TestJsonPathSet_InvalidOperation(t *testing.T) {
	obj := map[string]interface{}{"data": []interface{}{1, 2, 3}}
	_, err := JsonPathSet(obj, "$.data[?(@ > 1)]", 99)
	if err == nil {
		t.Error("Expected error for unsupported filter operation")
	}
}

func TestJsonPathSet_NonMapKey(t *testing.T) {
	obj := map[string]interface{}{"name": "John"}
	_, err := JsonPathSet(obj, "$.name.invalid", "value")
	if err == nil {
		t.Error("Expected error for setting key on non-map type")
	}
}

func TestJsonPathSet_NonSliceIndex(t *testing.T) {
	obj := map[string]interface{}{"name": "John"}
	_, err := JsonPathSet(obj, "$.name[0]", "value")
	if err == nil {
		t.Error("Expected error for indexing non-slice type")
	}
}

func TestJsonPathSet_NilMap(t *testing.T) {
	var obj map[string]interface{}
	_, err := JsonPathSet(obj, "$.name", "value")
	if err == nil {
		t.Error("Expected error for nil map")
	}
}

func TestJsonPathSet_NilSlice(t *testing.T) {
	obj := map[string]interface{}{"data": ([]interface{})(nil)}
	_, err := JsonPathSet(obj, "$.data[0]", "value")
	if err == nil {
		t.Error("Expected error for nil slice")
	}
}

func TestJsonPathSet_Range(t *testing.T) {
	obj := map[string]interface{}{"numbers": []interface{}{1, 2, 3, 4, 5}}
	_, err := JsonPathSet(obj, "$.numbers[1:3]", 0)
	if err != nil {
		t.Errorf("Expected no error for range, got %v", err)
	}
}

func TestJsonPathSet_StructField(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	obj := TestStruct{Name: "John", Value: 10}
	result, err := JsonPathSet(obj, "$.name", "Jane")
	if err != nil {
		t.Fatalf("JsonPathSet failed: %v", err)
	}

	resultStruct := result.(TestStruct)
	if resultStruct.Name != "Jane" {
		t.Errorf("Expected Name to be 'Jane', got '%v'", resultStruct.Name)
	}

	if obj.Name != "John" {
		t.Error("Original struct was modified")
	}
}

func TestJsonPathSet_StructJsonTag(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	obj := TestStruct{Name: "John", Value: 10}
	result, err := JsonPathSet(obj, "$.value", 99)
	if err != nil {
		t.Fatalf("JsonPathSet failed: %v", err)
	}

	resultStruct := result.(TestStruct)
	if resultStruct.Value != 99 {
		t.Errorf("Expected Value to be 99, got %v", resultStruct.Value)
	}
}

func TestJsonPathSet_NestedStruct(t *testing.T) {
	type Nested struct{ Value string }
	type Parent struct{ Child Nested }

	obj := Parent{Child: Nested{Value: "original"}}
	result, err := JsonPathSet(obj, "$.Child.Value", "modified")
	if err != nil {
		t.Fatalf("JsonPathSet failed: %v", err)
	}

	resultStruct := result.(Parent)
	if resultStruct.Child.Value != "modified" {
		t.Errorf("Expected nested value to be 'modified', got '%v'", resultStruct.Child.Value)
	}
}

func TestJsonPathSet_StructFieldNotFound(t *testing.T) {
	type TestStruct struct {
		Name string
	}
	obj := TestStruct{Name: "John"}
	_, err := JsonPathSet(obj, "$.invalid", "value")
	if err == nil {
		t.Error("Expected error for field not found in struct")
	}
}

func TestJsonPathSet_RangeNegativeFrom(t *testing.T) {
	obj := map[string]interface{}{"numbers": []interface{}{1, 2, 3, 4, 5}}
	_, err := JsonPathSet(obj, "$.numbers[-2:]", 0)
	if err != nil {
		t.Errorf("Expected no error for negative range, got %v", err)
	}
}

func TestJsonPathSet_RangeNegativeTo(t *testing.T) {
	obj := map[string]interface{}{"numbers": []interface{}{1, 2, 3, 4, 5}}
	_, err := JsonPathSet(obj, "$.numbers[:-2]", 0)
	if err != nil {
		t.Errorf("Expected no error for negative to range, got %v", err)
	}
}

func TestJsonPathSet_RangeClamped(t *testing.T) {
	obj := map[string]interface{}{"numbers": []interface{}{1, 2, 3}}
	_, err := JsonPathSet(obj, "$.numbers[1:100]", 0)
	if err != nil {
		t.Errorf("Expected no error for clamped range, got %v", err)
	}
}

func TestJsonPathSet_WildcardOnMap(t *testing.T) {
	obj := map[string]interface{}{
		"data": map[string]interface{}{"a": 1, "b": 2},
	}
	_, err := JsonPathSet(obj, "$.data[*]", 0)
	if err == nil {
		t.Error("Expected error for wildcard on map")
	}
}

func TestDeepCopy_Nil(t *testing.T) {
	result := deepCopy(nil)
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestDeepCopy_Primitive(t *testing.T) {
	result := deepCopy(42)
	if result != 42 {
		t.Errorf("Expected 42, got %v", result)
	}
}

func TestDeepCopy_String(t *testing.T) {
	result := deepCopy("hello")
	if result != "hello" {
		t.Errorf("Expected 'hello', got %v", result)
	}
}

func TestDeepCopy_Bool(t *testing.T) {
	result := deepCopy(true)
	if result != true {
		t.Errorf("Expected true, got %v", result)
	}
}

func TestSet_recursive_UnsupportedOp(t *testing.T) {
	steps := []step{{op: "recursive", key: "..", args: nil}}
	_, err := set_recursive(map[string]interface{}{}, steps, 0, "value")
	if err == nil {
		t.Error("Expected error for unsupported operation")
	}
}
