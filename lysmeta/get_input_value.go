package lysmeta

import (
	"reflect"

	"github.com/loveyourstack/lys/lystype"
)

func init() {
	// reflect lystype date types one time on startup for use in GetInputValue
	gLystypeDateType = reflect.TypeFor[lystype.Date]()
	gLystypeDateTypeP = reflect.TypeFor[*lystype.Date]()
	gLystypeDateTypeA = reflect.TypeFor[[]lystype.Date]()

	gLystypeTimeType = reflect.TypeFor[lystype.Time]()
	gLystypeTimeTypeP = reflect.TypeFor[*lystype.Time]()
	gLystypeTimeTypeA = reflect.TypeFor[[]lystype.Time]()

	gLystypeDatetimeType = reflect.TypeFor[lystype.Datetime]()
	gLystypeDatetimeTypeP = reflect.TypeFor[*lystype.Datetime]()
	gLystypeDatetimeTypeA = reflect.TypeFor[[]lystype.Datetime]()
}

var (
	gLystypeDateType  reflect.Type
	gLystypeDateTypeP reflect.Type
	gLystypeDateTypeA reflect.Type

	gLystypeTimeType  reflect.Type
	gLystypeTimeTypeP reflect.Type
	gLystypeTimeTypeA reflect.Type

	gLystypeDatetimeType  reflect.Type
	gLystypeDatetimeTypeP reflect.Type
	gLystypeDatetimeTypeA reflect.Type
)

// getInputValue returns val, but contains special handling for lystype date types
func GetInputValue(val any, reflType reflect.Type) (inputVal any) {

	switch reflType {

	// lystype.Date
	case gLystypeDateType:
		val := val.(lystype.Date)
		return val.Format(lystype.DateFormat)

	case gLystypeDateTypeP:
		val := val.(*lystype.Date)
		if val == nil {
			return nil
		}
		return (*val).Format(lystype.DateFormat)

	case gLystypeDateTypeA:
		val := val.([]lystype.Date)
		inputVal := make([]string, len(val))
		for i, v := range val {
			inputVal[i] = v.Format(lystype.DateFormat)
		}
		return inputVal

	// lystype.Time
	case gLystypeTimeType:
		val := val.(lystype.Time)
		return val.Format(lystype.TimeFormatDb)

	case gLystypeTimeTypeP:
		val := val.(*lystype.Time)
		if val == nil {
			return nil
		}
		return (*val).Format(lystype.TimeFormatDb)

	case gLystypeTimeTypeA:
		val := val.([]lystype.Time)
		inputVal := make([]string, len(val))
		for i, v := range val {
			inputVal[i] = v.Format(lystype.TimeFormatDb)
		}
		return inputVal

	// lystype.Datetime
	case gLystypeDatetimeType:
		val := val.(lystype.Datetime)
		return val.Format(lystype.DatetimeFormat)

	case gLystypeDatetimeTypeP:
		val := val.(*lystype.Datetime)
		if val == nil {
			return nil
		}
		return (*val).Format(lystype.DatetimeFormat)

	case gLystypeDatetimeTypeA:
		val := val.([]lystype.Datetime)
		inputVal := make([]string, len(val))
		for i, v := range val {
			inputVal[i] = v.Format(lystype.DatetimeFormat)
		}
		return inputVal

	default:
		return val
	}
}
