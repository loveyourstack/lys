package lyspg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeReservedSuccess(t *testing.T) {

	s := []string{"a", "b"}
	expect := []string{"a", "b"}
	EscapeReserved(s)
	assert.Equal(t, expect, s, "no reserved")

	s = []string{"a", "table"}
	expect = []string{"a", `"table"`}
	EscapeReserved(s)
	assert.Equal(t, expect, s, "1 reserved")
}
