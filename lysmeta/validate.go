package lysmeta

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validate uses the validator lib to validate the supplied T
// returns all errors joined by pipe (|)
func Validate[T any](validate *validator.Validate, input T) (err error) {

	err = validate.Struct(input)
	if err != nil {

		msgs := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			msg := ""
			switch err.Tag() {
			case "email":
				msg = fmt.Sprintf("%s must be a valid email address", err.Field())
			case "gt": // numbers
				msg = fmt.Sprintf("%s must be > %s", err.Field(), err.Param())
			case "gte": // numbers
				msg = fmt.Sprintf("%s must be >= %s", err.Field(), err.Param())
			case "len":
				msg = fmt.Sprintf("%s must be len %s", err.Field(), err.Param())
			case "lowercase", "uppercase":
				msg = fmt.Sprintf("%s must be %s", err.Field(), err.Tag())
			case "lt": // numbers
				msg = fmt.Sprintf("%s must be < %s", err.Field(), err.Param())
			case "lte": // numbers
				msg = fmt.Sprintf("%s must be <= %s", err.Field(), err.Param())
			case "min", "max": // strings or slices
				msg = fmt.Sprintf("%s must be %s len %s", err.Field(), err.Tag(), err.Param())
			case "required":
				msg = fmt.Sprintf("%s is required", err.Field())
			default:
				msg = fmt.Sprintf("%s is invalid: %s", err.Field(), err.Tag())
			}
			msgs = append(msgs, msg)
		}

		return fmt.Errorf(strings.Join(msgs, " | "))
	}

	return nil
}
