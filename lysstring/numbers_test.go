package lysstring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinInts(t *testing.T) {

	type comboS struct {
		Name     string
		Input    []int
		Sep      string
		Expected string
	}

	combos := []comboS{
		{Name: "Empty slice", Input: []int{}, Sep: ",", Expected: ""},
		{Name: "Single element", Input: []int{1}, Sep: ",", Expected: "1"},
		{Name: "Multiple elements", Input: []int{1, 2, 3}, Sep: ",", Expected: "1,2,3"},
		{Name: "Different separator", Input: []int{1, 2, 3}, Sep: "-", Expected: "1-2-3"},
		{Name: "Negative numbers", Input: []int{-1, -2, -3}, Sep: ",", Expected: "-1,-2,-3"},
	}
	for _, c := range combos {
		res := JoinInts(c.Input, c.Sep)
		assert.Equal(t, c.Expected, res, c.Name)
	}
}
