package lysmeta

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// analyzeStruct returns a Plan for the exported fields in struct t. It includes embedded structs with a db tag.
// if setValues is true, the Plan Field.Value will be set to the value from the struct, with special handling for lystype date types.
// if setValues is false, the Plan Field.Value will be nil and the Plan cached values will be populated.
func analyzeStruct(t any, setValues bool) (plan Plan, err error) {

	reflVal := reflect.ValueOf(t)

	if reflVal.Kind() != reflect.Struct {
		return Plan{}, fmt.Errorf("only struct types are accepted, but got %s", reflVal.Kind())
	}

	fields := getStructFields(reflVal, setValues)

	if len(fields) == 0 {
		return Plan{}, fmt.Errorf("struct has no exported fields")
	}

	plan = Plan{
		fields:    fields,
		hasValues: setValues,
	}

	// exit if only setting values (Analyze is generally called once per struct, but AnalyzeValues is called in loops)
	if setValues {
		return plan, nil
	}

	// set cached values for getters

	for _, field := range plan.fields {
		if field.DbName != "" {
			plan.dbNames = append(plan.dbNames, field.DbName)
		}
	}

	for _, field := range plan.fields {
		if field.JsonKey != "" {
			plan.jsonKeys = append(plan.jsonKeys, field.JsonKey)
		}
	}

	plan.jsonKeyDbNameMap = make(map[string]string)
	for _, field := range plan.fields {
		if field.JsonKey != "" && field.DbName != "" {
			plan.jsonKeyDbNameMap[field.JsonKey] = field.DbName
		}
	}

	plan.jsonKeyTypeMap = make(map[string]reflect.Type)
	for _, field := range plan.fields {
		if field.JsonKey != "" {
			plan.jsonKeyTypeMap[field.JsonKey] = field.Type
		}
	}

	return plan, nil
}

// AnalyzeValues returns a Plan for the exported fields in struct t, including embedded structs with a db tag.
// The Plan Field.Value will be set to the value from the struct, with special handling for lystype date types.
func AnalyzeValues(t any) (plan Plan, err error) {
	return analyzeStruct(t, true)
}

// Analyze returns a Plan for the exported fields in struct t, including embedded structs with a db tag.
// The Plan Field.Value will be nil, and the Plan cached values will be populated for use in getters.
// It checks for duplicate db names and json keys, and returns an error if any are found.
func Analyze(t any) (plan Plan, err error) {

	plan, err = analyzeStruct(t, false)
	if err != nil {
		return Plan{}, err
	}

	// use name/count maps to help with duplication checks
	dbNameMap := make(map[string]int)
	jsonKeyMap := make(map[string]int)

	for _, field := range plan.Fields() {
		if field.DbName != "" {
			dbNameMap[field.DbName]++
		}
		if field.JsonKey != "" && field.JsonKey != "-" {
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
func getStructFields(reflVal reflect.Value, setValues bool) (fields []Field) {

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
			innerFields := getStructFields(reflVal.Field(i), setValues)
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
		if setValues {
			f.Value = GetInputValue(reflVal.Field(i).Interface(), field.Type)
		}

		fields = append(fields, f)
	}

	return fields
}
