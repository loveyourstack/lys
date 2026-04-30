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
	Metadata string // optional data passed via API query param
}

// SelectParams holds the fields needed to modify a SELECT query
type SelectParams struct {
	Fields             []string
	Conditions         []Condition
	OrConditionSets    [][]Condition // sets of OR conditions. To be used by stores: is currently not available via API query params
	Sorts              []string
	Limit              int
	Offset             int
	SetFuncParamValues []any // if selecting from a setFunc, the param values that will be passed
	GetUnpagedCount    bool  // if true, will estimate the total number of records returned by this query regardless of paging
}

// GetLimit returns the LIMIT for a select.
func GetLimit(limitParam int) int {

	// if limit param was not sent (is 0), return all records
	// note that pg uses "LIMIT ALL" to return all records, but instead return the max int32 value so the datatype is always int
	// should never be negative since the API should validate this, but if it is, treat it as 0 (no limit)

	if limitParam <= 0 {
		return math.MaxInt32
	}

	return limitParam
}

// GetLimitOffsetClause returns a SELECT statement's LIMIT and OFFSET clauses
func GetLimitOffsetClause(placeholderCount int) string {

	return fmt.Sprintf(" LIMIT $%d OFFSET $%d;", placeholderCount+1, placeholderCount+2)
}

// GetOrderBy returns an SQL ORDER BY clause from a slice of sort strings
func GetOrderBy(sorts []string, defaultOrderBy string) string {

	if len(sorts) == 0 {
		if defaultOrderBy == "" {
			return ""
		}
		return " ORDER BY " + defaultOrderBy
	}
	return " ORDER BY " + strings.Join(sorts, ", ")
}

// getSelectParamValue returns the param value(s) for a given condition, handling special cases for certain operators.
func getSelectParamValue(cond Condition) []any {

	switch cond.Operator {

	// ContainsAny gets split into multiple OR statements
	case OpContainsAny:
		var paramValues []any
		for _, val := range cond.InValues {
			paramValues = append(paramValues, val)
		}
		return paramValues

	// In/NotIn uses InValues. Pgx uses the pq.Array wrapper to handle array params, so we can pass the InValues slice directly
	case OpIn, OpNotIn:
		return []any{cond.InValues}

	// Null/NotNull uses special syntax to allow nil to be passed
	case OpNull, OpNotNull:
		return []any{nil}

	// otherwise just allow pgx to handle the value
	default:
		return []any{cond.Value}
	}
}

// GetSelectParamValues returns the array of param values needed for a SELECT query.
func GetSelectParamValues(setFuncParamValues []any, conds []Condition, orCondSets [][]Condition, includeLimitOffset bool, limit, offset int) (paramValues []any) {

	// setFunc param values need to be first in the array since they are used in the FROM clause
	paramValues = append(paramValues, setFuncParamValues...)

	// regular (AND) conditions
	for _, cond := range conds {
		paramValues = append(paramValues, getSelectParamValue(cond)...)
	}

	// sets of OR conditions
	for _, orCondSet := range orCondSets {
		for _, orCond := range orCondSet {
			paramValues = append(paramValues, getSelectParamValue(orCond)...)
		}
	}

	if includeLimitOffset {
		paramValues = append(paramValues, limit)
		paramValues = append(paramValues, offset)
	}

	return paramValues
}

// GetSelectStem returns the stem of a SELECT statement using the supplied params
func GetSelectStem(selectCols string, schemaName string, sourceName string, whereClause string) string {

	return fmt.Sprintf("SELECT %s FROM %s.%s WHERE 1=1%s", selectCols, schemaName, sourceName, whereClause)
}

// GetSourceName returns viewName unless a set-returning func is used, in which case it assumes viewName is a func and appends the func param placeholders, e.g. myfunc($1,$2)
func GetSourceName(viewName string, setFuncParamValuesCount int) string {

	// if setFunc params are needed
	if setFuncParamValuesCount > 0 {

		// viewName is the setFunc: append param placeholders
		paramPlaceholders := make([]string, setFuncParamValuesCount)
		for i := range setFuncParamValuesCount {
			paramPlaceholders[i] = fmt.Sprintf("$%v", i+1) // PG params are indexed from 1
		}
		viewName = fmt.Sprintf("%s(%s)", viewName, strings.Join(paramPlaceholders, ","))
	}

	return viewName
}
