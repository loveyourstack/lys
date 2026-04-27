package lysmeta

import (
	"fmt"
	"reflect"
)

type Field struct {
	Name string
	Type reflect.Type

	DbName  string // does not use pgx rules: if db tag is missing or "-", the field is ignored in db operations
	JsonKey string // uses same rules as encoding/json: if json tag is missing or empty, use field name as json key; if json tag is "-", omit field from json

	Value any // only set if getValues is true in AnalyzeT
}

type Plan struct {
	Fields    []Field
	HasValues bool
}

func (p Plan) JsonKeys() (jsonKeys []string) {
	for _, field := range p.Fields {
		if field.JsonKey != "" {
			jsonKeys = append(jsonKeys, field.JsonKey)
		}
	}
	return jsonKeys
}

func (p Plan) JsonKeyTypeMap() (jsonKeyTypeMap map[string]reflect.Type) {
	jsonKeyTypeMap = make(map[string]reflect.Type)
	for _, field := range p.Fields {
		if field.JsonKey != "" {
			jsonKeyTypeMap[field.JsonKey] = field.Type
		}
	}
	return jsonKeyTypeMap
}

func (p Plan) DbNames() (dbNames []string) {
	for _, field := range p.Fields {
		if field.DbName != "" {
			dbNames = append(dbNames, field.DbName)
		}
	}
	return dbNames
}

// DbValues returns slices of db column names and corresponding values, for fields that have a db tag.
// The Plan must have been created with getValues=true in AnalyzeT.
func (p Plan) DbValues() (dbNames []string, values []any, err error) {

	if !p.HasValues {
		return nil, nil, fmt.Errorf("Plan was analyzed without values")
	}

	for _, field := range p.Fields {
		if field.DbName != "" {
			dbNames = append(dbNames, field.DbName)
			values = append(values, field.Value)
		}
	}

	if len(dbNames) == 0 {
		return nil, nil, fmt.Errorf("no fields have db tags")
	}

	return dbNames, values, nil
}
