JsonPath
----------------

![Build Status](https://travis-ci.org/oliveagle/jsonpath.svg?branch=master)

A golang implementation of JsonPath syntax.
Follow the majority rules in http://goessner.net/articles/JsonPath/
but also with some minor differences.

This library is till bleeding edge, so use it at your own risk. :D

**Golang Version Required**: 1.15+

**Dependencies**: None! This library uses only Go standard library.

Get Started
------------

```bash
go get github.com/oliveagle/jsonpath
```

Example code:

```go
import (
    "github.com/oliveagle/jsonpath"
    "encoding/json"
)

var json_data interface{}
json.Unmarshal([]byte(data), &json_data)

res, err := jsonpath.JsonPathLookup(json_data, "$.expensive")

//or reuse lookup pattern
pat, _ := jsonpath.Compile(`$.store.book[?(@.price < $.expensive)].price`)
res, err := pat.Lookup(json_data)
```

Operators
--------
referenced from github.com/jayway/JsonPath

| Operator | Supported | Description |
| ---- | :---: | ---------- |
| `$` | Y | The root element to query. This starts all path expressions. |
| `@` | Y | The current node being processed by a filter predicate. |
| `*` | Y | Wildcard. Available anywhere a name or numeric are required. |
| `..` | Y | Deep scan. Available anywhere a name is required. |
| `.<name>` | Y | Dot-notated child |
| `['<name>' (, '<name>')]` | X | Bracket-notated child or children |
| `[<number> (, <number>)]` | Y | Array index or indexes |
| `[start:end]` | Y | Array slice operator (end is exclusive per RFC 9535) |
| `[?(<expression>)]` | Y | Filter expression. Expression must evaluate to a boolean value. |
| `length()` | Y | RFC 9535 function: returns length of array, string, or map |
| `count()` | Y | RFC 9535 function: returns count of items in array |
| `match()` | Y | RFC 9535 function: regex match with implicit anchoring (`^pattern$`) |
| `search()` | Y | RFC 9535 function: regex search without anchoring |

Examples
--------
Given these example data.

```javascript
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
```

Example json path syntax
----

| jsonpath | result |
| :--------- | :-------|
| `$.expensive` | 10 |
| `$.store.book[0].price` | 8.95 |
| `$.store.book[-1].isbn` | "0-395-19395-8" |
| `$.store.book[0,1].price` | [8.95, 12.99] |
| `$.store.book[0:2].price` | [8.95, 12.99] (slice end is exclusive) |
| `$.store.book[?(@.isbn)].price` | [8.99, 22.99] |
| `$.store.book[?(@.price > 10)].title` | ["Sword of Honour", "The Lord of the Rings"] |
| `$.store.book[?(@.price < $.expensive)].price` | [8.95, 8.99] |
| `$.store.book[:].price` | [8.95, 12.99, 8.99, 22.99] |
| `$.store.book[?(@.author =~ /(?i).*REES/)].author` | "Nigel Rees" |
| `$..author` | ["Nigel Rees", "Evelyn Waugh", "Herman Melville", "J. R. R. Tolkien"] |
| `$.store.book[*].price` | [8.95, 12.99, 8.99, 22.99] |

> Note: golang support regular expression flags in form of `(?imsU)pattern`
>
> RFC 9535 functions supported:
> - `length()` - returns length of array, string, or map
> - `count()` - returns count of items in array (used in filter expressions)
> - `match()` - regex match with implicit anchoring (`^pattern$`)
> - `search()` - regex search without anchoring
