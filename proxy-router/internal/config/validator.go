package config

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

func NewValidator() (*validator.Validate, error) {
	valid := validator.New()

	err := valid.RegisterValidation("duration", func(fl validator.FieldLevel) bool {
		kind := fl.Field().Kind()
		if kind != reflect.Int64 {
			return false
		}

		value := fl.Field().Int()
		return value != 0
	})

	if err != nil {
		return nil, err
	}

	return valid, nil
}
