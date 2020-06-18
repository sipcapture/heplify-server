package decoder

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/sipcapture/heplify-server/config"
)

type ScriptEngine interface {
	Run(hep *HEP) error
	Close()
}

// NewScriptEngine returns a script interface
func NewScriptEngine() (ScriptEngine, error) {
	switch strings.ToLower(config.Setting.ScriptEngine) {
	case "lua":
		return NewLuaEngine()
	case "expr":
		return NewExprEngine()
	}
	return nil, fmt.Errorf("unknown script engine %s\n", config.Setting.ScriptEngine)
}

func scanCode() ([]string, *bytes.Buffer, error) {
	var files []string
	buf := bytes.NewBuffer(nil)
	path := config.Setting.ScriptFolder
	b64 := config.Setting.ScriptBase64

	if path != "" {
		dir, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, nil, err
		}

		for _, file := range dir {
			if !file.IsDir() && strings.HasSuffix(file.Name(), "."+config.Setting.ScriptEngine) {
				p := filepath.Join(path, file.Name())
				f, err := os.Open(p)
				if err != nil {
					return nil, nil, err
				}
				_, err = io.Copy(buf, f)
				if err != nil {
					return nil, nil, err
				}
				err = f.Close()
				if err != nil {
					return nil, nil, err
				}
				s, err := ioutil.ReadFile(p)
				if err != nil {
					return nil, nil, err
				}
				if len(s) > 4 {
					files = append(files, string(s))
				}
			}
		}
	}

	if len(b64) > 20 {
		buf.WriteString("\n")

		b, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, nil, err
		}

		buf.Write(b)
	}

	return files, buf, nil
}

func extractFunc(r io.Reader) []string {
	var funcs []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := cutSpace(scanner.Text())
		if strings.HasPrefix(line, "--") {
			continue
		}
		if strings.HasPrefix(line, "function") {
			if b, e := strings.Index(line, "("), strings.Index(line, ")"); b > -1 && e > -1 && b < e {
				funcs = append(funcs, line[len("function"):e+1])
			}
		}
	}
	return funcs
}

func cutSpace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}
