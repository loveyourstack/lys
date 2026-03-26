package lysstring

import (
	"fmt"
	"strings"

	"golang.org/x/exp/constraints"
)

// JoinInts joins the supplied int slice into a single string separated by sep
func JoinInts[T constraints.Integer](ints []T, sep string) string {
	if len(ints) == 0 {
		return ""
	}

	strA := make([]string, len(ints))
	for i, intVal := range ints {
		strA[i] = fmt.Sprintf("%d", intVal)
	}

	return strings.Join(strA, sep)
}
