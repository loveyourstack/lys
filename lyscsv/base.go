package lyscsv

import (
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/loveyourstack/lys/lystype"
	"golang.org/x/exp/maps"
)

// WriteItemsToFile creates a csv file from items
// T must have json tags set. Only the fields with a json tag get written
// jsonTagTypeMap is a map of [json tag]type name
// if filePath exists, it gets overwritten
func WriteItemsToFile[T any](items []T, jsonTagTypeMap map[string]string, filePath string, delimiter rune) (err error) {

	if len(items) == 0 {
		return fmt.Errorf("items is empty")
	}
	if len(jsonTagTypeMap) == 0 {
		return fmt.Errorf("jsonTagMap is empty")
	}
	if filePath == "" {
		return fmt.Errorf("filePath is mandatory")
	}
	if delimiter == 0 {
		return fmt.Errorf("delimiter is mandatory")
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

	// open file for writing
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return fmt.Errorf("os.OpenFile failed: %w", err)
	}
	defer f.Close()

	// write file
	w := csv.NewWriter(f)
	w.Comma = delimiter
	for _, rec := range data {
		if err := w.Write(rec); err != nil {
			return fmt.Errorf("failed to write line: %s: %w", rec, err)
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return fmt.Errorf("csv.NewWriter: %w", err)
	}

	return nil
}

func getStrData(recsMap []map[string]any, jsonTagTypeMap map[string]string) (data [][]string, err error) {

	// get sorted keys
	keys := maps.Keys(jsonTagTypeMap)
	slices.Sort(keys)

	// add header row
	data = append(data, keys)

	// add data

	// for each record
	for i := range recsMap {

		// add a row
		row := []string{}

		// for each key
		for _, key := range keys {

			val := recsMap[i][key]

			// use jsonTagTypeMap to call the appropriate cell.SetType func for each type
			switch jsonTagTypeMap[key] {

			case "bool":
				boolVal, ok := val.(bool)
				if ok {
					row = append(row, strconv.FormatBool(boolVal))
				} else {
					row = append(row, "")
				}

			case "float32", "float64":
				f64Val, ok := val.(float64)
				if ok {
					row = append(row, strconv.FormatFloat(f64Val, 'f', -1, 64))
				} else {
					row = append(row, "")
				}

			case "int", "int32", "int64": // needs float64 fallback
				intVal, ok := val.(int)
				if ok {
					row = append(row, strconv.Itoa(intVal))
				}
				f64Val, ok := val.(float64)
				if ok {
					row = append(row, strconv.FormatFloat(f64Val, 'f', -1, 64))
				} else {
					row = append(row, "")
				}

			case "lystype.Date":
				strVal, ok := val.(string)
				if !ok {
					continue
				}
				timeVal, err := time.Parse(lystype.DateFormat, strVal)
				if err != nil {
					return nil, fmt.Errorf("time.Parse failed: %w", err)
				}
				row = append(row, timeVal.Format(lystype.DateFormat))

			case "lystype.Datetime":
				strVal, ok := val.(string)
				if !ok {
					continue
				}
				timeVal, err := time.Parse(lystype.DatetimeFormat, strVal)
				if err != nil {
					return nil, fmt.Errorf("time.Parse failed: %w", err)
				}
				row = append(row, timeVal.Format(lystype.DatetimeFormat))

			default:
				strVal, ok := val.(string)
				if ok {
					row = append(row, strVal)
				} else {
					row = append(row, "")
				}
			}

		} // next key

		data = append(data, row)

	} // next record

	return data, nil
}
