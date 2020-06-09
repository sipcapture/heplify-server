package decoder

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/negbie/logp"
	"github.com/sipcapture/golua/lua"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder/luar"
)

/// structure for Script Engine
type ScriptEngine struct {
	functions []string
	LuaEngine *lua.State
	mapObj    luar.Map
	/* pointer to modify */
	hepPkt **HEP
}

func (d *ScriptEngine) GetHEPStruct() interface{} {
	if (*d.hepPkt) == nil {
		return ""
	}
	return (*d.hepPkt)
}

func (d *ScriptEngine) GetSIPStruct() interface{} {
	if (*d.hepPkt).SIP == nil {
		return ""
	}
	return (*d.hepPkt).SIP
}

func (d *ScriptEngine) GetHEPProtoType() uint32 {
	return (*d.hepPkt).GetProtoType()
}

func (d *ScriptEngine) GetHEPSrcIP() string {
	return (*d.hepPkt).GetSrcIP()
}

func (d *ScriptEngine) GetHEPSrcPort() uint32 {
	return (*d.hepPkt).GetSrcPort()
}

func (d *ScriptEngine) GetHEPDstIP() string {
	return (*d.hepPkt).GetDstIP()
}

func (d *ScriptEngine) GetHEPDstPort() uint32 {
	return (*d.hepPkt).GetDstPort()
}

func (d *ScriptEngine) GetHEPTimeSeconds() uint32 {
	return (*d.hepPkt).GetTsec()
}

func (d *ScriptEngine) GetHEPTimeUseconds() uint32 {
	return (*d.hepPkt).GetTmsec()
}

func (d *ScriptEngine) GetHEPNodeID() uint32 {
	return (*d.hepPkt).GetNodeID()
}

func (d *ScriptEngine) GetRawMessage() string {
	return (*d.hepPkt).GetPayload()
}

func (d *ScriptEngine) SetRawMessage(value string) {
	hepPkt := *d.hepPkt
	hepPkt.Payload = value
}

func (d *ScriptEngine) SetCustomHeader(m *map[string]string) {
	hepPkt := *d.hepPkt

	/* not SIP */
	if hepPkt.SIP == nil {
		return
	}

	if hepPkt.SIP.CustomHeader == nil {
		hepPkt.SIP.CustomHeader = make(map[string]string)
	}

	for k, v := range *m {
		hepPkt.SIP.CustomHeader[k] = v
	}
}

func (d *ScriptEngine) SetHEPField(field string, value string) {
	hepPkt := *d.hepPkt

	switch field {
	case "ProtoType":
		if i, err := strconv.Atoi(value); err == nil {
			hepPkt.ProtoType = uint32(i)
		}
	case "SrcIP":
		hepPkt.SrcIP = value
	case "SrcPort":
		if i, err := strconv.Atoi(value); err == nil {
			hepPkt.SrcPort = uint32(i)
		}
	case "DstIP":
		hepPkt.DstIP = value
	case "DstPort":
		if i, err := strconv.Atoi(value); err == nil {
			hepPkt.DstPort = uint32(i)
		}
	case "NodeID":
		if i, err := strconv.Atoi(value); err == nil {
			hepPkt.NodeID = uint32(i)
		}
	case "CID":
		hepPkt.CID = value
	case "SID":
		hepPkt.SID = value
	case "NodeName":
		hepPkt.NodeName = value
	case "TargetName":
		hepPkt.TargetName = value
	}
}

func (d *ScriptEngine) SetSIPHeader(header string, value string) {
	hepPkt := *d.hepPkt

	switch header {
	case "ViaOne":
		hepPkt.SIP.ViaOne = value
	case "FromUser":
		hepPkt.SIP.FromUser = value
	case "FromHost":
		hepPkt.SIP.FromHost = value
	case "FromTag":
		hepPkt.SIP.FromTag = value
	case "ToUser":
		hepPkt.SIP.ToUser = value
	case "ToHost":
		hepPkt.SIP.ToHost = value
	case "ToTag":
		hepPkt.SIP.ToTag = value
	case "CallID":
		hepPkt.SIP.CallID = value
	case "XCallID":
		hepPkt.SIP.XCallID = value
	case "ContactUser":
		hepPkt.SIP.ContactUser = value
	case "ContactHost":
		hepPkt.SIP.ContactHost = value
	case "UserAgent":
		hepPkt.SIP.UserAgent = value
	case "Server":
		hepPkt.SIP.Server = value
	case "Authorization.Username":
		hepPkt.SIP.Authorization.Username = value
	case "PaiUser":
		hepPkt.SIP.PaiUser = value
	case "PaiHost":
		hepPkt.SIP.PaiHost = value
	}
}

func (d *ScriptEngine) Logp(level string, message string, data interface{}) {
	if level == "ERROR" {
		logp.Err("[script] %s: %v", message, data)
	} else {
		logp.Debug("[script] %s: %v", message, data)
	}
}

// RegisteredScriptEngine returns a script interface
func RegisteredScriptEngine() (*ScriptEngine, error) {
	logp.Debug("script", "Init script engine...")

	dec := &ScriptEngine{}
	dec.LuaEngine = lua.NewState()
	dec.LuaEngine.OpenLibs()

	luar.Register(dec.LuaEngine, "scriptEngine", luar.Map{
		"GetHEPStruct":       dec.GetHEPStruct,
		"GetSIPStruct":       dec.GetSIPStruct,
		"GetHEPProtoType":    dec.GetHEPProtoType,
		"GetHEPSrcIP":        dec.GetHEPSrcIP,
		"GetHEPSrcPort":      dec.GetHEPSrcPort,
		"GetHEPDstIP":        dec.GetHEPDstIP,
		"GetHEPDstPort":      dec.GetHEPDstPort,
		"GetHEPTimeSeconds":  dec.GetHEPTimeSeconds,
		"GetHEPTimeUseconds": dec.GetHEPTimeUseconds,
		"GetHEPNodeID":       dec.GetHEPNodeID,
		"GetRawMessage":      dec.GetRawMessage,
		"SetRawMessage":      dec.SetRawMessage,
		"SetCustomHeader":    dec.SetCustomHeader,
		"SetHEPField":        dec.SetHEPField,
		"SetSIPHeader":       dec.SetSIPHeader,
		"Logp":               dec.Logp,
		"Print":              fmt.Println,
	})

	code, funcs, err := scanScripts(config.Setting.ScriptFolder)
	if err != nil {
		return nil, err
	}

	dec.functions = funcs
	if len(dec.functions) < 1 {
		return nil, fmt.Errorf("no function name found in lua scripts")
	}

	err = dec.LuaEngine.DoString(code)
	if err != nil {
		return nil, err
	}

	return dec, nil
}

/* our main point ExecuteScriptEngine */
func (d *ScriptEngine) ExecuteScriptEngine(hep *HEP) error {
	/* preload */
	d.hepPkt = &hep
	for _, v := range d.functions {
		err := d.LuaEngine.DoString(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func scanScripts(path string) (string, []string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return "", nil, err
	}

	buf := bytes.NewBuffer(nil)
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".lua") {
			f, err := os.Open(filepath.Join(path, file.Name()))
			if err != nil {
				return "", nil, err
			}
			_, err = io.Copy(buf, f)
			if err != nil {
				return "", nil, err
			}
			err = f.Close()
			if err != nil {
				return "", nil, err
			}
		}
	}

	code := buf.String()

	var funcs []string
	scanner := bufio.NewScanner(buf)
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
	return code, funcs, nil
}

func cutSpace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}
