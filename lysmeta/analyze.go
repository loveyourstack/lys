package lysmeta

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// AnalyzeT returns a Plan for the exported fields in struct t. It includes embedded structs with a db tag.
// if getValues is true, the Plan Field.Value will be set to the value from the struct, with special handling for lystype date types.
func AnalyzeT(t any, getValues bool) (plan Plan, err error) {

	reflVal := reflect.ValueOf(t)

	if reflVal.Kind() != reflect.Struct {
		return Plan{}, fmt.Errorf("AnalyzeT only accepts struct types, but got %s", reflVal.Kind())
	}

	fields := getStructFields(reflVal, getValues)

	if len(fields) == 0 {
		return Plan{}, fmt.Errorf("struct has no exported fields")
	}

	plan = Plan{
		Fields:    fields,
		HasValues: getValues,
	}

	return plan, nil
}

// AnalyzeAndCheckT is a wrapper around AnalyzeT that also checks for duplicate db and json tags across the struct hierarchy.
// It should be called by each Store on startup.
func AnalyzeAndCheckT(t any) (plan Plan, err error) {

	plan, err = AnalyzeT(t, false)
	if err != nil {
		return Plan{}, fmt.Errorf("AnalyzeT failed: %w", err)
	}

	// use name/count maps to help with duplication checks
	dbNameMap := make(map[string]int)
	jsonKeyMap := make(map[string]int)

	for _, field := range plan.Fields {
		if field.DbName != "" {
			dbNameMap[field.DbName]++
		}
		if field.JsonKey != "" {
			jsonKeyMap[field.JsonKey]++
		}
	}

	// get slices of dups from maps for sorting
	type dup struct {
		name string
		n    int
	}
	dbDups := []dup{}
	jsonDups := []dup{}

	for k, v := range dbNameMap {
		if v > 1 {
			dbDups = append(dbDups, dup{name: k, n: v})
		}
	}
	for k, v := range jsonKeyMap {
		if v > 1 {
			jsonDups = append(jsonDups, dup{name: k, n: v})
		}
	}

	// sort dups slices by name for deterministic error messages
	sort.Slice(dbDups, func(i, j int) bool {
		return dbDups[i].name < dbDups[j].name
	})
	sort.Slice(jsonDups, func(i, j int) bool {
		return jsonDups[i].name < jsonDups[j].name
	})

	// build error messages for dups, if any
	errA := []string{}
	for _, v := range dbDups {
		errA = append(errA, fmt.Sprintf("db name '%s' is set on %d fields", v.name, v.n))
	}
	for _, v := range jsonDups {
		errA = append(errA, fmt.Sprintf("json key '%s' is set on %d fields", v.name, v.n))
	}

	// return joined errors, if any
	if len(errA) > 0 {
		return Plan{}, fmt.Errorf("%s", strings.Join(errA, ", "))
	}

	// success
	return plan, nil
}

// getDbName returns the db name from the struct tag, or empty string if the field should be ignored in db operations.
// Unlike pgx, missing tags also result in empty db name, not just "-" tags.
func getDbName(dbTag string) string {

	// if db tag is missing or "-", return empty key
	if dbTag == "" || dbTag == "-" {
		return ""
	}

	return dbTag
}

// getJsonKey implements the same rules as encoding/json for determining the json key from the struct tag.
func getJsonKey(jsonTag, fieldName string) string {

	// if json tag is missing, use the field name as the json key
	if jsonTag == "" {
		return fieldName
	}

	// split json tag by comma
	parts := strings.Split(jsonTag, ",")

	// if json name is empty (e.g. ",omitempty"), return field name
	if parts[0] == "" {
		return fieldName
	}

	// if field is omitted in json, return empty key
	if parts[0] == "-" {
		return ""
	}

	return parts[0]
}

// getStructFields recursively gets the fields of a struct, including embedded structs. It skips unexported fields.
func getStructFields(reflVal reflect.Value, getValues bool) (fields []Field) {

	reflType := reflVal.Type()

	// for each struct field
	for i := 0; i < reflVal.NumField(); i++ {

		field := reflType.Field(i)

		// skip unexported fields
		if !field.IsExported() {
			continue
		}

		// if this is an embedded struct field
		if field.Type.Kind() == reflect.Struct && field.Anonymous {

			// recurse into it
			innerFields := getStructFields(reflVal.Field(i), getValues)
			fields = append(fields, innerFields...)
			continue

		}

		// add field
		f := Field{
			Name: field.Name,
			Type: field.Type,

			DbName:  getDbName(field.Tag.Get("db")),
			JsonKey: getJsonKey(field.Tag.Get("json"), field.Name),
		}
		if getValues {
			f.Value = GetInputValue(reflVal.Field(i).Interface(), field.Type)
		}

		fields = append(fields, f)
	}

	return fields
}
