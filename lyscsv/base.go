package lyscsv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"strconv"

	"github.com/loveyourstack/lys/lystype"
	"golang.org/x/exp/maps"
)

// WriteItems writes csv data to a writer from items.
// T must have json tags set. Only the fields with a json tag get written.
// jsonTagTypeMap is a map of [json tag]type.
func WriteItems[T any](items []T, jsonTagTypeMap map[string]reflect.Type, delimiter rune, w io.Writer) (err error) {

	if len(items) == 0 {
		return fmt.Errorf("items is empty")
	}
	if len(jsonTagTypeMap) == 0 {
		return fmt.Errorf("jsonTagTypeMap is empty")
	}
	if delimiter == 0 {
		return fmt.Errorf("delimiter is mandatory")
	}
	if w == nil {
		return fmt.Errorf("writer is mandatory")
	}

	// convert items to []map[string]any
	recsMap, err := lystype.RecsToMap(items)
	if err != nil {
		return fmt.Errorf("lystype.RecsToMap failed: %w", err)
	}

	// get [][]string
	data, err := getStrData(recsMap, jsonTagTypeMap)
	if err != nil {
		return fmt.Errorf("getStrData failed: %w", err)
	}

	// write csv to writer
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = delimiter
	for _, rec := range data {
		if err := csvWriter.Write(rec); err != nil {
			return fmt.Errorf("failed to write line: %s: %w", rec, err)
		}
	}
	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("csv.NewWriter: flush: %w", err)
	}

	return nil
}

// WriteItemsToFile creates a csv file from items.
// T must have json tags set. Only the fields with a json tag get written.
// jsonTagTypeMap is a map of [json tag]type.
func WriteItemsToFile[T any](items []T, jsonTagTypeMap map[string]reflect.Type, delimiter rune, filePath string) (err error) {

	if filePath == "" {
		return fmt.Errorf("filePath is mandatory")
	}

	// open file for writing
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return fmt.Errorf("os.OpenFile failed: %w", err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			fmt.Printf("f.Close failed: %s", err.Error())
		}
	}()

	if err := WriteItems(items, jsonTagTypeMap, delimiter, f); err != nil {
		return fmt.Errorf("WriteItems failed: %w", err)
	}

	return nil
}

func getStrData(recsMap []map[string]any, jsonTagTypeMap map[string]reflect.Type) (data [][]string, err error) {

	// return 1 row per record, plus 1 for the header row
	data = make([][]string, len(recsMap)+1)

	// get sorted keys
	keys := maps.Keys(jsonTagTypeMap)
	slices.Sort(keys)

	// assign header row
	data[0] = keys

	// add data

	// for each record
	for i := range recsMap {

		// add a row with 1 column per key
		row := make([]string, len(keys))

		// for each key
		for j, key := range keys {

			val := recsMap[i][key]

			// use jsonTagTypeMap to call the appropriate formatting func for each type
			// to allow for optional fields, skip values that cannot be asserted rather than returning an error
			switch jsonTagTypeMap[key] {

			case reflect.TypeFor[bool]():
				boolVal, ok := val.(bool)
				if ok {
					row[j] = strconv.FormatBool(boolVal)
				} else {
					row[j] = ""
				}

			case reflect.TypeFor[float32](), reflect.TypeFor[float64]():
				f64Val, ok := val.(float64)
				if ok {
					row[j] = strconv.FormatFloat(f64Val, 'f', -1, 64)
				} else {
					row[j] = ""
				}

			case reflect.TypeFor[int](), reflect.TypeFor[int32](), reflect.TypeFor[int64]():
				intVal, ok := val.(int64)
				if ok {
					row[j] = strconv.FormatInt(intVal, 10)
				} else {
					row[j] = ""
				}

			case reflect.TypeFor[lystype.Date]():
				timeVal, ok := val.(lystype.Date)
				if !ok {
					continue
				}
				row[j] = timeVal.Format(lystype.DateFormat)

			case reflect.TypeFor[lystype.Datetime]():
				timeVal, ok := val.(lystype.Datetime)
				if !ok {
					continue
				}
				row[j] = timeVal.Format(lystype.DatetimeFormat)

			case reflect.TypeFor[lystype.Time]():
				timeVal, ok := val.(lystype.Time)
				if !ok {
					continue
				}
				row[j] = timeVal.Format(lystype.TimeFormat)

			default:
				strVal, ok := val.(string)
				if ok {
					row[j] = strVal
				} else {
					row[j] = ""
				}
			}

		} // next key

		data[i+1] = row

	} // next record

	return data, nil
}
