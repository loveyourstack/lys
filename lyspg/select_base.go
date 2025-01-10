package lyspg

import (
	"fmt"
	"math"
	"strings"
)

type Operator string

// Valid condition operators
const (
	OpEquals            Operator = "="
	OpNotEquals         Operator = "!="
	OpLessThan          Operator = "<"
	OpLessThanEquals    Operator = "<="
	OpGreaterThan       Operator = ">"
	OpGreaterThanEquals Operator = ">="
	OpIn                Operator = "IN"
	OpNotIn             Operator = "NOT IN"
	OpStartsWith        Operator = "StartsWith"
	OpEndsWith          Operator = "EndsWith"
	OpContains          Operator = "Contains"
	OpNotContains       Operator = "NotContains"
	OpContainsAny       Operator = "ContainsAny"
	OpEmpty             Operator = "Empty"
	OpNotEmpty          Operator = "NotEmpty"
	OpNull              Operator = "Null"
	OpNotNull           Operator = "NotNull"
)

// Condition is a condition passed to a SELECT stmt
type Condition struct {
	Field    string
	Operator Operator // must be one of the Operator consts. if "IN" or "NOT IN", fill InValues, not Value
	Value    string
	InValues []string
}

// SelectParams holds the fields needed to modify a SELECT query
type SelectParams struct {
	Fields          []string
	Conditions      []Condition
	OrConditionSets [][]Condition // sets of OR conditions. To be used by stores: is currently not available via API query params
	Sorts           []string
	Limit           int
	Offset          int
	GetUnpagedCount bool // if true, will estimate the total number of records returned by this query regardless of paging
}

// getSelectStem returns the stem of a SELECT statement using the supplied params
func GetSelectStem(selectCols string, schemaName string, viewName string, whereClause string) string {

	return fmt.Sprintf("SELECT %s FROM %s.%s WHERE 1=1%s", selectCols, schemaName, viewName, whereClause)
}

// getOrderBy returns an SQL ORDER BY clause from a slice of sort strings
func GetOrderBy(sorts []string, defaultOrderBy string) string {

	if len(sorts) == 0 {
		if defaultOrderBy == "" {
			return ""
		}
		return " ORDER BY " + defaultOrderBy
	}
	return " ORDER BY " + strings.Join(sorts, ", ")
}

// getLimitOffsetClause returns a SELECT statement's LIMIT and OFFSET clauses
func GetLimitOffsetClause(placeholderCount int) string {

	return fmt.Sprintf(" LIMIT $%d OFFSET $%d;", placeholderCount+1, placeholderCount+2)
}

// getSelectParamValues returns the array of param values needed for a SELECT query
func GetSelectParamValues(conds []Condition, orCondSets [][]Condition, includeLimitOffset bool, limit, offset int) (paramValues []any) {

	// regular (AND) conditions
	for _, cond := range conds {
		switch cond.Operator {

		// ContainsAny gets split into multiple OR statements
		case OpContainsAny:
			for _, val := range cond.InValues {
				paramValues = append(paramValues, val)
			}

		// In/NotIn uses InValues
		case OpIn, OpNotIn:
			paramValues = append(paramValues, cond.InValues)

		// Null/NotNull uses special syntax to allow nil to be passed
		case OpNull, OpNotNull:
			paramValues = append(paramValues, nil)

		// otherwise just allow pgx to handle the value
		default:
			paramValues = append(paramValues, cond.Value)
		}
	}

	// sets of OR conditions
	for _, orCondSet := range orCondSets {
		for _, orCond := range orCondSet {

			switch orCond.Operator {
			case OpContainsAny:
				for _, val := range orCond.InValues {
					paramValues = append(paramValues, val)
				}
			case OpIn, OpNotIn:
				paramValues = append(paramValues, orCond.InValues)
			case OpNull, OpNotNull:
				paramValues = append(paramValues, nil)
			default:
				paramValues = append(paramValues, orCond.Value)
			}
		}
	}

	if includeLimitOffset {
		paramValues = append(paramValues, limit)
		paramValues = append(paramValues, offset)
	}

	return paramValues
}

// getLimit returns the LIMIT for a select
func GetLimit(limitParam int) int {

	// if limit param was not sent (is 0), return all records
	// note that pg uses "LIMIT ALL" to return all records, but instead return the max int32 value so the datatype is always int
	if limitParam == 0 {
		return math.MaxInt32
	}

	return limitParam
}
