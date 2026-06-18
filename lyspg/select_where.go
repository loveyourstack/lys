package lyspg

import (
	"fmt"
	"strings"
)

// GetWhereClause returns an SQL WHERE clause using placeholders such as $1, $2 etc from the supplied conds.
// existingParamCount is the number of placeholders already used by other parts of the query, e.g. a set-returning function in the FROM clause.
// orCondSets is a slice of slices of Conditions, where the inner slices represent sets of OR conditions which should be grouped together in the resulting clause.
// The outer slice allows for multiple sets of OR conditions which will be ANDed together in the resulting clause.
func GetWhereClause(existingParamCount int, conds []Condition, orCondSets [][]Condition) (res string, numPlaceholders int) {

	idx := existingParamCount
	clause := strings.Builder{}

	// append each regular AND condition to the WHERE clause
	for _, cond := range conds {
		idx++
		part, newIdx := getWherePart(cond, idx)
		clause.WriteString(" AND ")
		clause.WriteString(part)
		idx = newIdx
	}

	// for each OR condition set
	for _, orCondSet := range orCondSets {

		// append a single AND which contains each part joined with OR
		orParts := []string{}
		for _, orCond := range orCondSet {
			idx++
			part, newIdx := getWherePart(orCond, idx)
			orParts = append(orParts, part)
			idx = newIdx
		}
		clause.WriteString(" AND (")
		clause.WriteString(strings.Join(orParts, " OR "))
		clause.WriteString(")")
	}

	return clause.String(), idx
}

// getWherePart returns the SQL clause part and updated placeholder index for a single Condition
func getWherePart(cond Condition, idx int) (clause string, newIdx int) {

	switch cond.Operator {

	case OpIn:
		return fmt.Sprintf("%s = ANY($%d)", cond.Field, idx), idx
	case OpNotIn:
		return fmt.Sprintf("NOT %s = ANY($%d)", cond.Field, idx), idx

	case OpContainsAny:
		switch len(cond.InValues) {
		case 0:
			// no values means the condition can never be true, so return a clause that is always false
			return "1=0", idx
		case 1:
			// if only one value, use the simpler single-placeholder syntax
			return fmt.Sprintf("%s::text ILIKE '%%' || $%d || '%%'", cond.Field, idx), idx
		default:
			// if multiple values, generate a clause with multiple placeholders joined by OR
			clauses := []string{}
			for valNum := range cond.InValues {

				if valNum > 0 {
					idx++
				}
				clause := fmt.Sprintf("%s::text ILIKE '%%' || $%d || '%%'", cond.Field, idx)
				clauses = append(clauses, clause)
			}
			return "(" + strings.Join(clauses, " OR ") + ")", idx
		}

	// cast field to text so that a LIKE search also works for dates/times
	// use ILIKE for case insensitivity
	case OpStartsWith:
		return fmt.Sprintf("%s::text ILIKE $%d || '%%'", cond.Field, idx), idx
	case OpEndsWith:
		return fmt.Sprintf("%s::text ILIKE '%%' || $%d", cond.Field, idx), idx
	case OpContains:
		return fmt.Sprintf("%s::text ILIKE '%%' || $%d || '%%'", cond.Field, idx), idx
	case OpNotContains:
		return fmt.Sprintf("%s::text NOT ILIKE '%%' || $%d || '%%'", cond.Field, idx), idx

	// empty / notempty: use length check, don't use a placeholder idx
	case OpEmpty:
		return fmt.Sprintf("LENGTH(%s) = 0", cond.Field), idx - 1
	case OpNotEmpty:
		return fmt.Sprintf("LENGTH(%s) > 0", cond.Field), idx - 1

	// NULL / NOT NULL: straight check, don't use a placeholder idx
	case OpNull:
		return fmt.Sprintf("%s IS NULL", cond.Field), idx - 1
	case OpNotNull:
		return fmt.Sprintf("%s IS NOT NULL", cond.Field), idx - 1

	default:
		return fmt.Sprintf("%s %s $%d", cond.Field, cond.Operator, idx), idx
	}
}
