package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/joho/godotenv"
	"github.com/omeid/uconfig/flat"
)

const (
	TagEnv  = "env"
	TagFlag = "flag"
	TagDesc = "desc"

	StrFlagNotDefined = "flag provided but not defined"
)

func isErrFlagNotDefined(err error) bool {
	return strings.Contains(err.Error(), StrFlagNotDefined)
}

var (
	ErrEnvLoad          = errors.New("error during loading .env file")
	ErrEnvParse         = errors.New("cannot parse env variable")
	ErrFlagParse        = errors.New("cannot parse flag")
	ErrConfigInvalid    = errors.New("invalid config struct")
	ErrConfigValidation = errors.New("config validation error")
)

type Validator interface {
	Struct(s interface{}) error
}

func LoadConfig(cfg ConfigInterface, osArgs *[]string, validator Validator) error {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(lib.WrapError(ErrEnvLoad, err))
	}

	// recursively iterates over each field of the nested struct
	fields, err := flat.View(cfg)
	if err != nil {
		return lib.WrapError(ErrConfigInvalid, err)
	}

	flagset := flag.NewFlagSet("", flag.ContinueOnError)
	flagset.Usage = func() {}

	for _, field := range fields {
		envName, ok := field.Tag(TagEnv)
		if !ok {
			continue
		}

		envValue := os.Getenv(envName)
		_ = field.Set(envValue)
		// if err != nil {
		// replace all bool envs to lib.Bool to uncomment this and error on invalid env var configuration
		// fmt.Printf("invalid value for env variable %s: %s. Error: %s\n", envName, envValue, err)
		// TODO: set default value on error
		// 	return lib.WrapError(ErrEnvParse, fmt.Errorf("%s: %w", envName, err))
		// }

		flagName, ok := field.Tag(TagFlag)
		if !ok {
			continue
		}

		flagDesc, _ := field.Tag(TagDesc)

		// writes flag value to variable
		flagset.Var(field, flagName, flagDesc)
	}

	var args []string
	if osArgs != nil {
		args = *osArgs
	} else {
		// if flargs not provided use global os.Args
		args = os.Args
	}

	// skipping program name
	args = args[1:]

	// flags override .env variables
	for {
		if len(args) == 0 {
			break
		}

		err = flagset.Parse(args)
		if err == nil {
			break
		}

		if !isErrFlagNotDefined(err) {
			return lib.WrapError(ErrFlagParse, err)
		}
		if len(flagset.Args()) == 0 {
			break
		}
		args = flagset.Args()[1:]
	}

	cfg.SetDefaults()

	err = validator.Struct(cfg)
	if err != nil {
		return lib.WrapError(ErrConfigValidation, err)
	}

	return nil
}
