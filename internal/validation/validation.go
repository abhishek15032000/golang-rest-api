package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(i interface{}) error {
	err := validate.Struct(i)
	if err != nil {
		var errMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errMessages = append(errMessages, fmt.Sprintf("%s is required", strings.ToLower(e.Field())))
		}
		return fmt.Errorf("%s", strings.Join(errMessages, ", "))
	}
	return nil
}
