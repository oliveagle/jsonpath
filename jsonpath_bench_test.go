package jsonpath

import "testing"

// Benchmarks for JsonPath lookup performance

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
