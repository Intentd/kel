package anymap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type named map[string]interface{}

func TestOf(t *testing.T) {
	m := map[string]interface{}{"a": "1"}
	got, ok := Of(m)
	assert.True(t, ok)
	assert.Equal(t, m, got)

	got, ok = Of(named{"a": "1"})
	assert.True(t, ok, "a named map type converts")
	assert.Equal(t, m, got)

	for _, v := range []interface{}{nil, "x", map[string]string{"a": "1"}, []interface{}{}} {
		_, ok := Of(v)
		assert.False(t, ok, "%T", v)
	}
}

func TestStrings(t *testing.T) {
	assert.Equal(t, map[string]string{"app": "web"}, Strings(map[string]interface{}{"app": "web", "n": 4, "nil": nil}))
	assert.Equal(t, map[string]string{}, Strings(nil))
}
