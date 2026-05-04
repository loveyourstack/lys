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

	t.Run("Standard", func(t *testing.T) {
		val := "My string"
		combos := []comboS{
			{Name: val + ": Camel", Func: Camel, Expected: "myString"},
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
	})

	t.Run("Accents removed", func(t *testing.T) {
		val := "Accénted strïng"
		combos := []comboS{
			{Name: val + ": Camel", Func: Camel, Expected: "accentedString"},
			{Name: val + ": Joined", Func: Joined, Expected: "accentedstring"},
			{Name: val + ": Kebab", Func: Kebab, Expected: "accented-string"},
			{Name: val + ": Pascal", Func: Pascal, Expected: "AccentedString"},
			{Name: val + ": Snake", Func: Snake, Expected: "accented_string"},
		}
		for _, c := range combos {
			res := c.Func(val)
			assert.Equal(t, c.Expected, res, c.Name)
		}
	})

	t.Run("Whitespace removed", func(t *testing.T) {
		val := "  My   string  "
		combos := []comboS{
			{Name: val + ": Camel", Func: Camel, Expected: "myString"},
			{Name: val + ": Joined", Func: Joined, Expected: "mystring"},
			{Name: val + ": Kebab", Func: Kebab, Expected: "my-string"},
			{Name: val + ": Pascal", Func: Pascal, Expected: "MyString"},
			{Name: val + ": Snake", Func: Snake, Expected: "my_string"},
		}
		for _, c := range combos {
			res := c.Func(val)
			assert.Equal(t, c.Expected, res, c.Name)
		}
	})

	t.Run("Punctuation removed", func(t *testing.T) {
		val := ", My- string! . "
		combos := []comboS{
			{Name: val + ": Camel", Func: Camel, Expected: "myString"},
			{Name: val + ": Joined", Func: Joined, Expected: "mystring"},
			{Name: val + ": Kebab", Func: Kebab, Expected: "my-string"},
			{Name: val + ": Pascal", Func: Pascal, Expected: "MyString"},
			{Name: val + ": Snake", Func: Snake, Expected: "my_string"},
		}
		for _, c := range combos {
			res := c.Func(val)
			assert.Equal(t, c.Expected, res, c.Name)
		}
	})
}

func TestFirstLower(t *testing.T) {
	type comboS struct {
		Name     string
		Input    string
		Expected string
	}

	combos := []comboS{
		{Name: "Empty", Input: "", Expected: ""},
		{Name: "Single char", Input: "A", Expected: "a"},
		{Name: "Already first lower", Input: "my string", Expected: "my string"},
		{Name: "Standard", Input: "My string", Expected: "my string"},
		{Name: "Accents preserved", Input: "Áccénted Strïng", Expected: "áccénted Strïng"},
	}

	for _, c := range combos {
		res := FirstLower(c.Input)
		assert.Equal(t, c.Expected, res, c.Name)
	}
}
