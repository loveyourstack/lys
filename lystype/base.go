package lystype

import (
	"fmt"
	"reflect"
	"strings"
)

// RecsToMap converts a slice of structs to map[string]any using reflection.
// It only includes fields with a json tag and uses the json tag name as the map key.
// Embedded structs with no json tag are flattened recursively.
// Values are written with their native Go types.
func RecsToMap[T any](recs []T) (recsMap []map[string]any, err error) {

	if len(recs) == 0 {
		return nil, fmt.Errorf("recs is empty")
	}

	// ensure T is struct or pointer to struct
	if !isStructOrPtrToStruct(reflect.ValueOf(recs[0])) {
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
		jsonTagName, omitEmpty, omitZero := parseJSONTag(fieldType.Tag.Get("json"))

		// skip fields with json tag "-"
		if jsonTagName == "-" {
			continue
		}

		// flatten anonymous embedded structs when there is no explicit json tag name
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
		if omitZero && fieldVal.IsZero() {
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
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}

	return false
}

func ToPtr[T any](a T) *T {
	return &a
}
