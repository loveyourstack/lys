package lyspg

import (
	"fmt"
	"strings"
)

// GetWhereClause returns an SQL WHERE clause using placeholders such as $1, $2 etc from the supplied conds
func GetWhereClause(existingParamCount int, conds []Condition, orCondSets [][]Condition) (res string, numPlaceholders int) {

	i := existingParamCount
	clause := strings.Builder{}

	// append each regular AND condition to the WHERE clause
	for _, cond := range conds {
		i++
		part, newI := getWherePart(cond, i)
		clause.WriteString(" AND " + part)
		i = newI
	}

	// for each OR condition set
	for _, orCondSet := range orCondSets {

		// append a single AND which contains each part joined with OR
		orParts := []string{}
		for _, orCond := range orCondSet {
			i++
			part, newI := getWherePart(orCond, i)
			orParts = append(orParts, part)
			i = newI
		}
		clause.WriteString(" AND (" + strings.Join(orParts, " OR ") + ")")
	}

	return clause.String(), i
}

func getWherePart(cond Condition, i int) (string, int) {

	switch cond.Operator {

	case OpIn:
		return fmt.Sprintf(" (%s = ANY($%d))", cond.Field, i), i
	case OpNotIn:
		return fmt.Sprintf(" NOT (%s = ANY($%d))", cond.Field, i), i

	case OpContainsAny:
		clauses := []string{}
		for valNum := range cond.InValues {

			if valNum > 0 {
				i++
			}
			clause := fmt.Sprintf("%s::text ILIKE '%%' || $%d || '%%'", cond.Field, i)
			clauses = append(clauses, clause)
		}
		return " (" + strings.Join(clauses, " OR ") + ")", i

	// cast field to text so that a LIKE search also works for dates/times
	// use ILIKE for case insensitivity
	case OpStartsWith:
		return fmt.Sprintf(" (%s::text ILIKE $%d || '%%')", cond.Field, i), i
	case OpEndsWith:
		return fmt.Sprintf(" (%s::text ILIKE '%%' || $%d)", cond.Field, i), i
	case OpContains:
		return fmt.Sprintf(" (%s::text ILIKE '%%' || $%d || '%%')", cond.Field, i), i
	case OpNotContains:
		return fmt.Sprintf(" (%s::text NOT ILIKE '%%' || $%d || '%%')", cond.Field, i), i

	// empty / notempty: using = '' doesn't work when param '' is a placeholder $, using LENGTH() as workaround
	case OpEmpty:
		return fmt.Sprintf(" (LENGTH(%s) = $%d)", cond.Field, i), i
	case OpNotEmpty:
		return fmt.Sprintf(" (LENGTH(%s) > $%d)", cond.Field, i), i

	// NULL: use special syntax which allows NULL param to be treated as a value
	case OpNull:
		return fmt.Sprintf(" (%s IS NOT DISTINCT FROM $%d)", cond.Field, i), i
	case OpNotNull:
		return fmt.Sprintf(" (%s IS DISTINCT FROM $%d)", cond.Field, i), i

	default:
		return fmt.Sprintf(" (%s %s $%d)", cond.Field, cond.Operator, i), i
	}
}
