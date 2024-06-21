package config

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func NewValidator() (*validator.Validate, error) {
	valid := validator.New()

	err := RegisterHex32(valid)
	if err != nil {
		return nil, err
	}

	err = RegisterDuration(valid)
	if err != nil {
		return nil, err
	}

	return valid, nil
}

func RegisterHex32(v *validator.Validate) error {
	return v.RegisterValidation("hex32", func(fl validator.FieldLevel) bool {
		if kind := fl.Field().Kind(); kind != reflect.String {
			return false
		}
		s := fl.Field().String()

		errs := v.Var(s, "hexadecimal")
		if errs != nil {
			return false
		}
		trimmed := strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
		return len(trimmed) == 32
	})
}

func RegisterDuration(v *validator.Validate) error {
	return v.RegisterValidation("duration", func(fl validator.FieldLevel) bool {
		if kind := fl.Field().Kind(); kind != reflect.Int64 {
			return false
		}

		return fl.Field().Int() != 0
	})
}
