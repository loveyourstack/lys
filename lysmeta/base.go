package lysmeta

import (
	"fmt"
	"maps"
	"reflect"
	"strings"
)

// Result contains struct metadata
type Result struct {
	DbTags         []string                // all combined "db" tags in the struct(s) passed
	JsonTags       []string                // all combined "json" tags in the struct(s) passed, excluding "-"
	JsonTagTypeMap map[string]reflect.Type // a map of [json tag]type
}

// AnalyzeStruct reflects the supplied struct and returns a Result.
// It recursively analyzes embedded and named structs if they have db or json tags.
// It checks for duplicate db and json tags across the entire struct hierarchy.
func AnalyzeStruct(reflVal reflect.Value) (res Result, err error) {

	// use maps to help with duplication checks
	dbTagMap := make(map[string]int)
	jsonTagMap := make(map[string]int)

	res = getStructResult(reflVal)
	for _, dbTag := range res.DbTags {
		dbTagMap[dbTag]++
	}
	for _, jsonTag := range res.JsonTags {
		jsonTagMap[jsonTag]++
	}

	// check for dups
	errA := []string{}
	for k, v := range dbTagMap {
		if v > 1 {
			errA = append(errA, fmt.Sprintf("db tag '%s' is set on %d fields", k, v))
		}
	}
	for k, v := range jsonTagMap {
		if v > 1 {
			errA = append(errA, fmt.Sprintf("json tag '%s' is set on %d fields", k, v))
		}
	}

	// return errors, if any
	if len(errA) > 0 {
		return Result{}, fmt.Errorf("%s", strings.Join(errA, ", "))
	}

	return res, nil
}

func getStructResult(reflVal reflect.Value) (res Result) {

	res.JsonTagTypeMap = make(map[string]reflect.Type)

	reflType := reflVal.Type()

	// for each struct field
	for i := 0; i < reflVal.NumField(); i++ {

		field := reflType.Field(i)

		// if this field is a struct (embedded or named) and has db or json tags (omits structs like time.Time that would cause a panic)
		if field.Type.Kind() == reflect.Struct && HasDbOrJsonTags(field.Type) {

			// recurse into it
			innerRes := getStructResult(reflVal.Field(i))

			for _, v := range innerRes.DbTags {
				res.DbTags = append(res.DbTags, v)
			}
			for _, v := range innerRes.JsonTags {
				res.JsonTags = append(res.JsonTags, v)
			}
			maps.Copy(res.JsonTagTypeMap, innerRes.JsonTagTypeMap)

			continue
		}

		// get the whole tag string
		structTag := field.Tag

		// get "db" tag if set
		dbTag := structTag.Get("db")
		if dbTag != "" && dbTag != "-" {
			res.DbTags = append(res.DbTags, dbTag)
		}

		// get "json" tag if set, but strip out omitempty and omitzero
		jsonTag := strings.ReplaceAll(structTag.Get("json"), ",omitempty", "")
		jsonTag = strings.ReplaceAll(jsonTag, ",omitzero", "")
		if jsonTag != "" && jsonTag != "-" {
			res.JsonTags = append(res.JsonTags, jsonTag)

			res.JsonTagTypeMap[jsonTag] = field.Type
		}

	} // next field

	return res
}

func HasDbOrJsonTags(reflType reflect.Type) bool {

	for i := 0; i < reflType.NumField(); i++ {
		field := reflType.Field(i)
		structTag := field.Tag
		dbTag := structTag.Get("db")
		jsonTag := structTag.Get("json")
		if (dbTag != "" && dbTag != "-") || (jsonTag != "" && jsonTag != "-") {
			return true
		}
	}

	return false
}
