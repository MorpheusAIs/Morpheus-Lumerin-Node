package config

import (
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/common"
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

	err = RegisterEthAddr(valid)
	if err != nil {
		return nil, err
	}

	return valid, nil
}

func RegisterHex32(v *validator.Validate) error {
	return v.RegisterValidation("hex32", func(fl validator.FieldLevel) bool {
		kind := fl.Field().Kind()

		if kind != reflect.String {
			s := fl.Field().String()

			errs := v.Var(s, "hexadecimal")
			if errs != nil {
				return false
			}
			trimmed := strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
			return len(trimmed) == 32
		}

		if kind == reflect.Array {
			arr := fl.Field().Interface().(common.Hash)
			return len(arr) == 32
		}

		return false
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

func RegisterEthAddr(v *validator.Validate) error {
	return v.RegisterValidation("eth_addr", func(fl validator.FieldLevel) bool {
		kind := fl.Field().Kind()
		if kind == reflect.String {
			s := fl.Field().String()

			errs := v.Var(s, "eth_addr")
			if errs != nil {
				return false
			}
			return true
		}

		// if stored as common.Address
		if kind == reflect.Array {
			arr := fl.Field().Interface().(common.Address)
			if len(arr) != 20 {
				return false
			}
			return true
		}

		return false
	})
}
