package lib

import (
	"errors"
	"strings"
)

type Bool struct {
	Bool *bool
}

func (b *Bool) UnmarshalText(data []byte) error {
	normalized := strings.ToLower(strings.TrimSpace(string(data)))
	if normalized == "true" {
		val := true
		*b = Bool{Bool: &val}
		return nil
	}

	if normalized == "false" {
		val := false
		*b = Bool{Bool: &val}
		return nil
	}

	if normalized == "" {
		*b = Bool{Bool: nil}
		return nil
	}

	return errors.New("invalid boolean value")
}

func (b *Bool) String() string {
	if b == nil {
		return "nil"
	}
	if b.Bool == nil {
		return "nil"
	}
	if *b.Bool {
		return "true"
	}
	return "false"
}
