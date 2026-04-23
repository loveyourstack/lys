package lysexcel

import (
	"fmt"
	"reflect"
	"slices"
	"time"

	"codeberg.org/tealeg/xlsx/v4"
	"github.com/loveyourstack/lys/lystype"
	"golang.org/x/exp/maps"
)

// WriteItemsToFile creates an Excel workbook from items
// T must have json tags set. Only the fields with a json tag get written
// jsonTagTypeMap is a map of [json tag]type name
// if filePath exists, it gets overwritten
// sheetName is optional and defaults to "data"
func WriteItemsToFile[T any](items []T, jsonTagTypeMap map[string]reflect.Type, filePath, sheetName string) (err error) {

	if len(items) == 0 {
		return fmt.Errorf("items is empty")
	}
	if len(jsonTagTypeMap) == 0 {
		return fmt.Errorf("jsonTagTypeMap is empty")
	}
	if filePath == "" {
		return fmt.Errorf("filePath is mandatory")
	}

	// convert items to []map[string]any
	recsMap, err := lystype.RecsToMap(items)
	if err != nil {
		return fmt.Errorf("lystype.RecsToMap failed: %w", err)
	}

	// write to Excel file in memory, return workbook
	wb, sh, err := writeData(recsMap, jsonTagTypeMap, sheetName)
	if err != nil {
		return fmt.Errorf("writeData failed: %w", err)
	}
	defer sh.Close()

	// save workbook to disk
	err = wb.Save(filePath)
	if err != nil {
		return fmt.Errorf("wb.Save failed: %w", err)
	}

	return nil
}

func writeData(recsMap []map[string]any, jsonTagTypeMap map[string]reflect.Type, sheetName string) (wb *xlsx.File, sh *xlsx.Sheet, err error) {

	if sheetName == "" {
		sheetName = "data"
	}

	// get sorted keys
	keys := maps.Keys(jsonTagTypeMap)
	slices.Sort(keys)

	// create workbook
	wb = xlsx.NewFile()

	// add sheet
	sh, err = wb.AddSheet(sheetName)
	if err != nil {
		return nil, nil, fmt.Errorf("wb.AddSheet failed: %w", err)
	}

	// add header row
	row := sh.AddRow()
	headerStyle := xlsx.NewStyle()
	headerStyle.Font.Name = "Arial"
	headerStyle.Font.Size = 11
	headerStyle.Font.Bold = true
	for _, key := range keys {
		cell := row.AddCell()
		cell.SetString(key)
		cell.SetStyle(headerStyle)
	}

	// add data

	// for each record
	for i := range recsMap {

		// add a row
		row := sh.AddRow()

		// for each key
		for _, key := range keys {

			// add a cell
			cell := row.AddCell()

			val := recsMap[i][key]

			// use jsonTagTypeMap to call the appropriate cell.SetType func for each type
			switch jsonTagTypeMap[key] {

			case reflect.TypeFor[bool]():
				boolVal, ok := val.(bool)
				if ok {
					cell.SetBool(boolVal)
				}

			case reflect.TypeFor[float32](), reflect.TypeFor[float64]():
				f64Val, ok := val.(float64)
				if ok {
					cell.SetFloat(f64Val)
				}

			case reflect.TypeFor[int](), reflect.TypeFor[int32](), reflect.TypeFor[int64]():
				intVal, ok := val.(int64)
				if ok {
					cell.SetInt64(intVal)
				}

			case reflect.TypeFor[lystype.Date]():
				timeVal, ok := val.(lystype.Date)
				if ok {
					cell.SetDate(time.Time(timeVal))
				}

			case reflect.TypeFor[lystype.Datetime]():
				timeVal, ok := val.(lystype.Datetime)
				if ok {
					cell.SetDateTime(time.Time(timeVal))
				}

			case reflect.TypeFor[lystype.Time]():
				timeVal, ok := val.(lystype.Time)
				if ok {
					cell.SetString(timeVal.Format(lystype.TimeFormat))
				}

			default:
				strVal, ok := val.(string)
				if ok {
					cell.SetString(strVal)
				}
			}

		} // next key

	} // next record

	return wb, sh, nil
}
