package validation

import "github.com/go-playground/validator/v10"

var v = validator.New()

func Validator() *validator.Validate {
	// İleride custom kural: v.RegisterValidation("password", passwordRule)
	return v
}
