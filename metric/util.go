package metric

import (
	"strings"
	"unicode"
)

func cutSpace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func normMax(val float64) float64 {
	if val > 10000000 {
		return 0
	}
	return val
}
