package lysexcel

import (
	"fmt"
	"slices"
	"time"

	"github.com/loveyourstack/lys/lystype"
	"github.com/tealeg/xlsx/v3"
	"golang.org/x/exp/maps"
)

// WriteItemsToFile creates an Excel workbook from items
// T must have json tags set. Only the fields with a json tag get written
// jsonTagTypeMap is a map of [json tag]type name
// if filePath exists, it gets overwritten
// sheetName is optional and defaults to "data"
func WriteItemsToFile[T any](items []T, jsonTagTypeMap map[string]string, filePath, sheetName string) (err error) {

	if len(items) == 0 {
		return fmt.Errorf("items is empty")
	}
	if len(jsonTagTypeMap) == 0 {
		return fmt.Errorf("jsonTagMap is empty")
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

func writeData(recsMap []map[string]any, jsonTagTypeMap map[string]string, sheetName string) (wb *xlsx.File, sh *xlsx.Sheet, err error) {

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

			case "bool":
				boolVal, ok := val.(bool)
				if ok {
					cell.SetBool(boolVal)
				}

			case "float32", "float64":
				f64Val, ok := val.(float64)
				if ok {
					cell.SetFloat(f64Val)
				}

			case "int", "int32", "int64": // needs float64 fallback
				intVal, ok := val.(int)
				if ok {
					cell.SetInt(intVal)
				}
				f64Val, ok := val.(float64)
				if ok {
					cell.SetFloat(f64Val)
				}

			case "lystype.Date":
				strVal, ok := val.(string)
				if !ok {
					continue
				}
				timeVal, err := time.Parse(lystype.DateFormat, strVal)
				if err != nil {
					return nil, nil, fmt.Errorf("time.Parse failed: %w", err)
				}
				cell.SetDate(timeVal)

			case "lystype.Datetime":
				strVal, ok := val.(string)
				if !ok {
					continue
				}
				timeVal, err := time.Parse(lystype.DatetimeFormat, strVal)
				if err != nil {
					return nil, nil, fmt.Errorf("time.Parse failed: %w", err)
				}
				cell.SetDateTime(timeVal)

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
