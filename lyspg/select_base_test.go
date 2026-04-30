package lyspg

import (
	"math"
	"reflect"
	"testing"
)

func TestGetLimit(t *testing.T) {
	t.Run("zero returns max int32", func(t *testing.T) {
		got := GetLimit(0)
		if got != math.MaxInt32 {
			t.Fatalf("unexpected limit: got %d, want %d", got, math.MaxInt32)
		}
	})

	t.Run("positive returns unchanged", func(t *testing.T) {
		got := GetLimit(25)
		if got != 25 {
			t.Fatalf("unexpected limit: got %d, want 25", got)
		}
	})

	t.Run("negative returns max int32", func(t *testing.T) {
		got := GetLimit(-5)
		if got != math.MaxInt32 {
			t.Fatalf("unexpected limit: got %d, want %d", got, math.MaxInt32)
		}
	})
}

func TestGetLimitOffsetClause(t *testing.T) {
	t.Run("placeholder 0", func(t *testing.T) {
		got := GetLimitOffsetClause(0)
		want := " LIMIT $1 OFFSET $2;"
		if got != want {
			t.Fatalf("unexpected clause: got %q, want %q", got, want)
		}
	})

	t.Run("placeholder 1", func(t *testing.T) {
		got := GetLimitOffsetClause(1)
		want := " LIMIT $2 OFFSET $3;"
		if got != want {
			t.Fatalf("unexpected clause: got %q, want %q", got, want)
		}
	})

	t.Run("placeholder 5", func(t *testing.T) {
		got := GetLimitOffsetClause(5)
		want := " LIMIT $6 OFFSET $7;"
		if got != want {
			t.Fatalf("unexpected clause: got %q, want %q", got, want)
		}
	})
}

func TestGetOrderBy(t *testing.T) {
	t.Run("empty sorts and empty default", func(t *testing.T) {
		got := GetOrderBy(nil, "")
		if got != "" {
			t.Fatalf("unexpected order by: got %q, want empty string", got)
		}
	})

	t.Run("empty sorts with default", func(t *testing.T) {
		got := GetOrderBy(nil, "id")
		want := " ORDER BY id"
		if got != want {
			t.Fatalf("unexpected order by: got %q, want %q", got, want)
		}
	})

	t.Run("non-empty sorts overrides default", func(t *testing.T) {
		got := GetOrderBy([]string{"c_text", "id DESC"}, "id")
		want := " ORDER BY c_text, id DESC"
		if got != want {
			t.Fatalf("unexpected order by: got %q, want %q", got, want)
		}
	})
}

func TestGetSelectParamValues_basics(t *testing.T) {
	setFuncVals := []any{"sf1", 2}
	conds := []Condition{{Field: "a", Operator: OpEquals, Value: "x"}}
	orCondSets := [][]Condition{{{Field: "b", Operator: OpEquals, Value: "y"}}}

	got := GetSelectParamValues(setFuncVals, conds, orCondSets, true, 50, 10)
	want := []any{"sf1", 2, "x", "y", 50, 10}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected params: got %#v, want %#v", got, want)
	}
}

func TestGetSelectParamValues_operatorEncodings(t *testing.T) {
	conds := []Condition{
		{Field: "a", Operator: OpEquals, Value: "eq"},
		{Field: "b", Operator: OpIn, InValues: []string{"i1", "i2"}},
		{Field: "c", Operator: OpNotIn, InValues: []string{"n1", "n2"}},
		{Field: "d", Operator: OpNull},
		{Field: "e", Operator: OpNotNull},
		{Field: "f", Operator: OpContainsAny, InValues: []string{"c1", "c2"}},
	}

	got := GetSelectParamValues(nil, conds, nil, false, 0, 0)
	want := []any{"eq", []string{"i1", "i2"}, []string{"n1", "n2"}, nil, nil, "c1", "c2"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected params: got %#v, want %#v", got, want)
	}
}

func TestGetSelectParamValues_orCondSetsOrder(t *testing.T) {
	orCondSets := [][]Condition{
		{
			{Field: "a", Operator: OpEquals, Value: "a1"},
			{Field: "b", Operator: OpContainsAny, InValues: []string{"b1", "b2"}},
		},
		{
			{Field: "c", Operator: OpNotIn, InValues: []string{"c1", "c2"}},
		},
	}

	got := GetSelectParamValues(nil, nil, orCondSets, false, 0, 0)
	want := []any{"a1", "b1", "b2", []string{"c1", "c2"}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected params: got %#v, want %#v", got, want)
	}
}

func TestGetSelectParamValues_noLimitOffset(t *testing.T) {
	got := GetSelectParamValues(nil, []Condition{{Field: "a", Operator: OpEquals, Value: "x"}}, nil, false, 99, 88)
	want := []any{"x"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected params: got %#v, want %#v", got, want)
	}
}

func TestGetSelectStem(t *testing.T) {
	got := GetSelectStem("id, c_text", "core", "v_tag_test", " AND c_text = $1")
	want := "SELECT id, c_text FROM core.v_tag_test WHERE 1=1 AND c_text = $1"
	if got != want {
		t.Fatalf("unexpected select stem: got %q, want %q", got, want)
	}
}

func TestGetSourceName(t *testing.T) {
	t.Run("no setfunc params", func(t *testing.T) {
		got := GetSourceName("core.my_view", 0)
		want := "core.my_view"
		if got != want {
			t.Fatalf("unexpected source name: got %q, want %q", got, want)
		}
	})

	t.Run("single setfunc param", func(t *testing.T) {
		got := GetSourceName("core.my_func", 1)
		want := "core.my_func($1)"
		if got != want {
			t.Fatalf("unexpected source name: got %q, want %q", got, want)
		}
	})

	t.Run("multiple setfunc params", func(t *testing.T) {
		got := GetSourceName("core.my_func", 3)
		want := "core.my_func($1,$2,$3)"
		if got != want {
			t.Fatalf("unexpected source name: got %q, want %q", got, want)
		}
	})
}
