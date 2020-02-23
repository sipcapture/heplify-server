package config

import (
	"regexp"
	"strings"
)

var CompileStore RegexData

type RegexData struct {
	RegexMap map[string]*regexp.Regexp
}

func GenerateRegexMap(AlegIDs []string) []string {
	if len(AlegIDs) > 0 {
		CompileStore = RegexData{RegexMap: make(map[string]*regexp.Regexp)}

		for i := range AlegIDs {
			AlegSplit := strings.Split(AlegIDs[i], ",")
			//assigned back the AlegIDs without the regex
			AlegIDs[i] = AlegSplit[0]
			if len(AlegSplit) > 1 {
				CompileStore.RegexMap[AlegSplit[0]] = regexp.MustCompile(AlegSplit[1])
			}
		}
	}
	return AlegIDs
}
