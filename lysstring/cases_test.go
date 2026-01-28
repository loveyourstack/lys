package lysstring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCases(t *testing.T) {

	type comboS struct {
		Name     string
		Func     func(string) string
		Expected string
	}

	val := "My string"
	combos := []comboS{
		{Name: val + ": Camel", Func: Camel, Expected: "myString"},
		{Name: val + ": FirstLower", Func: FirstLower, Expected: "my string"},
		{Name: val + ": Joined", Func: Joined, Expected: "mystring"},
		{Name: val + ": Kebab", Func: Kebab, Expected: "my-string"},
		{Name: val + ": Pascal", Func: Pascal, Expected: "MyString"},
		{Name: val + ": Snake", Func: Snake, Expected: "my_string"},
	}
	for _, c := range combos {
		res := c.Func(val)
		assert.Equal(t, c.Expected, res, c.Name)
	}

	val = "my string"
	combos = []comboS{
		{Name: val + ": Title", Func: Title, Expected: "My String"},
	}
	for _, c := range combos {
		res := c.Func(val)
		assert.Equal(t, c.Expected, res, c.Name)
	}
}
