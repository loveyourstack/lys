package lyspg

import "testing"

func TestGetWhereClause_emptyInputs(t *testing.T) {
	clause, n := GetWhereClause(0, nil, nil)
	if clause != "" {
		t.Fatalf("unexpected clause: got %q, want empty", clause)
	}
	if n != 0 {
		t.Fatalf("unexpected placeholder count: got %d, want 0", n)
	}
}

func TestGetWhereClause_existingParamOffset(t *testing.T) {
	conds := []Condition{{Field: "a", Operator: OpEquals, Value: "x"}}
	clause, n := GetWhereClause(2, conds, nil)

	wantClause := " AND a = $3"
	if clause != wantClause {
		t.Fatalf("unexpected clause: got %q, want %q", clause, wantClause)
	}
	if n != 3 {
		t.Fatalf("unexpected placeholder count: got %d, want 3", n)
	}
}

func TestGetWhereClause_andConditionsOrdering(t *testing.T) {
	conds := []Condition{
		{Field: "a", Operator: OpEquals, Value: "x"},
		{Field: "b", Operator: OpLessThan, Value: "10"},
	}
	clause, n := GetWhereClause(0, conds, nil)

	wantClause := " AND a = $1 AND b < $2"
	if clause != wantClause {
		t.Fatalf("unexpected clause: got %q, want %q", clause, wantClause)
	}
	if n != 2 {
		t.Fatalf("unexpected placeholder count: got %d, want 2", n)
	}
}

func TestGetWhereClause_orConditionSetsGrouping(t *testing.T) {
	orCondSets := [][]Condition{{
		{Field: "a", Operator: OpEquals, Value: "x"},
		{Field: "b", Operator: OpGreaterThan, Value: "1"},
	}}
	clause, n := GetWhereClause(0, nil, orCondSets)

	wantClause := " AND (a = $1 OR b > $2)"
	if clause != wantClause {
		t.Fatalf("unexpected clause: got %q, want %q", clause, wantClause)
	}
	if n != 2 {
		t.Fatalf("unexpected placeholder count: got %d, want 2", n)
	}
}

func TestGetWhereClause_mixedAndAndOr(t *testing.T) {
	conds := []Condition{{Field: "a", Operator: OpEquals, Value: "x"}}
	orCondSets := [][]Condition{{
		{Field: "b", Operator: OpEquals, Value: "y"},
		{Field: "c", Operator: OpNotEquals, Value: "z"},
	}}
	clause, n := GetWhereClause(1, conds, orCondSets)

	wantClause := " AND a = $2 AND (b = $3 OR c != $4)"
	if clause != wantClause {
		t.Fatalf("unexpected clause: got %q, want %q", clause, wantClause)
	}
	if n != 4 {
		t.Fatalf("unexpected placeholder count: got %d, want 4", n)
	}
}

func TestGetWhereClause_operatorShapes(t *testing.T) {
	testsWithPlaceholder := []struct {
		name   string
		cond   Condition
		clause string
	}{
		{name: "in", cond: Condition{Field: "f", Operator: OpIn, InValues: []string{"a"}}, clause: " AND f = ANY($1)"},
		{name: "not in", cond: Condition{Field: "f", Operator: OpNotIn, InValues: []string{"a"}}, clause: " AND NOT f = ANY($1)"},

		{name: "starts with", cond: Condition{Field: "f", Operator: OpStartsWith, Value: "a"}, clause: " AND f::text ILIKE $1 || '%'"},
		{name: "ends with", cond: Condition{Field: "f", Operator: OpEndsWith, Value: "a"}, clause: " AND f::text ILIKE '%' || $1"},
		{name: "contains", cond: Condition{Field: "f", Operator: OpContains, Value: "a"}, clause: " AND f::text ILIKE '%' || $1 || '%'"},
		{name: "not contains", cond: Condition{Field: "f", Operator: OpNotContains, Value: "a"}, clause: " AND f::text NOT ILIKE '%' || $1 || '%'"},

		{name: "default (equals)", cond: Condition{Field: "f", Operator: OpEquals, Value: "a"}, clause: " AND f = $1"},
		{name: "default (gt)", cond: Condition{Field: "f", Operator: OpGreaterThan, Value: "1"}, clause: " AND f > $1"},
		{name: "default (lt)", cond: Condition{Field: "f", Operator: OpLessThan, Value: "1"}, clause: " AND f < $1"},
	}

	for _, tc := range testsWithPlaceholder {
		t.Run(tc.name, func(t *testing.T) {
			clause, n := GetWhereClause(0, []Condition{tc.cond}, nil)
			if clause != tc.clause {
				t.Fatalf("unexpected clause: got %q, want %q", clause, tc.clause)
			}
			if n != 1 {
				t.Fatalf("unexpected placeholder count: got %d, want 1", n)
			}
		})
	}

	testsWithoutPlaceholder := []struct {
		name   string
		cond   Condition
		clause string
	}{
		{name: "empty", cond: Condition{Field: "f", Operator: OpEmpty, Value: ""}, clause: " AND LENGTH(f) = 0"},
		{name: "not empty", cond: Condition{Field: "f", Operator: OpNotEmpty, Value: ""}, clause: " AND LENGTH(f) > 0"},

		{name: "null", cond: Condition{Field: "f", Operator: OpNull}, clause: " AND f IS NULL"},
		{name: "not null", cond: Condition{Field: "f", Operator: OpNotNull}, clause: " AND f IS NOT NULL"},
	}

	for _, tc := range testsWithoutPlaceholder {
		t.Run(tc.name, func(t *testing.T) {
			clause, n := GetWhereClause(0, []Condition{tc.cond}, nil)
			if clause != tc.clause {
				t.Fatalf("unexpected clause: got %q, want %q", clause, tc.clause)
			}
			if n != 0 {
				t.Fatalf("unexpected placeholder count: got %d, want 0", n)
			}
		})
	}
}

func TestGetWhereClause_containsAnySingleValue(t *testing.T) {
	conds := []Condition{{Field: "f", Operator: OpContainsAny, InValues: []string{"a"}}}
	clause, n := GetWhereClause(0, conds, nil)

	wantClause := " AND f::text ILIKE '%' || $1 || '%'"
	if clause != wantClause {
		t.Fatalf("unexpected clause: got %q, want %q", clause, wantClause)
	}
	if n != 1 {
		t.Fatalf("unexpected placeholder count: got %d, want 1", n)
	}
}

func TestGetWhereClause_containsAnyMultipleValues(t *testing.T) {
	conds := []Condition{{Field: "f", Operator: OpContainsAny, InValues: []string{"a", "b", "c"}}}
	clause, n := GetWhereClause(0, conds, nil)

	wantClause := " AND (f::text ILIKE '%' || $1 || '%' OR f::text ILIKE '%' || $2 || '%' OR f::text ILIKE '%' || $3 || '%')"
	if clause != wantClause {
		t.Fatalf("unexpected clause: got %q, want %q", clause, wantClause)
	}
	if n != 3 {
		t.Fatalf("unexpected placeholder count: got %d, want 3", n)
	}
}

func TestGetWhereClause_containsAnyEmptyInValues(t *testing.T) {
	conds := []Condition{{Field: "f", Operator: OpContainsAny, InValues: nil}}
	clause, n := GetWhereClause(0, conds, nil)

	// Current behavior: empty InValues renders an empty inner expression and still consumes one placeholder index.
	wantClause := " AND 1=0"
	if clause != wantClause {
		t.Fatalf("unexpected clause: got %q, want %q", clause, wantClause)
	}
	if n != 1 {
		t.Fatalf("unexpected placeholder count: got %d, want 1", n)
	}
}

func TestGetWhereClause_placeholderProgressionAfterContainsAny(t *testing.T) {
	conds := []Condition{
		{Field: "f", Operator: OpContainsAny, InValues: []string{"a", "b"}},
		{Field: "g", Operator: OpEquals, Value: "z"},
	}
	clause, n := GetWhereClause(0, conds, nil)

	wantClause := " AND (f::text ILIKE '%' || $1 || '%' OR f::text ILIKE '%' || $2 || '%') AND g = $3"
	if clause != wantClause {
		t.Fatalf("unexpected clause: got %q, want %q", clause, wantClause)
	}
	if n != 3 {
		t.Fatalf("unexpected placeholder count: got %d, want 3", n)
	}
}
