package config

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

func NewValidator() (*validator.Validate, error) {
	valid := validator.New()

	err := valid.RegisterValidation("duration", func(fl validator.FieldLevel) bool {
		if kind := fl.Field().Kind(); kind != reflect.Int64 {
			return false
		}

		return fl.Field().Int() != 0
	})

	if err != nil {
		return nil, err
	}

	return valid, nil
}
