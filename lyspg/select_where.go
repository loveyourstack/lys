package lyspg

import (
	"fmt"
	"strings"
)

// GetWhereClause returns an SQL WHERE clause using placeholders such as $1, $2 etc from the supplied conds
func GetWhereClause(conds []Condition) (res string, numPlaceholders int) {

	i := 0
	for _, cond := range conds {

		i++
		switch cond.Operator {

		case OpIn:
			res += fmt.Sprintf(" AND (%s = ANY($%d))", cond.Field, i)
		case OpNotIn:
			res += fmt.Sprintf(" AND NOT (%s = ANY($%d))", cond.Field, i)

		case OpContainsAny:
			clauses := []string{}
			for valNum := range cond.InValues {

				if valNum > 0 {
					i++
				}
				clause := fmt.Sprintf("%s::text ILIKE '%%' || $%d || '%%'", cond.Field, i)
				clauses = append(clauses, clause)
			}
			res += " AND (" + strings.Join(clauses, " OR ") + ")"

		// cast field to text so that a LIKE search also works for dates/times
		// use ILIKE for case insensitivity
		case OpStartsWith:
			res += fmt.Sprintf(" AND (%s::text ILIKE $%d || '%%')", cond.Field, i)
		case OpEndsWith:
			res += fmt.Sprintf(" AND (%s::text ILIKE '%%' || $%d)", cond.Field, i)
		case OpContains:
			res += fmt.Sprintf(" AND (%s::text ILIKE '%%' || $%d || '%%')", cond.Field, i)
		case OpNotContains:
			res += fmt.Sprintf(" AND (%s::text NOT ILIKE '%%' || $%d || '%%')", cond.Field, i)

		// empty / notempty: using = '' doesn't work when param '' is a placeholder $, using LENGTH() as workaround
		case OpEmpty:
			res += fmt.Sprintf(" AND (LENGTH(%s) = $%d)", cond.Field, i)
		case OpNotEmpty:
			res += fmt.Sprintf(" AND (LENGTH(%s) > $%d)", cond.Field, i)

		// NULL: use special syntax which allows NULL param to be treated as a value
		case OpNull:
			res += fmt.Sprintf(" AND (%s IS NOT DISTINCT FROM $%d)", cond.Field, i)
		case OpNotNull:
			res += fmt.Sprintf(" AND (%s IS DISTINCT FROM $%d)", cond.Field, i)

		default:
			res += fmt.Sprintf(" AND (%s %s $%d)", cond.Field, cond.Operator, i)
		}
	}
	return res, i
}
