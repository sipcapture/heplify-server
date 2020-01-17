package config

import (
	"regexp"
)

var CompileStore RegexData

type RegexData struct {
	RegexMap            map[string]*regexp.Regexp
}
