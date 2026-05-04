package lysstring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandLength(t *testing.T) {
	type comboS struct {
		Name  string
		Input int
	}

	combos := []comboS{
		{Name: "Zero", Input: 0},
		{Name: "Negative", Input: -1},
		{Name: "Small", Input: 1},
		{Name: "Medium", Input: 16},
		{Name: "Large", Input: 256},
	}

	for _, c := range combos {
		res := Rand(c.Input)
		expectedLen := max(c.Input, 0)
		assert.Equal(t, expectedLen, len(res), c.Name)
	}
}

func TestRandCharset(t *testing.T) {
	res := Rand(1024)
	for _, r := range res {
		assert.Contains(t, letterBytes, string(r), "unexpected rune")
	}
}

func TestRandRandomnessSanity(t *testing.T) {
	a := Rand(32)
	b := Rand(32)

	assert.NotEqual(t, "", a)
	assert.NotEqual(t, "", b)
	assert.NotEqual(t, a, b, "two consecutive random values should almost always differ")
}
