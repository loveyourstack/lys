package lysmeta

import (
	"fmt"
	"reflect"
)

// Plan represents the metadata of a struct, used for db and json operations.
// Embedded structs are flattened into the Plan, so the fields slice includes all fields from the struct and its embedded structs.
// Values are only included if the Plan was created with setValues=true in AnalyzeT.
// The Plan is created by AnalyzeT and is immutable thereafter.
type Plan struct {
	fields []Field

	// see Getters for comments
	dbNames          []string
	hasValues        bool
	jsonKeys         []string
	jsonKeyDbNameMap map[string]string
	jsonKeyTypeMap   map[string]reflect.Type
}

// Field represents a struct field with db and json metadata, and optionally its value.
type Field struct {
	Name string
	Type reflect.Type

	DbName  string // if db tag is missing or "-", the field is ignored in db operations (unlike pgx, where only "-" is ignored)
	JsonKey string // uses same rules as encoding/json: if json tag is missing or empty, use field name as json key; if json tag is "-", omit field from json

	Value any // only set if setValues is true in AnalyzeT
}

// AllDbNamesSet returns true if all fields in the Plan have a db tag (i.e. no fields are missing a db tag or have db tag "-").
func (p Plan) AllDbNamesSet() bool {
	return len(p.dbNames) == len(p.fields)
}

// DbNames returns the db column names of the fields in the Plan. Fields missing a db tag or with db tag "-" are omitted.
func (p Plan) DbNames() []string {
	return p.dbNames
}

// DbValues returns slices of db column names and corresponding values, for fields that have a db tag.
// The Plan must have been created with setValues=true in AnalyzeT.
func (p Plan) DbValues() (dbNames []string, values []any, err error) {

	if !p.hasValues {
		return nil, nil, fmt.Errorf("Plan was created without values")
	}

	for _, field := range p.fields {
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

// Fields returns the fields in the Plan, including their metadata and optionally their values if the Plan was created with setValues=true in AnalyzeT.
func (p Plan) Fields() []Field {
	return p.fields
}

// HasValues returns true if the Plan was created with setValues=true in AnalyzeT, meaning the Field.Value is populated with values from the struct.
func (p Plan) HasValues() bool {
	return p.hasValues
}

// JsonKeys returns the json keys of the fields in the Plan. Fields with json tag "-" are omitted.
func (p Plan) JsonKeys() []string {
	return p.jsonKeys
}

// JsonKeyDbNameMap returns a map of json keys to their corresponding db column names, for fields that have both json and db tags.
// Fields with json tag "-" or missing db tag are omitted.
func (p Plan) JsonKeyDbNameMap() map[string]string {
	return p.jsonKeyDbNameMap
}

// JsonKeyTypeMap returns a map of json keys to their corresponding field types. Fields with json tag "-" are omitted.
func (p Plan) JsonKeyTypeMap() map[string]reflect.Type {
	return p.jsonKeyTypeMap
}
