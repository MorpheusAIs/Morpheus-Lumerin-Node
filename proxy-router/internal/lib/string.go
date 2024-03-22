package lib

import (
	"fmt"
)

// StrShort returns a short version of the string in a format "aaaaa..aa"
func StrShort(str string) string {
	return StrShortConf(str, 5, 3)
}

// StrShortConf returns a short version of the string in a format "{prefix chars}..{suffix chars}"
func StrShortConf(str string, prefix, suffix int) string {
	if len(str) <= (prefix + suffix + 2) {
		return str
	}
	l := len(str)
	if l >= (prefix + suffix + 2) {
		return fmt.Sprintf("%s..%s", str[:prefix], str[l-suffix:l])
	}
	return str
}
