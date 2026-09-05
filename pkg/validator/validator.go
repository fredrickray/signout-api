package apivalidator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Struct(v any) map[string]string {
	err := validate.Struct(v)
	if err == nil {
		return nil
	}

	out := make(map[string]string)
	if verrs, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range verrs {
			field := strings.ToLower(fe.Field()[:1]) + fe.Field()[1:]
			out[field] = messageFor(fe)
		}
		return out
	}

	out["_"] = err.Error()
	return out
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	default:
		return "is invalid"
	}
}
