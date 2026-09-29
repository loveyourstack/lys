package lysmap

import (
	"fmt"
	"reflect"
	"strings"
)

// FromRecs converts a slice of recs (structs from db) to map[string]any using reflection.
// It only includes fields with a json tag and uses the json tag name as the map key.
// Embedded structs with no json tag are flattened recursively.
// Values are written with their native Go types.
func FromRecs[T any](recs []T) (recsMap []map[string]any, err error) {

	// ensure T is struct or pointer to struct
	if len(recs) > 0 && !isStructOrPtrToStruct(reflect.ValueOf(recs[0])) {
		return nil, fmt.Errorf("T must be a struct or pointer to struct")
	}

	recsMap = make([]map[string]any, len(recs))

	for i, rec := range recs {
		reflVal := reflect.ValueOf(rec)

		// dereference pointer if needed
		if reflVal.Kind() == reflect.Pointer {
			if reflVal.IsNil() {
				return nil, fmt.Errorf("recs[%d] is nil", i)
			}
			reflVal = reflVal.Elem()
		}

		rowMap := make(map[string]any)
		err := structToMapByJSONTags(reflVal, rowMap)
		if err != nil {
			return nil, fmt.Errorf("structToMapByJSONTags failed for recs[%d]: %w", i, err)
		}

		recsMap[i] = rowMap
	}

	return recsMap, nil
}

func isStructOrPtrToStruct(reflVal reflect.Value) bool {
	if reflVal.Kind() == reflect.Pointer {
		if reflVal.IsNil() {
			return false
		}
		reflVal = reflVal.Elem()
	}
	return reflVal.Kind() == reflect.Struct
}

func structToMapByJSONTags(reflVal reflect.Value, out map[string]any) error {

	reflType := reflVal.Type()

	for i := 0; i < reflVal.NumField(); i++ {

		fieldType := reflType.Field(i)
		fieldVal := reflVal.Field(i)

		// skip unexported non-embedded fields
		if fieldType.PkgPath != "" && !fieldType.Anonymous {
			continue
		}

		// get json tag details
		jsonTag := fieldType.Tag.Get("json")
		jsonTagName, omitEmpty, omitZero := parseJSONTag(jsonTag)

		// use field name for json tag if the tag is empty but not omitted explicitly
		if jsonTagName == "" && jsonTag != "" && !fieldType.Anonymous {
			jsonTagName = fieldType.Name
		}

		// skip fields with json tag "-"
		if jsonTagName == "-" {
			continue
		}

		// flatten embedded structs when there is no explicit json tag name
		if fieldType.Anonymous && jsonTagName == "" {
			embVal := fieldVal

			// dereference pointer if needed
			if embVal.Kind() == reflect.Pointer {
				if embVal.IsNil() {
					continue
				}
				embVal = embVal.Elem()
			}

			if embVal.Kind() == reflect.Struct {
				err := structToMapByJSONTags(embVal, out)
				if err != nil {
					return err
				}
				continue
			}
		}

		if jsonTagName == "" {
			continue
		}

		if omitEmpty && isEmptyValue(fieldVal) {
			continue
		}
		if omitZero && isZeroValue(fieldVal) {
			continue
		}

		out[jsonTagName] = valueForMap(fieldVal)
	}

	return nil
}

func parseJSONTag(tag string) (name string, omitEmpty bool, omitZero bool) {

	if tag == "" {
		return "", false, false
	}

	parts := strings.Split(tag, ",")
	name = parts[0]

	for _, part := range parts[1:] {
		switch part {
		case "omitempty":
			omitEmpty = true
		case "omitzero":
			omitZero = true
		}
	}

	return name, omitEmpty, omitZero
}

func valueForMap(v reflect.Value) any {

	// dereference pointer if needed
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		return v.Elem().Interface()
	}

	return v.Interface()
}

func isEmptyValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	// recursively unwrap pointers and interfaces to check if the underlying value is empty
	switch v.Kind() {
	case reflect.Interface, reflect.Pointer:
		if v.IsNil() {
			return true
		}
		return isEmptyValue(v.Elem())

	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	}

	return false
}

func isZeroValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	// An interface's zero value is nil. Inspect its dynamic value so
	// custom IsZero methods are still found.
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}

	// Do not invoke methods on nil pointers.
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return true
	}

	// check for custom IsZero method on the value (e.g. lystype datetime custom types)
	if method := v.MethodByName("IsZero"); method.IsValid() &&
		method.Type().NumIn() == 0 &&
		method.Type().NumOut() == 1 &&
		method.Type().Out(0).Kind() == reflect.Bool {
		return method.Call(nil)[0].Bool()
	}

	// supports structs, arrays, complex values, unsafe pointers, etc.
	return v.IsZero()
}
