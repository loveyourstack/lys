package lysmeta

import (
	"fmt"
	"reflect"
	"strings"
)

// Result contains struct metadata
type Result struct {
	DbTags         []string          // all combined "db" tags in the struct(s) passed
	JsonTags       []string          // all combined "json" tags in the struct(s) passed, excluding "-"
	JsonTagTypeMap map[string]string // a map of [json tag]type name
}

// AnalyzeStructs reflects the supplied struct(s) and returns a Result
// checks for duplicate db and json tags
func AnalyzeStructs(reflVals ...reflect.Value) (res Result, err error) {

	res.JsonTagTypeMap = make(map[string]string)

	// use maps to help with duplication checks
	dbTagMap := make(map[string]int)
	jsonTagMap := make(map[string]int)

	// for each struct passed
	for _, reflVal := range reflVals {

		reflType := reflVal.Type()

		// for each struct field
		for i := 0; i < reflVal.NumField(); i++ {

			// get the whole tag string
			structTag := reflType.Field(i).Tag

			// get "db" tag if set
			dbTag := structTag.Get("db")
			if dbTag != "" && dbTag != "-" {
				res.DbTags = append(res.DbTags, dbTag)
				dbTagMap[dbTag]++
			}

			// get "json" tag if set, but strip out omitempty
			jsonTag := strings.Replace(structTag.Get("json"), ",omitempty", "", -1)
			if jsonTag != "" && jsonTag != "-" {
				res.JsonTags = append(res.JsonTags, jsonTag)
				jsonTagMap[jsonTag]++

				res.JsonTagTypeMap[jsonTag] = fmt.Sprintf("%v", reflType.Field(i).Type)
			}

		} // next field

	} // next struct

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
