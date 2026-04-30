package lyspg

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/loveyourstack/lys/lyserr"
)

func TestSelectEnum(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	t.Run("default order", func(t *testing.T) {
		vals, err := SelectEnum(ctx, db, "core.weekday", nil, nil, "")
		if err != nil {
			t.Fatalf("SelectEnum failed: %v", err)
		}

		expected := []string{"None", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
		if !slices.Equal(vals, expected) {
			t.Fatalf("unexpected vals: got %v, want %v", vals, expected)
		}
	})

	t.Run("include filter", func(t *testing.T) {
		vals, err := SelectEnum(ctx, db, "core.weekday", []string{"Sunday", "Monday"}, nil, "")
		if err != nil {
			t.Fatalf("SelectEnum failed: %v", err)
		}

		expected := []string{"Sunday", "Monday"}
		if !slices.Equal(vals, expected) {
			t.Fatalf("unexpected vals: got %v, want %v", vals, expected)
		}
	})

	t.Run("exclude filter", func(t *testing.T) {
		vals, err := SelectEnum(ctx, db, "core.weekday", nil, []string{"None", "Sunday"}, "")
		if err != nil {
			t.Fatalf("SelectEnum failed: %v", err)
		}

		expected := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
		if !slices.Equal(vals, expected) {
			t.Fatalf("unexpected vals: got %v, want %v", vals, expected)
		}
	})

	t.Run("ascending sort", func(t *testing.T) {
		vals, err := SelectEnum(ctx, db, "core.weekday", nil, nil, "val")
		if err != nil {
			t.Fatalf("SelectEnum failed: %v", err)
		}

		expected := []string{"Friday", "Monday", "None", "Saturday", "Sunday", "Thursday", "Tuesday", "Wednesday"}
		if !slices.Equal(vals, expected) {
			t.Fatalf("unexpected vals: got %v, want %v", vals, expected)
		}
	})

	t.Run("descending sort", func(t *testing.T) {
		vals, err := SelectEnum(ctx, db, "core.weekday", nil, nil, "-val")
		if err != nil {
			t.Fatalf("SelectEnum failed: %v", err)
		}

		expected := []string{"Wednesday", "Tuesday", "Thursday", "Sunday", "Saturday", "None", "Monday", "Friday"}
		if !slices.Equal(vals, expected) {
			t.Fatalf("unexpected vals: got %v, want %v", vals, expected)
		}
	})

	t.Run("combined filters", func(t *testing.T) {
		vals, err := SelectEnum(ctx, db, "core.weekday", []string{"Sunday", "Monday", "Tuesday"}, []string{"Monday"}, "")
		if err != nil {
			t.Fatalf("SelectEnum failed: %v", err)
		}

		expected := []string{"Sunday", "Tuesday"}
		if !slices.Equal(vals, expected) {
			t.Fatalf("unexpected vals: got %v, want %v", vals, expected)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		vals, err := SelectEnum(ctx, db, "core.weekday", []string{"Sunday", "Monday"}, []string{"Sunday", "Monday"}, "")
		if err != nil {
			t.Fatalf("SelectEnum failed: %v", err)
		}
		if len(vals) != 0 {
			t.Fatalf("unexpected vals: got %v, want empty result", vals)
		}
	})
}

func TestSelectEnum_invalidSort(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	_, err := SelectEnum(ctx, db, "core.weekday", nil, nil, "x")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var userErr lyserr.User
	if !errors.As(err, &userErr) {
		t.Fatalf("expected lyserr.User, got %T", err)
	}
	if userErr.Message != "invalid sort val 'x'" {
		t.Fatalf("unexpected user error: got %q", userErr.Message)
	}
}

func TestSelectEnum_invalidEnumName(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	_, err := SelectEnum(ctx, db, "core.missing_enum", nil, nil, "")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dbErr lyserr.Db
	if !errors.As(err, &dbErr) {
		t.Fatalf("expected lyserr.Db, got %T", err)
	}
}

func TestCheckEnumValue(t *testing.T) {

	ctx := context.Background()
	db := mustGetDb(ctx, t)
	defer db.Close()

	t.Run("success", func(t *testing.T) {
		err := CheckEnumValue(ctx, db, "core.weekday", "Monday", "weekday")
		if err != nil {
			t.Fatalf("CheckEnumValue failed: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := CheckEnumValue(ctx, db, "core.weekday", "Funday", "weekday")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		expected := "value Funday not found in enum weekday"
		if err.Error() != expected {
			t.Fatalf("unexpected error: got %q, want %q", err.Error(), expected)
		}
	})

	t.Run("bad db enum", func(t *testing.T) {
		err := CheckEnumValue(ctx, db, "core.missing_enum", "Monday", "weekday")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error()[:18] != "SelectEnum failed:" {
			t.Fatalf("unexpected error: got %q", err.Error())
		}
	})
}
