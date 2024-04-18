package lysmeta

import (
	"fmt"
	"reflect"
	"strings"
)

// GetStructTags reflects the supplied struct(s) and returns the combined "db" and "json" tags
// checks for duplicate tags
func GetStructTags(reflVals ...reflect.Value) (dbTags, jsonTags []string, err error) {

	// use maps to help with duplication checks
	dbTagMap := make(map[string]int)
	jsonTagMap := make(map[string]int)

	// for each struct passed
	for _, reflVal := range reflVals {

		// for each struct field
		for i := 0; i < reflVal.NumField(); i++ {

			// get the whole tag string
			structTag := reflVal.Type().Field(i).Tag

			// get "db" tag if set
			dbTag := structTag.Get("db")
			if dbTag != "" && dbTag != "-" {
				dbTags = append(dbTags, dbTag)
				dbTagMap[dbTag]++
			}

			// get "json" tag if set, but strip out omitempty
			jsonTag := strings.Replace(structTag.Get("json"), ",omitempty", "", -1)
			if jsonTag != "" && jsonTag != "-" {
				jsonTags = append(jsonTags, jsonTag)
				jsonTagMap[jsonTag]++
			}
		}
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
		return nil, nil, fmt.Errorf(strings.Join(errA, ", "))
	}

	// success: don't use maps.Keys, since some callers depend on field order
	return dbTags, jsonTags, nil
}
