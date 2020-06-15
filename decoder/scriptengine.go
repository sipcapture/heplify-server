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

func scanCode() (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	path := config.Setting.ScriptFolder
	b64 := config.Setting.ScriptBase64

	if path != "" {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), "."+config.Setting.ScriptEngine) {
				f, err := os.Open(filepath.Join(path, file.Name()))
				if err != nil {
					return nil, err
				}
				_, err = io.Copy(buf, f)
				if err != nil {
					return nil, err
				}
				err = f.Close()
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if len(b64) > 20 {
		buf.WriteString("\n")

		b, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, err
		}

		buf.Write(b)
	}

	return buf, nil
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
