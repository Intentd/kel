package core

import (
	"fmt"
	"sort"
	"text/template"

	"github.com/intentd/kel/internal/anymap"
)

// listMapFuncs — Collections
//
// Domain: collections
//
// Build lists and maps, and read them without guarding every step: make a
// list or a map, get an element by index or a value by key, test membership, take keys
// and values, or keep and drop keys. Out-of-range and missing keys give zero values, not errors.
func listMapFuncs() template.FuncMap {
	return template.FuncMap{
		"list":      List,
		"map":       Map,
		"listHas":   ListHas,
		"listGet":   ListGet,
		"listLen":   ListLen,
		"mapGet":    MapGet,
		"mapKeys":   MapKeys,
		"mapValues": MapValues,
		"mapPick":   MapPick,
		"mapOmit":   MapOmit,
	}
}

// List returns its arguments as a list, for functions that take several
// values in one argument, such as the failed values of expectState. With no
// arguments it returns an empty list.
//
// Keywords: list, collection, build, make, slice, array, values, literal
//
// Example:
//
//	{{ list "Failed" "Unknown" }}  → [Failed Unknown]
//	{{ listLen (list 1 2 3) }}  → 3
func List(items ...interface{}) []interface{} {
	if items == nil {
		return []interface{}{}
	}
	return items
}

// Map returns a map from alternating keys and values, the counterpart of list.
// Keys that are not strings are written as text. With no arguments it returns
// an empty map. An odd number of arguments is an error: a key without a value.
//
// Keywords: map, dict, collection, build, make, object, keys, values, literal
//
// Example:
//
//	{{ mapGet (map "region" "eu-west-1" "replicas" 3) "replicas" }}  → 3
//	{{ mapKeys (map "b" 2 "a" 1) }}  → [a b]
func Map(pairs ...interface{}) (map[string]interface{}, error) {
	pairsCount := len(pairs)
	if (pairsCount % 2) != 0 {
		return nil, fmt.Errorf("map: %d arguments, want key and value pairs", pairsCount)
	}
	m := make(map[string]interface{}, pairsCount/2)
	for i := 0; i < pairsCount; i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			key = fmt.Sprint(pairs[i])
		}
		m[key] = pairs[i+1]
	}
	return m, nil
}

// ListHas reports whether a list contains a value. Comparison is by `==`.
//
// Keywords: list, collection, contains, check, slice, boolean, membership
//
// Example:
//
//	{{ listHas .spec.regions "eu-west-1" }}  → true
func ListHas(list interface{}, val interface{}) bool {
	items, ok := list.([]interface{})
	if !ok {
		return false
	}
	for _, v := range items {
		if v == val {
			return true
		}
	}
	return false
}

// ListGet returns the element at a given index. Returns `nil` for out-of-range
// indices — no panic.
//
// Keywords: list, collection, get, index, slice, access, element
//
// Example:
//
//	{{ listGet .spec.regions 0 }}  → "us-east-1"
//	{{ listGet .spec.regions 9 }}  → <no value>
func ListGet(list interface{}, index int) interface{} {
	items, ok := list.([]interface{})
	if !ok || index < 0 || index >= len(items) {
		return nil
	}
	return items[index]
}

// ListLen returns the number of elements in a list. Returns `0` for non-lists.
//
// Keywords: list, collection, length, count, slice, size
//
// Example:
//
//	{{ listLen .spec.regions }}  → 2
func ListLen(list interface{}) int {
	items, ok := list.([]interface{})
	if !ok {
		return 0
	}
	return len(items)
}

// MapGet returns a map value by key. Returns `nil` when the map is absent or
// the key is missing.
//
// Keywords: map, collection, get, key, access, lookup
//
// Example:
//
//	{{ mapGet .deployment.metadata.labels "app" }}      → "web"
//	{{ mapGet .deployment.metadata.labels "missing" }}  → <no value>
func MapGet(m interface{}, key string) interface{} {
	mp, ok := anymap.Of(m)
	if !ok {
		return nil
	}
	return mp[key]
}

// MapKeys returns all keys of a map as sorted `[]string`. Returns an empty
// slice for non-maps.
//
// Keywords: map, collection, keys, slice, iterate, list
//
// Example:
//
//	{{ mapKeys .spec.selector }}  → [app tier]
func MapKeys(m interface{}) []string {
	mp, ok := anymap.Of(m)
	if !ok {
		return []string{}
	}
	keys := make([]string, 0, len(mp))
	for k := range mp {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// MapValues returns all values of a map as `[]interface{}`. Returns an empty
// slice for non-maps.
//
// Keywords: map, collection, values, slice, iterate, list
//
// Example:
//
//	{{ len (mapValues .spec.selector) }}  → 2
func MapValues(m interface{}) []interface{} {
	mp, ok := anymap.Of(m)
	if !ok {
		return []interface{}{}
	}
	values := make([]interface{}, 0, len(mp))
	for _, v := range mp {
		values = append(values, v)
	}
	return values
}

// MapPick returns a new map containing only the specified keys. Missing keys
// are silently omitted.
//
// Keywords: map, pick
//
// Example:
//
//	{{ mapPick .spec "image" "replicas" }}
func MapPick(m interface{}, keys ...string) map[string]interface{} {
	mp, ok := anymap.Of(m)
	out := make(map[string]interface{}, len(keys))
	if !ok {
		return out
	}
	for _, k := range keys {
		if v, exists := mp[k]; exists {
			out[k] = v
		}
	}
	return out
}

// MapOmit returns a new map with the specified keys removed.
//
// Keywords: map, omit
//
// Example:
//
//	{{ mapOmit .metadata.labels "internal-key" "debug" }}
func MapOmit(m interface{}, keys ...string) map[string]interface{} {
	mp, ok := anymap.Of(m)
	if !ok {
		return map[string]interface{}{}
	}
	skip := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		skip[k] = struct{}{}
	}
	out := make(map[string]interface{}, len(mp))
	for k, v := range mp {
		if _, omit := skip[k]; !omit {
			out[k] = v
		}
	}
	return out
}
