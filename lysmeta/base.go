package lysmeta

import (
	"fmt"
	"reflect"
)

// Field represents a struct field with db and json metadata, and optionally its value.
type Field struct {
	Name string
	Type reflect.Type

	DbName  string // if db tag is missing or "-", the field is ignored in db operations (unlike pgx, where only "-" is ignored)
	JsonKey string // uses same rules as encoding/json: if json tag is missing or empty, use field name as json key; if json tag is "-", omit field from json

	Value any // only set if getValues is true in AnalyzeT
}

// Plan represents the metadata of a struct, used for db and json operations. Values are only included if the Plan was created with getValues=true in AnalyzeT.
// Embedded structs are flattened into the Plan, so the Fields slice includes all fields from the struct and its embedded structs.
type Plan struct {
	Fields    []Field
	HasValues bool
}

// DbNames returns the db column names of the fields in the Plan. Fields missing a db tag or with db tag "-" are omitted.
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
		return nil, nil, fmt.Errorf("Plan was created without values")
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

// JsonKeys returns the json keys of the fields in the Plan. Fields with json tag "-" are omitted.
func (p Plan) JsonKeys() (jsonKeys []string) {
	for _, field := range p.Fields {
		if field.JsonKey != "" {
			jsonKeys = append(jsonKeys, field.JsonKey)
		}
	}
	return jsonKeys
}

// JsonKeyDbNameMap returns a map of json keys to their corresponding db column names, for fields that have both json and db tags.
// Fields with json tag "-" or missing db tag are omitted.
func (p Plan) JsonKeyDbNameMap() (jsonKeyDbNameMap map[string]string) {
	jsonKeyDbNameMap = make(map[string]string)
	for _, field := range p.Fields {
		if field.JsonKey != "" && field.DbName != "" {
			jsonKeyDbNameMap[field.JsonKey] = field.DbName
		}
	}
	return jsonKeyDbNameMap
}

// JsonKeyTypeMap returns a map of json keys to their corresponding field types. Fields with json tag "-" are omitted.
func (p Plan) JsonKeyTypeMap() (jsonKeyTypeMap map[string]reflect.Type) {
	jsonKeyTypeMap = make(map[string]reflect.Type)
	for _, field := range p.Fields {
		if field.JsonKey != "" {
			jsonKeyTypeMap[field.JsonKey] = field.Type
		}
	}
	return jsonKeyTypeMap
}
